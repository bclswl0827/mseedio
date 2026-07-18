package mseedio

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

// dumpRecord renders a MiniSeedData into a stable, human-readable string so we
// can characterize the current behavior and detect any drift after refactoring.
func dumpRecord(m *MiniSeedData) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Type=%d Order=%d Records=%d Samples=%d\n", m.Type, m.Order, m.Records, m.Samples)
	fmt.Fprintf(&b, "StartTime=%s EndTime=%s\n", m.StartTime.UTC().Format(time.RFC3339Nano), m.EndTime.UTC().Format(time.RFC3339Nano))
	for i, s := range m.Series {
		f := s.FixedSection
		bl := s.BlocketteSection
		fmt.Fprintf(&b, "--- series %d ---\n", i)
		fmt.Fprintf(&b, "Seq=%q DQ=%q STA=%q LOC=%q CHA=%q NET=%q\n",
			f.SequenceNumber, f.DataQuality, f.StationCode, f.LocationCode, f.ChannelCode, f.NetworkCode)
		fmt.Fprintf(&b, "Start=%s Samples=%d Factor=%d Mult=%d\n",
			f.StartTime.UTC().Format(time.RFC3339Nano), f.SamplesNumber, f.SampleFactor, f.SampleMultiplier)
		fmt.Fprintf(&b, "Blk=%d Enc=%d BitOrder=%d RecLen=%d\n",
			bl.BlocketteCode, bl.EncodingFormat, bl.BitOrder, bl.RecordLength)
		fmt.Fprintf(&b, "Decoded=%v\n", s.DataSection.Decoded)
	}
	return b.String()
}

// TestGoldenRead reads every fixture and compares the rendered output against a
// committed golden file. Run with -update to regenerate the goldens.
func TestGoldenRead(t *testing.T) {
	files, err := filepath.Glob("example/reader/testdata/*.mseed")
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(files)
	if len(files) == 0 {
		t.Fatal("no testdata fixtures found")
	}

	var out strings.Builder
	for _, f := range files {
		var m MiniSeedData
		if err := m.Read(f); err != nil {
			t.Fatalf("Read(%s): %v", f, err)
		}
		fmt.Fprintf(&out, "=== %s ===\n%s\n", filepath.Base(f), dumpRecord(&m))
	}

	goldenPath := "example/reader/testdata/golden_read.txt"
	if os.Getenv("UPDATE_GOLDEN") != "" {
		if err := os.WriteFile(goldenPath, []byte(out.String()), 0644); err != nil {
			t.Fatal(err)
		}
		t.Log("golden updated")
		return
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden (set UPDATE_GOLDEN=1 to create): %v", err)
	}
	if string(want) != out.String() {
		t.Errorf("output drifted from golden.\n--- got ---\n%s", out.String())
	}
}

// TestReadFromReaderMatchesRead ensures the io.Reader path and file path agree.
func TestReadFromReaderMatchesRead(t *testing.T) {
	files, _ := filepath.Glob("example/reader/testdata/*.mseed")
	for _, f := range files {
		var mFile MiniSeedData
		if err := mFile.Read(f); err != nil {
			t.Fatalf("Read(%s): %v", f, err)
		}
		fh, err := os.Open(f)
		if err != nil {
			t.Fatal(err)
		}
		var mReader MiniSeedData
		err = mReader.ReadFromReader(fh)
		fh.Close()
		if err != nil {
			t.Fatalf("ReadFromReader(%s): %v", f, err)
		}
		if dumpRecord(&mFile) != dumpRecord(&mReader) {
			t.Errorf("%s: Read and ReadFromReader disagree", f)
		}
	}
}

// TestInt16NegativeSignExtension verifies that negative samples encoded as INT16
// are sign-extended on decode (previously they came back as large positives).
func TestInt16NegativeSignExtension(t *testing.T) {
	var m MiniSeedData
	_ = m.Init(INT16, MSBFIRST)
	_ = m.Append([]int32{-50}, &AppendOptions{
		SampleRate: 1, StartTime: time.Unix(0, 0).UTC(), SequenceNumber: "000001",
		StationCode: "AAAAA", LocationCode: "BB", ChannelCode: "EHZ", NetworkCode: "CC",
	})
	b, _ := m.Encode(OVERWRITE, MSBFIRST)
	tmp := filepath.Join(t.TempDir(), "n.mseed")
	_ = m.Write(tmp, OVERWRITE, b)
	var got MiniSeedData
	_ = got.Read(tmp)
	if v := got.Series[0].DataSection.Decoded[0]; v != int32(-50) {
		t.Fatalf("want -50, got %v", v)
	}
}

// TestRoundTrip exercises Init -> Append -> Encode -> Read for each encoding and
// verifies the decoded samples survive a full write/read cycle.
func TestRoundTrip(t *testing.T) {
	encodings := []struct {
		name string
		typ  int
	}{
		{"INT16", INT16},
		{"INT32", INT32},
		{"FLOAT32", FLOAT32},
		{"FLOAT64", FLOAT64},
		{"STEIM1", STEIM1},
		{"STEIM2", STEIM2},
	}

	// Non-negative values only: the current INT16/INT24 decoders do not
	// sign-extend, so negatives would not survive a round trip (see
	// TestInt16NegativeSignExtension for that known issue).
	sample := make([]int32, 100)
	for i := range sample {
		sample[i] = int32(i * 3)
	}

	for _, enc := range encodings {
		t.Run(enc.name, func(t *testing.T) {
			var m MiniSeedData
			if err := m.Init(enc.typ, MSBFIRST); err != nil {
				t.Fatal(err)
			}
			err := m.Append(sample, &AppendOptions{
				SampleRate:     100,
				StartTime:      time.Date(2020, 1, 2, 3, 4, 5, 600000000, time.UTC),
				SequenceNumber: "000001",
				StationCode:    "AAAAA",
				LocationCode:   "BB",
				ChannelCode:    "EHZ",
				NetworkCode:    "CC",
			})
			if err != nil {
				t.Fatal(err)
			}

			dataBytes, err := m.Encode(OVERWRITE, MSBFIRST)
			if err != nil {
				t.Fatal(err)
			}

			tmp := filepath.Join(t.TempDir(), "rt.mseed")
			if err := m.Write(tmp, OVERWRITE, dataBytes); err != nil {
				t.Fatal(err)
			}

			var got MiniSeedData
			if err := got.Read(tmp); err != nil {
				t.Fatal(err)
			}
			if len(got.Series) != 1 {
				t.Fatalf("want 1 series, got %d", len(got.Series))
			}
			decoded := got.Series[0].DataSection.Decoded
			if len(decoded) != len(sample) {
				t.Fatalf("want %d samples, got %d", len(sample), len(decoded))
			}
			for i, v := range decoded {
				var iv int32
				switch n := v.(type) {
				case int32:
					iv = n
				case float64:
					iv = int32(n)
				case float32:
					iv = int32(n)
				default:
					t.Fatalf("unexpected decoded type %T", v)
				}
				if iv != sample[i] {
					t.Fatalf("sample %d: want %d got %d", i, sample[i], iv)
				}
			}
		})
	}
}

package mseedio

import "fmt"

// knownBlockettes lists blockette types the parser recognizes but does not
// decode field-by-field. They parse successfully with only BlocketteCode set,
// preserving the original permissive behavior for records that carry them.
var knownBlockettes = map[int32]bool{
	100: true, 200: true, 201: true, 300: true, 310: true, 320: true,
	390: true, 395: true, 400: true, 405: true, 500: true, 2000: true,
}

// Parse decodes a fixed data-record header from the fixed section of a miniSEED
// record. buffer must contain at least FIXED_SECTION_LENGTH bytes.
func (f *FixedSection) Parse(buffer []byte, bitOrder int) error {
	if len(buffer) < FIXED_SECTION_LENGTH {
		return fmt.Errorf("fixed section requires %d bytes, got %d", FIXED_SECTION_LENGTH, len(buffer))
	}

	r := &byteReader{buf: buffer, order: bitOrder}
	f.SequenceNumber = r.string(6)
	f.DataQuality = r.string(1)
	r.skip(1) // reserved
	f.StationCode = r.string(5)
	f.LocationCode = r.string(2)
	f.ChannelCode = r.string(3)
	f.NetworkCode = r.string(2)
	f.StartTime = r.time()
	f.SamplesNumber = r.int(2)
	f.SampleFactor = r.int(2)
	f.SampleMultiplier = r.int(2)
	f.ActivityFlags = r.int(1)
	f.IOClockFlags = r.int(1)
	f.DataQualityFlags = r.int(1)
	f.BlockettesFollow = r.int(1)
	f.TimeCorrection = r.int(4)
	f.DataStartOffset = r.int(2)
	f.SectionEndOffset = r.int(2)
	return nil
}

// Parse decodes a blockette section. Only blockette 1000 (Data Only SEED) and
// 1001 (Data Extension) are decoded in full; other standard blockette types are
// accepted with only BlocketteCode populated, and unknown types return an error.
func (b *BlocketteSection) Parse(buffer []byte, bitOrder int) error {
	code, err := getBlocketteType(buffer, bitOrder)
	if err != nil {
		return err
	}
	b.BlocketteCode = code

	switch code {
	case 1000:
		if len(buffer) < 7 {
			return fmt.Errorf("blockette 1000 requires 7 bytes, got %d", len(buffer))
		}
		r := &byteReader{buf: buffer, pos: 2, order: bitOrder}
		b.NextBlockette = r.int(2)
		b.EncodingFormat = r.int(1)
		b.BitOrder = r.int(1)
		b.RecordLength = r.int(1)
	case 1001:
		if len(buffer) < 8 {
			return fmt.Errorf("blockette 1001 requires 8 bytes, got %d", len(buffer))
		}
		r := &byteReader{buf: buffer, pos: 2, order: bitOrder}
		b.NextBlockette = r.int(2)
		b.TimingQuality = r.int(1)
		b.Microseconds = r.int(1)
		r.skip(1) // reserved
		b.FrameCount = r.int(1)
	default:
		if !knownBlockettes[code] {
			return fmt.Errorf("blockette type %d is not supported", code)
		}
	}
	return nil
}

// Parse decodes the data section into DataSection.Decoded according to the
// record's encoding format, keeping the original bytes in RawData.
func (d *DataSection) Parse(buffer []byte, samples, blockette, encoding, bitOrder int) error {
	d.RawData = buffer

	switch encoding {
	case ASCII:
		d.Decoded = append(d.Decoded, unpackAscii(buffer))
	case INT16:
		appendDecoded(d, unpackInt(buffer, samples, 16, bitOrder))
	case INT24:
		appendDecoded(d, unpackInt(buffer, samples, 24, bitOrder))
	case INT32:
		appendDecoded(d, unpackInt(buffer, samples, 32, bitOrder))
	case FLOAT32:
		appendDecoded(d, unpackFloat(buffer, samples, 32, bitOrder))
	case FLOAT64:
		appendDecoded(d, unpackFloat(buffer, samples, 64, bitOrder))
	case STEIM1:
		result, err := unpackSteim1(buffer, samples, bitOrder)
		if err != nil {
			return err
		}
		appendDecoded(d, result)
	case STEIM2:
		result, err := unpackSteim2(buffer, samples, bitOrder)
		if err != nil {
			return err
		}
		appendDecoded(d, result)
	default:
		return fmt.Errorf("encoding %d is not supported", encoding)
	}
	return nil
}

// appendDecoded appends every element of vals to DataSection.Decoded, boxing
// each into the any-typed slice that callers expect.
func appendDecoded[T any](d *DataSection, vals []T) {
	for _, v := range vals {
		d.Decoded = append(d.Decoded, v)
	}
}

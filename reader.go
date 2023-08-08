package mseedio

import (
	"bufio"
	"os"
)

// Read miniSEED file to structured MiniSeedData
func (m *MiniSeedData) Read(filename string) error {
	// Open miniSEED file
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read file to bytes
	var bytes []byte
	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)
	for {
		n, err := reader.Read(buffer)
		if err != nil {
			break
		}

		bytes = append(bytes, buffer[:n]...)
	}

	// Guess data bit order
	bitOrder, err := getBitOrder(bytes[46:48])
	if err != nil {
		return err
	}

	// Parse fixed and blockette sections
	var (
		fixedSections     = []fixedSection{}
		blocketteSections = []blocketteSection{}
		samplesNumber     = 0 // Total number of samples
	)
	for i := 0; i < len(bytes); i += 64 {
		var (
			fs = fixedSection{}
			bs = blocketteSection{}
		)

		// Parse fixed section
		fsOffset := i + FIXED_SECTION_LENGTH
		err := fs.Parse(bytes[i:fsOffset], bitOrder)
		if fs.SectionEndOffset != FIXED_SECTION_LENGTH ||
			err != nil || (fs.DataQuality != "D" &&
			fs.DataQuality != "R" &&
			fs.DataQuality != "Q" &&
			fs.DataQuality != "M") {
			continue
		}

		// Parse blockette
		bsOffset := i + int(fs.DataStartOffset)
		err = bs.Parse(bytes[fsOffset:bsOffset], bitOrder)
		if err != nil {
			continue
		}

		// Determine data encoding for 1001-blockettes
		if bs.BlocketteCode == 1001 {
			bs.EncodingFormat = int32(bytes[fsOffset:bsOffset][12])
		}

		// Set position [start:end]
		fs.ReaderOffset = sectionOffset{
			i, fsOffset,
		}
		bs.ReaderOffset = sectionOffset{
			fsOffset, bsOffset,
		}

		// Set start time and sample
		if i == 0 {
			m.StartTime = fs.StartTime
		}

		// Add samples and append
		samplesNumber += int(fs.SamplesNumber)
		fixedSections = append(fixedSections, fs)
		blocketteSections = append(blocketteSections, bs)
	}

	// Detect initial frame length automatically
	var initLength int
	for i := 64; i < len(bytes); i += 64 {
		// Parse fixed section again
		var fs = fixedSection{}
		var fsOffset = i + FIXED_SECTION_LENGTH
		err := fs.Parse(bytes[i:fsOffset], bitOrder)
		if fs.SectionEndOffset == FIXED_SECTION_LENGTH &&
			err == nil && (fs.DataQuality == "D" ||
			fs.DataQuality == "R" ||
			fs.DataQuality == "Q" ||
			fs.DataQuality == "M") {
			initLength = i
			break
		}
	}
	if initLength == 0 {
		initLength = len(bytes)
	}

	// Detect each frame length automatically
	var (
		frameLength []int
		lastOffset  sectionOffset
	)
	for i, v := range fixedSections {
		if i == 0 {
			frameLength = append(frameLength, initLength)
		} else {
			frameLength = append(frameLength, v.ReaderOffset.Start-lastOffset.Start)
		}
		lastOffset = v.ReaderOffset
	}

	// Parse data series section
	for i, v := range blocketteSections {
		// Parse data section
		var ds = dataSection{}
		var dsOffset = fixedSections[i].ReaderOffset.Start + frameLength[i]
		err = ds.Parse(
			bytes[v.ReaderOffset.End:dsOffset],
			int(fixedSections[i].SamplesNumber),
			int(v.BlocketteCode),
			int(v.EncodingFormat),
			bitOrder,
		)
		if err != nil {
			return err
		}

		// Append data series
		m.Series = append(m.Series, dataSeries{
			BlocketteSection: v,
			DataSection:      ds,
			FixedSection:     fixedSections[i],
		})
	}

	// Set file info
	m.Order = bitOrder
	m.Samples = samplesNumber
	m.Records = len(fixedSections)
	m.Type = int(blocketteSections[0].BlocketteCode)
	m.EndTime = fixedSections[len(fixedSections)-1].StartTime

	return nil
}

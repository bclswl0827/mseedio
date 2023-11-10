package mseedio

import (
	"bufio"
	"fmt"
	"os"
)

// m.Read() reads miniSEED file to structured MiniSeedData
func (m *MiniSeedData) Read(filePath string) error {
	// Open miniSEED file
	file, err := os.Open(filePath)
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

	// Return error if length is less than 48 bytes
	if len(bytes) < 48 {
		err := fmt.Errorf("file length is less than 48 bytes")
		return err
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
		if err != nil ||
			fs.SectionEndOffset != FIXED_SECTION_LENGTH ||
			(fs.DataQuality != "D" && fs.DataQuality != "R" &&
				fs.DataQuality != "Q" && fs.DataQuality != "M") {
			continue
		}

		// Parse blockette
		bsOffset := i + int(fs.DataStartOffset)
		err = bs.Parse(bytes[fsOffset:bsOffset], bitOrder)
		if err != nil {
			continue
		}

		// Determine encoding for non 100-blockettes
		if bs.BlocketteCode == 1001 {
			// Encoding is usually in bytes[fsOffset:bsOffset][12]
			bs.EncodingFormat = int32(bytes[fsOffset:bsOffset][12])
		}

		// Set slice position [start:end]
		fs.ReaderOffset = sectionOffset{
			i, fsOffset,
		}
		bs.ReaderOffset = sectionOffset{
			fsOffset, bsOffset,
		}

		// Set global start time
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
		if err == nil && fs.SectionEndOffset == FIXED_SECTION_LENGTH {
			initLength = i
			break
		}
	}
	if initLength == 0 {
		initLength = len(bytes)
	}

	// Detect rest each frame length automatically
	var (
		frameLength = []int{initLength}
		lastOffset  sectionOffset
	)
	for i := 1; i < len(fixedSections); i++ {
		readerOffset := fixedSections[i].ReaderOffset
		frameLength = append(frameLength, readerOffset.Start-lastOffset.Start)
		lastOffset = readerOffset
	}

	// Parse data series section
	for i := 0; i < len(fixedSections); i++ {
		var (
			endIndex   = len(bytes)
			startIndex = blocketteSections[i].ReaderOffset.End
		)
		if i != len(fixedSections)-1 {
			endIndex = fixedSections[i].ReaderOffset.Start + frameLength[i+1]
		}

		var ds dataSection
		err = ds.Parse(
			bytes[startIndex:endIndex],
			int(fixedSections[i].SamplesNumber),
			int(blocketteSections[i].BlocketteCode),
			int(blocketteSections[i].EncodingFormat),
			bitOrder,
		)
		if err != nil {
			return err
		}

		// Append data series
		m.Series = append(m.Series, dataSeries{
			DataSection:      ds,
			FixedSection:     fixedSections[i],
			BlocketteSection: blocketteSections[i],
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

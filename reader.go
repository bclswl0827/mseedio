package mseedio

import (
	"bufio"
	"fmt"
	"math"
	"os"
)

// Read miniSEED file to structured MiniSeedData
func (m *MiniSeedData) Read(filename string) error {
	// Open SAC file
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

	bitOrder, err := getBitOrder(bytes[46:48])
	if err != nil {
		return err
	}

	for i, j := 0, 0; ; i++ {
		// Parse first fixed section
		var (
			fixedSection      = fixedSection{}
			fixedSectionStart = 0
			fixedSectionEnd   = FIXED_SECTION_LENGTH
		)
		err = fixedSection.Parse(bytes[fixedSectionStart:fixedSectionEnd], bitOrder)
		if err != nil {
			return err
		}

		// Set start time and sample
		if i == 0 {
			m.StartTime = fixedSection.StartTime
		}
		j += int(fixedSection.SamplesNumber)

		// Parse blockette section
		var (
			blocketteSection = blocketteSection{}
			blocketteStart   = FIXED_SECTION_LENGTH
			blocketteEnd     = fixedSection.DataStartOffset
		)
		err = blocketteSection.Parse(bytes[blocketteStart:blocketteEnd], bitOrder)
		if err != nil {
			return err
		}

		// Get frame length
		var frameLen int
		switch blocketteSection.BlocketteCode {
		case 1000:
			length := float64(blocketteSection.RecordLength)
			frameLen = int(math.Pow(2.0, length))
		case 1001:
			frameLen = 512
		default:
			return fmt.Errorf("blockette type %d is not supported", blocketteSection.BlocketteCode)
		}

		// Parse data section
		var (
			dataSection = dataSection{}
			dataStart   = blocketteEnd
			dataEnd     = frameLen
		)
		err = dataSection.Parse(
			bytes[dataStart:dataEnd],
			int(fixedSection.SamplesNumber),
			int(blocketteSection.BlocketteCode),
			int(blocketteSection.EncodingFormat),
			bitOrder,
		)
		if err != nil {
			return err
		}

		// Append data series
		m.Series = append(m.Series, dataSeries{
			FixedSection:     fixedSection,
			BlocketteSection: blocketteSection,
			DataSection:      dataSection,
		})

		// Update bytes
		bytes = bytes[frameLen:]
		if len(bytes) == 0 {
			m.Records = i
			m.Samples = j
			m.EndTime = fixedSection.StartTime
			m.Order = int(blocketteSection.BitOrder)
			m.Type = int(blocketteSection.BlocketteCode)
			break
		}
	}

	return nil
}

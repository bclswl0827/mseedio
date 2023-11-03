package mseedio

import (
	"fmt"
	"math"
)

// m.Append() appends data to 1000 blockette MiniSeedData
func (m *MiniSeedData) Append(data []int32, options *AppendOptions) error {
	// Check if sequence number is valid
	for _, v := range m.Series {
		if v.FixedSection.SequenceNumber == options.SequenceNumber &&
			v.FixedSection.ChannelCode == options.ChannelCode {
			return fmt.Errorf("sequence number %s already exists", options.SequenceNumber)
		}
	}

	// Data length must be greater than sample rate
	if int(options.SampleRate) >= len(data) {
		return fmt.Errorf("data length must be greater than sample rate")
	}

	// Pack the data
	var dataBytes []byte
	switch m.Type {
	case ASCII:
		dataBytes = append(dataBytes, packAscii(data)...)
	case INT16:
		dataBytes = append(dataBytes, packInt(data, 16, m.Order)...)
	case INT24:
		dataBytes = append(dataBytes, packInt(data, 24, m.Order)...)
	case INT32:
		dataBytes = append(dataBytes, packInt(data, 32, m.Order)...)
	case FLOAT32:
		dataBytes = append(dataBytes, packFloat(data, 32, m.Order)...)
	case FLOAT64:
		dataBytes = append(dataBytes, packFloat(data, 64, m.Order)...)
	case STEIM1:
		result, err := packSteim1(data, m.Order)
		if err != nil {
			return err
		}
		dataBytes = append(dataBytes, result...)
	case STEIM2:
		result, err := packSteim2(data, m.Order)
		if err != nil {
			return err
		}
		dataBytes = append(dataBytes, result...)
	default:
		return fmt.Errorf("%d is not a valid encoding format", m.Type)
	}

	// Get entire record length
	recordLength := math.Log2(float64(
		nextPow2(FIXED_SECTION_LENGTH + BLOCKETTE100X_SECTION_LENGTH + len(dataBytes)),
	))
	if recordLength < 8 {
		recordLength = 8
	}

	// Set blockette section
	bs := blocketteSection{
		BlocketteCode:  1000,
		NextBlockette:  0,
		EncodingFormat: int32(m.Type),
		BitOrder:       int32(m.Order),
		RecordLength:   int32(recordLength),
		TimingQuality:  0,
		Microseconds:   0,
		FrameCount:     0,
	}

	// Get SampleFactor and SampleMultiplier
	var (
		sampleFactor     int32
		SampleMultiplier int32
	)
	if options.SampleRate != math.Floor(options.SampleRate) {
		_, f := getDigitsFloat64(options.SampleRate)
		sampleFactor = int32(options.SampleRate * math.Pow10(f))
		SampleMultiplier = int32(-math.Pow10(f))
	} else {
		SampleMultiplier = 1
		sampleFactor = int32(options.SampleRate)
	}

	// Set fixed section
	fs := fixedSection{
		DataQuality:      "D",
		SequenceNumber:   options.SequenceNumber,
		StationCode:      options.StationCode,
		LocationCode:     options.LocationCode,
		ChannelCode:      options.ChannelCode,
		NetworkCode:      options.NetworkCode,
		StartTime:        options.StartTime,
		SampleFactor:     sampleFactor,
		SampleMultiplier: SampleMultiplier,
		SamplesNumber:    int32(len(data)),
		ActivityFlags:    0,
		IOClockFlags:     0,
		DataQualityFlags: 0,
		BlockettesFollow: 1,
		TimeCorrection:   0,
		DataStartOffset:  64,
		SectionEndOffset: 48,
	}

	// Set start time on first append
	if len(m.Series) == 0 {
		m.StartTime = options.StartTime
	}
	m.EndTime = options.StartTime

	// Appending the new data series
	ds := dataSection{}
	ds.Decoded = append(ds.Decoded, data)
	ds.RawData = append(ds.RawData, dataBytes...)
	m.Series = append(m.Series, dataSeries{
		FixedSection:     &fs,
		BlocketteSection: &bs,
		DataSection:      &ds,
	})

	// Updating counters
	m.Records++
	m.Samples += len(data)
	return nil
}

package mseedio

import (
	"fmt"
	"math"
)

// m.Encode() encodes record(s) with 1000-blockette
func (m *MiniSeedData) Encode(encodeMode, bitOrder int) ([]byte, error) {
	// STEIM-* does not support LSBFIRST
	if bitOrder == LSBFIRST {
		for _, v := range m.Series {
			if v.BlocketteSection.EncodingFormat == STEIM1 ||
				v.BlocketteSection.EncodingFormat == STEIM2 {
				return nil, fmt.Errorf("STEIM-* does not support LSBFIRST")
			}
		}
	}

	// Append mode only encode last record
	if encodeMode == APPEND {
		lastRecord := m.Series[len(m.Series)-1]
		if lastRecord.BlocketteSection.BlocketteCode != 1000 {
			return nil, fmt.Errorf("only 1000-blockette is supported")
		}

		// Create data bytes with fixed length
		dataLength := int(math.Pow(2, float64(lastRecord.BlocketteSection.RecordLength)))
		dataBytes := make([]byte, dataLength)

		// Compose fixed section data bytes
		fs, err := lastRecord.FixedSection.Compose(bitOrder)
		if err != nil {
			return nil, err
		}

		// Compose blockette section data bytes
		bs, err := lastRecord.BlocketteSection.Compose(bitOrder)
		if err != nil {
			return nil, err
		}

		// Copy raw data to data bytes
		copy(dataBytes, fs)
		copy(dataBytes[FIXED_SECTION_LENGTH:], bs)
		copy(dataBytes[FIXED_SECTION_LENGTH+BLOCKETTE100X_SECTION_LENGTH:], lastRecord.DataSection.RawData)

		return dataBytes, nil
	}

	// Go through all record and encode
	var dataBytes []byte
	for _, v := range m.Series {
		if v.BlocketteSection.BlocketteCode != 1000 {
			return nil, fmt.Errorf("only 1000-blockette is supported")
		}

		// Create data slice with fixed length
		dataLength := int(math.Pow(2, float64(v.BlocketteSection.RecordLength)))
		dataSlice := make([]byte, dataLength)

		// Compose fixed section data bytes
		fs, err := v.FixedSection.Compose(bitOrder)
		if err != nil {
			return nil, err
		}

		// Compose blockette section data bytes
		bs, err := v.BlocketteSection.Compose(bitOrder)
		if err != nil {
			return nil, err
		}

		// Copy raw data to data bytes
		copy(dataSlice, fs)
		copy(dataSlice[FIXED_SECTION_LENGTH:], bs)
		copy(dataSlice[FIXED_SECTION_LENGTH+BLOCKETTE100X_SECTION_LENGTH:], v.DataSection.RawData)

		// Append slice to data bytes
		dataBytes = append(dataBytes, dataSlice...)
	}

	return dataBytes, nil
}

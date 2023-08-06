package mseedio

import (
	"fmt"
	"reflect"
)

// f.Parse parses MiniSeedData fixed section
func (f *fixedSection) Parse(buffer []byte, bitOrder int) error {
	t := reflect.ValueOf(f).Elem()

	for i, j := 0, 0; i < FIXED_SECTION_LENGTH; j++ {
		var (
			field      = fixedSectionMap[j]
			fieldName  = field.FieldName
			fieldSize  = field.FieldSize
			fieldSlice = buffer[i : i+fieldSize]
		)

		var err error
		i += field.FieldSize

		switch field.FieldType {
		case "int32":
			result := assembleInt(fieldSlice, fieldSize, bitOrder)
			err = setStructFieldValue(t, fieldName, result)
		case "string":
			result := assembleString(fieldSlice)
			err = setStructFieldValue(t, fieldName, result)
		case "time.Time":
			result := assembleTime(fieldSlice, bitOrder)
			err = setStructFieldValue(t, fieldName, result)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// b.Parse parses MiniSeedData blockette section, only blockette 1000 is supported
func (b *blocketteSection) Parse(buffer []byte, bitOrder int) error {
	t := reflect.ValueOf(b).Elem()

	blkTyp, err := getBlocketteType(buffer, bitOrder)
	if err != nil {
		return err
	}
	b.BlocketteCode = blkTyp

	var fieldLen int
	switch blkTyp {
	case 1000:
		fieldLen = len(blockette1000SectionMap)
	case 1001:
		fieldLen = len(blockette1001SectionMap)
	default:
		return fmt.Errorf("blockette type %d is not supported", blkTyp)
	}

	for i, j := 2, 1; j < fieldLen; j++ {
		var field sectionMap
		switch blkTyp {
		case 1000:
			field = blockette1000SectionMap[j]
		case 1001:
			field = blockette1001SectionMap[j]
		default:
			return fmt.Errorf("blockette type %d is not supported", blkTyp)
		}

		var (
			err        error
			fieldName  = field.FieldName
			fieldSize  = field.FieldSize
			fieldSlice = buffer[i : i+fieldSize]
		)

		i += field.FieldSize
		switch field.FieldType {
		case "int32":
			result := assembleInt(fieldSlice, fieldSize, bitOrder)
			err = setStructFieldValue(t, fieldName, result)
		case "string":
			result := assembleString(fieldSlice)
			err = setStructFieldValue(t, fieldName, result)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// d.Parse parses MiniSeedData data section, only encoding 0, 1, 2, 3, 11 are supported
func (d *dataSection) Parse(buffer []byte, samples, blockette, encoding, bitOrder int) error {
	d.RawData = buffer

	if blockette == 1001 {
		encoding = 11
	}

	switch encoding {
	case 0: // ASCII text
		d.Decoded = append(d.Decoded, string(buffer))
	case 1: // 16-bit integer
		for i := 2; i < len(buffer); i += 2 {
			d.Decoded = append(d.Decoded, assembleInt(buffer[i-2:i], 2, bitOrder))
		}
	case 2: // 24-bit integer
		for i := 3; i < len(buffer); i += 3 {
			d.Decoded = append(d.Decoded, assembleInt(buffer[i-3:i], 3, bitOrder))
		}
	case 3: // 32-bit integer
		for i := 4; i < len(buffer); i += 4 {
			d.Decoded = append(d.Decoded, assembleInt(buffer[i-4:i], 4, bitOrder))
		}
	case 4: // IEEE 32-bit floating point
		for i := 4; i < len(buffer); i += 4 {
			d.Decoded = append(d.Decoded, assembleFloat32(buffer[i-4:i], bitOrder))
		}
	case 5: // IEEE 64-bit floating point
		for i := 8; i < len(buffer); i += 8 {
			d.Decoded = append(d.Decoded, assembleFloat64(buffer[i-8:i], bitOrder))
		}
	case 11: // Steim-2
		result, err := unpackSteim2(buffer, samples, bitOrder)
		if err != nil {
			return err
		}
		for _, v := range result {
			d.Decoded = append(d.Decoded, v)
		}
	default:
		return fmt.Errorf("encoding %d is not supported", encoding)
	}

	return nil
}

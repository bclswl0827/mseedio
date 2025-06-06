package mseedio

import (
	"fmt"
	"reflect"
)

// f.Parse() parses MiniSeedData fixed section
func (f *FixedSection) Parse(buffer []byte, bitOrder int) error {
	t := reflect.ValueOf(f).Elem()

	for i, j := 0, 0; i < FIXED_SECTION_LENGTH; j++ {
		var (
			field      = fixedSectionMap[j]
			fieldName  = field.FieldName
			fieldSize  = field.FieldSize
			fieldSlice = buffer[i : i+fieldSize]
		)

		var err error
		i += fieldSize

		switch field.FieldType {
		case "int32":
			// "int32" bit width should actually be determined by fieldSize
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

// b.Parse() parses MiniSeedData blockette section
func (b *BlocketteSection) Parse(buffer []byte, bitOrder int) error {
	t := reflect.ValueOf(b).Elem()

	blkTyp, err := getBlocketteType(buffer, bitOrder)
	if err != nil {
		return err
	}
	b.BlocketteCode = blkTyp

	var fieldLength int
	switch blkTyp {
	case 100:
	case 200:
	case 201:
	case 300:
	case 310:
	case 320:
	case 390:
	case 395:
	case 400:
	case 405:
	case 500:
	case 1000:
		fieldLength = len(blockette1000SectionMap)
	case 1001:
		fieldLength = len(blockette1001SectionMap)
	case 2000:
	default:
		return fmt.Errorf("blockette type %d is not supported", blkTyp)
	}

	for i, j := 2, 1; j < fieldLength; j++ {
		var field sectionMap
		switch blkTyp {
		case 100:
		case 200:
		case 201:
		case 300:
		case 310:
		case 320:
		case 390:
		case 395:
		case 400:
		case 405:
		case 500:
		case 1000:
			field = blockette1000SectionMap[j]
		case 1001:
			field = blockette1001SectionMap[j]
		case 2000:
		default:
			return fmt.Errorf("blockette type %d is not supported", blkTyp)
		}

		var (
			err        error
			fieldName  = field.FieldName
			fieldSize  = field.FieldSize
			fieldSlice = buffer[i : i+fieldSize]
		)

		i += fieldSize
		switch field.FieldType {
		case "int32":
			// "int32" bit width should actually be determined by fieldSize
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

// d.Parse() parses MiniSeedData data section
func (d *DataSection) Parse(buffer []byte, samples, blockette, encoding, bitOrder int) error {
	d.RawData = buffer

	switch encoding {
	case ASCII: // ASCII text
		d.Decoded = append(d.Decoded, unpackAscii(buffer))
	case INT16: // 16-bit integer
		result := unpackInt(buffer, samples, 16, bitOrder)
		for _, v := range result {
			d.Decoded = append(d.Decoded, v)
		}
	case INT24: // 24-bit integer
		result := unpackInt(buffer, samples, 24, bitOrder)
		for _, v := range result {
			d.Decoded = append(d.Decoded, v)
		}
	case INT32: // 32-bit integer
		result := unpackInt(buffer, samples, 32, bitOrder)
		for _, v := range result {
			d.Decoded = append(d.Decoded, v)
		}
	case FLOAT32: // IEEE 32-bit floating point
		result := unpackFloat(buffer, samples, 32, bitOrder)
		for _, v := range result {
			d.Decoded = append(d.Decoded, v)
		}
	case FLOAT64: // IEEE 64-bit floating point
		result := unpackFloat(buffer, samples, 64, bitOrder)
		for _, v := range result {
			d.Decoded = append(d.Decoded, v)
		}
	case STEIM1: // Steim-1 compression
		result, err := unpackSteim1(buffer, samples, bitOrder)
		if err != nil {
			return err
		}
		for _, v := range result {
			d.Decoded = append(d.Decoded, v)
		}
	case STEIM2: // Steim-2 compression
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

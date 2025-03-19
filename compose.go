package mseedio

import (
	"reflect"
	"time"
)

// f.Compose() compose a fixed section to bytes
func (f *FixedSection) Compose(bitOrder int) ([]byte, error) {

	var (
		result    []byte
		dataBytes []byte
		values    = reflect.ValueOf(f).Elem()
	)
	for i, j := 0, 0; i < FIXED_SECTION_LENGTH; j++ {
		var (
			field     = fixedSectionMap[j]
			fieldName = field.FieldName
		)
		var (
			fieldSize       = field.FieldSize
			filedValue, err = getStructFieldValue(values, fieldName)
		)

		if err != nil && fieldName != "Reserved" {
			return nil, err
		}

		i += fieldSize
		switch field.FieldType {
		case "int32":
			// "int32" bit width should actually be determined by fieldSize
			result = disassembleInt(filedValue.(int32), fieldSize, bitOrder)
		case "string":
			result = disassembleString(filedValue.(string), fieldSize, ' ')
		case "time.Time":
			result = disassembleTime(filedValue.(time.Time), bitOrder)
		}

		// Space as reserved field padding
		if fieldName == "Reserved" {
			result = []byte{' '}
		}

		dataBytes = append(dataBytes, result...)
	}

	return dataBytes, nil
}

// b.Compose() compose an 1000-blockette section to bytes
func (b *BlocketteSection) Compose(bitOrder int) ([]byte, error) {
	var blocketteLength int
	for _, field := range blockette1000SectionMap {
		blocketteLength += field.FieldSize
	}

	var (
		result    []byte
		dataBytes []byte
		values    = reflect.ValueOf(b).Elem()
	)
	for i, j := 0, 0; i < blocketteLength; j++ {
		var (
			field     = blockette1000SectionMap[j]
			fieldName = field.FieldName
		)
		var (
			fieldSize       = field.FieldSize
			filedValue, err = getStructFieldValue(values, fieldName)
		)

		if err != nil && fieldName != "Reserved" {
			return nil, err
		}

		i += fieldSize
		switch field.FieldType {
		case "int32":
			// "int32" bit width should actually be determined by fieldSize
			result = disassembleInt(filedValue.(int32), fieldSize, bitOrder)
		case "string":
			result = disassembleString(filedValue.(string), fieldSize, ' ')
		}

		// Space as reserved field padding
		if fieldName == "Reserved" {
			result = []byte{0}
		}

		dataBytes = append(dataBytes, result...)
	}

	// Padding reserved field with 0 (56:64)
	for i := 0; i < BLOCKETTE100X_SECTION_LENGTH-blocketteLength; i++ {
		dataBytes = append(dataBytes, 0)
	}

	return dataBytes, nil
}

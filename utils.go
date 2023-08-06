package mseedio

import (
	"fmt"
	"math"
	"reflect"
	"time"
)

// setSignToUint sets signed integer to byte array
func setSignToUint(value, bitWidth uint32) int32 {
	if value>>(bitWidth-1) == 1 {
		offset := int32(math.Pow(2, float64(bitWidth)))
		return int32(value) - offset
	}

	return int32(value)
}

// getSplitedByteArray returns bytes from a number
func getSplitedByteArray(number uint, space, bitOrder int) ([]byte, error) {
	if space <= 0 || space > 32 {
		return nil, fmt.Errorf("invalid bits space value")
	}

	numSegments := 32 / space
	dataArray := make([]byte, 0, numSegments)
	mask := (1 << space) - 1

	if bitOrder == LSBFIRST {
		for i := 0; i < numSegments; i++ {
			data := byte(number & uint(mask))
			dataArray = append(dataArray, data)
			number >>= space
		}

		return dataArray, nil
	}

	for i := numSegments - 1; i >= 0; i-- {
		data := byte((number >> (space * i)) & uint(mask))
		dataArray = append(dataArray, data)
	}

	return dataArray, nil
}

// getBitOrder returns bit order from SectionEndOffset
func getBitOrder(buffer []byte) (int, error) {
	if len(buffer) < 2 {
		return -1, fmt.Errorf("buffer is too short")
	}

	bitOrder := assembleInt(buffer, 2, MSBFIRST)
	if bitOrder == FIXED_SECTION_LENGTH {
		return MSBFIRST, nil
	}

	bitOrder = assembleInt(buffer, 2, LSBFIRST)
	if bitOrder == FIXED_SECTION_LENGTH {
		return LSBFIRST, nil
	}

	return -1, fmt.Errorf("buffer is not SectionEndOffset")
}

// getBlocketteType returns blockette type
func getBlocketteType(buffer []byte, bitOrder int) (int32, error) {
	if len(buffer) < 2 {
		return 0, fmt.Errorf("buffer is too short")
	}

	typ := assembleInt(buffer, 2, bitOrder)
	return typ, nil
}

// getMonthByDays returns month by days of year
func getMonthByDays(year, days int) time.Time {
	if days < 1 || days > 366 {
		return time.Time{}
	}

	return time.Date(year, time.January, days, 0, 0, 0, 0, time.UTC)
}

// getStructFieldValue gets the value of a struct field with reflection
func getStructFieldValue(v reflect.Value, fieldName string) (any, error) {
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return nil, fmt.Errorf("field %s does not exist", fieldName)
	}

	return field.Interface(), nil
}

// setStructFieldValue sets the value of a struct field with reflection
func setStructFieldValue(v reflect.Value, fieldName string, fieldValue any) error {
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return fmt.Errorf("field %s does not exist", fieldName)
	}

	if !field.CanSet() {
		return fmt.Errorf("cannot set field %s", fieldName)
	}

	if field.Type().Kind() != reflect.TypeOf(fieldValue).Kind() {
		return fmt.Errorf("type mismatch for field %s", fieldName)
	}

	field.Set(reflect.ValueOf(fieldValue))
	return nil
}

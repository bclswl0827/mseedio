package mseedio

import (
	"fmt"
	"strconv"
	"strings"
)

func packAscii(buffer []int32) []byte {
	var strSlice []string
	for _, num := range buffer {
		strSlice = append(strSlice, strconv.Itoa(int(num)))
	}

	joinedStr := strings.Join(strSlice, " ")
	return []byte(joinedStr)
}

func packInt(buffer []int32, bitWidth, bitOrder int) (data []byte) {
	n := bitWidth / 8
	for _, v := range buffer {
		data = append(data, disassembleInt(v, n, bitOrder)...)
	}

	return data
}

func packFloat(buffer []int32, bitWidth, bitOrder int) (data []byte) {
	for _, v := range buffer {
		if bitWidth == 32 {
			data = append(data, disassembleFloat(float32(v), bitOrder)...)
		} else {
			data = append(data, disassembleFloat(float64(v), bitOrder)...)
		}
	}

	return data
}

func packSteim1(buffer []int32, bitOrder int) ([]byte, error) {
	if bitOrder == LSBFIRST {
		return nil, fmt.Errorf("Steim1 with LSBFIRST is not allowed")
	}

	// dataLength := len(buffer)
	// x0 := buffer[0]            // Fisrt absolute value
	// xn := buffer[dataLength-1] // Last absolute value

	err := fmt.Errorf("Steim1 encoding is not suuported yet")
	return nil, err
}

func packSteim2(buffer []int32, bitOrder int) ([]byte, error) {
	if bitOrder == LSBFIRST {
		return nil, fmt.Errorf("Steim1 with LSBFIRST is not allowed")
	}

	// dataLength := len(buffer)
	// x0 := buffer[0]            // Fisrt absolute value
	// xn := buffer[dataLength-1] // Last absolute value

	err := fmt.Errorf("Steim1 encoding is not suuported yet")
	return nil, err
}

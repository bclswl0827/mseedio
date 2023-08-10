package mseedio

import (
	"strconv"
	"strings"
)

func packAscii(samples []int32) []byte {
	var strSlice []string
	for _, num := range samples {
		strSlice = append(strSlice, strconv.Itoa(int(num)))
	}

	joinedStr := strings.Join(strSlice, " ")
	return []byte(joinedStr)
}

func packInt(samples []int32, bitWidth, bitOrder int) (data []byte) {
	n := bitWidth / 8
	for _, v := range samples {
		data = append(data, disassembleInt(v, n, bitOrder)...)
	}

	return data
}

func packFloat(samples []int32, bitWidth, bitOrder int) (data []byte) {
	for _, v := range samples {
		if bitWidth == 32 {
			data = append(data, disassembleFloat(float32(v), bitOrder)...)
		} else {
			data = append(data, disassembleFloat(float64(v), bitOrder)...)
		}
	}

	return data
}

func packSteim1(samples []int32, bitOrder int) ([]byte, error) {
	return nil, nil
}

func packSteim2(samples []int32, bitOrder int) ([]byte, error) {
	return nil, nil
}

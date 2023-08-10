package mseedio

import (
	"fmt"
)

// unpackAscii unpacks ASCII data from buffer
func unpackAscii(buffer []byte) string {
	return string(buffer)
}

// unpackInt unpacks int32 array from buffer
func unpackInt(buffer []byte, samples, bitWidth, bitOrder int) (data []int32) {
	space := bitWidth / 8
	for i := space; i < len(buffer); i += space {
		data = append(data, assembleInt(buffer[i-space:i], space, bitOrder))
	}

	return data[:samples]
}

// unpackFloat unpacks float64 array from buffer
func unpackFloat(buffer []byte, samples, bitWidth, bitOrder int) (data []float64) {
	space := bitWidth / 8
	for i := space; i < len(buffer); i += space {
		switch space {
		case 4:
			data = append(data, float64(assembleFloat32(buffer[i-space:i], bitOrder)))
		case 8:
			data = append(data, assembleFloat64(buffer[i-space:i], bitOrder))
		}
	}

	return data[:samples]
}

// unpackSteim1 unpacks Steim1 data from buffer
func unpackSteim1(buffer []byte, samples, bitOrder int) ([]int32, error) {
	if bitOrder == LSBFIRST {
		return nil, fmt.Errorf("Steim1 with LSBFIRST is not allowed")
	}

	dataLength := len(buffer)
	x0 := assembleInt(buffer[4:8], 4, bitOrder)  // Fisrt absolute value
	xn := assembleInt(buffer[8:12], 4, bitOrder) // Last absolute value

	// Get encoding nibbles
	var w0 []uint32
	for i := 4; i < dataLength; i += 64 {
		value := assembleUint(buffer[i-4:i], 4, bitOrder)
		if value != 0 {
			w0 = append(w0, value)
		}
	}

	// Get compression flags
	var cf [][]byte
	for i := 0; i < len(w0); i++ {
		n, _ := getSplitedArray(uint(w0[i]), 2, bitOrder)
		cf = append(cf, n[1:])
	}

	// Get differential raw data
	var wn [][]uint32
	for i := 0; i < dataLength/64; i++ {
		wn = append(wn, []uint32{})
		for j := 4; j < 64; j += 4 {
			offset := i*64 + j
			wn[i] = append(wn[i], assembleUint(buffer[offset:offset+4], 4, bitOrder))
		}
	}

	// Recover from differential nibbles
	var df []int32
	// Go through frames by compression flags
	for i, v := range cf {
		// Go through 2-bit nibble codes
		for ii, vv := range v {
			dat := wn[i][ii]
			switch vv {
			case 0: // Non-data information (x0, xn)
				switch ii {
				case 0:
					x0 = int32(dat)
				case 1:
					xn = int32(dat)
				}
			case 1: // Contains four 8-bit samples
				for idx := 0; idx < 4; idx++ {
					value := (dat >> (24 - idx*8)) & 0xff
					df = append(df, setSignToUint(value, 8))
				}
			case 2: // Contains two 16-bit samples
				for idx := 0; idx < 2; idx++ {
					value := (dat >> (16 - idx*16)) & 0xffff
					df = append(df, setSignToUint(value, 16))
				}
			case 3: // Contains one 32-bit sample
				df = append(df, setSignToUint(dat, 32))
			default:
				err := fmt.Errorf("unknown compression flag")
				return nil, err
			}
		}
	}

	// Recover from differential samples
	res := make([]int32, len(df))
	for i, v := range df {
		if i == 0 {
			res[i] = x0
		} else {
			res[i] = res[i-1] + v
		}
	}

	// Compare xn
	if res[samples-1] != xn {
		err := fmt.Errorf("unpacked samples does not match xn")
		return nil, err
	}

	return res[:samples], nil
}

// unpackSteim2 unpacks Steim2 data from buffer
func unpackSteim2(buffer []byte, samples, bitOrder int) ([]int32, error) {
	if bitOrder == LSBFIRST {
		return nil, fmt.Errorf("Steim2 with LSBFIRST is not allowed")
	}

	dataLength := len(buffer)
	x0 := assembleInt(buffer[4:8], 4, bitOrder)  // Fisrt absolute value
	xn := assembleInt(buffer[8:12], 4, bitOrder) // Last absolute value

	// Get encoding nibbles
	var w0 []uint32
	for i := 4; i < dataLength; i += 64 {
		value := assembleUint(buffer[i-4:i], 4, bitOrder)
		if value != 0 {
			w0 = append(w0, value)
		}
	}

	// Get compression flags
	var cf [][]byte
	for i := 0; i < len(w0); i++ {
		n, _ := getSplitedArray(uint(w0[i]), 2, bitOrder)
		cf = append(cf, n[1:])
	}

	// Get differential raw data
	var wn [][]uint32
	for i := 0; i < dataLength/64; i++ {
		wn = append(wn, []uint32{})
		for j := 4; j < 64; j += 4 {
			offset := i*64 + j
			wn[i] = append(wn[i], assembleUint(buffer[offset:offset+4], 4, bitOrder))
		}
	}

	// Recover from differential nibbles
	var df []int32
	// Go through frames by compression flags
	for i, v := range cf {
		// Go through 2-bit nibble codes
		for ii, vv := range v {
			dat := wn[i][ii]
			switch vv {
			case 0: // Non-data information (x0, xn)
				switch ii {
				case 0:
					x0 = int32(dat)
				case 1:
					xn = int32(dat)
				}
			case 1: // Contains four 8-bit differences
				arr, err := getSplitedArray(uint(dat), 8, bitOrder)
				if err != nil {
					return nil, err
				}
				for _, v := range arr {
					df = append(df, setSignToUint(uint32(v), 8))
				}
			case 2: // get dnib
				dnib := (dat >> 30) & 0x03
				switch dnib {
				case 1: // Wn contains one 30-bit difference
					value := dat & 0x3fffffff
					df = append(df, setSignToUint(value, 30))
				case 2: // Wn contains two 15-bit differences
					for idx := 0; idx < 2; idx++ {
						value := (dat >> (15 - idx*15)) & 0x7fff
						df = append(df, setSignToUint(value, 15))
					}
				case 3: // Wn contains three 10-bit differences
					for idx := 0; idx < 3; idx++ {
						value := (dat >> (20 - idx*10)) & 0x3ff
						df = append(df, setSignToUint(value, 10))
					}
				default:
					err := fmt.Errorf("illegal decode nibble")
					return nil, err
				}
			case 3: // get dnib
				dnib := (dat >> 30) & 0x3fffffff
				switch dnib {
				case 0: // Wn contains five 6-bit differences
					for idx := 0; idx < 5; idx++ {
						value := (dat >> (24 - idx*6)) & 0x3f
						df = append(df, setSignToUint(value, 6))
					}
				case 1: // Wn contains six 5-bit differences
					for idx := 0; idx < 6; idx++ {
						value := (dat >> (25 - idx*5)) & 0x1f
						df = append(df, setSignToUint(value, 5))
					}
				case 2: // Wk contains seven 4-bit differences
					for idx := 0; idx < 7; idx++ {
						value := (dat >> (24 - idx*4)) & 0x0f
						df = append(df, setSignToUint(value, 4))
					}
				default:
					err := fmt.Errorf("illegal decode nibble")
					return nil, err
				}
			default:
				err := fmt.Errorf("unknown compression flag")
				return nil, err
			}
		}
	}

	// Recover from differential samples
	res := make([]int32, len(df))
	for i, v := range df {
		if i == 0 {
			res[i] = x0
		} else {
			res[i] = res[i-1] + v
		}
	}

	// Compare xn
	if res[samples-1] != xn {
		err := fmt.Errorf("unpacked samples does not match xn")
		return nil, err
	}

	return res[:samples], nil
}

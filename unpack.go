package mseedio

import (
	"fmt"
)

func unpackSteim2(data []byte, samples, bitOrder int) ([]int32, error) {
	dataLength := len(data)
	x0 := assembleInt(data[4:8], 4, bitOrder)  // Fisrt absolute value
	xn := assembleInt(data[8:12], 4, bitOrder) // Last absolute value

	// Get encoding nibbles
	var w0 []uint32
	for i := 4; i < dataLength; i += 64 {
		w0 = append(w0, assembleUint(data[i-4:i], 4, bitOrder))
	}

	// Get compression flags
	var cf [][]byte
	for i := 0; i < len(w0); i++ {
		n, _ := getSplitedByteArray(uint(w0[i]), 2, bitOrder)
		cf = append(cf, n[1:])
	}

	// Get differential data
	var wn [][]uint32
	for i := 0; i < dataLength/64; i++ {
		wn = append(wn, []uint32{})
		for j := 4; j < 64; j += 4 {
			offset := i*64 + j
			wn[i] = append(wn[i], assembleUint(data[offset:offset+4], 4, bitOrder))
		}
	}

	// Recover from differential nibbles
	var (
		df  []int32
		res [][]int32
	)
	err := fmt.Errorf("unknown compression flag")
	// Go through frames by compression flags
	for i, v := range cf {
		// Go through 2-bit nibble codes
		for ii, vv := range v {
			switch vv {
			case 0: // Non-data information (x0, xn)
				switch ii {
				case 0:
					x0 = int32(wn[i][ii])
				case 1:
					xn = int32(wn[i][ii])
				}
			case 1: // Contains 4 8-bit differences
				arr, err := getSplitedByteArray(uint(wn[i][ii]), 8, bitOrder)
				if err != nil {
					return nil, err
				}
				for _, v := range arr {
					df = append(df, setSignToUint(uint32(v), 8))
				}
			case 2: // get dnib
				dat := wn[i][ii]
				dnib := (wn[i][ii] >> 30) & 0x03
				switch dnib {
				case 1: // Wn contains one 30-bit difference
					value := (dat >> 0) & ((1 << 30) - 1)
					df = append(df, setSignToUint(value, 30))
				case 2: // Wn contains two 15-bit differences
					for idx := 0; idx < 2; idx++ {
						value := (dat >> (15 - idx*15)) & ((1 << 15) - 1)
						df = append(df, setSignToUint(value, 15))
					}
				case 3: // Wn contains three 10-bit differences
					for idx := 0; idx < 3; idx++ {
						value := (dat >> (20 - idx*10)) & ((1 << 10) - 1)
						df = append(df, setSignToUint(value, 10))
					}
				default:
					err = fmt.Errorf("illegal decode nibble")
					return nil, err
				}
			case 3: // get dnib
				dat := wn[i][ii]
				dnib := (wn[i][ii] >> 30) & 0x03
				switch dnib {
				case 0: // Wn contains five 6-bit differences
					for idx := 0; idx < 5; idx++ {
						value := (dat >> (24 - idx*6)) & ((1 << 6) - 1)
						df = append(df, setSignToUint(value, 6))
					}
				case 1: // Wn contains six 5-bit differences
					for idx := 0; idx < 6; idx++ {
						value := (dat >> (25 - idx*5)) & ((1 << 5) - 1)
						df = append(df, setSignToUint(value, 5))
					}
				case 2: // Wk contains seven 4-bit differences
					for idx := 0; idx < 6; idx++ {
						value := (dat >> (24 - idx*4)) & ((1 << 4) - 1)
						df = append(df, setSignToUint(value, 4))
					}
				default:
					err = fmt.Errorf("illegal decode nibble")
					return nil, err
				}
			default:
				return nil, err
			}
		}

		// Append result slice
		res = append(res, []int32{})
		for i, v := range df {
			if i == 0 {
				res[len(res)-1] = append(res[len(res)-1], x0)
			} else {
				res[len(res)-1] = append(res[len(res)-1], res[len(res)-1][i-1]+v)
			}
		}
	}

	// Compare samples
	if len(df) != samples || len(res[len(res)-1]) != samples {
		err = fmt.Errorf("samples does not match")
		return nil, err
	}

	// Compare x0 and xn
	if res[len(res)-1][0] != x0 || res[len(res)-1][len(res[len(res)-1])-1] != xn {
		err = fmt.Errorf("x0 or xn does not match")
		return nil, err
	}

	return res[len(res)-1], nil
}

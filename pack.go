package mseedio

import (
	"fmt"
)

// packAscii packs ASCII data from buffer
func packAscii(buffer []int32) []byte {
	var strSlice []byte
	for _, num := range buffer {
		strSlice = append(strSlice, byte(num))
	}

	return strSlice
}

// packInt packs int32 array from buffer
func packInt(buffer []int32, bitWidth, bitOrder int) (data []byte) {
	n := bitWidth / 8
	for _, v := range buffer {
		data = append(data, disassembleInt(v, n, bitOrder)...)
	}

	return data
}

// packFloat packs float64 array from buffer
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

// packSteim1 packs Steim1 data from buffer
func packSteim1(buffer []int32, bitOrder int) ([]byte, error) {
	if bitOrder == LSBFIRST {
		return nil, fmt.Errorf("Steim1 with LSBFIRST is not allowed")
	}

	// Get absolute raw data
	dataLength := len(buffer)
	x0 := buffer[0]            // Fisrt absolute value
	xn := buffer[dataLength-1] // Last absolute value

	// Get differential raw data
	df := []int32{0}
	for i := 0; i < dataLength-1; i++ {
		df = append(df, buffer[i+1]-buffer[i])
	}

	// Go through wn to get encoding modes
	var em []byte
	for i := 0; i < len(df); i++ {
		// First 2 elements should be x0 & xn
		if i == 0 {
			em = append(em, []byte{0, 0}...)
		}

		// Append encoding mode by differential value
		if i+3 < len(df) &&
			(df[i] >= -128 && df[i] <= 127) &&
			(df[i+1] >= -128 && df[i+1] <= 127) &&
			(df[i+2] >= -128 && df[i+2] <= 127) &&
			(df[i+3] >= -128 && df[i+3] <= 127) {
			// Four 8-bit samples
			em = append(em, 1)
			i += 3
		} else if i+1 < len(df) &&
			(df[i] >= -32768 && df[i] <= 32767) &&
			(df[i+1] >= -32768 && df[i+1] <= 32767) {
			// Two 16-bit samples
			em = append(em, 2)
			i += 1
		} else {
			// One 32-bit sample
			em = append(em, 3)
		}
	}

	// Split encoding modes to get compression flags
	var cf [][]byte
	for i := 0; i < len(em); i += 15 {
		// Set start & end index of each group
		var (
			startIndex = i
			endIndex   = i + 15
			cfWithW0   []byte
		)
		if endIndex > len(em) {
			endIndex = len(em)
		}

		// Append 0 as placeholder of w0 to first byte of each group
		cfWithW0 = append(cfWithW0, append([]byte{0}, em[startIndex:endIndex]...)...)
		if endIndex < len(em) {
			cf = append(cf, cfWithW0)
		} else {
			// Padding 0 to the remaining bytes
			padding := make([]byte, 15-len(em[i:]))
			cf = append(cf, append(cfWithW0, padding...))
		}
	}

	// Encode compression flags to w0
	var w0 []uint32
	for _, v := range cf {
		value, err := getMergedUint(v, 2, bitOrder)
		if err != nil {
			return nil, err
		}
		if value != 0 {
			w0 = append(w0, uint32(value))
		}
	}

	// Go through and append frames by compression flags
	var res []byte
	var dataOffset int
	for i, v := range cf {
		// Append w0 firstly
		res = append(res, disassembleInt(int32(w0[i]), 4, bitOrder)...)
		// Go through 2-bit nibble codes
		for ii, vv := range v {
			switch vv {
			case 0: // Non-data information (x0, xn)
				switch {
				case i == 0 && ii == 0:
					res = append(res, disassembleInt(x0, 4, bitOrder)...)
				case i == 0 && ii == 1:
					res = append(res, disassembleInt(xn, 4, bitOrder)...)
				}
			case 1: // Contains four 8-bit samples
				dataOffset += 4
				for idx := 4; idx > 0; idx-- {
					res = append(res, disassembleInt(df[dataOffset-idx], 1, bitOrder)...)
				}
			case 2: // Contains two 16-bit samples
				dataOffset += 2
				for idx := 2; idx > 0; idx-- {
					res = append(res, disassembleInt(df[dataOffset-idx], 2, bitOrder)...)
				}
			case 3: // Contains one 32-bit sample
				dataOffset += 1
				res = append(res, disassembleInt(df[dataOffset-1], 4, bitOrder)...)
			default:
				err := fmt.Errorf("unknown compression flag")
				return nil, err
			}
		}
	}

	return res, nil
}

// packSteim2 packs Steim2 data from buffer
func packSteim2(buffer []int32, bitOrder int) ([]byte, error) {
	if bitOrder == LSBFIRST {
		return nil, fmt.Errorf("Steim-2 with LSBFIRST is not allowed")
	}

	// Get absolute raw data
	dataLength := len(buffer)
	x0 := buffer[0]            // Fisrt absolute value
	xn := buffer[dataLength-1] // Last absolute value

	// Get differential raw data
	df := []int32{0}
	for i := 0; i < dataLength-1; i++ {
		df = append(df, buffer[i+1]-buffer[i])
	}

	// Go through wn to get encoding methods & decode nibbles
	var (
		em []byte
		dn []byte
	)
	for i := 0; i < len(df); i++ {
		// First 2 elements should be x0 & xn
		if i == 0 {
			em = append(em, []byte{0, 0}...)
		}

		// Append encoding method by differential value
		if i+6 < len(df) &&
			(df[i] >= -8 && df[i] <= 7) &&
			(df[i+1] >= -8 && df[i+1] <= 7) &&
			(df[i+2] >= -8 && df[i+2] <= 7) &&
			(df[i+3] >= -8 && df[i+3] <= 7) &&
			(df[i+4] >= -8 && df[i+4] <= 7) &&
			(df[i+5] >= -8 && df[i+5] <= 7) &&
			(df[i+6] >= -8 && df[i+6] <= 7) {
			// Seven 4-bit samples
			em = append(em, 3)
			dn = append(dn, 2)
			i += 6
		} else if i+5 < len(df) &&
			(df[i] >= -16 && df[i] <= 15) &&
			(df[i+1] >= -16 && df[i+1] <= 15) &&
			(df[i+2] >= -16 && df[i+2] <= 15) &&
			(df[i+3] >= -16 && df[i+3] <= 15) &&
			(df[i+4] >= -16 && df[i+4] <= 15) &&
			(df[i+5] >= -16 && df[i+5] <= 15) {
			// Six 5-bit samples
			em = append(em, 3)
			dn = append(dn, 1)
			i += 5
		} else if i+4 < len(df) &&
			(df[i] >= -32 && df[i] <= 31) &&
			(df[i+1] >= -32 && df[i+1] <= 31) &&
			(df[i+2] >= -32 && df[i+2] <= 31) &&
			(df[i+3] >= -32 && df[i+3] <= 31) &&
			(df[i+4] >= -32 && df[i+4] <= 31) {
			// Five 6-bit samples
			em = append(em, 3)
			dn = append(dn, 0)
			i += 4
		} else if i+3 < len(df) &&
			(df[i] >= -128 && df[i] <= 127) &&
			(df[i+1] >= -128 && df[i+1] <= 127) &&
			(df[i+2] >= -128 && df[i+2] <= 127) &&
			(df[i+3] >= -128 && df[i+3] <= 127) {
			// Four 8-bit samples
			em = append(em, 1)
			i += 3
		} else if i+2 < len(df) &&
			(df[i] >= -512 && df[i] <= 511) &&
			(df[i+1] >= -512 && df[i+1] <= 511) &&
			(df[i+2] >= -512 && df[i+2] <= 511) {
			// Three 10-bit samples
			em = append(em, 2)
			dn = append(dn, 3)
			i += 2
		} else if i+1 < len(df) &&
			(df[i] >= -16384 && df[i] <= 16383) &&
			(df[i+1] >= -16384 && df[i+1] <= 16383) {
			// Two 15-bit samples
			em = append(em, 2)
			dn = append(dn, 2)
			i += 1
		} else {
			// One 30-bit samples
			em = append(em, 2)
			dn = append(dn, 1)
		}
	}

	// Split encoding methods to get compression flags
	var cf [][]byte
	for i := 0; i < len(em); i += 15 {
		// Set start & end index of each group
		var (
			startIndex = i
			endIndex   = i + 15
			cfWithW0   []byte
		)
		if endIndex > len(em) {
			endIndex = len(em)
		}

		// Append 0 as placeholder of w0 to first byte of each group
		cfWithW0 = append(cfWithW0, append([]byte{0}, em[startIndex:endIndex]...)...)
		if endIndex < len(em) {
			cf = append(cf, cfWithW0)
		} else {
			// Padding 0 to the remaining bytes
			padding := make([]byte, 15-len(em[i:]))
			cf = append(cf, append(cfWithW0, padding...))
		}
	}

	// Encode compression flags to w0
	var w0 []uint32
	for _, v := range cf {
		value, err := getMergedUint(v, 2, bitOrder)
		if err != nil {
			return nil, err
		}
		if value != 0 {
			w0 = append(w0, uint32(value))
		}
	}

	// Go through and append frames by compression flags
	var (
		dataOffset int
		dnibOffset int
		res        []byte
	)
	for i, v := range cf {
		// Append w0 firstly
		res = append(res, disassembleInt(int32(w0[i]), 4, bitOrder)...)
		// Go through 2-bit nibble codes
		for ii, vv := range v {
			switch vv {
			case 0: // Non-data information (x0, xn)
				switch {
				case i == 0 && ii == 0:
					res = append(res, disassembleInt(x0, 4, bitOrder)...)
				case i == 0 && ii == 1:
					res = append(res, disassembleInt(xn, 4, bitOrder)...)
				}
			case 1: // Contains four 8-bit samples
				dataOffset += 4
				for idx := 4; idx > 0; idx-- {
					res = append(res, disassembleInt(df[dataOffset-idx], 1, bitOrder)...)
				}
			case 2: // Determine from dnib
				dnib := dn[dnibOffset]
				value := int32(dnib) << 30
				switch dnib {
				case 1: // Wn contains one 30-bit difference
					dataOffset += 1
					value |= df[dataOffset-1] & 0x3fffffff
				case 2: // Wn contains two 15-bit differences
					dataOffset += 2
					for idx := 2; idx > 0; idx-- {
						value |= (df[dataOffset-idx] & 0x7fff) << ((idx - 1) * 15)
					}
				case 3: // Wn contains three 10-bit differences
					dataOffset += 3
					for idx := 3; idx > 0; idx-- {
						value |= (df[dataOffset-idx] & 0x3ff) << ((idx - 1) * 10)
					}
				default:
					err := fmt.Errorf("illegal decode nibble")
					return nil, err
				}
				dnibOffset += 1
				res = append(res, disassembleInt(value, 4, bitOrder)...)
			case 3: // Determine from dnib
				dnib := dn[dnibOffset]
				value := int32(dnib) << 30
				switch dnib {
				case 0: // Wn contains five 6-bit differences
					dataOffset += 5
					for idx := 5; idx > 0; idx-- {
						value |= (df[dataOffset-idx] & 0x3f) << ((idx - 1) * 6)
					}
				case 1: // Wn contains six 5-bit differences
					dataOffset += 6
					for idx := 6; idx > 0; idx-- {
						value |= (df[dataOffset-idx] & 0x1f) << ((idx - 1) * 5)
					}
				case 2: // Wn contains seven 4-bit differences
					dataOffset += 7
					for idx := 7; idx > 0; idx-- {
						value |= (df[dataOffset-idx] & 0x0f) << ((idx - 1) * 4)
					}
				default:
					err := fmt.Errorf("illegal decode nibble")
					return nil, err
				}
				dnibOffset += 1
				res = append(res, disassembleInt(value, 4, bitOrder)...)
			default:
				err := fmt.Errorf("unknown compression flag")
				return nil, err
			}
		}
	}

	return res, nil
}

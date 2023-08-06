package mseedio

import (
	"math"
	"time"
)

// assembleString assembles a string from bytes
func assembleString(data []byte) string {
	return string(data)
}

// assembleInt assembles a int from n bytes
func assembleInt(data []byte, n int, bitOrder int) int32 {
	var result int32
	if bitOrder == LSBFIRST {
		for i := 0; i < n; i++ {
			result |= int32(data[i]) << uint(i*8)
		}

		return result
	}

	for i := 0; i < n; i++ {
		result |= int32(data[n-i-1]) << uint(i*8)
	}

	return result
}

// assembleUint assembles a uint from n bytes
func assembleUint(data []byte, n int, bitOrder int) uint32 {
	var result uint32
	if bitOrder == LSBFIRST {
		for i := 0; i < n; i++ {
			result |= uint32(data[i]) << uint(i*8)
		}

		return result
	}

	for i := 0; i < n; i++ {
		result |= uint32(data[n-i-1]) << uint(i*8)
	}

	return result
}

// assembleTime assembles a Time from 10 bytes
func assembleTime(data []byte, bitOrder int) time.Time {
	var (
		year int
		days int
		hour int
		min  int
		sec  int
		nsec int
	)

	if bitOrder == LSBFIRST {
		year = int(data[1])<<8 | int(data[0])
		days = int(data[3])<<8 | int(data[2])
		hour = int(data[4])
		min = int(data[5])
		sec = int(data[6])
		nsec = int(data[9])<<16 | int(data[8])<<8 | int(data[7])
	} else {
		year = int(data[0])<<8 | int(data[1])
		days = int(data[2])<<8 | int(data[3])
		hour = int(data[4])
		min = int(data[5])
		sec = int(data[6])
		nsec = int(data[7])<<16 | int(data[8])<<8 | int(data[9])
	}
	nsec *= 100000

	md := getMonthByDays(year, days)
	offset := time.Duration(nsec) * time.Nanosecond
	return time.Date(year, md.Month(), md.Day(), hour, min, sec, 0, time.UTC).Add(offset)
}

// disassembleInt disassembles a int32 to n bytes
func disassembleInt(data int32, n int, bitOrder int) []byte {
	bytes := make([]byte, n)
	if bitOrder == LSBFIRST {
		for i := 0; i < n; i++ {
			bytes[i] = byte(data >> uint(i*8))
		}

		return bytes
	}

	for i := 0; i < n; i++ {
		bytes[i] = byte(data >> uint((n-i-1)*8))
	}

	return bytes
}

// assembleFloat32 assembles a float32 from 4 bytes
func assembleFloat32(data []byte, bitOrder int) float32 {
	var bits uint32
	if bitOrder == LSBFIRST {
		bits = uint32(data[3])<<24 | uint32(data[2])<<16 | uint32(data[1])<<8 | uint32(data[0])
		return math.Float32frombits(bits)
	}

	bits = uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])
	return math.Float32frombits(bits)
}

// assembleFloat64 assembles a float64 from 8 bytes
func assembleFloat64(data []byte, bitOrder int) float64 {
	var bits uint64
	if bitOrder == LSBFIRST {
		bits = uint64(data[7])<<56 | uint64(data[6])<<48 | uint64(data[5])<<40 | uint64(data[4])<<32 | uint64(data[3])<<24 | uint64(data[2])<<16 | uint64(data[1])<<8 | uint64(data[0])
		return math.Float64frombits(bits)
	}

	bits = uint64(data[0])<<56 | uint64(data[1])<<48 | uint64(data[2])<<40 | uint64(data[3])<<32 | uint64(data[4])<<24 | uint64(data[5])<<16 | uint64(data[6])<<8 | uint64(data[7])
	return math.Float64frombits(bits)
}

// disassembleString disassembles a string to bytes
func disassembleString(data string) []byte {
	bytes := []byte(data)
	if len(bytes) > 8 {
		bytes = bytes[:8]
	}

	return bytes
}

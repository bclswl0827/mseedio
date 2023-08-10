package mseedio

import (
	"encoding/binary"
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

// assembleTime assembles a time.Time from 10 bytes
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

// disassembleInt disassembles n int to bytes
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

// disassembleFloat disassembles a float to bytes
func disassembleFloat(data any, bitOrder int) []byte {
	switch v := data.(type) {
	case float32:
		bytes := make([]byte, 4)
		if bitOrder == LSBFIRST {
			binary.LittleEndian.PutUint32(bytes, math.Float32bits(v))
			return bytes
		}
		binary.BigEndian.PutUint32(bytes, math.Float32bits(v))
		return bytes
	case float64:
		bytes := make([]byte, 8)
		if bitOrder == LSBFIRST {
			binary.LittleEndian.PutUint64(bytes, math.Float64bits(v))
			return bytes
		}
		binary.BigEndian.PutUint64(bytes, math.Float64bits(v))
		return bytes
	}

	return nil
}

// disassembleString disassembles a string to bytes
func disassembleString(data string, length int, padding byte) []byte {
	if len(data) > length {
		data = data[:length]
	} else if len(data) < length {
		zeroPadding := make([]byte, length-len(data))
		// fill padding
		for i := range zeroPadding {
			zeroPadding[i] = padding
		}
		data += string(zeroPadding)
	}

	return []byte(data)
}

// disassembleTime disassembles a time.Time to 10 bytes
func disassembleTime(t time.Time, bitOrder int) []byte {
	year := t.Year()
	days := getDaysByDate(t)
	hour, min, sec := t.Hour(), t.Minute(), t.Second()
	nsec := t.Nanosecond() / 100000

	var data []byte
	if bitOrder == LSBFIRST {
		data = []byte{
			byte(year & 0xFF), byte(year >> 8),
			byte(days & 0xFF), byte(days >> 8),
			byte(hour), byte(min), byte(sec),
			byte(nsec & 0xFF), byte((nsec >> 8) & 0xFF), byte(nsec >> 16),
		}
	} else {
		data = []byte{
			byte(year >> 8), byte(year & 0xFF),
			byte(days >> 8), byte(days & 0xFF),
			byte(hour), byte(min), byte(sec),
			byte(nsec >> 16), byte((nsec >> 8) & 0xFF), byte(nsec & 0xFF),
		}
	}

	return data
}

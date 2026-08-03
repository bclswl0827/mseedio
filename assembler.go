package mseedio

import (
	"encoding/binary"
	"math"
	"time"
)

// byteOrder maps the package bit-order flag to a binary.ByteOrder.
func byteOrder(bitOrder int) binary.ByteOrder {
	if bitOrder == LSBFIRST {
		return binary.LittleEndian
	}
	return binary.BigEndian
}

// assembleString assembles a string from bytes.
func assembleString(data []byte) string {
	return string(data)
}

// assembleUint assembles an unsigned integer from the first n bytes of data,
// honoring the given bit order.
func assembleUint(data []byte, n, bitOrder int) uint32 {
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

// assembleInt assembles a signed integer from the first n bytes of data. The
// value is sign-extended from its n-byte width, so widths narrower than 32 bits
// (INT16, INT24, signed header fields such as SampleMultiplier) decode negative
// values correctly.
func assembleInt(data []byte, n, bitOrder int) int32 {
	value := assembleUint(data, n, bitOrder)

	bits := uint(n * 8)
	if bits < 32 && value>>(bits-1)&1 == 1 {
		value |= ^uint32(0) << bits
	}
	return int32(value)
}

// assembleTime assembles a time.Time from a 10-byte BTIME structure.
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

	md := getMonthByDays(year, days)
	return time.Date(year, md.Month(), md.Day(), hour, min, sec, nsec, time.UTC)
}

// assembleFloat32 assembles a float32 from 4 bytes.
func assembleFloat32(data []byte, bitOrder int) float32 {
	return math.Float32frombits(byteOrder(bitOrder).Uint32(data))
}

// assembleFloat64 assembles a float64 from 8 bytes.
func assembleFloat64(data []byte, bitOrder int) float64 {
	return math.Float64frombits(byteOrder(bitOrder).Uint64(data))
}

// disassembleInt disassembles a signed integer into its low n bytes.
func disassembleInt(data int32, n, bitOrder int) []byte {
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

// disassembleFloat disassembles a float32 or float64 into bytes.
func disassembleFloat(data any, bitOrder int) []byte {
	order := byteOrder(bitOrder)
	switch v := data.(type) {
	case float32:
		bytes := make([]byte, 4)
		order.PutUint32(bytes, math.Float32bits(v))
		return bytes
	case float64:
		bytes := make([]byte, 8)
		order.PutUint64(bytes, math.Float64bits(v))
		return bytes
	}
	return nil
}

// disassembleString disassembles a string into exactly length bytes, truncating
// or right-padding with the given padding byte as needed.
func disassembleString(data string, length int, padding byte) []byte {
	if len(data) >= length {
		return []byte(data[:length])
	}

	out := make([]byte, length)
	n := copy(out, data)
	for i := n; i < length; i++ {
		out[i] = padding
	}
	return out
}

// disassembleTime disassembles a time.Time into a 10-byte BTIME structure.
func disassembleTime(t time.Time, bitOrder int) []byte {
	year := t.Year()
	days := getDaysByDate(t)
	hour, min, sec := t.Hour(), t.Minute(), t.Second()
	nsec := t.Nanosecond() / 100000
	if bitOrder == LSBFIRST {
		return []byte{
			byte(year & 0xFF), byte(year >> 8),
			byte(days & 0xFF), byte(days >> 8),
			byte(hour), byte(min), byte(sec),
			byte(nsec & 0xFF), byte((nsec >> 8) & 0xFF), byte(nsec >> 16),
		}
	}

	return []byte{
		byte(year >> 8), byte(year & 0xFF),
		byte(days >> 8), byte(days & 0xFF),
		byte(hour), byte(min), byte(sec),
		byte(nsec >> 16), byte((nsec >> 8) & 0xFF), byte(nsec & 0xFF),
	}
}

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
	order := byteOrder(bitOrder)
	var (
		year = int(order.Uint16(data[0:2]))
		days = int(order.Uint16(data[2:4]))
		hour = int(data[4])
		min  = int(data[5])
		sec  = int(data[6])
		// 0.0001-second ticks stored little-/big-endian across bytes 7..9.
		ticks = int(assembleUint(data[7:10], 3, bitOrder))
	)

	md := getMonthByDays(year, days)
	return time.Date(year, md.Month(), md.Day(), hour, min, sec, ticks*100000, time.UTC)
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
	order := byteOrder(bitOrder)
	ticks := t.Nanosecond() / 100000 // 0.0001-second units

	out := make([]byte, 10)
	order.PutUint16(out[0:2], uint16(t.Year()))
	order.PutUint16(out[2:4], uint16(getDaysByDate(t)))
	out[4] = byte(t.Hour())
	out[5] = byte(t.Minute())
	out[6] = byte(t.Second())
	copy(out[7:10], disassembleInt(int32(ticks), 3, bitOrder))
	return out
}

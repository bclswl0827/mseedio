package mseedio

import "time"

// byteReader is a sequential cursor over a byte buffer that decodes the
// fixed-width fields of a miniSEED record header.
type byteReader struct {
	buf   []byte
	pos   int
	order int
}

// remaining reports how many bytes are still unread.
func (r *byteReader) remaining() int { return len(r.buf) - r.pos }

// string reads n bytes as a string.
func (r *byteReader) string(n int) string {
	s := assembleString(r.buf[r.pos : r.pos+n])
	r.pos += n
	return s
}

// int reads a signed integer from the next n bytes.
func (r *byteReader) int(n int) int32 {
	v := assembleInt(r.buf[r.pos:r.pos+n], n, r.order)
	r.pos += n
	return v
}

// time reads a 10-byte BTIME value.
func (r *byteReader) time() time.Time {
	t := assembleTime(r.buf[r.pos:r.pos+10], r.order)
	r.pos += 10
	return t
}

// skip advances the cursor past n bytes (e.g. reserved fields).
func (r *byteReader) skip(n int) { r.pos += n }

// byteWriter is a sequential cursor that encodes fixed-width header fields.
type byteWriter struct {
	buf   []byte
	order int
}

// string writes s as exactly n bytes, right-padded with pad.
func (w *byteWriter) string(s string, n int, pad byte) {
	w.buf = append(w.buf, disassembleString(s, n, pad)...)
}

// int writes v into n bytes.
func (w *byteWriter) int(v int32, n int) {
	w.buf = append(w.buf, disassembleInt(v, n, w.order)...)
}

// time writes t as a 10-byte BTIME value.
func (w *byteWriter) time(t time.Time) {
	w.buf = append(w.buf, disassembleTime(t, w.order)...)
}

// pad writes n bytes of value b (e.g. reserved fields).
func (w *byteWriter) pad(n int, b byte) {
	for i := 0; i < n; i++ {
		w.buf = append(w.buf, b)
	}
}

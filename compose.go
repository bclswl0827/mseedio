package mseedio

// Compose serializes the fixed section into its 48-byte on-disk form. String
// fields are space-padded and the reserved byte is written as a space, matching
// the SEED fixed-header convention.
func (f *FixedSection) Compose(bitOrder int) ([]byte, error) {
	w := &byteWriter{order: bitOrder}
	w.string(f.SequenceNumber, 6, ' ')
	w.string(f.DataQuality, 1, ' ')
	w.pad(1, ' ') // reserved
	w.string(f.StationCode, 5, ' ')
	w.string(f.LocationCode, 2, ' ')
	w.string(f.ChannelCode, 3, ' ')
	w.string(f.NetworkCode, 2, ' ')
	w.time(f.StartTime)
	w.int(f.SamplesNumber, 2)
	w.int(f.SampleFactor, 2)
	w.int(f.SampleMultiplier, 2)
	w.int(f.ActivityFlags, 1)
	w.int(f.IOClockFlags, 1)
	w.int(f.DataQualityFlags, 1)
	w.int(f.BlockettesFollow, 1)
	w.int(f.TimeCorrection, 4)
	w.int(f.DataStartOffset, 2)
	w.int(f.SectionEndOffset, 2)
	return w.buf, nil
}

// Compose serializes a blockette 1000 (Data Only SEED) into the fixed
// BLOCKETTE100X_SECTION_LENGTH-byte area, zero-padding the reserved tail.
func (b *BlocketteSection) Compose(bitOrder int) ([]byte, error) {
	w := &byteWriter{order: bitOrder}
	w.int(b.BlocketteCode, 2)
	w.int(b.NextBlockette, 2)
	w.int(b.EncodingFormat, 1)
	w.int(b.BitOrder, 1)
	w.int(b.RecordLength, 1)
	w.pad(1, 0) // reserved
	w.pad(BLOCKETTE100X_SECTION_LENGTH-len(w.buf), 0)
	return w.buf, nil
}

// Package mseedio reads and writes miniSEED, the data-record format used for
// seismological time series (SEED, the Standard for the Exchange of Earthquake
// Data).
//
// A miniSEED stream is a sequence of fixed-length data records. Each record
// begins with a 48-byte fixed header (FixedSection), followed by one or more
// blockettes (BlocketteSection) — this package fully supports blockette 1000
// (Data Only SEED) and 1001 (Data Extension) — and then the encoded samples
// (DataSection).
//
// # Reading
//
//	var ms mseedio.MiniSeedData
//	if err := ms.Read("record.mseed"); err != nil {
//		// handle error
//	}
//	for _, s := range ms.Series {
//		fmt.Println(s.DataSection.Decoded)
//	}
//
// Read (or ReadFromReader for an arbitrary io.Reader) auto-detects the byte
// order and decodes every supported sample encoding: ASCII, INT16, INT24,
// INT32, FLOAT32, FLOAT64, and the Steim-1/Steim-2 compressions.
//
// # Writing
//
//	var ms mseedio.MiniSeedData
//	ms.Init(mseedio.STEIM2, mseedio.MSBFIRST)
//	ms.Append(samples, &mseedio.AppendOptions{ /* station, timing, ... */ })
//	out, _ := ms.Encode(mseedio.OVERWRITE, mseedio.MSBFIRST)
//	ms.Write("record.mseed", mseedio.OVERWRITE, out)
//
// Init sets the encoding and byte order, Append adds one blockette-1000 record
// per call, Encode serializes the records to bytes, and Write persists them.
// Note that the Steim compressions require MSBFIRST (big-endian) byte order.
package mseedio

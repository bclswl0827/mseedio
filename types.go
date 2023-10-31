package mseedio

import (
	"time"
)

// The length of the fixed header section
const (
	FIXED_SECTION_LENGTH         = 48
	BLOCKETTE100X_SECTION_LENGTH = 16
)

// First significant bit
const (
	LSBFIRST = 0
	MSBFIRST = 1
)

// Encoding Types
const (
	ASCII   = 0
	INT16   = 1
	INT24   = 2
	INT32   = 3
	FLOAT32 = 4
	FLOAT64 = 5
	STEIM1  = 10
	STEIM2  = 11
)

// Writing mode
const (
	APPEND    = 0
	OVERWRITE = 1
)

// sectionOffset is used when parsing a MiniSeed record
type sectionOffset struct {
	Start int
	End   int
}

// fixedSection is the fixed header section of a MiniSeed record
type fixedSection struct {
	SequenceNumber   string
	DataQuality      string
	StationCode      string
	LocationCode     string
	ChannelCode      string
	NetworkCode      string
	StartTime        time.Time
	SamplesNumber    int32
	SampleFactor     int32
	SampleMultiplier int32
	ActivityFlags    int32
	IOClockFlags     int32
	DataQualityFlags int32
	BlockettesFollow int32
	TimeCorrection   int32
	DataStartOffset  int32
	SectionEndOffset int32
	ReaderOffset     sectionOffset // Used when parsing
}

// blocketteSection is the blockette header section of a MiniSeed record
type blocketteSection struct {
	BlocketteCode  int32         // Blockette 100*
	NextBlockette  int32         // Blockette 100*
	EncodingFormat int32         // Blockette 1000
	BitOrder       int32         // Blockette 1000
	RecordLength   int32         // Blockette 1000
	TimingQuality  int32         // Blockette 1001
	Microseconds   int32         // Blockette 1001
	FrameCount     int32         // Blockette 1001
	ReaderOffset   sectionOffset // Used when parsing
}

// dataSection includes the decoded data and the raw data
type dataSection struct {
	Decoded []any
	RawData []byte
}

// dataSeries corresponds to a single data series in a MiniSeed record
type dataSeries struct {
	DataSection      *dataSection
	FixedSection     *fixedSection
	BlocketteSection *blocketteSection
}

// sectionMap is used when parsing a MiniSeed record
type sectionMap struct {
	FieldName string
	FieldType string
	FieldSize int
}

// MiniSeedData is the main struct for a MiniSeed record
type MiniSeedData struct {
	Type      int
	Order     int
	Records   int
	Samples   int
	StartTime time.Time
	EndTime   time.Time
	Series    []dataSeries
}

// AppendOptions is used when appending a MiniSeed record
type AppendOptions struct {
	SampleRate     float64
	SequenceNumber string
	StationCode    string
	LocationCode   string
	ChannelCode    string
	NetworkCode    string
	StartTime      time.Time
}

var (
	// field name, field type, field size
	fixedSectionMap = []sectionMap{
		{"SequenceNumber", "string", 6},
		{"DataQuality", "string", 1},
		{"Reserved", "", 1},
		{"StationCode", "string", 5},
		{"LocationCode", "string", 2},
		{"ChannelCode", "string", 3},
		{"NetworkCode", "string", 2},
		{"StartTime", "time.Time", 10},
		{"SamplesNumber", "int32", 2},
		{"SampleFactor", "int32", 2},
		{"SampleMultiplier", "int32", 2},
		{"ActivityFlags", "int32", 1},
		{"IOClockFlags", "int32", 1},
		{"DataQualityFlags", "int32", 1},
		{"BlockettesFollow", "int32", 1},
		{"TimeCorrection", "int32", 4},
		{"DataStartOffset", "int32", 2},
		{"SectionEndOffset", "int32", 2},
	}
	// field name, field type, field size
	blockette1000SectionMap = []sectionMap{
		{"BlocketteCode", "int32", 2},
		{"NextBlockette", "int32", 2},
		{"EncodingFormat", "int32", 1},
		{"BitOrder", "int32", 1},
		{"RecordLength", "int32", 1},
		{"Reserved", "", 1},
	}
	// field name, field type, field size
	blockette1001SectionMap = []sectionMap{
		{"BlocketteCode", "int32", 2},
		{"NextBlockette", "int32", 2},
		{"TimingQuality", "int32", 1},
		{"Microseconds", "int32", 1},
		{"Reserved", "", 1},
		{"FrameCount", "int32", 1},
	}
)

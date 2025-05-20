package main

import (
	"fmt"
	"math"
	"time"

	"github.com/bclswl0827/mseedio"
)

func main() {
	var miniseed mseedio.MiniSeedData

	// Init header fields
	miniseed.Init(mseedio.INT32, mseedio.MSBFIRST)

	// For generating test waveform
	const (
		sampleRate       = 100
		sineFreq         = 2.0
		amplitude        = 1000.0
		samplesPerRecord = 100
	)

	startTime := time.Now()
	for i := 0; i < 600; i++ {
		data := getSineWaveData(samplesPerRecord, sampleRate, sineFreq, amplitude, float64(i*samplesPerRecord))
		t := startTime.Add(time.Duration(i*samplesPerRecord) * time.Second / time.Duration(sampleRate))

		// Append records
		err := miniseed.Append(data, &mseedio.AppendOptions{
			SampleRate:     sampleRate,
			StartTime:      t,
			SequenceNumber: fmt.Sprintf("%06d", i+1),
			StationCode:    "AAAAA",
			LocationCode:   "BB",
			ChannelCode:    "EHZ",
			NetworkCode:    "CC",
		})
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	// Encode data
	dataBytes, err := miniseed.Encode(mseedio.OVERWRITE, mseedio.MSBFIRST)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Write to file
	err = miniseed.Write("test.mseed", mseedio.OVERWRITE, dataBytes)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Write success")
}

func getSineWaveData(length int, sampleRate int, frequency float64, amplitude float64, phaseOffset float64) []int32 {
	data := make([]int32, length)
	for i := 0; i < length; i++ {
		t := float64(i) / float64(sampleRate)
		theta := 2 * math.Pi * frequency * (t + phaseOffset/float64(sampleRate))
		data[i] = int32(amplitude * math.Sin(theta))
	}
	return data
}

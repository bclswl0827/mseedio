package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/bclswl0827/mseedio"
)

func main() {
	var miniseed mseedio.MiniSeedData

	// Init header fields
	miniseed.Init(mseedio.INT32, mseedio.MSBFIRST)

	// Append records and increment sequence number
	startTime := time.Now()
	for i := 0; i < 60; i++ {
		// Generate random data
		data, err := getRandomData(101, -1000, 1000)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Append records
		t := startTime.Add(time.Second)
		err = miniseed.Append(data, mseedio.AppendOptions{
			SampleRate:     100,
			StartTime:      t,
			SequenceNumber: fmt.Sprintf("%06d", i),
			StationCode:    "AAAAA",
			LocationCode:   "BB",
			ChannelCode:    "EHZ",
			NetworkCode:    "CC",
		})
		if err != nil {
			fmt.Println(err)
			return
		}

		startTime = t
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

func getRandomData(length int, min, max int32) ([]int32, error) {
	if max <= min {
		return nil, fmt.Errorf("invalid range")
	}

	if length < 0 {
		return nil, fmt.Errorf("invalid length")
	}

	data := make([]int32, length)
	rangeSize := int64(max - min + 1)

	for i := 0; i < length; i++ {
		randomBigInt, err := rand.Int(rand.Reader, big.NewInt(rangeSize))
		if err != nil {
			return nil, err
		}
		data[i] = min + int32(randomBigInt.Int64())
	}

	return data, nil
}

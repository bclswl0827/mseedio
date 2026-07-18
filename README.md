# mseedio

`mseedio` is a lightweight Go module for reading and writing MiniSEED files, a commonly used format for seismological time series data.

## Features

- Read MiniSEED files and arbitrary `io.Reader` sources
- Auto-detects byte order
- Supports encoded sample formats:
  - `ASCII`
  - `INT16`, `INT24`, `INT32`
  - `FLOAT32`, `FLOAT64`
  - `Steim-1`, `Steim-2`
- Write MiniSEED records with blockette 1000 and 1001 support
- Includes example reader and writer programs

## Installation

```shell
$ go get github.com/bclswl0827/mseedio
```

## Basic usage

### Read MiniSEED files

```go
package main

import (
    "fmt"
    "github.com/bclswl0827/mseedio"
)

func main() {
    var ms mseedio.MiniSeedData
    if err := ms.Read("record.mseed"); err != nil {
        panic(err)
    }

    for _, s := range ms.Series {
        fmt.Println(s.DataSection.Decoded)
    }
}
```

Use `ReadFromReader` to decode a MiniSEED stream from any `io.Reader`.

### Write MiniSEED files

```go
package main

import (
    "fmt"
    "github.com/bclswl0827/mseedio"
)

func main() {
    var ms mseedio.MiniSeedData
    ms.Init(mseedio.INT32, mseedio.MSBFIRST)

    data := []int32{0, 1, 2, 3, 4}
    err := ms.Append(data, &mseedio.AppendOptions{
        SampleRate:     100,
        StartTime:      time.Now(),
        SequenceNumber: "000001",
        StationCode:    "AAAAA",
        LocationCode:   "BB",
        ChannelCode:    "EHZ",
        NetworkCode:    "CC",
    })
    if err != nil {
        panic(err)
    }

    out, err := ms.Encode(mseedio.OVERWRITE, mseedio.MSBFIRST)
    if err != nil {
        panic(err)
    }

    if err := ms.Write("test.mseed", mseedio.OVERWRITE, out); err != nil {
        panic(err)
    }

    fmt.Println("written test.mseed")
}
```

## Examples

See the `example/reader` and `example/writer` directories for working sample programs.

## License

This project is licensed under the MIT License.

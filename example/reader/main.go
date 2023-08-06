package main

import (
	"fmt"
	"reflect"

	"github.com/bclswl0827/mseedio"
)

func main() {
	var miniseed mseedio.MiniSeedData
	err := miniseed.Read("./testdata.mseed")
	if err != nil {
		fmt.Println(err)
	}

	for _, v := range miniseed.Series {
		printFields(v.FixedSection)
		printFields(v.BlocketteSection)
		fmt.Println("DataSeries:", v.DataSection.Decoded)
		fmt.Println()
	}
}

func printFields(obj any) {
	value := reflect.ValueOf(obj)
	typ := reflect.TypeOf(obj)

	if typ.Kind() != reflect.Struct {
		fmt.Println("Object is not a struct")
		return
	}

	for i := 0; i < value.NumField(); i++ {
		fieldValue := value.Field(i)
		fieldType := typ.Field(i)

		fmt.Printf("%s: %v\n", fieldType.Name, fieldValue.Interface())
	}
}

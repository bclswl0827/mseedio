package mseedio

import "fmt"

// m.Init() Initialize MiniSeedData with data type and bit order
func (m *MiniSeedData) Init(dataType, bitOrder int) error {
	if len(m.Series) > 0 {
		return fmt.Errorf("empty data series is required")
	}

	m.Type = dataType
	m.Order = bitOrder
	return nil
}

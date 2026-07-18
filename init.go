package mseedio

import "fmt"

// Init prepares an empty MiniSeedData with the given encoding type and bit
// order, ready for Append. It fails if the record already holds data series.
func (m *MiniSeedData) Init(dataType, bitOrder int) error {
	if len(m.Series) > 0 {
		return fmt.Errorf("cannot initialize a MiniSeedData that already has data series")
	}

	m.Type = dataType
	m.Order = bitOrder
	return nil
}

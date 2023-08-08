package mseedio

func packString(samples string) []byte {
	return []byte(samples)
}

func packInt(samples []int32, bitWidth, bitOrder int) (data []byte) {
	return data
}

func packFloat(samples []byte, bitWidth, bitOrder int) (data []byte) {
	return data
}

func packSteim1(samples []byte, bitOrder int) ([]byte, error) {
	return nil, nil
}

func packSteim2(samples []byte, bitOrder int) ([]byte, error) {
	return nil, nil
}

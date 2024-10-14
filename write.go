package mseedio

import (
	"bufio"
	"os"
)

// m.Write() writes the MiniSeedData to file depending on the mode (APPEND, OVERWRITE)
func (m *MiniSeedData) Write(filePath string, writeMode int, dataBytes []byte) error {
	var (
		err  error
		file *os.File
	)
	if writeMode == APPEND {
		file, err = os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	} else {
		file, err = os.Create(filePath)
	}

	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.Write(dataBytes)
	if err != nil {
		return err
	}

	err = writer.Flush()
	if err != nil {
		return err
	}

	return nil
}

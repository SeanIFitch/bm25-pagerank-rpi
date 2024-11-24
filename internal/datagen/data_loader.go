package datagen

import (
	"encoding/gob"
	"log"
	"os"
	"path/filepath"
)

// SaveData saves X and Y data to a file.
func SaveData(filename string, X, Y interface{}) error {
	// Ensure the directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	// Create the file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		// Close the file and handle errors if they occur
		if closeErr := file.Close(); closeErr != nil {
			log.Printf("warning: failed to close file: %v\n", closeErr)
		}
	}()

	// Initialize the gob encoder and encode the data
	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(X); err != nil {
		return err
	}
	if err := encoder.Encode(Y); err != nil {
		return err
	}

	return nil
}

// LoadData loads X and Y data from a file.
func LoadData(filename string, X, Y interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("warning: failed to close file: %v\n", err)
		}
	}(file)

	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(X); err != nil {
		return err
	}
	if err := decoder.Decode(Y); err != nil {
		return err
	}

	return nil
}

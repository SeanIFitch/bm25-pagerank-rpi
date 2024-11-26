package ranking

import (
	"encoding/gob"
	"fmt"
	"github.com/google/uuid"
	"log"
	"os"
	"path/filepath"
	"time"
)

// saveData saves X and Y data to a file.
func saveData(filename string, X interface{}) error {
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

	return nil
}

func generateUniqueFilename(base string) string {
	timestamp := time.Now().Format("20060102_150405.000000000")
	uniqueID := uuid.New().String()
	return fmt.Sprintf("%s_%s_%s.gob", base, timestamp, uniqueID)
}

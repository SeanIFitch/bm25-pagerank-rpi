package datagen

import (
	"encoding/csv"
	"encoding/gob"
	"log"
	"os"
	"path/filepath"
	"rpi-search-ranking/internal/ranking"
	"strconv"
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

func SaveDataToCSV(filename string, X []ranking.Features, Y []int) error {
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

	// Initialize the CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header (columns)
	header := []string{
		"CoveredQueryTermNumber", "CoveredQueryTermRatio",
		"SumTermFrequency", "MinTermFrequency", "MaxTermFrequency", "MeanTermFrequency", "VarianceTermFrequency",
		"StreamLength", "SumStreamLengthNormalizedTF", "MinStreamLengthNormalizedTF", "MaxStreamLengthNormalizedTF",
		"MeanStreamLengthNormalizedTF", "VarianceStreamLengthNormalizedTF",
		"SumTFIDF", "MinTFIDF", "MaxTFIDF", "MeanTFIDF", "VarianceTFIDF",
		"BM25", "NumSlashesInURL", "LengthOfURL",
		"InlinkCount", "OutlinkCount", "PageRank", "Y",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data rows
	for i := 0; i < len(X); i++ {
		// Convert Features values to strings
		record := []string{
			strconv.Itoa(X[i].CoveredQueryTermNumber),
			strconv.FormatFloat(X[i].CoveredQueryTermRatio, 'f', 6, 64),
			strconv.Itoa(X[i].SumTermFrequency),
			strconv.Itoa(X[i].MinTermFrequency),
			strconv.Itoa(X[i].MaxTermFrequency),
			strconv.FormatFloat(X[i].MeanTermFrequency, 'f', 6, 64),
			strconv.FormatFloat(X[i].VarianceTermFrequency, 'f', 6, 64),
			strconv.Itoa(X[i].StreamLength),
			strconv.FormatFloat(X[i].SumStreamLengthNormalizedTF, 'f', 6, 64),
			strconv.FormatFloat(X[i].MinStreamLengthNormalizedTF, 'f', 6, 64),
			strconv.FormatFloat(X[i].MaxStreamLengthNormalizedTF, 'f', 6, 64),
			strconv.FormatFloat(X[i].MeanStreamLengthNormalizedTF, 'f', 6, 64),
			strconv.FormatFloat(X[i].VarianceStreamLengthNormalizedTF, 'f', 6, 64),
			strconv.FormatFloat(X[i].SumTFIDF, 'f', 6, 64),
			strconv.FormatFloat(X[i].MinTFIDF, 'f', 6, 64),
			strconv.FormatFloat(X[i].MaxTFIDF, 'f', 6, 64),
			strconv.FormatFloat(X[i].MeanTFIDF, 'f', 6, 64),
			strconv.FormatFloat(X[i].VarianceTFIDF, 'f', 6, 64),
			strconv.FormatFloat(X[i].BM25, 'f', 6, 64),
			strconv.Itoa(X[i].NumSlashesInURL),
			strconv.Itoa(X[i].LengthOfURL),
			strconv.Itoa(X[i].InlinkCount),
			strconv.Itoa(X[i].OutlinkCount),
			strconv.FormatFloat(X[i].PageRank, 'f', 6, 64),
			strconv.Itoa(Y[i]),
		}

		// Write the record to the CSV file
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

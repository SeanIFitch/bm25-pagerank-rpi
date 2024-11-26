package main

import (
	"flag"
	"log"
	"os"
	"rpi-search-ranking/internal/datagen"
)

// Create microsoft pairwise comparison dataset
// Designed to use data from https://www.microsoft.com/en-us/research/project/mslr/ by Tao Qin and Tie-Yan Liu
func main() {
	file := flag.String("file", "", "Path to the train dataset file (e.g., MSLR-WEB30K/Fold1/train.txt)")
	csvSave := flag.String("csvFile", "", "Path to the file in which to save the file as CSV (e.g., data/processed/MSLR-WEB30K/Fold1/1mil-train.csv)")
	gobSave := flag.String("gobFile", "", "Path to the file in which to save the file as gob (e.g., data/processed/MSLR-WEB30K/Fold1/1mil-train.gob)")
	exampleCount := flag.Int("trainCount", 1000000, "Number of examples to save")
	minDiff := flag.Int("minDiff", 3, "Minimum relevance difference for a valid example")

	flag.Parse()

	// Ensure required file paths are provided
	if *file == "" || (*csvSave == "" && *gobSave == "") {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Validate minDiff range
	if *minDiff < 1 || *minDiff > 4 {
		log.Fatal("Error: Minimum relevance difference must be between 1 and 4")
	}

	X, Y, err := datagen.CreateExamples(*file, *exampleCount, *minDiff)
	if err != nil {
		log.Fatal(err)
	}

	if *gobSave != "" {
		err = datagen.SaveData(*gobSave, X, Y)
		if err != nil {
			return
		}
	} else if *csvSave != "" {
		if err != nil {
			return
		}
		err = datagen.SaveDataToCSV(*csvSave, X, Y)
		if err != nil {
			return
		}
	}

}

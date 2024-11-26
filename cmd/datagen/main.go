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
	trainFile := flag.String("trainFile", "", "Path to the train dataset file (e.g., MSLR-WEB30K/Fold1/train.txt)")
	testFile := flag.String("testFile", "", "Path to the test dataset file (e.g., MSLR-WEB30K/Fold1/test.txt)")
	trainSave := flag.String("saveTrainFile", "", "Path to the file in which to save the train dataset (e.g., data/processed/MSLR-WEB30K/Fold1/1mil-train.gob)")
	testSave := flag.String("saveTestFile", "", "Path to the file in which to save the test dataset (e.g., data/processed/MSLR-WEB30K/Fold1/100k-test.gob)")
	trainCount := flag.Int("trainCount", 1000000, "Number of train examples to save")
	testCount := flag.Int("testCount", 100000, "Number of test examples to save")
	minDiff := flag.Int("minDiff", 3, "Minimum relevance difference for a valid example")
	fileType := flag.String("fileType", "gob", "Either gob or csv to save a go binary file or csv")

	flag.Parse()

	// Ensure required file paths are provided
	if *trainFile == "" || *testFile == "" || *trainSave == "" || *testSave == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *fileType != "gob" && *fileType != "csv" {
		log.Fatal("Error: fileType must be one of gob or csv")
	}

	// Validate minDiff range
	if *minDiff < 1 || *minDiff > 4 {
		log.Fatal("Error: Minimum relevance difference must be between 1 and 4")
	}

	XTrain, YTrain, err := datagen.CreateExamples(*trainFile, *trainCount, *minDiff)
	if err != nil {
		log.Fatal(err)
	}

	XTest, YTest, err := datagen.CreateExamples(*trainFile, *testCount, *minDiff)
	if err != nil {
		log.Fatal(err)
	}

	if *fileType == "gob" {
		err = datagen.SaveData(*trainSave, XTrain, YTrain)
		if err != nil {
			return
		}
		err = datagen.SaveData(*testSave, XTest, YTest)
		if err != nil {
			return
		}
	} else if *fileType == "csv" {
		err = datagen.SaveDataToCSV(*trainSave, XTrain, YTrain)
		if err != nil {
			return
		}
		err = datagen.SaveDataToCSV(*testSave, XTest, YTest)
		if err != nil {
			return
		}
	}

}

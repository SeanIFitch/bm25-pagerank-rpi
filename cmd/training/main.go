package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"rpi-search-ranking/internal/training"
)

// Create microsoft dataset
// Designed to use data from https://www.microsoft.com/en-us/research/project/mslr/ by Tao Qin and Tie-Yan Liu
func main() {
	// Define a CLI flag for the file path
	filePath := flag.String("file", "", "Path to the dataset file (e.g., Fold1/train.txt)")
	flag.Parse()

	// Check if the file path is provided
	if *filePath == "" {
		fmt.Println("Error: File path is required. Use -file to specify the dataset file.")
		os.Exit(1)
	}

	// Load the dataset from the specified file path
	relevances, qids, features, err := training.LoadDataset(*filePath)
	if err != nil {
		fmt.Println("Error loading dataset:", err)
		return
	}

	X, Y := training.CreateExamples(relevances, qids, features, 10000)

	// Shuffle the data
	training.ShuffleData(X, Y)

	// Define the split ratio
	trainSize := int(0.8 * float64(len(X))) // 80% for training

	// Split the data into training and test sets
	XTrain, YTrain := X[:trainSize], Y[:trainSize]
	XTest, YTest := X[trainSize:], Y[trainSize:]

	// Initialize the logistic regression model
	lr := &training.LogisticRegression{}

	// Train the model on the training data
	err = lr.Train(XTrain, YTrain, 0.01, 10000)
	if err != nil {
		log.Fatal(err)
	}

	// Confusion matrix variables
	TP, FP, TN, FN := 0, 0, 0, 0

	// Make predictions and evaluate on the test data
	for i, features := range XTest {
		class := lr.PredictClass(features)
		actual := YTest[i]

		// Update confusion matrix based on the prediction and actual class
		if class == 1 && actual == 1 {
			TP++ // True Positive
		} else if class == 1 && actual == -1 {
			FP++ // False Positive
		} else if class == -1 && actual == -1 {
			TN++ // True Negative
		} else if class == -1 && actual == 1 {
			FN++ // False Negative
		}
	}

	// Calculate accuracy
	accuracy := float64(TP+TN) / float64(TP+FP+TN+FN) * 100
	fmt.Printf("Test Accuracy: %.2f%%\n", accuracy)

	// Print Confusion Matrix
	fmt.Printf("Confusion Matrix:\n")
	fmt.Printf("              Predicted\n")
	fmt.Printf("              1     -1\n")
	fmt.Printf("Actual  1    %d    %d\n", TP, FN)
	fmt.Printf("        -1   %d    %d\n", FP, TN)
}

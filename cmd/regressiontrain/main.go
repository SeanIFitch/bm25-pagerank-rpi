package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"rpi-search-ranking/internal/datagen"
	"rpi-search-ranking/internal/ranking"
	"rpi-search-ranking/internal/training"
)

// Current output:
//
// Best Lambda: 1.5000, Cross-Validation Accuracy: 67.48%
// Early stopping at epoch 684
// Test Accuracy: 67.59%
// Confusion Matrix:
//				  Predicted
//				  1		 -1
//  Actual	 1	33858	16008
//			-1	16399	33735

// Train model
func main() {
	trainFile := flag.String("trainFile", "", "Path to the train dataset file (e.g., data/processed/MSLR-WEB30K/Fold1/train.gob)")
	testFile := flag.String("testFile", "", "Path to the test dataset file (e.g., data/processed/MSLR-WEB30K/Fold1/test.gob)")
	flag.Parse()

	// Ensure required file paths are provided
	if *trainFile == "" || *testFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Load train and test data back
	var XTrain, XTest []ranking.Features
	var YTrain, YTest []int
	if err := datagen.LoadData(*trainFile, &XTrain, &YTrain); err != nil {
		log.Fatalf("Error loading train data: %v", err)
	}
	if err := datagen.LoadData(*testFile, &XTest, &YTest); err != nil {
		log.Fatalf("Error loading test data: %v", err)
	}

	// Define lambda values to search through
	lambdaValues := []float64{1.0, 1.25, 1.5, 1.75, 2.0, 2.25}

	// Perform Grid Search CV to find the best lambda
	bestLambda, bestAcc := training.GridSearchCV(XTrain, YTrain, lambdaValues, 5)
	fmt.Printf("Best Lambda: %.4f, Best Cross-Validation Accuracy: %.2f%%\n", bestLambda, bestAcc)

	// Train the final model with the best lambda on the full training set
	lr := training.NewLogisticRegression(bestLambda)
	err := lr.Train(XTrain, YTrain, 0.02, 1000)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(lr.Weights)

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

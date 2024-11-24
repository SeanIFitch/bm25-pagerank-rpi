package training

import (
	"fmt"
	"gonum.org/v1/gonum/mat"
	"log"
	"math"
	"math/rand"
	"rpi-search-ranking/internal/ranking"
)

// LogisticRegression represents a logistic regression model
type LogisticRegression struct {
	Weights     *mat.VecDense
	bias        float64
	lambda      float64 // L2 regularization parameter
	featureMean []float64
	featureStd  []float64
}

// NewLogisticRegression creates a new logistic regression model with specified L2 strength
func NewLogisticRegression(lambda float64) *LogisticRegression {
	return &LogisticRegression{
		lambda: lambda,
	}
}

// featureToVector converts a Features struct to a slice of float64
func featureToVector(f ranking.Features) []float64 {
	return []float64{
		float64(f.CoveredQueryTermNumber),
		f.CoveredQueryTermRatio,
		float64(f.SumTermFrequency),
		float64(f.MinTermFrequency),
		float64(f.MaxTermFrequency),
		f.MeanTermFrequency,
		f.VarianceTermFrequency,
		float64(f.StreamLength),
		f.SumStreamLengthNormalizedTF,
		f.MinStreamLengthNormalizedTF,
		f.MaxStreamLengthNormalizedTF,
		f.MeanStreamLengthNormalizedTF,
		f.VarianceStreamLengthNormalizedTF,
		f.SumTFIDF,
		f.MinTFIDF,
		f.MaxTFIDF,
		f.MeanTFIDF,
		f.VarianceTFIDF,
		f.BM25,
		float64(f.NumSlashesInURL),
		float64(f.LengthOfURL),
		float64(f.InlinkCount),
		float64(f.OutlinkCount),
		f.PageRank,
	}
}

func (lr *LogisticRegression) standardizeFeatures(features []ranking.Features) (*mat.Dense, error) {
	numSamples := len(features)
	if numSamples == 0 {
		return nil, fmt.Errorf("empty feature set")
	}

	// Convert first feature to get dimensions
	firstVec := featureToVector(features[0])
	numFeatures := len(firstVec)

	// Initialize feature matrix
	X := mat.NewDense(numSamples, numFeatures, nil)

	// Fill feature matrix
	for i, f := range features {
		X.SetRow(i, featureToVector(f))
	}

	// Compute mean and std if not already computed (training phase)
	if lr.featureMean == nil {
		lr.featureMean = make([]float64, numFeatures)
		lr.featureStd = make([]float64, numFeatures)

		// Compute mean
		for j := 0; j < numFeatures; j++ {
			col := mat.Col(nil, j, X)
			sum := 0.0
			for _, val := range col {
				sum += val
			}
			lr.featureMean[j] = sum / float64(numSamples)
		}

		// Compute std
		for j := 0; j < numFeatures; j++ {
			col := mat.Col(nil, j, X)
			sumSquared := 0.0
			for _, val := range col {
				diff := val - lr.featureMean[j]
				sumSquared += diff * diff
			}
			lr.featureStd[j] = math.Sqrt(sumSquared / float64(numSamples))
			if lr.featureStd[j] == 0 {
				lr.featureStd[j] = 1 // Prevent division by zero
			}
		}
	}

	// Standardize features
	standardized := mat.NewDense(numSamples, numFeatures, nil)
	for i := 0; i < numSamples; i++ {
		for j := 0; j < numFeatures; j++ {
			val := X.At(i, j)
			standardizedVal := (val - lr.featureMean[j]) / lr.featureStd[j]
			standardized.Set(i, j, standardizedVal)
		}
	}

	return standardized, nil
}

// sigmoid computes the sigmoid function
func sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

// Train trains the logistic regression model with early stopping
func (lr *LogisticRegression) Train(features []ranking.Features, labels []int, learningRate float64, numEpochs int) error {
	if len(features) != len(labels) {
		return fmt.Errorf("number of features (%d) does not match number of labels (%d)", len(features), len(labels))
	}
	if len(features) == 0 {
		return fmt.Errorf("empty training data")
	}

	// Standardize features
	X, err := lr.standardizeFeatures(features)
	if err != nil {
		return err
	}

	numSamples, numFeatures := X.Dims()
	y := mat.NewVecDense(len(labels), nil)

	// Fill the label vector
	for i, label := range labels {
		if label == 1 {
			y.SetVec(i, 1.0)
		} else {
			y.SetVec(i, 0.0)
		}
	}

	// Initialize weights with Xavier/Glorot initialization
	weights := make([]float64, numFeatures)
	limit := math.Sqrt(6.0 / float64(numFeatures))
	for i := range weights {
		weights[i] = (2.0*rand.Float64() - 1.0) * limit
	}

	lr.Weights = mat.NewVecDense(numFeatures, weights)
	lr.bias = 0.0

	prevLoss := math.Inf(1)
	patience := 5
	noImprovement := 0

	// Gradient descent with early stopping
	for epoch := 0; epoch < numEpochs; epoch++ {
		// Forward pass
		predictions := mat.NewVecDense(numSamples, nil)
		for i := 0; i < numSamples; i++ {
			xRow := mat.Row(nil, i, X)
			vecXRow := mat.NewVecDense(len(xRow), xRow)
			z := mat.Dot(vecXRow, lr.Weights) + lr.bias
			predictions.SetVec(i, sigmoid(z))
		}

		// Compute loss with L2 regularization
		loss := 0.0
		for i := 0; i < numSamples; i++ {
			yi := y.AtVec(i)
			pi := predictions.AtVec(i)
			loss -= yi*math.Log(pi+1e-15) + (1-yi)*math.Log(1-pi+1e-15)
		}
		loss /= float64(numSamples)

		// Add L2 regularization term to loss
		l2Term := 0.0
		for j := 0; j < numFeatures; j++ {
			l2Term += lr.Weights.AtVec(j) * lr.Weights.AtVec(j)
		}
		loss += 0.5 * lr.lambda * l2Term

		// Early stopping check
		if loss >= prevLoss {
			noImprovement++
			if noImprovement >= patience {
				fmt.Printf("Early stopping at epoch %d\n", epoch)
				break
			}
		} else {
			noImprovement = 0
		}
		prevLoss = loss

		if epoch%10 == 0 {
			fmt.Printf("Epoch %d, Loss: %.4f\n", epoch, loss)
		}

		// Compute gradients with L2 regularization
		predError := mat.NewVecDense(numSamples, nil)
		predError.SubVec(predictions, y)

		// Update weights with L2 regularization
		gradW := mat.NewVecDense(numFeatures, nil)
		for j := 0; j < numFeatures; j++ {
			xCol := mat.Col(nil, j, X)
			vecXCol := mat.NewVecDense(numSamples, xCol)
			l2Term := lr.lambda * lr.Weights.AtVec(j)
			gradW.SetVec(j, (mat.Dot(predError, vecXCol)/float64(numSamples))+l2Term)
		}

		// Update bias with L2 regularization
		gradB := mat.Sum(predError)/float64(numSamples) + lr.lambda*lr.bias

		// Apply updates with adaptive learning rate
		lr.Weights.AddScaledVec(lr.Weights, -learningRate, gradW)
		lr.bias -= learningRate * gradB
	}

	return nil
}

// predict makes predictions for new features
func (lr *LogisticRegression) predict(features ranking.Features) float64 {
	if lr.Weights == nil {
		return 0.0
	}

	// Convert and standardize features
	x := featureToVector(features)
	standardizedX := make([]float64, len(x))
	for i := range x {
		standardizedX[i] = (x[i] - lr.featureMean[i]) / lr.featureStd[i]
	}

	vecX := mat.NewVecDense(len(standardizedX), standardizedX)
	z := mat.Dot(vecX, lr.Weights) + lr.bias
	return sigmoid(z)
}

// PredictClass predicts the class (1 or -1) for new features
func (lr *LogisticRegression) PredictClass(features ranking.Features) int {
	prob := lr.predict(features)
	if prob >= 0.5 {
		return 1
	}
	return -1
}

// GridSearchCV performs grid search with cross-validation for hyperparameter tuning
func GridSearchCV(XTrain []ranking.Features, YTrain []int, lambdaValues []float64, numFolds int) (bestLambda float64, bestAccuracy float64) {
	// Split data into folds for cross-validation
	foldSize := len(XTrain) / numFolds
	var bestAcc float64
	var bestLambdaValue float64

	for _, lambda := range lambdaValues {
		var totalAccuracy float64

		// Perform cross-validation
		for fold := 0; fold < numFolds; fold++ {
			// Split training data into train and validation sets
			startIdx := fold * foldSize
			endIdx := startIdx + foldSize
			if fold == numFolds-1 {
				endIdx = len(XTrain) // For the last fold, include the rest of the data
			}

			// Create training and validation sets
			XTrainFold := append(XTrain[:startIdx], XTrain[endIdx:]...)
			YTrainFold := append(YTrain[:startIdx], YTrain[endIdx:]...)
			XValFold := XTrain[startIdx:endIdx]
			YValFold := YTrain[startIdx:endIdx]

			// Train the model with the given lambda
			lr := NewLogisticRegression(lambda)
			err := lr.Train(XTrainFold, YTrainFold, 0.02, 100)
			if err != nil {
				log.Fatal(err)
			}

			// Evaluate the model on the validation set
			accuracy := evaluateModel(lr, XValFold, YValFold)
			totalAccuracy += accuracy
		}

		// Calculate average accuracy for this lambda
		avgAccuracy := totalAccuracy / float64(numFolds)
		fmt.Printf("Lambda: %.4f, Cross-Validation Accuracy: %.2f%%\n", lambda, avgAccuracy)

		// Track the best lambda value based on cross-validation accuracy
		if avgAccuracy > bestAcc {
			bestAcc = avgAccuracy
			bestLambdaValue = lambda
		}
	}

	return bestLambdaValue, bestAcc
}

// evaluateModel evaluates the trained model on a validation set
func evaluateModel(lr *LogisticRegression, XVal []ranking.Features, YVal []int) float64 {
	TP, FP, TN, FN := 0, 0, 0, 0

	// Make predictions and evaluate on the validation data
	for i, features := range XVal {
		class := lr.PredictClass(features)
		actual := YVal[i]

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
	return accuracy
}

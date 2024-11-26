package datagen

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"rpi-search-ranking/internal/ranking"
)

func CreateExamples(filePath string, maxExamples, minDiff int) ([]ranking.Features, []int, error) {
	relevances, qids, features, err := loadDataset(filePath)
	if err != nil {
		return nil, nil, err
	}

	X, Y := createComparisons(relevances, qids, features, maxExamples, minDiff)

	if len(Y) < maxExamples {
		return X, Y, fmt.Errorf("error: Not enough examples in dataset, found %v, expected %v", len(Y), maxExamples)
	}

	shuffleData(X, Y)

	return X, Y, nil
}

// shuffleData Helper function to shuffle the data
func shuffleData(X []ranking.Features, Y []int) {
	for i := len(X) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		X[i], X[j] = X[j], X[i]
		Y[i], Y[j] = Y[j], Y[i]
	}
}

func parseLine(line string) (relevance int, qid int, features ranking.Features, err error) {
	// Split the line into components
	parts := strings.Fields(line)

	// First column is the relevance label
	relevance, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, ranking.Features{}, err
	}

	// Parse the query ID (qid:X)
	qidParts := strings.Split(parts[1], ":")
	if len(qidParts) != 2 || qidParts[0] != "qid" {
		return 0, 0, ranking.Features{}, fmt.Errorf("invalid query ID format: %s", parts[1])
	}
	qid, err = strconv.Atoi(qidParts[1])
	if err != nil {
		return 0, 0, ranking.Features{}, err
	}

	// Create the Features struct
	features = ranking.Features{}

	// Iterate over the feature columns (1:136)
	for _, part := range parts[2:] {
		featureParts := strings.Split(part, ":")
		if len(featureParts) != 2 {
			return 0, 0, ranking.Features{}, fmt.Errorf("invalid feature format: %s", part)
		}
		featureID, err := strconv.Atoi(featureParts[0])
		if err != nil {
			return 0, 0, ranking.Features{}, err
		}
		featureValue, err := strconv.ParseFloat(featureParts[1], 64)
		if err != nil {
			return 0, 0, ranking.Features{}, fmt.Errorf("failed to parse feature value as float: %s", part)
		}

		// Map the feature values to the appropriate fields in the Features struct
		switch featureID {
		case 5:
			features.CoveredQueryTermNumber = int(featureValue)
		case 10:
			features.CoveredQueryTermRatio = featureValue
		case 15:
			features.StreamLength = int(featureValue)
		case 25:
			features.SumTermFrequency = int(featureValue)
		case 30:
			features.MinTermFrequency = int(featureValue)
		case 35:
			features.MaxTermFrequency = int(featureValue)
		case 40:
			features.MeanTermFrequency = featureValue
		case 45:
			features.VarianceTermFrequency = featureValue
		case 50:
			features.SumStreamLengthNormalizedTF = featureValue
		case 55:
			features.MinStreamLengthNormalizedTF = featureValue
		case 60:
			features.MaxStreamLengthNormalizedTF = featureValue
		case 65:
			features.MeanStreamLengthNormalizedTF = featureValue
		case 70:
			features.VarianceStreamLengthNormalizedTF = featureValue
		case 75:
			features.SumTFIDF = featureValue
		case 80:
			features.MinTFIDF = featureValue
		case 85:
			features.MaxTFIDF = featureValue
		case 90:
			features.MeanTFIDF = featureValue
		case 95:
			features.VarianceTFIDF = featureValue
		case 110:
			features.BM25 = featureValue
		case 126:
			features.NumSlashesInURL = int(featureValue)
		case 127:
			features.LengthOfURL = int(featureValue)
		case 128:
			features.InlinkCount = int(featureValue)
		case 129:
			features.OutlinkCount = int(featureValue)
		case 130:
			features.PageRank = featureValue
		}
	}

	return relevance, qid, features, nil
}

func loadDataset(filePath string) ([]int, []int, []ranking.Features, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("warning: failed to close file: %v\n", err)
		}
	}(file)

	var relevances []int
	var qids []int
	var featureVectors []ranking.Features

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		relevance, qid, features, err := parseLine(line)
		if err != nil {
			return nil, nil, nil, err
		}

		relevances = append(relevances, relevance)
		qids = append(qids, qid)
		featureVectors = append(featureVectors, features)
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, nil, err
	}

	return relevances, qids, featureVectors, nil
}

func createComparisons(relevances []int, qids []int, features []ranking.Features, maxExamples, minDiff int) ([]ranking.Features, []int) {
	var pairwiseFeatures []ranking.Features
	var labels []int

	// Group documents by QID
	qidGroups := make(map[int][]int) // QID -> indices
	for i, qid := range qids {
		qidGroups[qid] = append(qidGroups[qid], i)
	}

	// Reservoir to hold pairwise examples
	type example struct {
		feature ranking.Features
		label   int
	}
	reservoir := make([]example, 0, maxExamples)
	exampleCount := 0

	// Generate pairwise comparisons for each QID
	for _, indices := range qidGroups {
		n := len(indices)
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				if i == j || math.Abs(float64(relevances[indices[i]]-relevances[indices[j]])) < float64(minDiff) {
					continue // Skip same document or less than min relevance difference
				}

				// Compute pairwise features
				diff := ranking.Features{
					CoveredQueryTermNumber:           features[indices[i]].CoveredQueryTermNumber - features[indices[j]].CoveredQueryTermNumber,
					CoveredQueryTermRatio:            features[indices[i]].CoveredQueryTermRatio - features[indices[j]].CoveredQueryTermRatio,
					SumTermFrequency:                 features[indices[i]].SumTermFrequency - features[indices[j]].SumTermFrequency,
					MinTermFrequency:                 features[indices[i]].MinTermFrequency - features[indices[j]].MinTermFrequency,
					MaxTermFrequency:                 features[indices[i]].MaxTermFrequency - features[indices[j]].MaxTermFrequency,
					MeanTermFrequency:                features[indices[i]].MeanTermFrequency - features[indices[j]].MeanTermFrequency,
					VarianceTermFrequency:            features[indices[i]].VarianceTermFrequency - features[indices[j]].VarianceTermFrequency,
					StreamLength:                     features[indices[i]].StreamLength - features[indices[j]].StreamLength,
					SumStreamLengthNormalizedTF:      features[indices[i]].SumStreamLengthNormalizedTF - features[indices[j]].SumStreamLengthNormalizedTF,
					MinStreamLengthNormalizedTF:      features[indices[i]].MinStreamLengthNormalizedTF - features[indices[j]].MinStreamLengthNormalizedTF,
					MaxStreamLengthNormalizedTF:      features[indices[i]].MaxStreamLengthNormalizedTF - features[indices[j]].MaxStreamLengthNormalizedTF,
					MeanStreamLengthNormalizedTF:     features[indices[i]].MeanStreamLengthNormalizedTF - features[indices[j]].MeanStreamLengthNormalizedTF,
					VarianceStreamLengthNormalizedTF: features[indices[i]].VarianceStreamLengthNormalizedTF - features[indices[j]].VarianceStreamLengthNormalizedTF,
					SumTFIDF:                         features[indices[i]].SumTFIDF - features[indices[j]].SumTFIDF,
					MinTFIDF:                         features[indices[i]].MinTFIDF - features[indices[j]].MinTFIDF,
					MaxTFIDF:                         features[indices[i]].MaxTFIDF - features[indices[j]].MaxTFIDF,
					MeanTFIDF:                        features[indices[i]].MeanTFIDF - features[indices[j]].MeanTFIDF,
					VarianceTFIDF:                    features[indices[i]].VarianceTFIDF - features[indices[j]].VarianceTFIDF,
					BM25:                             features[indices[i]].BM25 - features[indices[j]].BM25,
					NumSlashesInURL:                  features[indices[i]].NumSlashesInURL - features[indices[j]].NumSlashesInURL,
					LengthOfURL:                      features[indices[i]].LengthOfURL - features[indices[j]].LengthOfURL,
					InlinkCount:                      features[indices[i]].InlinkCount - features[indices[j]].InlinkCount,
					OutlinkCount:                     features[indices[i]].OutlinkCount - features[indices[j]].OutlinkCount,
					PageRank:                         features[indices[i]].PageRank - features[indices[j]].PageRank,
				}

				// Determine label
				label := 1
				if relevances[indices[i]] < relevances[indices[j]] {
					label = -1
				}

				// Add example to the reservoir using reservoir sampling
				exampleCount++
				if len(reservoir) < maxExamples {
					// Fill the reservoir initially
					reservoir = append(reservoir, example{feature: diff, label: label})
				} else {
					// Replace an existing element with decreasing probability
					r := rand.Intn(exampleCount)
					if r < maxExamples {
						reservoir[r] = example{feature: diff, label: label}
					}
				}
			}
		}
	}

	// Separate reservoir into features and labels
	for _, ex := range reservoir {
		pairwiseFeatures = append(pairwiseFeatures, ex.feature)
		labels = append(labels, ex.label)
	}

	return pairwiseFeatures, labels
}

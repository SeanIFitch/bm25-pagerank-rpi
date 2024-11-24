package ranking

import (
	"fmt"
	"math"
	"net/http"
	"strings"
)

// Helper functions for specific calculations
func getIDF(index invertibleIndex, totalDocCount int) map[string]float64 {
	idf := make(map[string]float64)
	for term, postings := range index {
		docFrequency := len(postings)
		idf[term] = math.Log(float64(totalDocCount) / float64(docFrequency+1)) // Smoothed IDF
	}
	return idf
}

func calculateTermFrequencyStats(query Query, termFrequencies map[string]int) (sum, min, max int, mean, variance float64) {
	sum = 0
	min = math.MaxInt
	max = math.MinInt

	// Prepare for statistics calculation
	var tfValues []int
	queryTermCount := float64(len(query.Terms)) // Total terms in query

	for _, term := range query.Terms {
		tf := 0 // Default frequency is 0 if not found in termFrequencies
		if val, found := termFrequencies[term]; found {
			tf = val
		}
		tfValues = append(tfValues, tf) // Always include tf (even if 0)

		sum += tf
		if tf < min {
			min = tf
		}
		if tf > max {
			max = tf
		}
	}

	// Handle edge case: No query terms
	if queryTermCount == 0.0 {
		min, max = 0, 0
		return 0, 0, 0, 0.0, 0.0
	}

	// Compute mean
	mean = float64(sum) / queryTermCount

	// Compute variance
	var varianceSum float64
	for _, tf := range tfValues {
		diff := float64(tf) - mean
		varianceSum += diff * diff
	}
	variance = varianceSum / queryTermCount

	return sum, min, max, mean, variance
}

func calculateNormalizedTFStats(query Query, termFrequencies map[string]int, docLength int) (sum, min, max, mean, variance float64) {
	// Ensure docLength is not zero to avoid division by zero
	if docLength <= 0 {
		return 0, 0, 0, 0, 0
	}

	// Initialize variables
	sum = 0.0
	min = math.MaxFloat64
	max = -math.MaxFloat64

	// Collect normalized term frequencies
	var normalizedTFValues []float64
	queryTermCount := len(query.Terms)

	for _, term := range query.Terms {
		tf := 0 // Default term frequency is 0 if the term is not found
		if val, found := termFrequencies[term]; found {
			tf = val
		}
		normalizedTF := float64(tf) / float64(docLength)
		normalizedTFValues = append(normalizedTFValues, normalizedTF)

		// Update sum, min, and max
		sum += normalizedTF
		if normalizedTF < min {
			min = normalizedTF
		}
		if normalizedTF > max {
			max = normalizedTF
		}
	}

	// Handle edge case: No query terms
	if queryTermCount == 0 {
		return 0, 0, 0, 0, 0
	}

	// Compute mean
	count := float64(queryTermCount)
	mean = sum / count

	// Compute variance
	var varianceSum float64
	for _, normalizedTF := range normalizedTFValues {
		diff := normalizedTF - mean
		varianceSum += diff * diff
	}
	variance = varianceSum / count

	return sum, min, max, mean, variance
}

func calculateBM25(query Query, termFrequencies map[string]int, idf map[string]float64, docLength int, avgDocLength float64) float64 {
	var bm25Score float64

	// Loop over query terms and calculate BM25 contributions
	for _, term := range query.Terms {
		tf, exists := termFrequencies[term] // Term frequency in the document
		if !exists {
			continue // Skip terms with no tf
		}
		idfValue, exists := idf[term] // Precomputed IDF for the term
		if !exists {
			continue // Skip terms with no IDF value
		}

		// BM25 formula components
		numerator := float64(tf) * (k1 + 1)
		denominator := float64(tf) + k1*(1-b+b*(float64(docLength)/avgDocLength))
		bm25Score += idfValue * (numerator / denominator)
	}

	return bm25Score
}

func calculateIDFMetrics(query Query, termFrequencies map[string]int, idf map[string]float64) (sum, min, max, mean, variance float64) {
	// Initialize variables
	sum = 0.0
	min = math.MaxFloat64
	max = -math.MaxFloat64
	var tfidfValues []float64

	// Iterate over query terms
	for _, term := range query.Terms {
		// Default frequency is 0 if not found in termFrequencies
		tf := 0
		if val, found := termFrequencies[term]; found {
			tf = val
		}

		// Proceed if IDF value exists for the term
		if idfValue, exists := idf[term]; exists {
			// Calculate TF-IDF for the term
			tfidf := float64(tf) * idfValue
			tfidfValues = append(tfidfValues, tfidf)

			// Update sum, min, and max
			sum += tfidf
			if tfidf < min {
				min = tfidf
			}
			if tfidf > max {
				max = tfidf
			}
		}
	}

	// Handle edge case: No valid terms found
	if len(tfidfValues) == 0 {
		return 0.0, 0.0, 0.0, 0.0, 0.0
	}

	// Compute mean
	count := float64(len(tfidfValues))
	mean = sum / count

	// Compute variance
	var varianceSum float64
	for _, tfidf := range tfidfValues {
		diff := tfidf - mean
		varianceSum += diff * diff
	}
	variance = varianceSum / count

	return sum, min, max, mean, variance
}

func analyzeURL(url string) (numSlashes, length int) {
	numSlashes = strings.Count(url, "/")
	length = len(url)
	return
}

// Main feature initialization function
func (doc *Document) calculateFeatures(query Query, idf map[string]float64, avgDocLength float64, client *http.Client) error {
	// Query term coverage metrics
	coveredTerms := 0
	for _, term := range query.Terms {
		if _, found := doc.TermFrequencies[term]; found {
			coveredTerms++
		}
	}
	doc.Features.CoveredQueryTermNumber = coveredTerms
	doc.Features.CoveredQueryTermRatio = float64(coveredTerms) / float64(len(query.Terms))

	// Term frequency statistics
	sumTF, minTF, maxTF, meanTF, varianceTF := calculateTermFrequencyStats(query, doc.TermFrequencies)
	doc.Features.SumTermFrequency = sumTF
	doc.Features.MinTermFrequency = minTF
	doc.Features.MaxTermFrequency = maxTF
	doc.Features.MeanTermFrequency = meanTF
	doc.Features.VarianceTermFrequency = varianceTF

	// Normalized TF statistics
	sumNormTF, minNormTF, maxNormTF, meanNormTF, varNormTF := calculateNormalizedTFStats(query, doc.TermFrequencies, doc.Metadata.DocLength)
	doc.Features.StreamLength = doc.Metadata.DocLength
	doc.Features.SumStreamLengthNormalizedTF = sumNormTF
	doc.Features.MinStreamLengthNormalizedTF = minNormTF
	doc.Features.MaxStreamLengthNormalizedTF = maxNormTF
	doc.Features.MeanStreamLengthNormalizedTF = meanNormTF
	doc.Features.VarianceStreamLengthNormalizedTF = varNormTF

	// IDF-based metrics
	sumTFIDF, minTFIDF, maxTFIDF, meanTFIDF, varTFIDF := calculateIDFMetrics(query, doc.TermFrequencies, idf)
	doc.Features.SumTFIDF = sumTFIDF
	doc.Features.MinTFIDF = minTFIDF
	doc.Features.MaxTFIDF = maxTFIDF
	doc.Features.MeanTFIDF = meanTFIDF
	doc.Features.VarianceTFIDF = varTFIDF

	// BM25 score
	doc.Features.BM25 = calculateBM25(query, doc.TermFrequencies, idf, doc.Metadata.DocLength, avgDocLength)

	// URL characteristics
	numSlashes, urlLength := analyzeURL(doc.Metadata.URL)
	doc.Features.NumSlashesInURL = numSlashes
	doc.Features.LengthOfURL = urlLength

	// Link analysis
	// get document metadata
	pageRank, err := fetchPageRank(client, doc.Metadata.URL)
	if err != nil {
		return err
	}

	doc.Features.InlinkCount = pageRank.InLinkCount
	doc.Features.OutlinkCount = pageRank.OutLinkCount
	doc.Features.PageRank = pageRank.PageRank

	return nil
}

// Batch initialization for a list of documents
func (docs *Documents) initializeFeatures(query Query, docStatistics totalDocStatistics, index invertibleIndex, client *http.Client) error {
	idf := getIDF(index, docStatistics.DocCount)

	// Fetch metadata and calculate features
	var errList []error
	for i := range *docs {
		doc := &(*docs)[i]
		metadata, err := fetchDocumentMetadata(client, doc.DocID)
		if err != nil {
			errList = append(errList, err)
			continue
		}
		doc.Metadata = metadata

		if err := doc.calculateFeatures(query, idf, docStatistics.AvgDocLength, client); err != nil {
			errList = append(errList, err)
		}
	}

	if len(errList) > 0 {
		return fmt.Errorf("encountered errors during initialization: %v", errList)
	}

	return nil
}

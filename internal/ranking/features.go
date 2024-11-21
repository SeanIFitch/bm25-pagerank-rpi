package ranking

import (
	"math"
)

// GenerateFeatures generates features for a document
func (doc *Document) GenerateFeatures(query Query, invertibleIndex InvertibleIndex, docStatistics TotalDocStatistics) {
	// Generate BM25 score and set it in the Features struct
	doc.Features.BM25 = GetBM25(query, invertibleIndex, *doc, docStatistics.AvgDocLength)
}

// GetBM25 computes the BM25 score for a single document for the given query
func GetBM25(query Query, invertibleIndex InvertibleIndex, document Document, avgDocumentLength float64) float64 {
	// Initialize the BM25 score
	bm25Score := 0.0

	// Tokenize the query and calculate BM25 for each term
	// You can split the query text into terms and compute BM25 for each one
	for _, term := range query.Terms() { // Assume `Terms` is a function that splits the query into terms
		// Look up the term in the inverted index
		if documentIndexList, exists := invertibleIndex[term]; exists {
			// Find the document frequency and term frequency for the term
			for _, docIndex := range documentIndexList {
				if docIndex.DocID == document.DocID {
					termFrequency := docIndex.Frequency
					documentLength := document.Metadata.DocLength

					// Compute BM25 for this term
					idf := math.Log((float64(len(invertibleIndex))-float64(len(documentIndexList))+0.5)/(float64(len(documentIndexList))+0.5) + 1.0)
					numerator := float64(termFrequency) * (k1 + 1)
					denominator := float64(termFrequency) + k1*(1-b+b*float64(documentLength)/avgDocumentLength)
					bm25Score += idf * numerator / denominator
				}
			}
		}
	}

	return bm25Score
}

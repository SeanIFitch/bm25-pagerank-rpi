package ranking

import (
	"math"
)

// GetBM25 calculates and updates the BM25 score for each document in the provided documents
func GetBM25(query Query, invertibleIndex InvertibleIndex, documents Documents, docStatistics TotalDocStatistics) {
	// Iterate over each document to calculate BM25
	for i := range documents {
		doc := &documents[i] // Reference to modify the document in place

		// Initialize BM25 score for the document
		bm25Score := 0.0

		// Iterate over each term in the query
		for _, term := range query.Text {
			// Calculate IDF for the term
			df := len(invertibleIndex[string(term)]) // Document frequency
			idf := math.Log((float64(docStatistics.DocCount)-float64(df)+0.5)/(float64(df)+0.5) + 1.0)

			// Find the term in the document's index
			termIndex := findTermIndex(doc.DocID, string(term), invertibleIndex)

			// If the term is present in the document
			if termIndex != -1 {
				// Term frequency for this term in the document
				tf := invertibleIndex[string(term)][termIndex].Frequency

				// Length of the document
				docLength := doc.Metadata.DocLength

				// Calculate BM25 score for this term
				bm25TermScore := idf * float64(tf) * (k1 + 1) / (float64(tf) + k1*(1-b+b*float64(docLength)/docStatistics.AvgDocLength))
				bm25Score += bm25TermScore
			}
		}

		// Update BM25 score in the document's features
		doc.Features.BM25 = bm25Score
	}
}

// Helper function to find the index of the term in a document's list of term occurrences
func findTermIndex(docID string, term string, invertibleIndex InvertibleIndex) int {
	// Iterate through the list of documents for the given term in the invertible index
	for i, docIndex := range invertibleIndex[term] {
		if docIndex.DocID == docID {
			return i // Return the index of the document that contains the term
		}
	}
	return -1 // Return -1 if the term is not found in the document
}

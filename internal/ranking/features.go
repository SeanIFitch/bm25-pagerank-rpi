package ranking

import (
	"math"
	"strings"
)

// GetBM25 calculates the BM25 score for each document based on the query
func GetBM25(query Query, invertibleIndex InvertibleIndex, documentLengths map[string]int, avgDocumentLength float64) map[string]float64 {
	// Initialize a map to store BM25 scores for each document
	scores := make(map[string]float64)
	totalDocs := len(documentLengths)

	// Split the query text into terms (words)
	terms := strings.Fields(query.Text)

	// Iterate over each term in the query
	for _, term := range terms {
		// Calculate the document frequency (df) for the term
		df := len(invertibleIndex[term]) // Number of documents containing the term
		// Calculate the Inverse Document Frequency (IDF) for the term
		idf := math.Log((float64(totalDocs)-float64(df)+0.5)/(float64(df)+0.5) + 1.0)

		// For each document containing the term, calculate the BM25 score
		for _, doc := range invertibleIndex[term] {
			docID := doc.DocID
			// Get the term frequency in the document
			tf := doc.Frequency
			// Get the length of the document
			docLength := documentLengths[docID]
			// Calculate the BM25 score for this document and term
			score := idf * float64(tf) * (k1 + 1) / (float64(tf) + k1*(1-b+b*float64(docLength)/avgDocumentLength))
			// Add the score to the document's total score
			scores[docID] += score
		}
	}

	return scores
}

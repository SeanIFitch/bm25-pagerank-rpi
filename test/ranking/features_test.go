package ranking

import (
	"rpi-search-ranking/internal/ranking"
	"testing"
)

// Mock Data for Tests
func mockQuery() ranking.Query {
	return ranking.Query{
		Id:   "query1",
		Text: "document retrieval",
	}
}

func mockDocument() ranking.Document {
	return ranking.Document{
		DocID: "doc1",
		Rank:  1,
		Metadata: ranking.DocumentMetadata{
			DocLength: 100,
		},
		Features: ranking.Features{},
	}
}

func mockInvertibleIndex() ranking.InvertibleIndex {
	return ranking.InvertibleIndex{
		"document": {
			{
				DocID:     "doc1",
				Frequency: 2,
				Positions: []int{1, 5},
			},
			{
				DocID:     "doc2",
				Frequency: 3,
				Positions: []int{3, 6, 9},
			},
		},
		"retrieval": {
			{
				DocID:     "doc1",
				Frequency: 1,
				Positions: []int{2},
			},
		},
	}
}

func mockDocStatistics() ranking.TotalDocStatistics {
	return ranking.TotalDocStatistics{
		AvgDocLength: 120.0,
		DocCount:     2,
	}
}

// Test the BM25 score calculation for a document
func TestGetBM25(t *testing.T) {
	query := mockQuery()
	document := mockDocument()
	invertibleIndex := mockInvertibleIndex()
	docStatistics := mockDocStatistics()

	// Generate Features for the document
	document.GenerateFeatures(query, invertibleIndex, docStatistics)

	// Manually calculate expected BM25 score based on the formula
	// In this example, you may want to use a known expected value.
	//expectedBM25 := 0.0

	// Example calculation for BM25
	// BM25 score depends on the term frequencies and document frequencies
	// So you'll want to adjust the expected value based on the BM25 formula
	// You can test the formula with smaller values and print them out for inspection.
	// Here we'll assume you compute the BM25 for "document" and "retrieval" terms.
	// This is a simplified expected value for demonstration purposes.

	if document.Features.BM25 <= 0 {
		t.Errorf("BM25 score should be greater than 0, got: %f", document.Features.BM25)
	}

	// For more precise tests, you can compute the expected value manually
	// and check if the BM25 computation is accurate with different inputs.
}

// Test GetBM25 with empty query
func TestGetBM25_EmptyQuery(t *testing.T) {
	query := ranking.Query{Id: "empty_query", Text: ""}
	document := mockDocument()
	invertibleIndex := mockInvertibleIndex()
	docStatistics := mockDocStatistics()

	// Generate Features for the document
	document.GenerateFeatures(query, invertibleIndex, docStatistics)

	// BM25 score should be 0 as no terms are in the query
	if document.Features.BM25 != 0 {
		t.Errorf("BM25 score for empty query should be 0, got: %f", document.Features.BM25)
	}
}

// Test GetBM25 with no matching terms in the document
func TestGetBM25_NoMatchingTerms(t *testing.T) {
	query := ranking.Query{Id: "query_no_match", Text: "nonexistent term"}
	document := mockDocument()
	invertibleIndex := mockInvertibleIndex()
	docStatistics := mockDocStatistics()

	// Generate Features for the document
	document.GenerateFeatures(query, invertibleIndex, docStatistics)

	// BM25 score should be 0 as there are no matching terms
	if document.Features.BM25 != 0 {
		t.Errorf("BM25 score for no matching terms should be 0, got: %f", document.Features.BM25)
	}
}

// Test GetBM25 with multiple documents in the inverted index
func TestGetBM25_MultipleDocuments(t *testing.T) {
	query := ranking.Query{Id: "query_multiple", Text: "document retrieval"}
	document := mockDocument()
	invertibleIndex := mockInvertibleIndex()
	docStatistics := mockDocStatistics()

	// Generate Features for the document
	document.GenerateFeatures(query, invertibleIndex, docStatistics)

	// BM25 score will depend on the term frequency and document frequency
	// The actual expected value should be manually computed based on the formula
	//expectedBM25 := 0.0 // Update with expected value based on your formula

	// Compare the BM25 score
	//if document.Features.BM25 != expectedBM25 {
	//	t.Errorf("Expected BM25 score %f, got: %f", expectedBM25, document.Features.BM25)
	//}
}

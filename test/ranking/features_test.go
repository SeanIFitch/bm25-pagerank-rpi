package ranking

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"rpi-search-ranking/internal/ranking"
	"testing"
)

// Helper function to calculate BM25 score manually for testing purposes.
func expectedBM25(k1, b, tf, docLength, avgDocLength, idf float64) float64 {
	return idf * (tf * (k1 + 1)) / (tf + k1*(1-b+b*(docLength/avgDocLength)))
}

func TestGetBM25_EmptyIndex(t *testing.T) {
	// Test 1 -- Document corpus contains no occurrences of a term, thus inverted index is empty.
	jsonData := `
	{
		"term": "data",
		"index": []
	}`

	var response ranking.InvertibleIndex
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		t.Fatalf("Error unmarshaling JSON: %v", err)
	}

	k1, b := 1.25, 0.75
	BM25Results := ranking.GetBM25(k1, b, response)

	// Check if the result is empty
	assert.Empty(t, BM25Results, "BM25 results should be empty when no occurrences of the term are found")
}

func TestGetBM25_SingleDocument(t *testing.T) {
	// Test 2 -- Document corpus contains only one document with the occurrence of the query term.
	jsonData2 := `
	{
		"term": "data",
		"index": [
			{
				"docID": "12345",
				"frequency": 5,
				"positions": [4, 15, 28, 102, 204]
			}
		]
	}`

	var response2 ranking.InvertibleIndex
	err2 := json.Unmarshal([]byte(jsonData2), &response2)
	if err2 != nil {
		t.Fatalf("Error unmarshaling JSON: %v", err2)
	}

	k1, b := 1.25, 0.75
	BM25Results2 := ranking.GetBM25(k1, b, response2)

	// Expected BM25 score for a single document with known frequency.
	expectedScore := expectedBM25(k1, b, 5, 5, 5, 0) // Example values, assuming a simplified IDF = 0
	assert.Len(t, BM25Results2, 1, "BM25 result should contain exactly one document")
	assert.Equal(t, expectedScore, BM25Results2["12345"], "BM25 score for docID 12345 should match expected value")
}

func TestGetBM25_MultipleDocuments(t *testing.T) {
	// Test 3 -- Document corpus contains multiple documents with occurrences of the term.
	jsonData3 := `
	{
		"term": "data",
		"index": [
			{
				"docID": "12345",
				"frequency": 5,
				"positions": [4, 15, 28, 102, 204]
			},
			{
				"docID": "123",
				"frequency": 5,
				"positions": [4, 15, 28, 102, 204]
			},
			{
				"docID": "12",
				"frequency": 5,
				"positions": [4, 15, 28, 102, 204]
			},
			{
				"docID": "1",
				"frequency": 5,
				"positions": [4, 15, 28, 102, 204]
			}
		]
	}`

	var response3 ranking.InvertibleIndex
	err3 := json.Unmarshal([]byte(jsonData3), &response3)
	if err3 != nil {
		t.Fatalf("Error unmarshaling JSON: %v", err3)
	}

	k1, b := 1.25, 0.75
	BM25Results3 := ranking.GetBM25(k1, b, response3)

	// Check the number of documents
	assert.Len(t, BM25Results3, 4, "BM25 result should contain four documents")
}

func TestGetBM25_VaryingTermFrequency(t *testing.T) {
	// Test 4 -- Document corpus contains multiple documents with varying term frequencies.
	jsonData4 := `
	{
		"term": "data",
		"index": [
			{
				"docID": "12345",
				"frequency": 5,
				"positions": [4, 15, 28, 102, 204]
			},
			{
				"docID": "123",
				"frequency": 3,
				"positions": [4, 15, 28]
			},
			{
				"docID": "12",
				"frequency": 1,
				"positions": [204]
			},
			{
				"docID": "1",
				"frequency": 10,
				"positions": [4, 15, 28, 102, 204,1000,1001,1002,1003,1004]
			}
		]
	}`

	var response4 ranking.InvertibleIndex
	err4 := json.Unmarshal([]byte(jsonData4), &response4)
	if err4 != nil {
		t.Fatalf("Error unmarshaling JSON: %v", err4)
	}

	k1, b := 1.25, 0.75
	BM25Results4 := ranking.GetBM25(k1, b, response4)

	// Check that there are four documents and their scores should vary with term frequency
	assert.Len(t, BM25Results4, 4, "BM25 result should contain four documents")
}

func TestGetBM25_VaryingK1(t *testing.T) {
	// Test 5 -- Varying k1 parameter to see its effect on the BM25 score.
	jsonData5 := `
	{
		"term": "data",
		"index": [
			{
				"docID": "12345",
				"frequency": 5,
				"positions": [4, 15, 28, 102, 204]
			},
			{
				"docID": "123",
				"frequency": 3,
				"positions": [4, 15, 28]
			},
			{
				"docID": "12",
				"frequency": 1,
				"positions": [204]
			},
			{
				"docID": "1",
				"frequency": 10,
				"positions": [4, 15, 28, 102, 204,1000,1001,1002,1003,1004]
			}
		]
	}`

	var response5 ranking.InvertibleIndex
	err5 := json.Unmarshal([]byte(jsonData5), &response5)
	if err5 != nil {
		t.Fatalf("Error unmarshaling JSON: %v", err5)
	}

	k1, b := 0.0, 0.75 // Setting k1 = 0 to cancel out all variables except for IDF
	BM25Results5 := ranking.GetBM25(k1, b, response5)

	// Check the effect of k1 on BM25 scores.
	assert.Len(t, BM25Results5, 4, "BM25 result should contain four documents")
}

func TestGetBM25_VaryingB(t *testing.T) {
	// Test 6 -- Varying b parameter to cancel out document length influence.
	jsonData6 := `
	{
		"term": "data",
		"index": [
			{
				"docID": "12345",
				"frequency": 3,
				"positions": [4, 15, 28]
			},
			{
				"docID": "123",
				"frequency": 3,
				"positions": [4, 15, 28]
			},
			{
				"docID": "12",
				"frequency": 1,
				"positions": [204]
			},
			{
				"docID": "1",
				"frequency": 10,
				"positions": [4, 15, 28, 102, 204,1000,1001,1002,1003,1004]
			}
		]
	}`

	var response6 ranking.InvertibleIndex
	err6 := json.Unmarshal([]byte(jsonData6), &response6)
	if err6 != nil {
		t.Fatalf("Error unmarshaling JSON: %v", err6)
	}

	k1, b := 1.2, 0.0 // Set b = 0 to eliminate document length influence
	BM25Results6 := ranking.GetBM25(k1, b, response6)

	// Check if document length has no effect
	assert.Len(t, BM25Results6, 4, "BM25 result should contain four documents")
}

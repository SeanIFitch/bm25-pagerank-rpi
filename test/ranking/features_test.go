package ranking_test

import (
	"math"
	"rpi-search-ranking/internal/ranking"
	"testing"
)

func TestGetIDF(t *testing.T) {
	index := ranking.InvertibleIndex{
		"term1": {{DocID: "doc1", Frequency: 2}, {DocID: "doc2", Frequency: 1}},
		"term2": {{DocID: "doc3", Frequency: 1}},
	}
	totalDocCount := 5

	expectedIDF := map[string]float64{
		"term1": math.Log(float64(totalDocCount) / float64(2+1)), // Smoothed IDF
		"term2": math.Log(float64(totalDocCount) / float64(1+1)), // Smoothed IDF
	}

	idf := ranking.GetIDF(index, totalDocCount)

	for term, expected := range expectedIDF {
		if math.Abs(idf[term]-expected) > 1e-6 {
			t.Errorf("IDF for term %s was incorrect, got: %f, want: %f", term, idf[term], expected)
		}
	}

	// Test for terms not in index
	if _, exists := idf["nonexistent"]; exists {
		t.Errorf("IDF should not include nonexistent terms")
	}
}

func TestGetBM25(t *testing.T) {
	// Mock inputs
	query := ranking.Query{Terms: []string{"term1", "term2"}}
	docStatistics := ranking.TotalDocStatistics{
		AvgDocLength: 100.0,
		DocCount:     5,
	}
	idf := map[string]float64{
		"term1": 1.2,
		"term2": 0.8,
	}

	// Mock document
	doc := ranking.Document{
		DocID: "doc1",
		Metadata: ranking.DocumentMetadata{
			DocLength: 120,
		},
		TermFrequencies: map[string]int{
			"term1": 3,
			"term2": 1,
		},
	}

	// Set BM25 parameters
	k1 := 1.5
	b := 0.75

	// Expected BM25 calculation
	expectedBM25 := 0.0
	{
		// term1 calculation
		tf := 3
		idfValue := 1.2
		numerator := float64(tf) * (k1 + 1)
		denominator := float64(tf) + k1*(1-b+b*(120.0/100.0))
		expectedBM25 += idfValue * (numerator / denominator)

		// term2 calculation
		tf = 1
		idfValue = 0.8
		numerator = float64(tf) * (k1 + 1)
		denominator = float64(tf) + k1*(1-b+b*(120.0/100.0))
		expectedBM25 += idfValue * (numerator / denominator)
	}

	// Call the GetBM25 function
	actualBM25, err := doc.GetBM25(query, docStatistics, idf)
	if err != nil {
		t.Fatalf("GetBM25 returned an error: %v", err)
	}

	// Verify the result
	if math.Abs(actualBM25-expectedBM25) > 1e-6 {
		t.Errorf("BM25 score was incorrect, got: %f, want: %f", actualBM25, expectedBM25)
	}
}

func TestGetBM25_MissingTerm(t *testing.T) {
	// Mock inputs
	query := ranking.Query{Terms: []string{"term1", "term3"}}
	docStatistics := ranking.TotalDocStatistics{
		AvgDocLength: 100.0,
		DocCount:     5,
	}
	idf := map[string]float64{
		"term1": 1.2,
	}

	// Mock document
	doc := ranking.Document{
		DocID: "doc1",
		Metadata: ranking.DocumentMetadata{
			DocLength: 120,
		},
		TermFrequencies: map[string]int{
			"term1": 3,
		},
	}

	// Call the GetBM25 function
	actualBM25, err := doc.GetBM25(query, docStatistics, idf)
	if err != nil {
		t.Fatalf("GetBM25 returned an error: %v", err)
	}

	// Verify the result (term3 is missing in both document and idf)
	expectedBM25 := 0.0
	{
		// term1 calculation
		tf := 3
		idfValue := 1.2
		k1 := 1.5
		b := 0.75
		numerator := float64(tf) * (k1 + 1)
		denominator := float64(tf) + k1*(1-b+b*(120.0/100.0))
		expectedBM25 += idfValue * (numerator / denominator)
	}

	if math.Abs(actualBM25-expectedBM25) > 1e-6 {
		t.Errorf("BM25 score with missing term was incorrect, got: %f, want: %f", actualBM25, expectedBM25)
	}
}

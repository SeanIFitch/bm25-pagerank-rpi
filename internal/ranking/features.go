package ranking

import "math"

// computeFeatures generates features for a document
func (doc *Document) computeFeatures(query Query, docStatistics totalDocStatistics, index invertibleIndex) error {
	idf := getIDF(index, docStatistics.DocCount)

	// Generate BM25 score
	bm25, err := doc.getBM25(query, docStatistics, idf)
	if err != nil {
		return err
	}

	doc.Features.BM25 = bm25

	return nil
}

func getIDF(index invertibleIndex, totalDocCount int) map[string]float64 {
	idf := make(map[string]float64)
	for term, postings := range index {
		docFrequency := len(postings)
		idf[term] = math.Log(float64(totalDocCount) / float64(docFrequency+1)) // Smoothed IDF
	}
	return idf
}

func (doc *Document) getBM25(query Query, docStatistics totalDocStatistics, idf map[string]float64) (float64, error) {
	var bm25Score float64

	// Loop over query terms and calculate BM25 contributions
	for _, term := range query.Terms {
		tf := doc.TermFrequencies[term] // Term frequency in the document
		idfValue, exists := idf[term]   // Precomputed IDF for the term
		if !exists {
			continue // Skip terms with no IDF value
		}

		// BM25 formula components
		numerator := float64(tf) * (k1 + 1)
		denominator := float64(tf) + k1*(1-b+b*(float64(doc.Metadata.DocLength)/docStatistics.AvgDocLength))
		bm25Score += idfValue * (numerator / denominator)
	}

	return bm25Score, nil
}

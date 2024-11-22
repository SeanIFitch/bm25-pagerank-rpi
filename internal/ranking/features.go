package ranking

// ComputeFeatures generates features for a document
func (doc *Document) ComputeFeatures(query Query, docStatistics TotalDocStatistics) error {
	// Generate BM25 score
	err := doc.ComputeBM25(query, docStatistics)
	if err != nil {
		return err
	}

	return nil
}

func (doc *Document) ComputeBM25(query Query, docStatistics TotalDocStatistics) error {
	return nil
}

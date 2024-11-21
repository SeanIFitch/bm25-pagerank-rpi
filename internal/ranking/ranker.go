package ranking

import (
	"log"
)

// RankDocuments ranks the documents based on the query text
func RankDocuments(query Query) ([]Document, error) {
	index, err := getInvertibleIndex(query)
	if err != nil {
		return nil, err
	}

	_, err = GetDocuments(index)
	if err != nil {
		return nil, err
	}

	// get metadata
	// declare feature struct
	// add each feature to the feature struct

	// Example documents
	docScores := []Document{
		{
			DocID:    "12345",
			Rank:     5,
			Metadata: DocumentMetadata{},
		},
		{
			DocID:    "67890",
			Rank:     3,
			Metadata: DocumentMetadata{},
		},
	}

	log.Printf("Ranked documents for query: %s", query)

	// Return the ranked documents
	return docScores, nil
}

// GetDocuments returns a slice of all documents in the InvertibleIndex
func GetDocuments(index InvertibleIndex) ([]Document, error) {
	// Map to store aggregated term frequencies for each document
	documentsMap := make(map[string]Document)

	// Iterate over each term in the invertible index
	for term, docIndices := range index {
		for _, docIndex := range docIndices {
			// Check if the document already exists in the map
			doc, exists := documentsMap[docIndex.DocID]
			if !exists {
				// If the document doesn't exist, initialize it
				doc = Document{
					DocID:           docIndex.DocID,
					TermFrequencies: make(map[string]int),
					Features:        Features{}, // Placeholder; features can be calculated later
				}
			}

			// Add the term frequency to the document
			doc.TermFrequencies[term] += docIndex.Frequency
			documentsMap[docIndex.DocID] = doc
		}
	}

	// Convert the map to a slice
	documents := make([]Document, 0, len(documentsMap))
	for _, doc := range documentsMap {
		documents = append(documents, doc)
	}

	return documents, nil
}

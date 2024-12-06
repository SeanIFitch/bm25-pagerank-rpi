package ranking

import (
	"log"
	"net/http"
	"slices"
)

// RankDocuments ranks the documents based on the query text
func RankDocuments(query Query, client *http.Client) ([]Document, error) {
	query.tokenize()

	// Get invertible index for the query
	index, err := getInvertibleIndex(client, query)
	if err != nil {
		return nil, err
	}

	// Get slice of all relevant documents
	documents, err := getDocuments(index)
	if err != nil {
		return nil, err
	}

	// Return early if there are no documents
	if len(documents) == 0 {
		return nil, nil
	}

	// Count and avg length of all documents
	docStatistics, err := fetchTotalDocStatistics(client)
	if err != nil {
		return nil, err
	}

	// Add document metadata and features
	err = documents.initializeFeatures(query, docStatistics, index, client)

	// Sort by BM25
	slices.SortFunc(documents, func(a, b Document) int {
		if a.Features.BM25 > b.Features.BM25 {
			return -1
		} else if a.Features.BM25 < b.Features.BM25 {
			return 1
		}
		return 0
	})

	// Only consider top maxDocuments documents
	documents = documents[:min(maxDocuments, len(documents))]

	// TODO: sort by pairwise classification on all features

	// Save data for training
	filename := generateUniqueFilename("../../data/raw/examples")
	err = saveData(filename, documents)
	if err != nil {
		log.Printf("warning: failed to write documents to file: %v\n", err)
	}

	// rank
	for i := range documents {
		documents[i].Rank = i + 1
	}

	log.Printf("Ranked documents for query: %s", query)

	// Return the ranked documents
	return documents, nil
}

// getDocuments returns a slice of all documents in the invertibleIndex
func getDocuments(index invertibleIndex) (Documents, error) {
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
				}
			}

			// Add the term frequency to the document
			doc.TermFrequencies[term] += docIndex.Frequency
			documentsMap[docIndex.DocID] = doc
		}
	}

	// Convert the map to a slice
	documents := make(Documents, 0, len(documentsMap))
	for _, doc := range documentsMap {
		documents = append(documents, doc)
	}

	return documents, nil
}

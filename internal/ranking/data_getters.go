package ranking

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// getInvertibleIndex fetches the unique inverted index for all terms in the given query.
func getInvertibleIndex(query Query) (InvertibleIndex, error) {
	// Initialize the index and a map to track unique terms
	index := InvertibleIndex{}
	uniqueTerms := make(map[string]struct{})

	// Iterate over each term in the query
	for _, term := range query.Terms {
		// If the term has not been processed yet, process it
		if _, exists := uniqueTerms[term]; !exists {
			// Mark the term as processed
			uniqueTerms[term] = struct{}{}

			// Fetch the inverted index for this term
			invertedIndex, err := fetchInvertibleIndexForTerm(term)
			if err != nil {
				return nil, err
			}
			// Store the inverted index for this term
			index[term] = invertedIndex
		}
	}

	return index, nil
}

// fetchInvertibleIndex retrieves the inverted index for a given term from the Indexing API
func fetchInvertibleIndexForTerm(term string) ([]DocumentIndex, error) {
	// Construct the API URL using the term as a query parameter
	apiURL := fmt.Sprintf("http://your-api-url.com/get-invertible-index?term=%s", term)

	// Create a new HTTP client with a timeout of 10 seconds
	client := &http.Client{
		Timeout: httpTimeout,
	}

	// Make the HTTP GET request to fetch the inverted index
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Ensure the response status code is OK (200)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch invertible index: %v", resp.Status)
	}

	// Decode the JSON response from the API into a struct
	var result struct {
		Term  string          `json:"term"`
		Index []DocumentIndex `json:"index"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// Return the parsed list of document indices for the term
	return result.Index, nil
}

// fetchDocumentMetadata retrieves the metadata for a given document ID from the Indexing API
func fetchDocumentMetadata(docID string) (DocumentMetadata, error) {
	// Construct the API URL using the document ID as a query parameter
	apiURL := fmt.Sprintf("http://your-api-url.com/get-document-metadata?docID=%s", docID)

	// Create a new HTTP client with a timeout of 10 seconds
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Make the HTTP GET request to fetch the document metadata
	resp, err := client.Get(apiURL)
	if err != nil {
		return DocumentMetadata{}, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Ensure the response status code is OK (200)
	if resp.StatusCode != http.StatusOK {
		return DocumentMetadata{}, fmt.Errorf("failed to fetch document metadata: %v", resp.Status)
	}

	// Decode the JSON response from the API into a struct
	var result struct {
		DocID    string           `json:"docID"`
		Metadata DocumentMetadata `json:"metadata"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return DocumentMetadata{}, fmt.Errorf("failed to decode response: %v", err)
	}

	// Return the parsed metadata for the document
	return result.Metadata, nil
}

package ranking

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

const InvertibleIndexEndpoint = "http://lspt-index-ranking.cs.rpi.edu:8080/get-invertible-index?term="
const MetadataEndpoint = "http://lspt-index-ranking.cs.rpi.edu:8080/get-document-metadata?docID="
const StatisticsEndpoint = "http://lspt-index-ranking.cs.rpi.edu:8080/get-total-doc-statistics"
const PagerankEndpoint = "http://lspt-link-analysis.cs.rpi.edu:1234/ranking/"

// getInvertibleIndex fetches the unique inverted index for all terms in the given query.
func getInvertibleIndex(client *http.Client, query Query) (invertibleIndex, error) {
	// Initialize the index and a map to track unique terms
	index := invertibleIndex{}
	uniqueTerms := make(map[string]struct{})

	// Iterate over each term in the query
	for _, term := range query.Terms {
		// If the term has not been processed yet, process it
		if _, exists := uniqueTerms[term]; !exists {
			// Mark the term as processed
			uniqueTerms[term] = struct{}{}

			// Fetch the inverted index for this term
			invertedIndex, err := fetchInvertibleIndexForTerm(client, term)
			if err != nil {
				return nil, err
			}
			// Store the inverted index for this term
			index[term] = invertedIndex
		}
	}

	return index, nil
}

// fetchInvertibleIndexForTerm retrieves the inverted index for a given term from the Indexing API
func fetchInvertibleIndexForTerm(client *http.Client, term string) ([]documentIndex, error) {
	// Construct the API URL using the term as a query parameter
	apiURL := InvertibleIndexEndpoint + term

	// Make the HTTP GET request to fetch the inverted index
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("warning: failed to close response body: %v\n", err)
		}
	}(resp.Body)

	// Ensure the response status code is OK (200)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch invertible index: %v", resp.Status)
	}

	// Decode the JSON response from the API into a struct
	var result struct {
		Term  string          `json:"term"`
		Index []documentIndex `json:"index"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// Return the parsed list of document indices for the term
	return result.Index, nil
}

// fetchDocumentMetadata retrieves the metadata for a given document ID from the Indexing API
func fetchDocumentMetadata(client *http.Client, docID string) (DocumentMetadata, error) {
	// Construct the API URL using the document ID as a query parameter
	apiURL := MetadataEndpoint + docID

	// Make the HTTP GET request to fetch the document metadata
	resp, err := client.Get(apiURL)
	if err != nil {
		return DocumentMetadata{}, fmt.Errorf("failed to make request: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("warning: failed to close response body: %v\n", err)
		}
	}(resp.Body)

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

// fetchTotalDocStatistics retrieves the total document statistics from the Indexing API
func fetchTotalDocStatistics(client *http.Client) (totalDocStatistics, error) {
	// Construct the API URL
	apiURL := StatisticsEndpoint

	// Make the HTTP GET request to fetch the total document statistics
	resp, err := client.Get(apiURL)
	if err != nil {
		return totalDocStatistics{}, fmt.Errorf("failed to make request: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("warning: failed to close response body: %v\n", err)
		}
	}(resp.Body)

	// Ensure the response status code is OK (200)
	if resp.StatusCode != http.StatusOK {
		return totalDocStatistics{}, fmt.Errorf("failed to fetch total document statistics: %v", resp.Status)
	}

	// Decode the JSON response from the API into a struct
	var stats totalDocStatistics
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return totalDocStatistics{}, fmt.Errorf("failed to decode response: %v", err)
	}

	// Return the parsed statistics
	return stats, nil
}

// fetchPageRank retrieves the PageRank score and related link information for a given URL
func fetchPageRank(client *http.Client, url string) (PageRankInfo, error) {
	// Construct the API URL using the document URL as a query parameter
	apiURL := PagerankEndpoint + url

	// Make the HTTP GET request to fetch the PageRank information
	resp, err := client.Get(apiURL)
	if err != nil {
		return PageRankInfo{}, fmt.Errorf("failed to make request: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("warning: failed to close response body: %v\n", err)
		}
	}(resp.Body)

	// Ensure the response status code is OK (200)
	if resp.StatusCode != http.StatusOK {
		// Read the response body to include raw JSON in the error
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return PageRankInfo{}, fmt.Errorf("failed to fetch PageRank info: %v, and failed to read response body: %v", resp.Status, readErr)
		}
		return PageRankInfo{}, fmt.Errorf("failed to fetch PageRank info: %v, response body: %s", resp.Status, string(bodyBytes))
	}

	// Decode the JSON response from the API into a struct
	var result PageRankInfo
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return PageRankInfo{}, fmt.Errorf("failed to decode response: %v", err)
	}

	// Return the parsed PageRank information
	return result, nil
}

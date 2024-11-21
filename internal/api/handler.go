package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

// Query defines the struct to parse the incoming query
type Query struct {
	QueryText string `json:"queryText"`
}

// DocumentScore represents a document's score and metadata
type DocumentScore struct {
	DocID    string                 `json:"docID"`
	Rank     int                    `json:"rank"`
	Metadata map[string]interface{} `json:"metadata"`
}

// GetDocumentScores returns the scores and metadata for relevant documents based on the query
func GetDocumentScores(query Query) ([]DocumentScore, error) {
	// This is where you would implement the logic to process the query and rank documents
	// For this example, we are returning mock data.

	if query.QueryText == "" {
		return nil, errors.New("query text cannot be empty")
	}

	// Example mock data for document scores and metadata
	docScores := []DocumentScore{
		{
			DocID: "12345",
			Rank:  5,
			Metadata: map[string]interface{}{
				"title":  "Document 1",
				"author": "Author 1",
			},
		},
		{
			DocID: "67890",
			Rank:  3,
			Metadata: map[string]interface{}{
				"title":  "Document 2",
				"author": "Author 2",
			},
		},
	}

	log.Printf("Processed query: %s", query.QueryText)

	// Return the document scores
	return docScores, nil
}

// getDocumentScoresHandler handles the /getDocumentScores API endpoint
func getDocumentScoresHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the user query from the request body
	var query Query
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Call the internal function to get document scores
	docScores, err := GetDocumentScores(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the document scores as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(docScores); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

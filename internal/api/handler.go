package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"rpi-search-ranking/internal/ranking"
)

// HTTP timeout
const httpTimeout = 10 * time.Second

// GetDocumentScores returns the scores and metadata for relevant documents based on the query
func GetDocumentScores(query ranking.Query) ([]ranking.Document, error) {
	// Validate the query text
	if query.Text == "" {
		return nil, errors.New("query text cannot be empty")
	}

	// Create a new HTTP client
	client := &http.Client{
		Timeout: httpTimeout,
	}

	// Call the ranking logic from internal/rank
	docScores, err := ranking.RankDocuments(query, client)
	if err != nil {
		return nil, err
	}

	log.Printf("Processed query ID: %s, Query Text: %s", query.Id, query.Text)

	// Return the document scores
	return docScores, nil
}

// getDocumentScoresHandler handles the /getDocumentScores API endpoint
func getDocumentScoresHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the user query from the request body
	var query ranking.Query
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

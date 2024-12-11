package api

import (
	"errors"
	"log"
	"rpi-search-ranking/internal/ranking"
	"rpi-search-ranking/internal/utils"
	"time"
)

// HTTP timeout
const httpTimeout = 10 * time.Second

// GetDocumentScores returns the scores and metadata for relevant documents based on the query
// It also takes in an evaluation object to record metrics
func GetDocumentScores(query ranking.Query, eval *utils.Evaluation) ([]ranking.Document, error) {

	// Validate the query text
	if query.Text == "" {
		return nil, errors.New("query text cannot be empty")
	}

	// Create a new HTTP client
	// client := &http.Client{
	// 	Timeout: httpTimeout,
	// }

	// Start timer
	startTime := time.Now()

	// Call the ranking logic from internal/rank 
	// docScores, err := ranking.RankDocuments(query, client)

	
	docScores := []ranking.Document{
		{
			DocID: "doc2",
			Rank:  1,
			Metadata: ranking.DocumentMetadata{
				DocLength:       100,
				TimeLastUpdated: "2024-11-09T15:30:00Z",
				FileType:        "PDF",
				ImageCount:      3,
				DocTitle:        "Introduction to Data Science",
				URL:             "http://example2.com",
			},
		},
		{
			DocID: "doc1",
			Rank:  2,
			Metadata: ranking.DocumentMetadata{
				DocLength:       100,
				TimeLastUpdated: "2024-11-09T15:30:00Z",
				FileType:        "PDF",
				ImageCount:      3,
				DocTitle:        "Introduction to Data Science",
				URL:             "http://example1.com",
			},
		},
	} 
	
	// End timer 
	endTime := time.Since(startTime)

	// Record metric to evaluation
	eval.AlgorithmRunTime = endTime
	eval.QueryData.NumRankedDocuments = len(docScores) 

	// // Error for getting ranked documents 
	// if err != nil {
	// 	return nil, err
	// }

	log.Printf("Processed query ID: %s, Query Text: %s", query.Id, query.Text)
    
	// Return the document scores
	return docScores, nil
}

// // getDocumentScoresHandler handles the /getDocumentScores API endpoint
// func getDocumentScoresHandler(w http.ResponseWriter, r *http.Request) {
// 	// Parse the user query from the request body
// 	var query ranking.Query
// 	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	// Call the internal function to get document scores
// 	docScores, err := GetDocumentScores(query)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	// Return the document scores as JSON
// 	w.Header().Set("Content-Type", "application/json")
// 	if err := json.NewEncoder(w).Encode(docScores); err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 	}
// }

package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"rpi-search-ranking/internal/api"
)

func main() {
	// Initialize the API router
	r := mux.NewRouter()
	r.HandleFunc("/getDocumentScores", getDocumentScores).Methods("POST")

	// Start the server
	log.Println("Starting API server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// Handler function for the /getDocumentScores endpoint
func getDocumentScores(w http.ResponseWriter, r *http.Request) {
	// Parse the user query from the request body
	var query api.Query
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Call the internal function to get document scores
	docScores, err := api.GetDocumentScores(query)
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

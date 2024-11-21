package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"rpi-search-ranking/internal/ranking"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"rpi-search-ranking/internal/api"
)

func main() {
	// Initialize the API router
	r := mux.NewRouter()

	// Add middleware to log the request
	r.Use(loggingMiddleware)

	// Define the endpoint
	r.HandleFunc("/getDocumentScores", getDocumentScores).Methods("POST")

	// Start the server in a goroutine
	srv := &http.Server{
		Handler: r,
		Addr:    ":8080",
	}

	// Run the server in a separate goroutine
	go func() {
		log.Println("Starting API server on port 8080...")
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal("Server failed: ", err)
		}
	}()

	// Set up a channel to listen for termination signals (Ctrl+C or kill)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive a termination signal
	<-sig

	// Gracefully shut down the server with a 5-second timeout
	log.Println("Shutting down the server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server shutdown failed: ", err)
	}
	log.Println("Server gracefully stopped.")
}

// loggingMiddleware logs each incoming request for debugging purposes
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// Handler function for the /getDocumentScores endpoint
func getDocumentScores(w http.ResponseWriter, r *http.Request) {
	// Parse the user query from the request body
	var query ranking.Query
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the query
	if query.Id == "" || query.Text == "" {
		http.Error(w, "Id and Text are required", http.StatusBadRequest)
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

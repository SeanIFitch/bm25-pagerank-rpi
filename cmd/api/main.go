package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"rpi-search-ranking/internal/api"
	"rpi-search-ranking/internal/ranking"
	"rpi-search-ranking/internal/utils"
	"syscall"
	"time"
	"github.com/gorilla/mux"
)

func main() {
	// Initialize the API router
	r := mux.NewRouter()

	// Add middleware to log the request
	r.Use(loggingMiddleware)

	// Define the endpoint using GET method
	r.HandleFunc("/getDocumentScores", getDocumentScores).Methods("GET")

	// Start the server in a goroutine
	srv := &http.Server{
		Handler: r,
		Addr:    ":6060",
	}
	 
	// Run the server in a separate goroutine
	go func() {
		log.Println("Starting Ranking API server on port 6060...")
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal("Server failed: ", err)
		}
	}()

	// Set up at channel to listen for termination signals (Ctrl+C or kill)
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

// loggingMiddleware logs each incoming request with client IP/hostname for debugging purposes
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the client's IP address from the request
		clientIP := r.RemoteAddr
		host, _, err := net.SplitHostPort(clientIP)
		if err != nil {
			log.Printf("Error extracting IP address from %s: %v", clientIP, err)
		}

		// Try to perform a reverse DNS lookup to get the hostname
		hostname, err := net.LookupAddr(host)
		if err != nil || len(hostname) == 0 {
			// Fallback if reverse DNS lookup fails
			hostname = append(hostname, "unknown")
		}

		// Log the HTTP method, path, client IP, and resolved hostname
		log.Printf("Received request: %s %s from IP: %s, Hostname: %s", r.Method, r.URL.Path, host, hostname[0])

		// Pass the request along the chain
		next.ServeHTTP(w, r)
	})
}

// Handler function for the /getDocumentScores endpoint
func getDocumentScores(w http.ResponseWriter, r *http.Request) { 
	// Create evaluation for component 
	evalObj := utils.CreateEvaluation()
	
	// Extract parameters from the URL query string
	queryId := r.URL.Query().Get("id")
	queryText := r.URL.Query().Get("text")

	// Validate the query parameters
	if queryId == "" || queryText == "" {
		sendError(w, http.StatusBadRequest, "Id and Text are required")
		return
	}

	// Call the internal function to get document scores and geenerate evaluation 
	docScores, err := api.GetDocumentScores(ranking.Query{Id: queryId, Text: queryText}, evalObj )
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to retrieve document scores")
		return
	}

	// Return the document scores as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(docScores); err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to encode response")
	}

	// Get storage
	err = evalObj.UpdateStorageSize( "./data" ) 
	if err != nil {
		log.Println( err )
	}

	// Send the evalation object to evaluation component
	err = utils.SendEvaluation( evalObj )
	if err != nil {
		log.Println( err ) 
	}

}

// sendError sends a structured error response
func sendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

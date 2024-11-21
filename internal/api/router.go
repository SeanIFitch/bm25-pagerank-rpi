package api

import (
	"github.com/gorilla/mux"
)

// SetupRouter initializes the API router and routes
func SetupRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/getDocumentScores", getDocumentScoresHandler).Methods("POST")
	return r
}

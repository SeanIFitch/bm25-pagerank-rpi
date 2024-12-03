package datagen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Evaluation struct {
	TotalStorage     byte       `json:"total_storage"`         // Total number of bytes used for storage
	AlgorithmRunTime time.Time  `json:"algorithm_update_time"` // Time to update algorithm
	QueryData        *QueryInfo `json:"query_data"`            // Pointer to query information
}

type QueryInfo struct {
	NumDocumentsParsed int       `json:"num_documents_parsed"` // Total number of documents
	NumRankedDocuments int       `json:"num_ranked_documents"` // Total number of ranked documents returned
	ProcessTime        time.Time `json:"process_time"`         // Total time used to process
	CacheExists        bool      `json:"cache_exists"`         // Are we implementing a cache?
}

// Create a evaluation object, everything is initially set to thinging 
func CreateEvalation( ) *Evaluation {
	queryData := &QueryInfo{
		NumDocumentsParsed: 0,
		NumRankedDocuments: 0,
		ProcessTime: time.Time{},
		CacheExists: false,
	}

	return &Evaluation{
		TotalStorage: byte(0), 
		AlgorithmRunTime:  time.Time{},
		QueryData: queryData,
	}
}

// Function serializes the Evaluation object and it's contents to a json string
func (e *Evaluation)SerializeToJson( ) (string, error) {
	// Marshal the struct into JSON
	jsonData, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error serializing Evaluation struct: %v", err)
	}
	// Return the JSON string
	return string(jsonData), nil
}

// SendEvaluation sends the serialized Evaluation to a server via POST request
func SendEvaluation(evaluation *Evaluation) error {
	// Serialize the Evaluation struct to JSON
	jsonStr, err := evaluation.SerializeToJson()
	if err != nil {
		return fmt.Errorf("error serializing Evaluation: %v", err)
	}

	// Replace this with the correct endpoint path later [CHANGE ENDPOINT LATER]
	serverURL := "https://lspt-data-eval.cs.rpi.edu:8080/endpoint" // Base URL
	
	// Create a new POST request with JSON data in the body
	req, err := http.NewRequest("POST", serverURL, bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Set the content-type to application/json
	req.Header.Set("Content-Type", "application/json")

	// Send the request using the default HTTP client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Optionally, read and log the response body if needed
	// body, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println("Response body:", string(body))

	fmt.Println("Successfully sent evaluation to server")
	return nil
}
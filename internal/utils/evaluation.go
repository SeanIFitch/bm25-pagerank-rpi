package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Evaluation struct {
	TotalStorage     byte          `json:"total_storage"`         // Total number of bytes used for storage
	AlgorithmRunTime time.Duration `json:"algorithm_update_time"` // Time to update algorithm
	QueryData        *QueryInfo    `json:"query_data"`            // Pointer to query information
}

type QueryInfo struct {
	NumDocumentsParsed int           `json:"num_documents_parsed"` // Total number of documents
	NumRankedDocuments int           `json:"num_ranked_documents"` // Total number of ranked documents returned
	ProcessTime        time.Duration `json:"process_time"`         // Total time used to process
	CacheExists        bool          `json:"cache_exists"`         // Are we implementing a cache?
}

// Create a new evaluation object, everything is initially set to default values
func CreateEvaluation() *Evaluation {
	queryData := &QueryInfo{
		NumDocumentsParsed: 0,
		NumRankedDocuments: 0,
		ProcessTime:        0, 
		CacheExists:        false,
	}

	return &Evaluation{
		TotalStorage:     byte(0),
		AlgorithmRunTime: 0, 
		QueryData:        queryData,
	}
}

// Serialize the Evaluation object and its contents into a JSON string
func (e *Evaluation) SerializeToJson() (string, error) {
	// Marshal the struct into JSON with indentation for better readability
	jsonData, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error serializing Evaluation struct: %v", err)
	}
	// Return the formatted JSON string
	return string(jsonData), nil
}


// Counts the amount of storage that ranking component is storing. 
func (e *Evaluation) UpdateStorageSize(dirPath string) error {
	var count int
	stack := []string{dirPath}

	for len(stack) > 0 {
		// Pop the last directory off the stack
		currentDir := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		// Open the directory
		dirEntries, err := os.ReadDir(currentDir)
		if err != nil {
			return err
		}

		// Iterate over each entry in the directory
		for _, entry := range dirEntries {
			entryPath := filepath.Join(currentDir, entry.Name())
			count++

			// If the entry is a directory, add it to the stack to visit later
			if entry.IsDir() {
				stack = append(stack, entryPath)
			}
		}
	}
	return nil
}

// Print evaluation json 


// SendEvaluation sends the serialized Evaluation to the server via a POST request
func SendEvaluation(evaluation *Evaluation) error {
	// Serialize the Evaluation struct to JSON
	jsonStr, err := evaluation.SerializeToJson()
	if err != nil {
		return fmt.Errorf("error serializing Evaluation: %v", err)
	}

	// New endpoint URL
	serverURL := "http://lspt-link-analysis.cs.rpi.edu:1234/evaluation/add_node/update_node_info"

	// Create a new POST request with JSON data in the body
	req, err := http.NewRequest("POST", serverURL, bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Set the content-type header to application/json
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

	// Read and log the response body if needed
	// body, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println("Response body:", string(body))

	// Successful request
	fmt.Println("Successfully sent evaluation to server")
	return nil
}



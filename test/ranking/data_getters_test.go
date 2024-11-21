package ranking

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

type DocData struct {
	DocID     string `json:"docID"`
	Frequency int    `json:"frequency"`
	Positions []int  `json:"positions"`
}

// TermIndex represents the index for a single term, including the term field
type TermIndex struct {
	Term  string    `json:"term"`
	Index []DocData `json:"index"`
}

func generateSyntheticData(terms []string, numDocuments int, maxFrequency int) map[string]TermIndex {
	rand.Seed(0)
	data := make(map[string]TermIndex)

	for _, term := range terms {
		var docs []DocData
		for docID := 1; docID <= numDocuments; docID++ {
			// Generate a random frequency
			frequency := rand.Intn(maxFrequency) + 1

			// Generate positions array with exactly frequency elements
			positions := make([]int, frequency)
			for i := 0; i < frequency; i++ {
				positions[i] = rand.Intn(1000) + 1 // Random positions between 1 and 1000
			}

			// Add document data
			docs = append(docs, DocData{
				DocID:     fmt.Sprintf("%d", docID),
				Frequency: frequency,
				Positions: positions,
			})
		}

		// Add term index, including the term field
		data[term] = TermIndex{
			Term:  term,
			Index: docs,
		}
	}
	return data
}

func main() {
	// Terms to generate data for
	terms := []string{"cat", "dog", "bird", "fish"}
	numDocuments := 5
	maxFrequency := 10

	// Generate synthetic data
	data := generateSyntheticData(terms, numDocuments, maxFrequency)

	// Convert to JSON and print
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	fmt.Println(string(jsonData))
}

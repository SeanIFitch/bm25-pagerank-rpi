package ranking

// BM25 parameters
const k1 = 1.5
const b = 0.75

// Query defines the struct to parse the incoming query
type Query struct {
	Id   string `json:"queryID"`
	Text string `json:"queryText"`
}

// Document represents a document with its ID, rank, and metadata
type Document struct {
	DocID    string                 `json:"docID"`
	Rank     int                    `json:"rank"`
	Metadata map[string]interface{} `json:"metadata"`
}

// DocumentIndex represents a document and the frequency and positions of a term in that document
type DocumentIndex struct {
	DocID     string `json:"docID"`
	Frequency int    `json:"frequency"`
	Positions []int  `json:"positions"`
}

// InvertibleIndex represents the inverted index structure, mapping a term to its list of document occurrences
type InvertibleIndex map[string][]DocumentIndex

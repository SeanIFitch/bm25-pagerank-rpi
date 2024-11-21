package ranking

import "time"

// BM25 parameters
const k1 = 1.2
const b = 0.75

// HTTP timeout
const httpTimeout = 10 * time.Second

// Query defines the struct to parse the incoming query
type Query struct {
	Id   string `json:"queryID"`
	Text string `json:"queryText"`
}

// Document represents a document with its ID, rank, and metadata
type Document struct {
	DocID           string           `json:"docID"`
	Rank            int              `json:"rank"`
	Metadata        DocumentMetadata `json:"metadata"`
	TermFrequencies map[string]int
	Features        Features
}

type Documents []Document

// DocumentMetadata holds metadata information about a document
type DocumentMetadata struct {
	DocLength       int    `json:"docLength"`
	TimeLastUpdated string `json:"timeLastUpdated"`
	FileType        string `json:"fileType"`
	ImageCount      int    `json:"imageCount"`
	DocTitle        string `json:"docTitle"`
	URL             string `json:"URL"`
}

// Features holds various statistical and computed features related to a document/query.
type Features struct {
	// Covered Query Term Metrics
	CoveredQueryTermNumber int     // Number of query terms covered
	CoveredQueryTermRatio  float64 // Ratio of covered query terms to total query terms

	// Term Frequency Statistics
	SumTermFrequency      int     // Sum of term frequencies
	MinTermFrequency      int     // Minimum term frequency
	MaxTermFrequency      int     // Maximum term frequency
	MeanTermFrequency     float64 // Mean of term frequencies
	VarianceTermFrequency float64 // Variance of term frequencies

	// Stream Length Statistics (normalized term frequencies)
	StreamLength                     int     // Length of the stream (or document length)
	SumStreamLengthNormalizedTF      float64 // Sum of stream length normalized term frequency
	MinStreamLengthNormalizedTF      float64 // Min stream length normalized term frequency
	MaxStreamLengthNormalizedTF      float64 // Max stream length normalized term frequency
	MeanStreamLengthNormalizedTF     float64 // Mean stream length normalized term frequency
	VarianceStreamLengthNormalizedTF float64 // Variance of stream length normalized term frequency

	// Inverse Document Frequency (IDF)
	IDF           float64 // IDF for the query term
	SumTFIDF      float64 // Sum of tf*idf for all relevant documents
	MinTFIDF      float64 // Minimum tf*idf value
	MaxTFIDF      float64 // Maximum tf*idf value
	MeanTFIDF     float64 // Mean of tf*idf values
	VarianceTFIDF float64 // Variance of tf*idf values

	// BM25 score for the document/query
	BM25 float64 // BM25 score

	// URL characteristics
	NumSlashesInURL int // Number of slashes in the URL
	LengthOfURL     int // Length of the URL

	// Link Analysis Metrics
	InlinkCount  int     // Number of inlinks
	OutlinkCount int     // Number of outlinks
	PageRank     float64 // PageRank score
}

// DocumentIndex represents a document and the frequency and positions of a term in that document
type DocumentIndex struct {
	DocID     string `json:"docID"`
	Frequency int    `json:"frequency"`
	Positions []int  `json:"positions"`
}

// InvertibleIndex represents the inverted index structure, mapping a term to its list of document occurrences
type InvertibleIndex map[string][]DocumentIndex

// TotalDocStatistics represents the statistics for all documents in the database returned by getTotalDocStatistics
type TotalDocStatistics struct {
	AvgDocLength float64 `json:"avgDocLength"`
	DocCount     int     `json:"docCount"`
}

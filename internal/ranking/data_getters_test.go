package ranking

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

// Helper to handle errors and varied responses
func createMockHTTPClient(responses map[string]string, errors map[string]error, statusCode int) *http.Client {
	return &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if err, ok := errors[req.URL.String()]; ok {
				return nil, err
			}
			if res, ok := responses[req.URL.String()]; ok {
				return &http.Response{
					StatusCode: statusCode,
					Body:       io.NopCloser(strings.NewReader(res)),
					Header:     make(http.Header),
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("")),
				Header:     make(http.Header),
			}, nil
		}),
	}
}

// roundTripFunc allows creating a RoundTripper from a function
type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func Test_getInvertibleIndex(t *testing.T) {
	type args struct {
		client *http.Client
		query  Query
	}
	tests := []struct {
		name    string
		args    args
		want    invertibleIndex
		wantErr bool
	}{
		{
			name: "Successful retrieval of invertible index for multiple terms",
			args: args{
				client: createMockHTTPClient(
					map[string]string{
						"http://lspt-index-ranking.cs.rpi.edu/get-invertible-index?term=testterm1": `{
                            "term": "testterm1",
                            "index": [
                                {"docID": "doc1", "frequency": 3, "positions": [1, 4, 7]},
                                {"docID": "doc2", "frequency": 2, "positions": [2, 5]}
                            ]
                        }`,
						"http://lspt-index-ranking.cs.rpi.edu/get-invertible-index?term=testterm2": `{
                            "term": "testterm2",
                            "index": [
                                {"docID": "doc3", "frequency": 1, "positions": [8]}
                            ]
                        }`,
					},
					map[string]error{}, // No errors
					http.StatusOK,      // Status OK
				),
				query: Query{Terms: []string{"testterm1", "testterm2"}},
			},
			want: invertibleIndex{
				"testterm1": {
					{DocID: "doc1", Frequency: 3, Positions: []int{1, 4, 7}},
					{DocID: "doc2", Frequency: 2, Positions: []int{2, 5}},
				},
				"testterm2": {
					{DocID: "doc3", Frequency: 1, Positions: []int{8}},
				},
			},
			wantErr: false,
		},
		{
			name: "Network error for one of the terms",
			args: args{
				client: createMockHTTPClient(
					map[string]string{
						"http://lspt-index-ranking.cs.rpi.edu/get-invertible-index?term=testterm1": `{
                            "term": "testterm1",
                            "index": [
                                {"docID": "doc1", "frequency": 3, "positions": [1, 4, 7]}
                            ]
                        }`,
						"http://lspt-index-ranking.cs.rpi.edu/get-invertible-index?term=testterm2": "",
					},
					map[string]error{
						"http://lspt-index-ranking.cs.rpi.edu/get-invertible-index?term=testterm2": fmt.Errorf("network error"),
					},
					http.StatusOK,
				),
				query: Query{Terms: []string{"testterm1", "testterm2"}},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Empty query",
			args: args{
				client: createMockHTTPClient(
					map[string]string{}, // No responses
					map[string]error{},  // No errors
					http.StatusOK,
				),
				query: Query{Terms: []string{}},
			},
			want:    invertibleIndex{},
			wantErr: false,
		},
		{
			name: "Duplicate terms in query",
			args: args{
				client: createMockHTTPClient(
					map[string]string{
						"http://lspt-index-ranking.cs.rpi.edu/get-invertible-index?term=testterm1": `{
                            "term": "testterm1",
                            "index": [
                                {"docID": "doc1", "frequency": 3, "positions": [1, 4, 7]}
                            ]
                        }`,
					},
					map[string]error{}, // No errors
					http.StatusOK,
				),
				query: Query{Terms: []string{"testterm1", "testterm1"}},
			},
			want: invertibleIndex{
				"testterm1": {
					{DocID: "doc1", Frequency: 3, Positions: []int{1, 4, 7}},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getInvertibleIndex(tt.args.client, tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("getInvertibleIndex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getInvertibleIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_fetchInvertibleIndexForTerm(t *testing.T) {
	type args struct {
		client *http.Client
		term   string
	}
	tests := []struct {
		name    string
		args    args
		want    []documentIndex
		wantErr bool
	}{
		{
			name: "Successful retrieval of index",
			args: args{
				client: createMockHTTPClient(
					map[string]string{
						// Use the full URL for the term "testterm"
						"http://lspt-index-ranking.cs.rpi.edu/get-invertible-index?term=testterm": `{
                            "term": "testterm", 
                            "index": [
                                {"docID": "doc1", "frequency": 2, "positions": [5, 15]},
                                {"docID": "doc2", "frequency": 1, "positions": [10]}
                            ]
                        }`,
					},
					map[string]error{}, // No errors
					http.StatusOK,      // Status OK
				),
				term: "testterm",
			},
			want: []documentIndex{
				{DocID: "doc1", Frequency: 2, Positions: []int{5, 15}},
				{DocID: "doc2", Frequency: 1, Positions: []int{10}},
			},
			wantErr: false,
		},
		{
			name: "Network error",
			args: args{
				client: createMockHTTPClient(
					map[string]string{}, // No responses
					map[string]error{
						"http://lspt-index-ranking.cs.rpi.edu/get-invertible-index?term=errorterm": fmt.Errorf("network error"),
					},
					http.StatusOK,
				),
				term: "errorterm",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Invalid HTTP response",
			args: args{
				client: createMockHTTPClient(
					map[string]string{},            // No responses
					map[string]error{},             // No errors
					http.StatusInternalServerError, // Internal server error
				),
				term: "invalidresponse",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fetchInvertibleIndexForTerm(tt.args.client, tt.args.term)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchInvertibleIndexForTerm() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fetchInvertibleIndexForTerm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_fetchDocumentMetadata(t *testing.T) {
	type args struct {
		client *http.Client
		docID  string
	}
	tests := []struct {
		name    string
		args    args
		want    DocumentMetadata
		wantErr bool
	}{
		{
			name: "Successful retrieval of metadata",
			args: args{
				client: createMockHTTPClient(
					map[string]string{
						// Use the full URL for the docID "12345"
						"http://lspt-index-ranking.cs.rpi.edu/get-document-metadata?docID=12345": `{
							"docID": "12345",
							"metadata": {
								"docLength": 2450,
								"timeLastUpdated": "2024-11-09T15:30:00Z",
								"docType": "PDF",
								"imageCount": 3,
								"docTitle": "Introduction to Data Science",
								"URL": "https://example.com/documents/12345"
							}
						}`,
					},
					map[string]error{}, // No errors
					http.StatusOK,      // Status OK
				),
				docID: "12345",
			},
			want: DocumentMetadata{
				DocLength:       2450,
				TimeLastUpdated: "2024-11-09T15:30:00Z",
				FileType:        "PDF",
				ImageCount:      3,
				DocTitle:        "Introduction to Data Science",
				URL:             "https://example.com/documents/12345",
			},
			wantErr: false,
		},
		{
			name: "Network error",
			args: args{
				client: createMockHTTPClient(
					map[string]string{}, // No responses
					map[string]error{
						"http://lspt-index-ranking.cs.rpi.edu/get-document-metadata?docID=doc2": fmt.Errorf("network error"),
					},
					http.StatusOK,
				),
				docID: "doc2",
			},
			want:    DocumentMetadata{},
			wantErr: true,
		},
		{
			name: "Invalid HTTP response",
			args: args{
				client: createMockHTTPClient(
					map[string]string{},            // No responses
					map[string]error{},             // No errors
					http.StatusInternalServerError, // Internal server error
				),
				docID: "doc3",
			},
			want:    DocumentMetadata{},
			wantErr: true,
		},
		{
			name: "Malformed JSON response",
			args: args{
				client: createMockHTTPClient(
					map[string]string{
						"http://lspt-index-ranking.cs.rpi.edu/get-document-metadata?docID=doc4": "{invalid json}",
					},
					map[string]error{}, // No errors
					http.StatusOK,      // Status OK
				),
				docID: "doc4",
			},
			want:    DocumentMetadata{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fetchDocumentMetadata(tt.args.client, tt.args.docID)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchDocumentMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fetchDocumentMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_fetchTotalDocStatistics(t *testing.T) {
	type args struct {
		client *http.Client
	}
	tests := []struct {
		name    string
		args    args
		want    totalDocStatistics
		wantErr bool
	}{
		{
			name: "Successful retrieval of total document statistics",
			args: args{
				client: createMockHTTPClient(
					map[string]string{
						// Use the full URL for the total doc statistics endpoint
						"http://lspt-index-ranking.cs.rpi.edu/get-total-doc-statistics": `{
							"avgDocLength": 798.8730,
							"docCount": 456789
						}`,
					},
					map[string]error{}, // No errors
					http.StatusOK,      // Status OK
				),
			},
			want: totalDocStatistics{
				AvgDocLength: 798.8730,
				DocCount:     456789,
			},
			wantErr: false,
		},
		{
			name: "Network error",
			args: args{
				client: createMockHTTPClient(
					map[string]string{}, // No responses
					map[string]error{
						"http://lspt-index-ranking.cs.rpi.edu/get-total-doc-statistics": fmt.Errorf("network error"),
					},
					http.StatusOK,
				),
			},
			want:    totalDocStatistics{},
			wantErr: true,
		},
		{
			name: "Invalid HTTP response",
			args: args{
				client: createMockHTTPClient(
					map[string]string{},            // No responses
					map[string]error{},             // No errors
					http.StatusInternalServerError, // Internal server error
				),
			},
			want:    totalDocStatistics{},
			wantErr: true,
		},
		{
			name: "Malformed JSON response",
			args: args{
				client: createMockHTTPClient(
					map[string]string{
						"http://lspt-index-ranking.cs.rpi.edu/get-total-doc-statistics": "{invalid json}",
					},
					map[string]error{}, // No errors
					http.StatusOK,      // Status OK
				),
			},
			want:    totalDocStatistics{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fetchTotalDocStatistics(tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchTotalDocStatistics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fetchTotalDocStatistics() = %v, want %v", got, tt.want)
			}
		})
	}
}

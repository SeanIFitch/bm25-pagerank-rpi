package ranking

import (
	"net/http"
	"reflect"
	"testing"
)

func TestRankDocuments(t *testing.T) {
	type args struct {
		query  Query
		client *http.Client
	}
	tests := []struct {
		name    string
		args    args
		want    []Document
		wantErr bool
	}{
		{
			name: "Single Term and Document",
			args: args{
				query: Query{
					Id:   "query1",
					Text: "term1",
				},
				client: createMockHTTPClient(
					map[string]string{
						"https://lspt-index-ranking.cs.rpi.edu/get-invertible-index?term=term1": `{
                            "term": "term1",
                            "index": [
                                {"docID": "doc1", "frequency": 1, "positions": [8]}
                            ]
                        }`,
						"https://lspt-index-ranking.cs.rpi.edu/get-document-metadata?docID=doc1": `{
							"docID": "doc1",
							"metadata": {
								"docLength": 100,
								"timeLastUpdated": "2024-11-09T15:30:00Z",
								"docType": "PDF",
								"imageCount": 3,
								"docTitle": "Introduction to Data Science",
								"URL": "https://example.com"
							}
						}`,
						"https://lspt-index-ranking.cs.rpi.edu/get-total-doc-statistics": `{
							"avgDocLength": 120.0,
							"docCount": 10
						}`,
						"https://lspt-index-ranking.cs.rpi.edu/get-pagerank?URL=https://example.com": `{
							"pageRank": 0.85,
							"inLinkCount": 123,
							"outLinkCount": 45
						}`,
					},
					map[string]error{}, // No errors
					http.StatusOK,      // Status OK
				),
			},
			want: []Document{
				{
					DocID: "doc1",
					Rank:  1,
					Metadata: DocumentMetadata{
						DocLength:       100,
						TimeLastUpdated: "2024-11-09T15:30:00Z",
						FileType:        "PDF",
						ImageCount:      3,
						DocTitle:        "Introduction to Data Science",
						URL:             "https://example.com",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Single Term and Two Documents With differing term frequencies",
			args: args{
				query: Query{
					Id:   "query1",
					Text: "term1",
				},
				client: createMockHTTPClient(
					map[string]string{
						"https://lspt-index-ranking.cs.rpi.edu/get-invertible-index?term=term1": `{
                            "term": "term1",
                            "index": [
                                {"docID": "doc1", "frequency": 1, "positions": [8]},
								{"docID": "doc2", "frequency": 2, "positions": [8, 19]}
                            ]
                        }`,
						"https://lspt-index-ranking.cs.rpi.edu/get-document-metadata?docID=doc1": `{
							"docID": "doc1",
							"metadata": {
								"docLength": 100,
								"timeLastUpdated": "2024-11-09T15:30:00Z",
								"docType": "PDF",
								"imageCount": 3,
								"docTitle": "Introduction to Data Science",
								"URL": "https://example1.com"
							}
						}`,
						"https://lspt-index-ranking.cs.rpi.edu/get-document-metadata?docID=doc2": `{
							"docID": "doc2",
							"metadata": {
								"docLength": 100,
								"timeLastUpdated": "2024-11-09T15:30:00Z",
								"docType": "PDF",
								"imageCount": 3,
								"docTitle": "Introduction to Data Science",
								"URL": "https://example2.com"
							}
						}`,
						"https://lspt-index-ranking.cs.rpi.edu/get-total-doc-statistics": `{
							"avgDocLength": 120.0,
							"docCount": 10
						}`,
						"https://lspt-index-ranking.cs.rpi.edu/get-pagerank?URL=https://example1.com": `{
							"pageRank": 0.85,
							"inLinkCount": 123,
							"outLinkCount": 45
						}`,
						"https://lspt-index-ranking.cs.rpi.edu/get-pagerank?URL=https://example2.com": `{
							"pageRank": 0.85,
							"inLinkCount": 123,
							"outLinkCount": 45
						}`,
					},
					map[string]error{}, // No errors
					http.StatusOK,      // Status OK
				),
			},
			want: []Document{
				{
					DocID: "doc2",
					Rank:  1,
					Metadata: DocumentMetadata{
						DocLength:       100,
						TimeLastUpdated: "2024-11-09T15:30:00Z",
						FileType:        "PDF",
						ImageCount:      3,
						DocTitle:        "Introduction to Data Science",
						URL:             "https://example2.com",
					},
				},
				{
					DocID: "doc1",
					Rank:  2,
					Metadata: DocumentMetadata{
						DocLength:       100,
						TimeLastUpdated: "2024-11-09T15:30:00Z",
						FileType:        "PDF",
						ImageCount:      3,
						DocTitle:        "Introduction to Data Science",
						URL:             "https://example1.com",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Single Term and No Documents",
			args: args{
				query: Query{
					Id:   "query1",
					Text: "term1",
				},
				client: createMockHTTPClient(
					map[string]string{
						"https://lspt-index-ranking.cs.rpi.edu/get-invertible-index?term=term1": `{
							"term": "term1",
							"index": []
						}`,
					},
					map[string]error{}, // No errors
					http.StatusOK,      // Status OK
				),
			},
			want:    []Document{}, // No documents should be returned
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RankDocuments(tt.args.query, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("RankDocuments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i := range got {
				if got[i].DocID != tt.want[i].DocID {
					t.Errorf("RankDocuments() = %v, want %v", got, tt.want)
				}
				if got[i].Rank != tt.want[i].Rank {
					t.Errorf("RankDocuments() = %v, want %v", got, tt.want)
				}
				if !reflect.DeepEqual(got[i].Metadata, tt.want[i].Metadata) {
					t.Errorf("RankDocuments() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func Test_getDocuments(t *testing.T) {
	type args struct {
		index invertibleIndex
	}
	tests := []struct {
		name    string
		args    args
		want    Documents
		wantErr bool
	}{
		{
			name: "Empty index",
			args: args{
				index: invertibleIndex{},
			},
			want:    Documents{}, // Should return an empty slice
			wantErr: false,
		},
		{
			name: "Single term, single document",
			args: args{
				index: invertibleIndex{
					"term1": {
						{DocID: "doc1", Frequency: 1, Positions: []int{0}},
					},
				},
			},
			want: Documents{
				{
					DocID:           "doc1",
					TermFrequencies: map[string]int{"term1": 1},
				},
			},
			wantErr: false,
		},
		{
			name: "Multiple terms, single document",
			args: args{
				index: invertibleIndex{
					"term1": {
						{DocID: "doc1", Frequency: 1, Positions: []int{0}},
					},
					"term2": {
						{DocID: "doc1", Frequency: 2, Positions: []int{1, 2}},
					},
				},
			},
			want: Documents{
				{
					DocID:           "doc1",
					TermFrequencies: map[string]int{"term1": 1, "term2": 2},
				},
			},
			wantErr: false,
		},
		{
			name: "Multiple terms, multiple documents",
			args: args{
				index: invertibleIndex{
					"term1": {
						{DocID: "doc1", Frequency: 1, Positions: []int{0}},
					},
					"term2": {
						{DocID: "doc1", Frequency: 2, Positions: []int{1, 2}},
						{DocID: "doc2", Frequency: 1, Positions: []int{3}},
					},
				},
			},
			want: Documents{
				{
					DocID:           "doc1",
					TermFrequencies: map[string]int{"term1": 1, "term2": 2},
				},
				{
					DocID:           "doc2",
					TermFrequencies: map[string]int{"term2": 1},
				},
			},
			wantErr: false,
		},
		{
			name: "Term frequencies aggregation",
			args: args{
				index: invertibleIndex{
					"term1": {
						{DocID: "doc1", Frequency: 1, Positions: []int{0}},
						{DocID: "doc1", Frequency: 2, Positions: []int{1, 2}},
					},
					"term2": {
						{DocID: "doc1", Frequency: 1, Positions: []int{0}},
						{DocID: "doc2", Frequency: 1, Positions: []int{3}},
					},
				},
			},
			want: Documents{
				{
					DocID:           "doc1",
					TermFrequencies: map[string]int{"term1": 3, "term2": 1},
				},
				{
					DocID:           "doc2",
					TermFrequencies: map[string]int{"term2": 1},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getDocuments(tt.args.index)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDocuments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Compare documents without regard for order using maps
			gotMap := make(map[string]Document)
			wantMap := make(map[string]Document)
			for _, doc := range got {
				gotMap[doc.DocID] = doc
			}
			for _, doc := range tt.want {
				wantMap[doc.DocID] = doc
			}
			if !reflect.DeepEqual(gotMap, wantMap) {
				t.Errorf("getDocuments() = %v, want %v", got, tt.want)
			}
		})
	}
}

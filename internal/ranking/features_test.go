package ranking

import (
	"math"
	"net/http"
	"reflect"
	"testing"

	"gonum.org/v1/gonum/stat"
)

func compareFeatures(got, want Features, epsilon float64) bool {
	return got.CoveredQueryTermNumber == want.CoveredQueryTermNumber &&
		math.Abs(got.CoveredQueryTermRatio-want.CoveredQueryTermRatio) <= epsilon &&
		got.SumTermFrequency == want.SumTermFrequency &&
		got.MinTermFrequency == want.MinTermFrequency &&
		got.MaxTermFrequency == want.MaxTermFrequency &&
		math.Abs(got.MeanTermFrequency-want.MeanTermFrequency) <= epsilon &&
		math.Abs(got.VarianceTermFrequency-want.VarianceTermFrequency) <= epsilon &&
		got.StreamLength == want.StreamLength &&
		math.Abs(got.SumStreamLengthNormalizedTF-want.SumStreamLengthNormalizedTF) <= epsilon &&
		math.Abs(got.MinStreamLengthNormalizedTF-want.MinStreamLengthNormalizedTF) <= epsilon &&
		math.Abs(got.MaxStreamLengthNormalizedTF-want.MaxStreamLengthNormalizedTF) <= epsilon &&
		math.Abs(got.MeanStreamLengthNormalizedTF-want.MeanStreamLengthNormalizedTF) <= epsilon &&
		math.Abs(got.VarianceStreamLengthNormalizedTF-want.VarianceStreamLengthNormalizedTF) <= epsilon &&
		math.Abs(got.SumTFIDF-want.SumTFIDF) <= epsilon &&
		math.Abs(got.MinTFIDF-want.MinTFIDF) <= epsilon &&
		math.Abs(got.MaxTFIDF-want.MaxTFIDF) <= epsilon &&
		math.Abs(got.MeanTFIDF-want.MeanTFIDF) <= epsilon &&
		math.Abs(got.VarianceTFIDF-want.VarianceTFIDF) <= epsilon &&
		math.Abs(got.BM25-want.BM25) <= epsilon &&
		got.NumSlashesInURL == want.NumSlashesInURL &&
		got.LengthOfURL == want.LengthOfURL &&
		got.InlinkCount == want.InlinkCount &&
		got.OutlinkCount == want.OutlinkCount &&
		math.Abs(got.PageRank-want.PageRank) <= epsilon
}

func Test_getIDF(t *testing.T) {
	type args struct {
		index         invertibleIndex
		totalDocCount int
	}
	tests := []struct {
		name string
		args args
		want map[string]float64
	}{
		{
			name: "Basic Case",
			args: args{
				index: invertibleIndex{
					"term1": {
						{DocID: "doc1", Frequency: 2, Positions: []int{1, 2}},
						{DocID: "doc2", Frequency: 1, Positions: []int{1}},
					},
					"term2": {
						{DocID: "doc1", Frequency: 1, Positions: []int{1}},
					},
				},
				totalDocCount: 3,
			},
			want: map[string]float64{
				"term1": math.Log(3.0 / 3.0), // 3 documents, 2 document occurrences
				"term2": math.Log(3.0 / 2.0), // 3 documents, 1 document occurrence
			},
		},
		{
			name: "Edge Case with No Documents",
			args: args{
				index:         invertibleIndex{},
				totalDocCount: 0,
			},
			want: map[string]float64{},
		},
		{
			name: "Single Document Case",
			args: args{
				index: invertibleIndex{
					"term1": {
						{DocID: "doc1", Frequency: 5, Positions: []int{1, 2, 3, 4, 5}},
					},
				},
				totalDocCount: 1,
			},
			want: map[string]float64{
				"term1": math.Log(1.0 / 2.0), // 1 document, 1 document occurrence (smoothed)
			},
		},
		{
			name: "Smoothed IDF",
			args: args{
				index: invertibleIndex{
					"term1": {
						{DocID: "doc1", Frequency: 1, Positions: []int{1}},
						{DocID: "doc2", Frequency: 1, Positions: []int{1}},
					},
					"term2": {
						{DocID: "doc1", Frequency: 1, Positions: []int{1}},
					},
				},
				totalDocCount: 5,
			},
			want: map[string]float64{
				"term1": math.Log(5.0 / (2.0 + 1.0)), // 5 documents, 2 document occurrences (smoothed)
				"term2": math.Log(5.0 / (1.0 + 1.0)), // 5 documents, 1 document occurrence (smoothed)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getIDF(tt.args.index, tt.args.totalDocCount); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getIDF() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculateTermFrequencyStats(t *testing.T) {
	type args struct {
		query           Query
		termFrequencies map[string]int
	}
	tests := []struct {
		name         string
		args         args
		wantSum      int
		wantMin      int
		wantMax      int
		wantMean     float64
		wantVariance float64
	}{
		{
			name: "Basic Case",
			args: args{
				query: Query{
					Id:    "q1",
					Text:  "example query",
					Terms: []string{"term1", "term2", "term3"},
				},
				termFrequencies: map[string]int{
					"term1": 3,
					"term2": 5,
					"term3": 2,
				},
			},
			wantSum:      10,
			wantMin:      2,
			wantMax:      5,
			wantMean:     3.3333333333333335,
			wantVariance: stat.PopVariance([]float64{3, 5, 2}, nil),
		},
		{
			name: "Edge Case - No Terms Found",
			args: args{
				query: Query{
					Id:    "q2",
					Text:  "another query",
					Terms: []string{"term4", "term5"},
				},
				termFrequencies: map[string]int{
					"term1": 3,
					"term2": 5,
					"term3": 2,
				},
			},
			wantSum:      0,
			wantMin:      0,
			wantMax:      0,
			wantMean:     0.0,
			wantVariance: 0.0,
		},
		{
			name: "Single Term Query",
			args: args{
				query: Query{
					Id:    "q3",
					Text:  "single term query",
					Terms: []string{"term2"},
				},
				termFrequencies: map[string]int{
					"term2": 4,
				},
			},
			wantSum:      4,
			wantMin:      4,
			wantMax:      4,
			wantMean:     4.0,
			wantVariance: 0.0,
		},
		{
			name: "Multiple Terms with Same Frequency",
			args: args{
				query: Query{
					Id:    "q4",
					Text:  "same frequency terms",
					Terms: []string{"term1", "term2", "term3"},
				},
				termFrequencies: map[string]int{
					"term1": 3,
					"term2": 3,
					"term3": 3,
				},
			},
			wantSum:      9,
			wantMin:      3,
			wantMax:      3,
			wantMean:     3.0,
			wantVariance: 0.0,
		},
		{
			name: "Large Frequency Values",
			args: args{
				query: Query{
					Id:    "q5",
					Text:  "large frequency terms",
					Terms: []string{"term1", "term2"},
				},
				termFrequencies: map[string]int{
					"term1": 10000,
					"term2": 20000,
				},
			},
			wantSum:      30000,
			wantMin:      10000,
			wantMax:      20000,
			wantMean:     15000.0,
			wantVariance: 2.5e+07,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSum, gotMin, gotMax, gotMean, gotVariance := calculateTermFrequencyStats(tt.args.query, tt.args.termFrequencies)
			if gotSum != tt.wantSum {
				t.Errorf("calculateTermFrequencyStats() gotSum = %v, want %v", gotSum, tt.wantSum)
			}
			if gotMin != tt.wantMin {
				t.Errorf("calculateTermFrequencyStats() gotMin = %v, want %v", gotMin, tt.wantMin)
			}
			if gotMax != tt.wantMax {
				t.Errorf("calculateTermFrequencyStats() gotMax = %v, want %v", gotMax, tt.wantMax)
			}
			if gotMean != tt.wantMean {
				t.Errorf("calculateTermFrequencyStats() gotMean = %v, want %v", gotMean, tt.wantMean)
			}
			if gotVariance != tt.wantVariance {
				t.Errorf("calculateTermFrequencyStats() gotVariance = %v, want %v", gotVariance, tt.wantVariance)
			}
		})
	}
}

func Test_calculateNormalizedTFStats(t *testing.T) {
	type args struct {
		query           Query
		termFrequencies map[string]int
		docLength       int
	}
	tests := []struct {
		name         string
		args         args
		wantSum      float64
		wantMin      float64
		wantMax      float64
		wantMean     float64
		wantVariance float64
	}{
		{
			name: "Basic Case",
			args: args{
				query: Query{
					Id:    "q1",
					Text:  "basic query",
					Terms: []string{"term1", "term2", "term3"},
				},
				termFrequencies: map[string]int{
					"term1": 2,
					"term2": 3,
					"term3": 5,
				},
				docLength: 10,
			},
			wantSum:      1.0,                  // Normalized TF for each term
			wantMin:      0.2,                  // 2/10
			wantMax:      0.5,                  // 5/10
			wantMean:     0.3333333333333333,   // (0.2 + 0.3 + 0.5) / 3
			wantVariance: 0.015555555555555553, // Variance calculation
		},
		{
			name: "Edge Case - Zero Document Length",
			args: args{
				query: Query{
					Id:    "q2",
					Text:  "empty document",
					Terms: []string{"term1", "term2", "term3"},
				},
				termFrequencies: map[string]int{
					"term1": 1,
					"term2": 2,
					"term3": 3,
				},
				docLength: 0,
			},
			wantSum:      0,
			wantMin:      0,
			wantMax:      0,
			wantMean:     0,
			wantVariance: 0,
		},
		{
			name: "Edge Case - No Matching Terms",
			args: args{
				query: Query{
					Id:    "q3",
					Text:  "no match",
					Terms: []string{"termX", "termY"},
				},
				termFrequencies: map[string]int{
					"term1": 1,
					"term2": 2,
					"term3": 3,
				},
				docLength: 10,
			},
			wantSum:      0,
			wantMin:      0,
			wantMax:      0,
			wantMean:     0,
			wantVariance: 0,
		},
		{
			name: "Edge Case - Single Term Query",
			args: args{
				query: Query{
					Id:    "q4",
					Text:  "single term",
					Terms: []string{"term1"},
				},
				termFrequencies: map[string]int{
					"term1": 3,
				},
				docLength: 5,
			},
			wantSum:      0.6, // Normalized TF for the single term
			wantMin:      0.6, // Only one term, so min = max
			wantMax:      0.6, // Only one term, so min = max
			wantMean:     0.6, // Only one term, so mean = TF
			wantVariance: 0,   // Variance is 0 for a single term
		},
		{
			name: "Multiple Terms with Same Frequency",
			args: args{
				query: Query{
					Id:    "q5",
					Text:  "same frequency",
					Terms: []string{"term1", "term2"},
				},
				termFrequencies: map[string]int{
					"term1": 4,
					"term2": 4,
				},
				docLength: 10,
			},
			wantSum:      0.8, // Both terms have normalized TF of 0.4
			wantMin:      0.4, // Min = Max = 0.4
			wantMax:      0.4, // Min = Max = 0.4
			wantMean:     0.4, // (0.4 + 0.4) / 2
			wantVariance: 0,   // Variance is 0 for identical terms
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSum, gotMin, gotMax, gotMean, gotVariance := calculateNormalizedTFStats(tt.args.query, tt.args.termFrequencies, tt.args.docLength)
			if gotSum != tt.wantSum {
				t.Errorf("calculateNormalizedTFStats() gotSum = %v, want %v", gotSum, tt.wantSum)
			}
			if gotMin != tt.wantMin {
				t.Errorf("calculateNormalizedTFStats() gotMin = %v, want %v", gotMin, tt.wantMin)
			}
			if gotMax != tt.wantMax {
				t.Errorf("calculateNormalizedTFStats() gotMax = %v, want %v", gotMax, tt.wantMax)
			}
			if gotMean != tt.wantMean {
				t.Errorf("calculateNormalizedTFStats() gotMean = %v, want %v", gotMean, tt.wantMean)
			}
			if gotVariance != tt.wantVariance {
				t.Errorf("calculateNormalizedTFStats() gotVariance = %v, want %v", gotVariance, tt.wantVariance)
			}
		})
	}
}

func Test_calculateBM25(t *testing.T) {
	type args struct {
		query           Query
		termFrequencies map[string]int
		idf             map[string]float64
		docLength       int
		avgDocLength    float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "Basic Case",
			args: args{
				query: Query{
					Terms: []string{"term1"},
				},
				termFrequencies: map[string]int{
					"term1": 3,
				},
				idf: map[string]float64{
					"term1": 1.2,
				},
				docLength:    100,
				avgDocLength: 120.0,
			},
			want: 1.2 * ((3 * (k1 + 1)) / (3 + k1*((1-b)+b*(100.0/120.0)))),
		},
		{
			name: "Multiple Terms",
			args: args{
				query: Query{
					Terms: []string{"term1", "term2"},
				},
				termFrequencies: map[string]int{
					"term1": 3,
					"term2": 2,
				},
				idf: map[string]float64{
					"term1": 1.2,
					"term2": 1.5,
				},
				docLength:    100,
				avgDocLength: 120,
			},
			want: 1.2*((3*(k1+1))/(3+k1*((1-b)+b*(100.0/120.0)))) +
				1.5*((2*(k1+1))/(2+k1*((1-b)+b*(100.0/120.0)))),
		},
		{
			name: "No IDF for Term",
			args: args{
				query: Query{
					Terms: []string{"term1", "term2"},
				},
				termFrequencies: map[string]int{
					"term1": 3,
					"term2": 2,
				},
				idf: map[string]float64{
					"term1": 1.2, // term2 has no IDF
				},
				docLength:    100,
				avgDocLength: 120,
			},
			want: 1.2 * ((3 * (k1 + 1)) / (3 + k1*((1-b)+b*(100.0/120.0)))),
		},
		{
			name: "Zero Term Frequency",
			args: args{
				query: Query{
					Terms: []string{"term1", "term2"},
				},
				termFrequencies: map[string]int{
					"term1": 0,
					"term2": 2,
				},
				idf: map[string]float64{
					"term1": 1.2,
					"term2": 1.5,
				},
				docLength:    100,
				avgDocLength: 120,
			},
			want: 1.5 * ((2 * (k1 + 1)) / (2 + k1*((1-b)+b*(100.0/120.0)))),
		},
		{
			name: "Edge Case with Long Document",
			args: args{
				query: Query{
					Terms: []string{"term1"},
				},
				termFrequencies: map[string]int{
					"term1": 10,
				},
				idf: map[string]float64{
					"term1": 1.2,
				},
				docLength:    1000,
				avgDocLength: 500,
			},
			want: 1.2 * ((10 * (k1 + 1)) / (10 + k1*((1-b)+b*(1000.0/500.0)))),
		},
		{
			name: "Duplicate Query Terms",
			args: args{
				query: Query{
					Terms: []string{"term1", "term1"},
				},
				termFrequencies: map[string]int{
					"term1": 10,
				},
				idf: map[string]float64{
					"term1": 1.2,
				},
				docLength:    1000,
				avgDocLength: 500,
			},
			want: 2 * 1.2 * ((10 * (k1 + 1)) / (10 + k1*((1-b)+b*(1000.0/500.0)))),
		},
		{
			name: "No Term Frequency in Query",
			args: args{
				query: Query{
					Terms: []string{"term1", "term3"},
				},
				termFrequencies: map[string]int{
					"term1": 3,
					"term2": 2, // term3 has no tf
				},
				idf: map[string]float64{
					"term1": 1.2,
					"term3": 0.8,
				},
				docLength:    100,
				avgDocLength: 120,
			},
			want: 1.2 * ((3 * (k1 + 1)) / (3 + k1*((1-b)+b*(100.0/120.0)))),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateBM25(tt.args.query, tt.args.termFrequencies, tt.args.idf, tt.args.docLength, tt.args.avgDocLength)
			if diff := math.Abs(got - tt.want); diff > epsilon {
				t.Errorf("calculateBM25() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculateIDFMetrics(t *testing.T) {
	type args struct {
		query           Query
		termFrequencies map[string]int
		idf             map[string]float64
	}
	tests := []struct {
		name         string
		args         args
		wantSum      float64
		wantMin      float64
		wantMax      float64
		wantMean     float64
		wantVariance float64
	}{
		{
			name: "Basic Case",
			args: args{
				query: Query{
					Id:    "q1",
					Text:  "basic query",
					Terms: []string{"term1", "term2", "term3"},
				},
				termFrequencies: map[string]int{
					"term1": 2,
					"term2": 3,
					"term3": 5,
				},
				idf: map[string]float64{
					"term1": 1.0,
					"term2": 0.5,
					"term3": 2.0,
				},
			},
			wantSum:      2.0*1.0 + 3.0*0.5 + 5.0*2.0,
			wantMin:      3.0 * 0.5,
			wantMax:      5.0 * 2.0,
			wantMean:     (2.0*1.0 + 3.0*0.5 + 5.0*2.0) / 3.0,
			wantVariance: ((2.0-4.5)*(2.0-4.5) + (1.5-4.5)*(1.5-4.5) + (10.0-4.5)*(10.0-4.5)) / 3.0,
		},
		{
			name: "No Matching Terms",
			args: args{
				query: Query{
					Id:    "q2",
					Text:  "unmatched query",
					Terms: []string{"term4"},
				},
				termFrequencies: map[string]int{
					"term1": 2,
					"term2": 3,
					"term3": 5,
				},
				idf: map[string]float64{
					"term1": 1.0,
					"term2": 0.5,
					"term3": 2.0,
				},
			},
			wantSum:      0.0, // No matching terms
			wantMin:      0.0,
			wantMax:      0.0,
			wantMean:     0.0,
			wantVariance: 0.0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSum, gotMin, gotMax, gotMean, gotVariance := calculateIDFMetrics(tt.args.query, tt.args.termFrequencies, tt.args.idf)
			if gotSum != tt.wantSum {
				t.Errorf("calculateIDFMetrics() gotSum = %v, want %v", gotSum, tt.wantSum)
			}
			if gotMin != tt.wantMin {
				t.Errorf("calculateIDFMetrics() gotMin = %v, want %v", gotMin, tt.wantMin)
			}
			if gotMax != tt.wantMax {
				t.Errorf("calculateIDFMetrics() gotMax = %v, want %v", gotMax, tt.wantMax)
			}
			if gotMean != tt.wantMean {
				t.Errorf("calculateIDFMetrics() gotMean = %v, want %v", gotMean, tt.wantMean)
			}
			if gotVariance != tt.wantVariance {
				t.Errorf("calculateIDFMetrics() gotVariance = %v, want %v", gotVariance, tt.wantVariance)
			}
		})
	}
}

func Test_analyzeURL(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name           string
		args           args
		wantNumSlashes int
		wantLength     int
	}{
		{
			name: "Simple URL",
			args: args{
				url: "https://example.com",
			},
			wantNumSlashes: 2,
			wantLength:     19,
		},
		{
			name: "URL with path",
			args: args{
				url: "https://example.com/path/to/resource",
			},
			wantNumSlashes: 5,
			wantLength:     36,
		},
		{
			name: "URL with query parameters",
			args: args{
				url: "https://example.com/path?query=1",
			},
			wantNumSlashes: 3,
			wantLength:     32,
		},
		{
			name: "Root URL",
			args: args{
				url: "https://example.com/",
			},
			wantNumSlashes: 3,
			wantLength:     20,
		},
		{
			name: "Empty URL",
			args: args{
				url: "",
			},
			wantNumSlashes: 0,
			wantLength:     0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNumSlashes, gotLength := analyzeURL(tt.args.url)
			if gotNumSlashes != tt.wantNumSlashes {
				t.Errorf("analyzeURL() gotNumSlashes = %v, want %v", gotNumSlashes, tt.wantNumSlashes)
			}
			if gotLength != tt.wantLength {
				t.Errorf("analyzeURL() gotLength = %v, want %v", gotLength, tt.wantLength)
			}
		})
	}
}

func TestDocument_calculateFeatures(t *testing.T) {
	type fields struct {
		DocID           string
		Rank            int
		Metadata        DocumentMetadata
		TermFrequencies map[string]int
		Features        Features
	}
	type args struct {
		query        Query
		idf          map[string]float64
		avgDocLength float64
		client       *http.Client
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    Features
	}{
		{
			name: "Basic case",
			fields: fields{
				Metadata: DocumentMetadata{
					DocLength: 100,
					URL:       "http://example.com",
				},
				TermFrequencies: map[string]int{
					"term1": 2,
					"term2": 10,
				},
			},
			args: args{
				query: Query{
					Terms: []string{"term1", "term2", "term3"},
				},
				idf: map[string]float64{
					"term1": 1.0,
					"term2": 0.5,
					"term3": 2.0,
				},
				avgDocLength: 120.0,
				client: createMockHTTPClient(
					map[string]string{
						PagerankEndpoint + "https://example.com": `{
							"pageRank": 0.85,
							"inLinkCount": 123,
							"outLinkCount": 45
						}`,
					},
					map[string]error{}, // No errors
					http.StatusOK,      // Status OK
				),
			},
			wantErr: false,
			want: Features{
				CoveredQueryTermNumber:           2,
				CoveredQueryTermRatio:            2.0 / 3.0,
				SumTermFrequency:                 12,
				MinTermFrequency:                 0.0,
				MaxTermFrequency:                 10,
				MeanTermFrequency:                4.0,
				VarianceTermFrequency:            stat.PopVariance([]float64{2.0, 10.0, 0.0}, nil),
				StreamLength:                     100,
				SumStreamLengthNormalizedTF:      (2.0 / 100.0) + (10.0 / 100.0),
				MinStreamLengthNormalizedTF:      0.0,
				MaxStreamLengthNormalizedTF:      10.0 / 100.0,
				MeanStreamLengthNormalizedTF:     4.0 / 100.0,
				VarianceStreamLengthNormalizedTF: stat.PopVariance([]float64{0.02, 0.1, 0.0}, nil),
				SumTFIDF:                         2.0*1.0 + 10.0*0.5,
				MinTFIDF:                         0.0,
				MaxTFIDF:                         5.0,
				MeanTFIDF:                        (2.0*1.0 + 10.0*0.5) / 3.0,
				VarianceTFIDF:                    stat.PopVariance([]float64{2.0, 5.0, 0.0}, nil),
				BM25:                             1.0*((2*(k1+1))/(2+k1*((1-b)+b*(100.0/120.0)))) + 0.5*((10*(k1+1))/(10+k1*((1-b)+b*(100.0/120.0)))),
				NumSlashesInURL:                  2,
				LengthOfURL:                      19,
				InlinkCount:                      123,
				OutlinkCount:                     45,
				PageRank:                         0.85,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := &Document{
				DocID:           tt.fields.DocID,
				Rank:            tt.fields.Rank,
				Metadata:        tt.fields.Metadata,
				TermFrequencies: tt.fields.TermFrequencies,
				Features:        tt.fields.Features,
			}
			if err := doc.calculateFeatures(tt.args.query, tt.args.idf, tt.args.avgDocLength, tt.args.client); (err != nil) != tt.wantErr {
				t.Errorf("Document.calculateFeatures() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !compareFeatures(doc.Features, tt.want, epsilon) {
				t.Errorf("Document.calculateFeatures() doc.Features = %+v, want %+v", doc.Features, tt.want)
			}
		})
	}
}

func TestDocuments_initializeFeatures(t *testing.T) {
	type args struct {
		query         Query
		docStatistics totalDocStatistics
		index         invertibleIndex
		client        *http.Client
	}
	tests := []struct {
		name    string
		docs    *Documents
		args    args
		wantErr bool
		want    *Documents
	}{
		{
			name: "Basic case",
			docs: &Documents{
				{
					DocID: "doc1",
					TermFrequencies: map[string]int{
						"term1": 2,
						"term2": 10,
					},
				},
			},
			args: args{
				query: Query{
					Terms: []string{"term1", "term2", "term3"},
				},
				docStatistics: totalDocStatistics{
					AvgDocLength: 120.0,
					DocCount:     5,
				},
				index: invertibleIndex{
					"term1": {
						{DocID: "doc1", Frequency: 2, Positions: []int{1, 2}},
						{DocID: "doc2", Frequency: 1, Positions: []int{1}},
					},
					"term2": {
						{DocID: "doc1", Frequency: 1, Positions: []int{1}},
					},
					"term3": {
						{DocID: "doc1", Frequency: 1, Positions: []int{1}},
					},
				},
				client: createMockHTTPClient(
					map[string]string{
						PagerankEndpoint + "http://example.com": `{
												"pageRank": 0.85,
												"inLinkCount": 123,
												"outLinkCount": 45
											}`,
						MetadataEndpoint + "doc1": `{
												"docID": "12345",
												"metadata": {
													"docLength": 100,
													"timeLastUpdated": "2024-11-09T15:30:00Z",
													"docType": "PDF",
													"imageCount": 3,
													"docTitle": "Introduction to Data Science",
													"URL": "http://example.com"
												}
											}`,
					},
					map[string]error{}, // No errors
					http.StatusOK,      // Status OK
				),
			},
			wantErr: false,
			want: &Documents{
				{
					DocID: "doc1",
					TermFrequencies: map[string]int{
						"term1": 2,
						"term2": 10,
					},
					Features: Features{
						CoveredQueryTermNumber:           2,
						CoveredQueryTermRatio:            2.0 / 3.0,
						SumTermFrequency:                 12,
						MinTermFrequency:                 0.0,
						MaxTermFrequency:                 10,
						MeanTermFrequency:                4.0,
						VarianceTermFrequency:            stat.PopVariance([]float64{2.0, 10.0, 0.0}, nil),
						StreamLength:                     100,
						SumStreamLengthNormalizedTF:      (2.0 / 100.0) + (10.0 / 100.0),
						MinStreamLengthNormalizedTF:      0.0,
						MaxStreamLengthNormalizedTF:      10.0 / 100.0,
						MeanStreamLengthNormalizedTF:     4.0 / 100.0,
						VarianceStreamLengthNormalizedTF: stat.PopVariance([]float64{0.02, 0.1, 0.0}, nil),
						SumTFIDF:                         2.0*math.Log(5.0/(2.0+1.0)) + 10.0*math.Log(5.0/(1.0+1.0)),
						MinTFIDF:                         0.0,
						MaxTFIDF:                         max(2.0*math.Log(5.0/(2.0+1.0)), 10.0*math.Log(5.0/(1.0+1.0))),
						MeanTFIDF:                        (2.0*math.Log(5.0/(2.0+1.0)) + 10.0*math.Log(5.0/(1.0+1.0))) / 3.0,
						VarianceTFIDF:                    stat.PopVariance([]float64{2.0 * math.Log(5.0/(2.0+1.0)), 10.0 * math.Log(5.0/(1.0+1.0)), 0.0}, nil),
						BM25:                             math.Log(5.0/(2.0+1.0))*((2*(k1+1))/(2+k1*((1-b)+b*(100.0/120.0)))) + math.Log(5.0/(1.0+1.0))*((10*(k1+1))/(10+k1*((1-b)+b*(100.0/120.0)))),
						NumSlashesInURL:                  2,
						LengthOfURL:                      18,
						InlinkCount:                      123,
						OutlinkCount:                     45,
						PageRank:                         0.85,
					},
					Metadata: DocumentMetadata{
						DocLength:       100,
						TimeLastUpdated: "2024-11-09T15:30:00Z",
						FileType:        "PDF",
						ImageCount:      3,
						DocTitle:        "Introduction to Data Science",
						URL:             "http://example.com",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.docs.initializeFeatures(tt.args.query, tt.args.docStatistics, tt.args.index, tt.args.client); (err != nil) != tt.wantErr {
				t.Errorf("Documents.initializeFeatures() error = %v, wantErr %v", err, tt.wantErr)
			}

			docs := tt.docs

			for i, doc := range *docs {
				gotFeatures := doc.Features
				wantFeatures := (*tt.want)[i].Features
				if !compareFeatures(gotFeatures, wantFeatures, 1e-9) {
					t.Errorf("Features mismatch for doc %s: got %v, want %v", doc.DocID, gotFeatures, wantFeatures)
				}

				gotMetadata := doc.Metadata
				wantMetadata := (*tt.want)[i].Metadata
				if !reflect.DeepEqual(gotMetadata, wantMetadata) {
					t.Errorf("DocumentMetadata mismatch for doc %s: got %v, want %v", doc.DocID, gotMetadata, wantMetadata)
				}
			}
		})
	}
}

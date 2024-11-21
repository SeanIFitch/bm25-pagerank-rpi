package ranking

type DocData struct {
	DocID     string `json:"docID"`
	Frequency int    `json:"frequency"`
	Positions []int  `json:"positions"`
}

type InvertibleIndex struct {
	Term  string    `json:"term"`
	Index []DocData `json:"index"`
}

func GetBM25(k1 float64, b float64, inverIdx InvertibleIndex) map[string]int {
	return make(map[string]int)
}

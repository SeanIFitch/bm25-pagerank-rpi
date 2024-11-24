package training

import (
	"math/rand"
	"rpi-search-ranking/internal/ranking"
)

// ShuffleData Helper function to shuffle the data
func ShuffleData(X []ranking.Features, Y []int) {
	for i := len(X) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		X[i], X[j] = X[j], X[i]
		Y[i], Y[j] = Y[j], Y[i]
	}
}

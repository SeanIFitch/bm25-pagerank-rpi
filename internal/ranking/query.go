package ranking

import (
	"strings"
	"unicode"
)

// Terms returns a slice of terms (tokens) extracted from the query's text.
// It processes the text by lowercasing and removing punctuation.
func (q Query) Terms() []string {
	// Convert the query text to lowercase
	text := strings.ToLower(q.Text)

	// Create a slice to hold the terms
	var terms []string

	// Iterate over the text and extract terms (words)
	var currentTerm []rune
	for _, char := range text {
		// If the character is alphanumeric, add it to the current term
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			currentTerm = append(currentTerm, char)
		} else if len(currentTerm) > 0 {
			// If we reach a non-alphanumeric character, save the current term and reset
			terms = append(terms, string(currentTerm))
			currentTerm = nil
		}
	}

	// If the last term was being formed, add it
	if len(currentTerm) > 0 {
		terms = append(terms, string(currentTerm))
	}

	return terms
}

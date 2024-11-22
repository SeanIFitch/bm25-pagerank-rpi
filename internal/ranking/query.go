package ranking

import (
	"strings"
)

func (q *Query) Tokenize() {
	q.Terms = strings.Fields(q.Text)
}

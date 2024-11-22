package ranking

import (
	"strings"
)

func (q *Query) tokenize() {
	q.Terms = strings.Fields(q.Text)
}

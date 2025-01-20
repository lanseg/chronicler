package pikabu

import (
	"encoding/json"
	"fmt"
	"sort"
)

type tree struct {
	Min  int
	Tree []interface{}
}

func (t *tree) walk() []string {
	parts := map[int]bool{}
	toVisit := append([](interface{}){}, t.Tree...)
	for {
		tr := toVisit[0]
		toVisit = toVisit[1:]
		switch t := tr.(type) {
		case []interface{}:
			toVisit = append(toVisit, t...)
		case float64:
			parts[int(t)] = true
		}
		if len(toVisit) == 0 {
			break
		}
	}
	commentIds := []string{}
	for k := range parts {
		commentIds = append(commentIds, fmt.Sprintf("%d", k+t.Min))
	}
	sort.Strings(commentIds)
	return commentIds
}

func ResolveCommentTree(commentTree string) ([]string, error) {
	t := &tree{}
	if err := json.Unmarshal([]byte(commentTree), t); err != nil {
		return []string{}, err
	}
	return t.walk(), nil
}

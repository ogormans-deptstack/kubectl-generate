package fuzzy

import (
	"sort"
	"strings"
)

func Distance(a, b string) int {
	a = strings.ToLower(a)
	b = strings.ToLower(b)

	if a == b {
		return 0
	}

	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	prev := make([]int, lb+1)
	curr := make([]int, lb+1)

	for j := 0; j <= lb; j++ {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min3(
				prev[j]+1,
				curr[j-1]+1,
				prev[j-1]+cost,
			)
		}
		prev, curr = curr, prev
	}

	return prev[lb]
}

func Suggest(input string, candidates []string, maxResults int) []string {
	if input == "" || len(candidates) == 0 {
		return nil
	}

	inputLen := len(input)
	threshold := inputLen / 2
	if threshold < 3 {
		threshold = 3
	}

	type match struct {
		name     string
		distance int
	}

	var matches []match
	for _, c := range candidates {
		d := Distance(input, c)
		if d <= threshold {
			matches = append(matches, match{name: c, distance: d})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].distance != matches[j].distance {
			return matches[i].distance < matches[j].distance
		}
		return matches[i].name < matches[j].name
	})

	if len(matches) > maxResults {
		matches = matches[:maxResults]
	}

	result := make([]string, len(matches))
	for i, m := range matches {
		result[i] = m.name
	}
	return result
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

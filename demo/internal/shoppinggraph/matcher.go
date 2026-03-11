package shoppinggraph

import "strings"

var stopwords = map[string]bool{
	"the": true, "a": true, "an": true, "and": true, "or": true,
	"of": true, "for": true, "with": true, "in": true, "on": true,
}

// Tokenize splits a title into lowercase tokens, removing stopwords.
func Tokenize(title string) map[string]bool {
	words := strings.Fields(strings.ToLower(title))
	tokens := make(map[string]bool, len(words))
	for _, w := range words {
		w = strings.Trim(w, ".,!?()-")
		if w != "" && !stopwords[w] {
			tokens[w] = true
		}
	}
	return tokens
}

// JaccardSimilarity computes |A∩B| / |A∪B|.
func JaccardSimilarity(a, b map[string]bool) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 0
	}
	intersection := 0
	for k := range a {
		if b[k] {
			intersection++
		}
	}
	union := len(a) + len(b) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

// GroupProducts clusters products with Jaccard similarity >= 0.3.
func GroupProducts(products []*ProductNode) []*ProductGroup {
	assigned := make([]bool, len(products))
	var groups []*ProductGroup

	for i, p := range products {
		if assigned[i] {
			continue
		}
		if p.Tokens == nil {
			p.Tokens = Tokenize(p.Title)
		}
		group := &ProductGroup{
			CanonicalName: p.Title,
			Offers:        []*ProductNode{p},
		}
		assigned[i] = true

		for j := i + 1; j < len(products); j++ {
			if assigned[j] {
				continue
			}
			other := products[j]
			if other.Tokens == nil {
				other.Tokens = Tokenize(other.Title)
			}
			if JaccardSimilarity(p.Tokens, other.Tokens) >= 0.3 {
				group.Offers = append(group.Offers, other)
				assigned[j] = true
			}
		}
		groups = append(groups, group)
	}
	return groups
}

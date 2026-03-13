package shoppinggraph

import (
	"fmt"
	"math"
	"sort"
)

// SearchRequest is the input for a product search.
type SearchRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

// Search finds products matching a query, ranked by relevance and merchant score.
func (g *ShoppingGraph) Search(query string, limit int) []SearchResult {
	if limit <= 0 {
		limit = 10
	}

	queryTokens := Tokenize(query)
	if len(queryTokens) == 0 {
		return nil
	}

	g.mu.RLock()
	algo := g.RankAlgo
	defer g.mu.RUnlock()

	type scored struct {
		product *ProductNode
		score   float64
	}

	var candidates []scored
	for _, p := range g.Products {
		sim := JaccardSimilarity(queryTokens, p.Tokens)
		if sim < 0.1 {
			continue
		}
		merchantScore := 1.0
		if m, ok := g.Merchants[p.MerchantID]; ok {
			merchantScore = float64(m.Score) / 100.0
			if !m.Online {
				continue
			}
		}
		stockBoost := 1.0
		if p.Quantity > 0 {
			stockBoost = 1.5
		}

		var s float64
		switch algo {
		case RankJaccardPrice:
			s = sim * merchantScore * stockBoost * (1.0 / math.Log2(float64(p.Price)+2))
		case RankPriceOnly:
			// Use negative price so higher score = lower price
			s = -float64(p.Price)
		default: // RankJaccard
			s = sim * merchantScore * stockBoost
		}
		candidates = append(candidates, scored{
			product: p,
			score:   s,
		})
	}

	// Fallback: if no Jaccard matches, return all in-stock products
	if len(candidates) == 0 {
		for _, p := range g.Products {
			if m, ok := g.Merchants[p.MerchantID]; ok && !m.Online {
				continue
			}
			if p.Quantity > 0 {
				candidates = append(candidates, scored{product: p, score: 1.0})
			}
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	if len(candidates) > limit {
		candidates = candidates[:limit]
	}

	results := make([]SearchResult, len(candidates))
	for i, c := range candidates {
		var hints []string
		if m, ok := g.Merchants[c.product.MerchantID]; ok {
			hints = m.DiscountHints
		}
		results[i] = SearchResult{
			Rank:          i + 1,
			ProductID:     c.product.ProductID,
			Title:         c.product.Title,
			MerchantID:    c.product.MerchantID,
			MerchantName:  c.product.MerchantName,
			MerchantURL:   c.product.MerchantURL,
			Price:         c.product.Price,
			PriceDisplay:  fmt.Sprintf("$%.2f", float64(c.product.Price)/100),
			InStock:       c.product.Quantity > 0,
			DiscountHints: hints,
		}
	}
	return results
}

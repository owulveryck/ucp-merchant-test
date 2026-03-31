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

// scoreDefault computes a balanced quality score (range ~0-10).
// Relevance 3 + Price 4 + Stock 2 + Bid 1.
func scoreDefault(jaccardSim float64, price int, avgPrice float64, inStock bool, bid int) float64 {
	relevance := jaccardSim * 3.0
	pricePoints := 0.0
	if price > 0 {
		pricePoints = math.Min(4.0, 4.0*avgPrice/float64(price))
	}
	stockPoints := 0.0
	if inStock {
		stockPoints = 2.0
	}
	bidPoints := math.Min(1.0, float64(bid)/100.0)
	return relevance + pricePoints + stockPoints + bidPoints
}

// scoreJaccardPrice computes a price-dominant score with relevance tiebreaker (range ~0-10).
// Price 6 + Relevance 2 + Stock 2.
func scoreJaccardPrice(jaccardSim float64, price int, avgPrice float64, inStock bool) float64 {
	pricePoints := 0.0
	if price > 0 {
		pricePoints = math.Min(6.0, 6.0*avgPrice/float64(price))
	}
	relevance := jaccardSim * 2.0
	stockPoints := 0.0
	if inStock {
		stockPoints = 2.0
	}
	return relevance + pricePoints + stockPoints
}

// scorePriceOnly computes a pure price ranking with stock bonus (range ~0-10).
// Price 8 + Stock 2.
func scorePriceOnly(price int, avgPrice float64, inStock bool) float64 {
	pricePoints := 0.0
	if price > 0 {
		pricePoints = math.Min(8.0, 8.0*avgPrice/float64(price))
	}
	stockPoints := 0.0
	if inStock {
		stockPoints = 2.0
	}
	return pricePoints + stockPoints
}

// Search finds products matching a query using a Google Ads-style auction model.
// Sponsored results (merchants with bid > 0) are ranked by Ad Rank (bid * quality),
// priced via second-price auction, and returned first.
// Organic results (all candidates) are ranked by quality score only.
func (g *ShoppingGraph) Search(query string, limit int) []SearchResult {
	if limit <= 0 {
		limit = 10
	}

	queryTokens := Tokenize(query)
	if len(queryTokens) == 0 {
		return nil
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	type candidate struct {
		product      *ProductNode
		merchant     *MerchantNode
		jaccardSim   float64
		qualityScore float64
		adRank       float64
		bid          int
		sponsored    bool
		hints        []string
	}

	// Step 1: Find candidates with Jaccard >= 0.1, exclude offline
	var candidates []candidate
	for _, p := range g.Products {
		sim := JaccardSimilarity(queryTokens, p.Tokens)
		if sim < 0.1 {
			continue
		}
		m, ok := g.Merchants[p.MerchantID]
		if !ok || !m.Online {
			continue
		}
		candidates = append(candidates, candidate{
			product:    p,
			merchant:   m,
			jaccardSim: sim,
			bid:        m.MaxCPCBid,
			sponsored:  m.MaxCPCBid > 0,
			hints:      m.DiscountHints,
		})
	}

	// Fallback: if no Jaccard matches, return all in-stock products
	if len(candidates) == 0 {
		for _, p := range g.Products {
			m, ok := g.Merchants[p.MerchantID]
			if !ok || !m.Online {
				continue
			}
			if p.Quantity > 0 {
				candidates = append(candidates, candidate{
					product:    p,
					merchant:   m,
					jaccardSim: 1.0,
					bid:        m.MaxCPCBid,
					sponsored:  m.MaxCPCBid > 0,
					hints:      m.DiscountHints,
				})
			}
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	// Step 2: Compute average price
	totalPrice := 0
	for _, c := range candidates {
		totalPrice += c.product.Price
	}
	avgPrice := float64(totalPrice) / float64(len(candidates))

	// Step 3: Compute quality scores based on selected ranking algorithm
	algo := g.RankAlgo
	for i := range candidates {
		c := &candidates[i]
		switch algo {
		case RankJaccardPrice:
			c.qualityScore = scoreJaccardPrice(c.jaccardSim, c.product.Price, avgPrice, c.product.Quantity > 0)
		case RankPriceOnly:
			c.qualityScore = scorePriceOnly(c.product.Price, avgPrice, c.product.Quantity > 0)
		default: // RankJaccard
			c.qualityScore = scoreDefault(c.jaccardSim, c.product.Price, avgPrice, c.product.Quantity > 0, c.bid)
		}
		if c.sponsored {
			c.adRank = float64(c.bid) * c.qualityScore
		}
	}

	// Step 4: Separate sponsored and organic
	var sponsored []candidate
	var organic []candidate
	for _, c := range candidates {
		if c.sponsored {
			sponsored = append(sponsored, c)
		}
		// All candidates appear in organic (including sponsored ones)
		organic = append(organic, c)
	}

	// Sort sponsored by adRank desc, merchant ID as tiebreaker for stability
	sort.Slice(sponsored, func(i, j int) bool {
		if sponsored[i].adRank != sponsored[j].adRank {
			return sponsored[i].adRank > sponsored[j].adRank
		}
		return sponsored[i].merchant.ID < sponsored[j].merchant.ID
	})

	// Second-price auction for CPC
	for i := range sponsored {
		var actualCPC int
		if i == len(sponsored)-1 {
			// Last position: floor price of 1 cent
			actualCPC = 1
		} else {
			// actualCPC = min(bid, ceil(nextAdRank / qualityScore) + 1)
			nextAdRank := sponsored[i+1].adRank
			qs := sponsored[i].qualityScore
			if qs <= 0 {
				qs = 0.1
			}
			computed := int(math.Ceil(nextAdRank/qs)) + 1
			if computed > sponsored[i].bid {
				actualCPC = sponsored[i].bid
			} else {
				actualCPC = computed
			}
		}
		if actualCPC < 1 {
			actualCPC = 1
		}
		// Store LastActualCPC on the merchant node
		sponsored[i].merchant.LastActualCPC = actualCPC
	}

	// Sort organic by qualityScore desc, merchant ID as tiebreaker for stability
	sort.Slice(organic, func(i, j int) bool {
		if organic[i].qualityScore != organic[j].qualityScore {
			return organic[i].qualityScore > organic[j].qualityScore
		}
		return organic[i].merchant.ID < organic[j].merchant.ID
	})

	// Step 5: Merge results: sponsored first, then organic (deduplicated)
	seen := make(map[string]bool)
	var results []SearchResult
	rank := 1

	for _, c := range sponsored {
		key := c.merchant.ID + ":" + c.product.ProductID
		if seen[key] {
			continue
		}
		seen[key] = true
		results = append(results, SearchResult{
			Rank:          rank,
			ProductID:     c.product.ProductID,
			Title:         c.product.Title,
			MerchantID:    c.merchant.ID,
			MerchantName:  c.merchant.Name,
			MerchantURL:   c.merchant.Endpoint,
			Price:         c.product.Price,
			PriceDisplay:  fmt.Sprintf("$%.2f", float64(c.product.Price)/100),
			InStock:       c.product.Quantity > 0,
			DiscountHints: c.hints,
			Sponsored:     true,
			ActualCPC:     c.merchant.LastActualCPC,
			QualityScore:  c.qualityScore,
		})
		rank++
	}

	for _, c := range organic {
		key := c.merchant.ID + ":" + c.product.ProductID
		if seen[key] {
			continue
		}
		seen[key] = true
		results = append(results, SearchResult{
			Rank:          rank,
			ProductID:     c.product.ProductID,
			Title:         c.product.Title,
			MerchantID:    c.merchant.ID,
			MerchantName:  c.merchant.Name,
			MerchantURL:   c.merchant.Endpoint,
			Price:         c.product.Price,
			PriceDisplay:  fmt.Sprintf("$%.2f", float64(c.product.Price)/100),
			InStock:       c.product.Quantity > 0,
			DiscountHints: c.hints,
			Sponsored:     false,
			QualityScore:  c.qualityScore,
		})
		rank++
	}

	if len(results) > limit {
		results = results[:limit]
	}

	return results
}

package shoppinggraph

import (
	"sync"
	"time"
)

// ProductNode represents a product offering from a specific merchant.
type ProductNode struct {
	MerchantID   string          `json:"merchant_id"`
	MerchantName string          `json:"merchant_name"`
	MerchantURL  string          `json:"merchant_url"`
	ProductID    string          `json:"product_id"`
	Title        string          `json:"title"`
	ImageURL     string          `json:"image_url"`
	Price        int             `json:"price"`
	Quantity     int             `json:"quantity"`
	Tokens       map[string]bool `json:"-"`
}

// MerchantNode represents a known merchant.
type MerchantNode struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Endpoint      string    `json:"endpoint"`
	Score         int       `json:"score"`
	Online        bool      `json:"online"`
	LastPoll      time.Time `json:"last_poll"`
	DiscountHints []string  `json:"discount_hints,omitempty"`
}

// ProductGroup is a set of similar products across merchants.
type ProductGroup struct {
	CanonicalName string         `json:"canonical_name"`
	Offers        []*ProductNode `json:"offers"`
}

// SearchResult is a ranked product result from a search query.
type SearchResult struct {
	Rank          int      `json:"rank"`
	ProductID     string   `json:"product_id"`
	Title         string   `json:"title"`
	MerchantID    string   `json:"merchant_id"`
	MerchantName  string   `json:"merchant_name"`
	MerchantURL   string   `json:"merchant_url"`
	Price         int      `json:"price"`
	PriceDisplay  string   `json:"price_display"`
	InStock       bool     `json:"in_stock"`
	DiscountHints []string `json:"discount_hints,omitempty"`
}

// RankingAlgorithm identifies a search ranking strategy.
type RankingAlgorithm string

const (
	RankJaccard      RankingAlgorithm = "jaccard"
	RankJaccardPrice RankingAlgorithm = "jaccard_price"
	RankPriceOnly    RankingAlgorithm = "price"
)

// ShoppingGraph maintains an index of products across merchants.
type ShoppingGraph struct {
	mu          sync.RWMutex
	Products    []*ProductNode
	Merchants   map[string]*MerchantNode
	Groups      []*ProductGroup
	LastUpdated time.Time
	RankAlgo    RankingAlgorithm
}

// NewShoppingGraph creates an empty shopping graph.
func NewShoppingGraph() *ShoppingGraph {
	return &ShoppingGraph{
		Merchants: make(map[string]*MerchantNode),
		RankAlgo:  RankJaccard,
	}
}

// SetRankAlgo changes the ranking algorithm used for search.
func (g *ShoppingGraph) SetRankAlgo(algo RankingAlgorithm) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.RankAlgo = algo
}

// GetRankAlgo returns the current ranking algorithm.
func (g *ShoppingGraph) GetRankAlgo() RankingAlgorithm {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.RankAlgo
}

// UpdateMerchantProducts replaces all products for a merchant.
func (g *ShoppingGraph) UpdateMerchantProducts(merchantID string, products []*ProductNode) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Remove old products for this merchant
	var kept []*ProductNode
	for _, p := range g.Products {
		if p.MerchantID != merchantID {
			kept = append(kept, p)
		}
	}
	// Add new
	kept = append(kept, products...)
	g.Products = kept

	// Tokenize
	for _, p := range products {
		p.Tokens = Tokenize(p.Title)
	}

	g.Groups = GroupProducts(g.Products)
	g.LastUpdated = time.Now()

	if m, ok := g.Merchants[merchantID]; ok {
		m.Online = true
		m.LastPoll = time.Now()
	}
}

// AddMerchant registers a new merchant in the graph.
func (g *ShoppingGraph) AddMerchant(node *MerchantNode) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Merchants[node.ID] = node
}

// RemoveMerchant removes a merchant and its products from the graph.
func (g *ShoppingGraph) RemoveMerchant(id string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.Merchants, id)
	var kept []*ProductNode
	for _, p := range g.Products {
		if p.MerchantID != id {
			kept = append(kept, p)
		}
	}
	g.Products = kept
	g.Groups = GroupProducts(g.Products)
}

// SetBoost updates the boost score for a merchant.
func (g *ShoppingGraph) SetBoost(merchantID string, boost int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if m, ok := g.Merchants[merchantID]; ok {
		m.Score = boost
	}
}

// MarkOffline marks a merchant as offline.
func (g *ShoppingGraph) MarkOffline(merchantID string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if m, ok := g.Merchants[merchantID]; ok {
		m.Online = false
		m.LastPoll = time.Now()
	}
}

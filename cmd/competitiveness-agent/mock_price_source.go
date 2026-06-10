package main

import (
	"time"

	compModels "github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive/models"
)

// MockCompetitorPriceSource provides mock competitor prices for standalone mode.
type MockCompetitorPriceSource struct {
	// Mock data: product -> competitors
	mockPrices map[string][]compModels.CompetitorPrice
}

// NewMockCompetitorPriceSource creates a mock price source with sample data.
func NewMockCompetitorPriceSource() *MockCompetitorPriceSource {
	return &MockCompetitorPriceSource{
		mockPrices: map[string][]compModels.CompetitorPrice{
			"laptop": {
				{MerchantID: "competitor_a", MerchantName: "TechStore", Price: 95000, InStock: true},
				{MerchantID: "competitor_b", MerchantName: "BestBuy", Price: 105000, InStock: true},
				{MerchantID: "competitor_c", MerchantName: "Amazon", Price: 98000, InStock: true},
			},
			"mouse": {
				{MerchantID: "competitor_a", MerchantName: "TechStore", Price: 2500, InStock: true},
				{MerchantID: "competitor_b", MerchantName: "BestBuy", Price: 3000, InStock: true},
			},
			"keyboard": {
				{MerchantID: "competitor_a", MerchantName: "TechStore", Price: 7000, InStock: true},
				{MerchantID: "competitor_b", MerchantName: "BestBuy", Price: 7500, InStock: true},
				{MerchantID: "competitor_c", MerchantName: "Amazon", Price: 6800, InStock: true},
			},
			"monitor": {
				{MerchantID: "competitor_a", MerchantName: "TechStore", Price: 35000, InStock: true},
				{MerchantID: "competitor_b", MerchantName: "BestBuy", Price: 38000, InStock: true},
			},
		},
	}
}

// GetLowestPrice returns the lowest competitor price.
func (m *MockCompetitorPriceSource) GetLowestPrice(productID string) (price int, merchantID string, err error) {
	prices, exists := m.mockPrices[productID]
	if !exists || len(prices) == 0 {
		// No competitors found - return a default higher price
		return 999999, "", nil
	}

	lowest := prices[0]
	for _, p := range prices {
		if p.Price < lowest.Price {
			lowest = p
		}
	}

	return lowest.Price, lowest.MerchantID, nil
}

// GetCompetitorPrices returns all competitor prices.
func (m *MockCompetitorPriceSource) GetCompetitorPrices(productID string) ([]compModels.CompetitorPrice, error) {
	prices, exists := m.mockPrices[productID]
	if !exists {
		// No competitors - return empty list
		return []compModels.CompetitorPrice{}, nil
	}

	return prices, nil
}

// AddMockPrice adds a mock price for testing.
func (m *MockCompetitorPriceSource) AddMockPrice(productID string, price compModels.CompetitorPrice) {
	m.mockPrices[productID] = append(m.mockPrices[productID], price)
}

// SetMockPrices sets all mock prices for a product.
func (m *MockCompetitorPriceSource) SetMockPrices(productID string, prices []compModels.CompetitorPrice) {
	m.mockPrices[productID] = prices
}

// MockHistoryStore provides a no-op history store.
type MockHistoryStore struct{}

// NewMockHistoryStore creates a mock history store.
func NewMockHistoryStore() *MockHistoryStore {
	return &MockHistoryStore{}
}

// RecordPrice is a no-op in standalone mode.
func (m *MockHistoryStore) RecordPrice(productID string, price int, timestamp time.Time) error {
	// No-op for standalone mode
	return nil
}

// GetPriceHistory returns empty history.
func (m *MockHistoryStore) GetPriceHistory(productID string, limit int) ([]compModels.PricePoint, error) {
	return []compModels.PricePoint{}, nil
}

// GetTrend returns a neutral trend.
func (m *MockHistoryStore) GetTrend(productID string, duration time.Duration) (compModels.Trend, error) {
	return compModels.Trend{
		Direction:     "stable",
		PercentChange: 0.0,
		Period:        duration,
		DataPoints:    0,
		Volatility:    0.0,
	}, nil
}

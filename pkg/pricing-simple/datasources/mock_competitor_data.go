// Package datasources provides data access for the pricing system.
package datasources

import (
	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-simple/models"
)

// MockCompetitorDataSource provides mock competitor pricing data.
type MockCompetitorDataSource struct {
	prices map[string][]models.CompetitorPrice
}

// NewMockCompetitorData creates a mock competitor data source.
func NewMockCompetitorData() *MockCompetitorDataSource {
	return &MockCompetitorDataSource{
		prices: map[string][]models.CompetitorPrice{
			"headphones": {
				{
					MerchantID:     "competitor_a",
					MerchantName:   "MarchandA",
					Price:          6000,             // $60
					EffectivePrice: 5400,             // $54 after WELCOME10
					DiscountHints:  []string{"WELCOME10"},
					InStock:        true,
				},
				{
					MerchantID:     "competitor_b",
					MerchantName:   "MarchandB",
					Price:          6500, // $65
					EffectivePrice: 6500,
					DiscountHints:  []string{},
					InStock:        true,
				},
				{
					MerchantID:     "competitor_c",
					MerchantName:   "MarchandC",
					Price:          5900, // $59
					EffectivePrice: 5900,
					DiscountHints:  []string{},
					InStock:        true,
				},
			},
			"laptop": {
				{
					MerchantID:     "competitor_a",
					MerchantName:   "MarchandA",
					Price:          80000, // $800
					EffectivePrice: 80000,
					DiscountHints:  []string{},
					InStock:        true,
				},
			},
			"phone": {
				// No competitors
			},
		},
	}
}

// GetCompetitorPrices retrieves competitor prices for a product.
func (m *MockCompetitorDataSource) GetCompetitorPrices(productID string) ([]models.CompetitorPrice, error) {
	if prices, ok := m.prices[productID]; ok {
		return prices, nil
	}
	return []models.CompetitorPrice{}, nil
}

// AddCompetitorPrice adds a competitor price.
func (m *MockCompetitorDataSource) AddCompetitorPrice(productID string, price models.CompetitorPrice) {
	m.prices[productID] = append(m.prices[productID], price)
}

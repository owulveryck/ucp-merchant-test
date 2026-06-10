package competitive

// LegacyShoppingGraphAdapter adapts the new ShoppingGraphClient to the old CompetitorPriceSource interface.
// This is for backward compatibility with code that uses the old monolithic CompetitivePricingAgent.
type LegacyShoppingGraphAdapter struct {
	client *ShoppingGraphClient
}

// NewLegacyShoppingGraphAdapter creates an adapter for backward compatibility.
func NewLegacyShoppingGraphAdapter(baseURL string) *LegacyShoppingGraphAdapter {
	return &LegacyShoppingGraphAdapter{
		client: NewShoppingGraphClient(baseURL),
	}
}

// GetLowestPrice implements the old CompetitorPriceSource interface.
func (a *LegacyShoppingGraphAdapter) GetLowestPrice(productID string) (price int, merchantID string, err error) {
	return a.client.GetLowestPrice(productID)
}

// GetCompetitorPrices implements the old CompetitorPriceSource interface.
func (a *LegacyShoppingGraphAdapter) GetCompetitorPrices(productID string) ([]CompetitorPrice, error) {
	return a.client.GetCompetitorPricesLegacy(productID)
}

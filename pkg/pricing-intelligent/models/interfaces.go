// Package models contains interfaces for intelligent pricing agents.
package models

// LoyaltyAgent analyzes customer value and recommends VIP pricing.
type LoyaltyAgent interface {
	// AnalyzeCustomer analyzes customer profile and determines if VIP pricing applies.
	// Intention: "Ce client mérite-t-il un prix préférentiel ?"
	AnalyzeCustomer(customerID string) (LoyaltyDecision, error)
}

// CompetitivenessAgent analyzes market competition and recommends competitive pricing.
type CompetitivenessAgent interface {
	// AnalyzeMarket analyzes competitor prices and determines optimal competitive price.
	// Intention: "Sommes-nous compétitifs sur ce produit ?"
	AnalyzeMarket(productID string, basePrice int) (CompetitivenessDecision, error)
}

// PricingOrchestrator coordinates agents and makes final pricing decision.
type PricingOrchestrator interface {
	// CalculateOptimalPrice orchestrates agents to calculate optimal price.
	// Intention: "Quel prix optimiser pour maximiser profit et vente ?"
	CalculateOptimalPrice(request PricingRequest) (PricingResult, error)
}

// CustomerDataSource provides customer data for loyalty analysis.
type CustomerDataSource interface {
	// GetCustomerProfile retrieves customer profile data.
	GetCustomerProfile(customerID string) (CustomerProfile, error)
}

// CompetitorDataSource provides competitor pricing data.
type CompetitorDataSource interface {
	// GetCompetitorPrices retrieves competitor prices for a product.
	GetCompetitorPrices(productID string) ([]int, error)
}

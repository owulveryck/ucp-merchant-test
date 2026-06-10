// Package models contains interfaces for the 3-agent pricing system.
package models

// MarketIntelligenceAgent analyzes competitor prices and market position.
// INTENTION: "Sommes-nous compétitifs sur ce produit ?"
type MarketIntelligenceAgent interface {
	Analyze(productID string, ourPrice int) (MarketIntelligenceDecision, error)
}

// LoyaltyAgent analyzes customer value and determines VIP status.
// INTENTION: "Ce client mérite-t-il un prix préférentiel ?"
type LoyaltyAgent interface {
	AnalyzeCustomer(customerID string) (LoyaltyDecision, error)
}

// FinalDecisionAgent orchestrates market + loyalty decisions and validates pricing.
// INTENTION: "Quel prix optimiser pour maximiser profit ET vente ?"
type FinalDecisionAgent interface {
	Decide(
		marketDecision MarketIntelligenceDecision,
		loyaltyDecision LoyaltyDecision,
		request PricingRequest,
	) (FinalPricingDecision, error)
}

// CustomerDataSource provides customer profile data.
type CustomerDataSource interface {
	GetCustomerProfile(customerID string) (CustomerProfile, error)
}

// CompetitorDataSource provides competitor pricing data.
type CompetitorDataSource interface {
	GetCompetitorPrices(productID string) ([]CompetitorPrice, error)
}

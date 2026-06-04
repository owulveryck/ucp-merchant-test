// Package models defines types for the unified multi-agent pricing system.
package models

// PricingRequest represents a request from a customer (Acheteur Lambda).
type PricingRequest struct {
	ProductID  string // ID du produit demandé
	CustomerID string // ID du client
	BasePrice  int    // Prix de base en centimes
	CostPrice  int    // Prix de revient en centimes
}

// CustomerGrowthDecision represents Agent 2's decision about customer retention.
type CustomerGrowthDecision struct {
	ShouldRetain       bool     // OUI/NON - Garder ce client ?
	CustomerTier       string   // Tier: premium, gold, silver, standard
	SuggestedDiscount  int      // % de réduction suggérée (0-20)
	LifetimeValue      int      // Valeur vie client en centimes
	RetentionReasoning []string // Pourquoi garder/ne pas garder ce client
}

// CompetitivenessDecision represents Agent 3's market competitiveness analysis.
type CompetitivenessDecision struct {
	IsCompetitive      bool     // Sommes-nous compétitifs ?
	MarketPosition     int      // Position sur le marché (1 = leader)
	TotalCompetitors   int      // Nombre de concurrents
	LowestCompetitor   int      // Prix le plus bas en centimes
	RecommendedPrice   int      // Prix recommandé en centimes
	Strategy           string   // Stratégie: market_leader, beat_competition, minimum_viable
	Margin             int      // Marge finale en %
	CompetitiveReasoning []string // Analyse du marché
}

// VendorDecision represents Agent 1's final pricing decision.
type VendorDecision struct {
	FinalPrice         int      // Prix final offert au client
	OriginalPrice      int      // Prix de départ
	TotalDiscount      int      // Réduction totale en centimes
	DiscountPercent    int      // Réduction en %
	Strategy           string   // Stratégie finale appliquée
	Margin             int      // Marge en %
	CustomerGrowth     CustomerGrowthDecision
	Competitiveness    CompetitivenessDecision
	DecisionReasoning  []string // Raisonnement de l'agent vendeur
}

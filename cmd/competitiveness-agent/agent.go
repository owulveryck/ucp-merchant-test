package main

import (
	"context"
	"fmt"

	"github.com/owulveryck/ucp-merchant-test/pkg/a2a"
	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive"
	compAgents "github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive/agents"
	compModels "github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive/models"
	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-unified/agents"
	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-unified/models"
)

// CompletivenessA2AAgent wraps the existing CompletivenessAgent for A2A communication.
type CompletivenessA2AAgent struct {
	agent *agents.CompletivenessAgent
}

// NewCompletivenessA2AAgent creates a new A2A-enabled Competitiveness agent.
func NewCompletivenessA2AAgent() *CompletivenessA2AAgent {
	// For standalone mode, create the full 4-agent system with minimal config

	// Create business config
	businessConfig := compModels.BusinessConfig{
		// Add default business config values if needed
	}

	// Create margin config
	marginConfig := compModels.MarginConfig{
		MinMarginPercent: 10,
		ActualCost:       5000, // Default $50
	}

	// Create the 4 agents
	priceIntel := compAgents.NewPriceIntelligenceAgent(
		nil,                     // CompetitorPriceSource (nil = mock/no competitors)
		"standalone_merchant",   // merchantID
	)

	marketAnalyst := compAgents.NewMarketAnalysisAgent(
		nil, // HistoryStore (nil = no history)
	)

	strategyRec := compAgents.NewStrategyRecommenderAgent(businessConfig)
	marginVal := compAgents.NewMarginValidatorAgent(marginConfig)

	// Create orchestrator
	orchestrator := competitive.NewOrchestrator(
		priceIntel,
		marketAnalyst,
		strategyRec,
		marginVal,
	)

	return &CompletivenessA2AAgent{
		agent: agents.NewCompletivenessAgent(
			orchestrator,
			"standalone_merchant",
			5000, // Default cost price in cents ($50)
			businessConfig,
		),
	}
}

// Identity returns the agent's identity card.
func (a *CompletivenessA2AAgent) Identity() a2a.AgentIdentity {
	return a2a.AgentIdentity{
		Name:       "Competitiveness Agent",
		Department: "Stratégie Prix",
		Role:       "Analyser la compétitivité du prix et recommander une stratégie de pricing",
		Version:    "1.0.0",
	}
}

// SupportedMethods returns the list of methods this agent supports.
func (a *CompletivenessA2AAgent) SupportedMethods() []string {
	return []string{
		"analyze_competitiveness",
		"check_price_position",
		"recommend_strategy",
	}
}

// HandleRequest processes a method call and returns a response.
func (a *CompletivenessA2AAgent) HandleRequest(ctx context.Context, method string, params map[string]interface{}) (interface{}, error) {
	switch method {
	case "analyze_competitiveness":
		return a.handleAnalyzeCompetitiveness(params)
	case "check_price_position":
		return a.handleCheckPricePosition(params)
	case "recommend_strategy":
		return a.handleRecommendStrategy(params)
	default:
		return nil, fmt.Errorf("method not found: %s", method)
	}
}

func (a *CompletivenessA2AAgent) handleAnalyzeCompetitiveness(params map[string]interface{}) (interface{}, error) {
	productID, ok := params["product_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid product_id parameter")
	}

	// Get price (default to 10000 cents = $100 if not provided)
	priceFloat, ok := params["price"].(float64)
	if !ok {
		priceFloat = 10000
	}
	price := int(priceFloat)

	// Call the underlying agent
	decision, err := a.agent.Analyze(productID, price)
	if err != nil {
		return nil, err
	}

	// Format conversational response
	message := a.formatCompetitivenessMessage(productID, price, decision)

	return a2a.AgentResponse{
		Agent:    a.Identity(),
		Message:  message,
		Decision: decision,
	}, nil
}

func (a *CompletivenessA2AAgent) handleCheckPricePosition(params map[string]interface{}) (interface{}, error) {
	productID, ok := params["product_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid product_id parameter")
	}

	priceFloat, ok := params["price"].(float64)
	if !ok {
		priceFloat = 10000
	}
	price := int(priceFloat)

	decision, err := a.agent.Analyze(productID, price)
	if err != nil {
		return nil, err
	}

	message := a2a.FormatMessage(
		a.Identity().Name,
		a.Identity().Department,
		fmt.Sprintf("Le produit '%s' au prix de $%.2f est à la position %d/%d sur le marché.",
			productID, float64(price)/100, decision.MarketPosition, decision.TotalCompetitors),
	)

	return a2a.AgentResponse{
		Agent: a.Identity(),
		Message: message,
		Decision: map[string]interface{}{
			"product_id":       productID,
			"market_position":  decision.MarketPosition,
			"total_competitors": decision.TotalCompetitors,
			"is_competitive":   decision.IsCompetitive,
		},
	}, nil
}

func (a *CompletivenessA2AAgent) handleRecommendStrategy(params map[string]interface{}) (interface{}, error) {
	productID, ok := params["product_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid product_id parameter")
	}

	priceFloat, ok := params["price"].(float64)
	if !ok {
		priceFloat = 10000
	}
	price := int(priceFloat)

	decision, err := a.agent.Analyze(productID, price)
	if err != nil {
		return nil, err
	}

	message := a2a.FormatMessage(
		a.Identity().Name,
		a.Identity().Department,
		fmt.Sprintf("Pour le produit '%s', je recommande la stratégie '%s' avec un prix de $%.2f.",
			productID, decision.Strategy, float64(decision.RecommendedPrice)/100),
	)

	return a2a.AgentResponse{
		Agent: a.Identity(),
		Message: message,
		Decision: map[string]interface{}{
			"product_id":        productID,
			"strategy":          decision.Strategy,
			"recommended_price": decision.RecommendedPrice,
			"margin":            decision.Margin,
		},
	}, nil
}

func (a *CompletivenessA2AAgent) formatCompetitivenessMessage(productID string, price int, decision models.CompetitivenessDecision) string {
	identity := a.Identity()

	competitiveStatus := "NON"
	if decision.IsCompetitive {
		competitiveStatus = "OUI"
	}

	message := fmt.Sprintf(
		"Bonjour, je suis %s du département %s. "+
			"Le produit '%s' au prix actuel de $%.2f est à la position %d/%d sur le marché. "+
			"%s, ce prix est compétitif. "+
			"Je recommande la stratégie '%s' avec un prix de $%.2f (marge: %d%%).",
		identity.Name,
		identity.Department,
		productID,
		float64(price)/100,
		decision.MarketPosition,
		decision.TotalCompetitors,
		competitiveStatus,
		decision.Strategy,
		float64(decision.RecommendedPrice)/100,
		decision.Margin,
	)

	return message
}

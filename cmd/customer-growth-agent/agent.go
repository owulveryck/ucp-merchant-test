package main

import (
	"context"
	"fmt"

	"github.com/owulveryck/ucp-merchant-test/pkg/a2a"
	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-unified/agents"
	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-unified/datasources"
	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-unified/models"
)

// CustomerGrowthA2AAgent wraps the existing CustomerGrowthAgent for A2A communication.
type CustomerGrowthA2AAgent struct {
	agent *agents.CustomerGrowthAgent
}

// NewCustomerGrowthA2AAgent creates a new A2A-enabled Customer Growth agent.
func NewCustomerGrowthA2AAgent() *CustomerGrowthA2AAgent {
	// Use mock data source for standalone mode
	dataSource := datasources.NewMockCustomerDataSource()

	return &CustomerGrowthA2AAgent{
		agent: agents.NewCustomerGrowthAgent(dataSource),
	}
}

// Identity returns the agent's identity card.
func (a *CustomerGrowthA2AAgent) Identity() a2a.AgentIdentity {
	return a2a.AgentIdentity{
		Name:       "Customer Growth Agent",
		Department: "Fidélisation",
		Role:       "Analyser la valeur client et recommander des stratégies de rétention",
		Version:    "1.0.0",
	}
}

// SupportedMethods returns the list of methods this agent supports.
func (a *CustomerGrowthA2AAgent) SupportedMethods() []string {
	return []string{
		"analyze_customer",
		"get_customer_tier",
		"recommend_discount",
	}
}

// HandleRequest processes a method call and returns a response.
func (a *CustomerGrowthA2AAgent) HandleRequest(ctx context.Context, method string, params map[string]interface{}) (interface{}, error) {
	switch method {
	case "analyze_customer":
		return a.handleAnalyzeCustomer(params)
	case "get_customer_tier":
		return a.handleGetCustomerTier(params)
	case "recommend_discount":
		return a.handleRecommendDiscount(params)
	default:
		return nil, fmt.Errorf("method not found: %s", method)
	}
}

func (a *CustomerGrowthA2AAgent) handleAnalyzeCustomer(params map[string]interface{}) (interface{}, error) {
	customerID, ok := params["customer_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid customer_id parameter")
	}

	// Call the underlying agent
	decision, err := a.agent.Analyze(customerID)
	if err != nil {
		return nil, err
	}

	// Format conversational response
	message := a.formatAnalysisMessage(customerID, decision)

	return a2a.AgentResponse{
		Agent:    a.Identity(),
		Message:  message,
		Decision: decision,
	}, nil
}

func (a *CustomerGrowthA2AAgent) handleGetCustomerTier(params map[string]interface{}) (interface{}, error) {
	customerID, ok := params["customer_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid customer_id parameter")
	}

	decision, err := a.agent.Analyze(customerID)
	if err != nil {
		return nil, err
	}

	message := a2a.FormatMessage(
		a.Identity().Name,
		a.Identity().Department,
		fmt.Sprintf("Le client '%s' est de tier %s.", customerID, decision.CustomerTier),
	)

	return a2a.AgentResponse{
		Agent: a.Identity(),
		Message: message,
		Decision: map[string]interface{}{
			"customer_id": customerID,
			"tier":        decision.CustomerTier,
		},
	}, nil
}

func (a *CustomerGrowthA2AAgent) handleRecommendDiscount(params map[string]interface{}) (interface{}, error) {
	customerID, ok := params["customer_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid customer_id parameter")
	}

	decision, err := a.agent.Analyze(customerID)
	if err != nil {
		return nil, err
	}

	message := a2a.FormatMessage(
		a.Identity().Name,
		a.Identity().Department,
		fmt.Sprintf("Pour le client '%s', je recommande une réduction de %d%% pour optimiser sa fidélité.",
			customerID, decision.SuggestedDiscount),
	)

	return a2a.AgentResponse{
		Agent: a.Identity(),
		Message: message,
		Decision: map[string]interface{}{
			"customer_id":       customerID,
			"suggested_discount": decision.SuggestedDiscount,
			"tier":              decision.CustomerTier,
		},
	}, nil
}

func (a *CustomerGrowthA2AAgent) formatAnalysisMessage(customerID string, decision models.CustomerGrowthDecision) string {
	identity := a.Identity()

	importanceVerb := "NON"
	if decision.ShouldRetain {
		importanceVerb = "OUI"
	}

	message := fmt.Sprintf(
		"Bonjour, je suis %s du département %s. "+
			"Le client '%s' est un client %s ayant dépensé $%.2f. "+
			"%s, c'est un client important à conserver. "+
			"Je recommande une réduction de %d%% pour maximiser sa fidélité.",
		identity.Name,
		identity.Department,
		customerID,
		decision.CustomerTier,
		float64(decision.LifetimeValue)/100,
		importanceVerb,
		decision.SuggestedDiscount,
	)

	return message
}

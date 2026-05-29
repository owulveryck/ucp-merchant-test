package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"

	"github.com/owulveryck/ucp-merchant-test/pkg/auth"
	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive"
	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive/agents"
	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive/history"
	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive/models"
	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/transport/a2a"
	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/transport/discovery"
)

// Tenant represents a registered arena merchant.
type Tenant struct {
	ID       string
	Name     string
	Merchant *arenaMerchant
	A2A      *a2a.Server
	OAuth    *auth.OAuthServer
	Disc     *discovery.Server
	Mux      *http.ServeMux
	Notifier *Notifier
	Config   *MerchantConfig
}

// tenantBaseURL returns the public-facing base URL for a tenant.
// When s.baseURL is set, it returns the public URL; otherwise localhost.
func (s *ArenaServer) tenantBaseURL(id string) string {
	if s.baseURL != "" {
		return s.baseURL + "/" + id
	}
	return fmt.Sprintf("http://localhost:%d/%s", s.port, id)
}

// RegisterTenant creates a new tenant and registers it with the shopping graph.
func (s *ArenaServer) RegisterTenant(name string) *Tenant {
	id := uuid.New().String()[:8]

	baseURL := func() string {
		return s.tenantBaseURL(id)
	}

	notifier := NewNotifier()

	m := newArenaMerchant("casque_audio", s.productName, s.costPrice, baseURL, notifier)
	m.graphURL = s.graphURL
	m.merchantID = id

	// Set up competitive pricing with multi-agent architecture
	if s.competitivePricing && s.graphURL != "" {
		log.Printf("Enabling MULTI-AGENT competitive pricing for tenant %s (minMargin=%d%%)",
			name, s.minMargin)

		// Create Shopping Graph client (new architecture)
		sgClient := competitive.NewShoppingGraphClient(s.graphURL)

		// Agent 1: Price Intelligence
		priceIntel := agents.NewPriceIntelligenceAgent(sgClient, id)
		log.Printf("[%s] Agent 1 (Price Intelligence) initialized", name)

		// Agent 2: Market Analysis (with history)
		historyStore := history.NewInMemoryHistoryStore()
		marketAnalyst := agents.NewMarketAnalysisAgent(historyStore)
		log.Printf("[%s] Agent 2 (Market Analysis) initialized", name)

		// Agent 3: Strategy Recommender
		businessConfig := models.BusinessConfig{
			Objective:      "volume",      // Arena default: maximize sales
			StockThreshold: 20,            // Low stock threshold
			BrandPosition:  "mid",         // Mid-market positioning
			MinMargin:      s.minMargin,   // From server config
			CostPercent:    60,            // 60% cost = 40% base margin
		}
		strategyRec := agents.NewStrategyRecommenderAgent(businessConfig)
		log.Printf("[%s] Agent 3 (Strategy Recommender) initialized - objective: %s", name, businessConfig.Objective)

		// Agent 4: Margin Validator
		marginConfig := models.MarginConfig{
			MinMarginPercent: s.minMargin,
			CostPercent:      60,
			ActualCost:       s.costPrice, // CRITICAL: Use real cost, not estimated
			HardFloor:        true,        // Never sell below cost
		}
		marginVal := agents.NewMarginValidatorAgent(marginConfig)
		log.Printf("[%s] Agent 4 (Margin Validator) initialized - min margin: %d%%, actual cost: $%.2f",
			name, s.minMargin, float64(s.costPrice)/100)

		// Create Orchestrator
		orchestrator := competitive.NewOrchestrator(
			priceIntel,
			marketAnalyst,
			strategyRec,
			marginVal,
		)

		// Create Discount Adapter
		discountAdapter := competitive.NewDiscountAdapter(
			nil,            // No base discount lookup for arena merchants
			orchestrator,   // Multi-agent orchestrator
			businessConfig, // Business context
		)

		// Set callback to send agent decisions to dashboard via SSE
		discountAdapter.SetAgentDecisionsCallback(func(decisions *competitive.AgentDecisions) {
			log.Printf("[Tenant %s] Agent decisions callback triggered!", name)

			log.Printf("[Tenant %s] Building event map...", name)
			event := map[string]interface{}{
				"type": "agent_decisions",
				"agent1": map[string]interface{}{
					"rank":         decisions.Intel.OurRank,
					"total":        decisions.Intel.TotalCount,
					"lowest_price": decisions.Intel.LowestPrice,
					"lowest_by":    decisions.Intel.LowestBy,
					"avg_price":    decisions.Intel.AvgPrice,
				},
				"agent2": map[string]interface{}{
					"position":    decisions.Insight.Position,
					"trend":       decisions.Insight.Trend,
					"opportunity": decisions.Insight.Opportunity,
					"reasoning":   decisions.Insight.Reasoning,
				},
				"agent3": map[string]interface{}{
					"strategy":   decisions.Recommendation.Strategy,
					"target":     decisions.Recommendation.TargetPrice,
					"discount":   decisions.Recommendation.DiscountAmount,
					"confidence": decisions.Recommendation.Confidence,
					"reasoning":  decisions.Recommendation.Reasoning,
				},
				"agent4": map[string]interface{}{
					"approved": decisions.Validation.Approved,
					"rejected": decisions.Validation.Rejected,
					"final":    decisions.Validation.FinalPrice,
					"margin":   decisions.Validation.Margin,
					"warnings": decisions.Validation.Warnings,
				},
			}

			log.Printf("[Tenant %s] Marshaling to JSON...", name)
			data, err := json.Marshal(event)
			if err != nil {
				log.Printf("[Tenant %s] ERROR marshaling JSON: %v", name, err)
				return
			}

			log.Printf("[Tenant %s] Sending agent_decisions SSE event: %s", name, string(data))
			notifier.SendRaw(data)
			log.Printf("[Tenant %s] SSE event sent", name)
		})

		// Inject into merchant
		m.pricingAgent = discountAdapter

		log.Printf("✅ MULTI-AGENT competitive pricing configured for %s (4 agents active)", name)
	}

	m.onActivity = func(eventType, summary string) {
		s.forwardToObsHub(name, eventType, summary, nil)
		// Only forward actionable events to presenter (skip catalog polling noise)
		switch eventType {
		case "checkout_created", "checkout_updated", "checkout_canceled", "cart_created":
			s.presenterN.Send(SaleEvent{Type: eventType, MerchantID: id})
		}
	}
	m.onSale = func(ev SaleEvent) {
		ev.MerchantID = id
		s.presenterN.Send(ev)
		s.forwardToObsHub("arena", "sale_completed",
			fmt.Sprintf("SALE! Order %s - $%.2f (%s)", ev.OrderID, float64(ev.Total)/100, ev.Buyer),
			map[string]any{
				"order_id":       ev.OrderID,
				"buyer":          ev.Buyer,
				"total":          ev.Total,
				"total_revenue":  ev.TotalRevenue,
				"total_ad_spend": ev.TotalAdSpend,
				"net_profit":     ev.NetProfit,
			})
	}

	oauthSrv := auth.NewOAuthServer(
		name,
		func() string { return "http" },
		func() int { return s.port },
	)
	// When a public base URL is configured, override OAuth metadata URLs
	if s.baseURL != "" {
		oauthSrv.BaseURLFn = baseURL
	}

	a2aOpts := []a2a.Option{
		a2a.WithMerchantName(name),
		a2a.WithListenPort(func() int { return s.port }),
		a2a.WithScheme(func() string { return "http" }),
	}
	// When a public base URL is configured, override agent card URL
	if s.baseURL != "" {
		a2aOpts = append(a2aOpts, a2a.WithBaseURL(baseURL))
	}

	a2aSrv := a2a.New(m, oauthSrv, a2aOpts...)

	disc := discovery.New(baseURL)

	mux := http.NewServeMux()

	// Dashboard
	mux.HandleFunc("GET /dashboard", func(w http.ResponseWriter, r *http.Request) {
		serveDashboard(w, r, id, name, s.costPrice)
	})

	// Config API
	mux.HandleFunc("GET /api/config", func(w http.ResponseWriter, r *http.Request) {
		handleGetConfig(w, r, m.config, s.costPrice, m)
	})
	mux.HandleFunc("PUT /api/config", func(w http.ResponseWriter, r *http.Request) {
		handlePutConfig(w, r, m.config, s.costPrice, s, id, m)
	})

	// Test AUTO_COMPETE API
	mux.HandleFunc("POST /api/test-auto-compete", func(w http.ResponseWriter, r *http.Request) {
		setCORSHeaders(w)
		handleTestAutoCompete(w, r, m, id)
	})

	// Competitive Intelligence API
	mux.HandleFunc("GET /api/competitive-intel", func(w http.ResponseWriter, r *http.Request) {
		handleCompetitiveIntel(w, r, m, s.graphURL, id, s.costPrice)
	})

	// SSE notifications
	mux.Handle("GET /api/notifications", notifier)

	// A2A transport
	mux.Handle("/a2a", a2aSrv)
	mux.HandleFunc("/.well-known/agent-card.json", a2aSrv.HandleAgentCard)

	// Discovery
	mux.HandleFunc("/.well-known/ucp", disc.HandleDiscovery)
	mux.HandleFunc("/.well-known/oauth-authorization-server", func(w http.ResponseWriter, r *http.Request) {
		oauthSrv.MerchantName = name
		oauthSrv.HandleMetadata(w, r)
	})

	// OAuth
	mux.HandleFunc("/oauth2/authorize", func(w http.ResponseWriter, r *http.Request) {
		oauthSrv.MerchantName = name
		oauthSrv.HandleAuthorize(w, r)
	})
	mux.HandleFunc("/oauth2/token", oauthSrv.HandleToken)
	mux.HandleFunc("/oauth2/revoke", oauthSrv.HandleRevoke)

	// Leave arena
	mux.HandleFunc("POST /leave", func(w http.ResponseWriter, r *http.Request) {
		setCORSHeaders(w)
		s.RemoveTenant(id)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "removed", "id": id})
	})

	// Specs and schemas
	mux.HandleFunc("/specs/", disc.HandleSpecsAndSchemas)
	mux.HandleFunc("/schemas/", disc.HandleSpecsAndSchemas)

	tenant := &Tenant{
		ID:       id,
		Name:     name,
		Merchant: m,
		A2A:      a2aSrv,
		OAuth:    oauthSrv,
		Disc:     disc,
		Mux:      mux,
		Notifier: notifier,
		Config:   m.config,
	}

	s.mu.Lock()
	s.tenants[id] = tenant
	s.mu.Unlock()

	// Register with shopping graph
	go s.registerWithGraph(tenant)

	// Notify presenter
	event := SaleEvent{Type: "registration", MerchantID: id, Buyer: name, OrderID: id}
	s.presenterN.Send(event)

	// Forward to obs-hub
	s.forwardToObsHub("arena", "merchant_registered", "New merchant: "+name, nil)

	return tenant
}

func (s *ArenaServer) registerWithGraph(t *Tenant) {
	if s.graphURL == "" {
		return
	}

	body, _ := json.Marshal(map[string]any{
		"id":          t.ID,
		"name":        t.Name,
		"endpoint":    s.tenantBaseURL(t.ID),
		"max_cpc_bid": t.Config.MaxCPCBid,
	})

	resp, err := httpClient.Post(s.graphURL+"/merchants", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("register with graph: %v", err)
		return
	}
	resp.Body.Close()
	log.Printf("registered tenant %s (%s) with shopping graph", t.Name, t.ID)
}

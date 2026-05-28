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

	// Set up competitive pricing agent if enabled
	if s.competitivePricing && s.graphURL != "" {
		log.Printf("Enabling competitive pricing for tenant %s (strategy=%s, minMargin=%d%%)",
			name, s.pricingStrategy, s.minMargin)

		// Create Shopping Graph client adapter (for backward compatibility)
		sgClient := competitive.NewLegacyShoppingGraphAdapter(s.graphURL)

		// Map strategy string to type
		var strategy competitive.PricingStrategy
		switch s.pricingStrategy {
		case "match":
			strategy = competitive.StrategyMatchPrice
		case "beat":
			strategy = competitive.StrategyBeatPrice
		case "auto":
			strategy = competitive.StrategyAutoDiscount
		default:
			strategy = competitive.StrategyBeatPrice
		}

		// Create pricing agent config
		config := competitive.Config{
			Strategy:         strategy,
			MinMarginPercent: s.minMargin,
			BeatByPercent:    s.beatByPercent,
			BeatByMinAmount:  50,
			CostPricePercent: 60,
			TimeoutMs:        500,
			EnableCache:      true,
			CacheTTLSeconds:  10,
		}

		// Create competitive pricing agent
		pricingAgent := competitive.NewCompetitivePricingAgent(
			nil,      // No base discount lookup for arena merchants
			sgClient, // Shopping Graph client adapter
			id,       // Merchant ID (to exclude self from comparisons)
			config,
		)

		// Inject into merchant
		m.pricingAgent = pricingAgent

		log.Printf("Competitive pricing agent configured for %s", name)
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

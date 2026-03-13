package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"

	"github.com/owulveryck/ucp-merchant-test/pkg/auth"
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

// RegisterTenant creates a new tenant and registers it with the shopping graph.
func (s *ArenaServer) RegisterTenant(name string) *Tenant {
	id := uuid.New().String()[:8]

	baseURL := func() string {
		return fmt.Sprintf("http://localhost:%d/%s", s.port, id)
	}

	notifier := NewNotifier()

	m := newArenaMerchant("casque_audio", s.productName, s.costPrice, baseURL, notifier)
	m.onActivity = func(eventType, summary string) {
		s.forwardToObsHub(name, eventType, summary, nil)
	}
	m.onSale = func(ev SaleEvent) {
		s.presenterN.Send(ev)
		s.forwardToObsHub("arena", "sale_completed",
			fmt.Sprintf("SALE! Order %s - $%.2f (%s)", ev.OrderID, float64(ev.Total)/100, ev.Buyer),
			map[string]any{
				"order_id":     ev.OrderID,
				"buyer":        ev.Buyer,
				"total":        ev.Total,
				"net_profit":   ev.NetProfit,
				"total_profit": ev.TotalProfit,
			})
	}

	oauthSrv := auth.NewOAuthServer(
		name,
		func() string { return "http" },
		func() int { return s.port },
	)

	a2aSrv := a2a.New(m, oauthSrv,
		a2a.WithMerchantName(name),
		a2a.WithListenPort(func() int { return s.port }),
		a2a.WithScheme(func() string { return "http" }),
	)

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
	event := SaleEvent{Type: "registration", Buyer: name, OrderID: id}
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
		"id":       t.ID,
		"name":     t.Name,
		"endpoint": fmt.Sprintf("http://localhost:%d/%s", s.port, t.ID),
		"score":    t.Config.BoostScore,
	})

	resp, err := http.Post(s.graphURL+"/merchants", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("register with graph: %v", err)
		return
	}
	resp.Body.Close()
	log.Printf("registered tenant %s (%s) with shopping graph", t.Name, t.ID)
}

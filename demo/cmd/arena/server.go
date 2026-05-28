package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// httpClient is used for all outgoing HTTP calls (shopping graph, obs-hub) with a timeout.
var httpClient = &http.Client{Timeout: 5 * time.Second}

// ArenaServer manages multiple tenant merchants for the conference demo.
type ArenaServer struct {
	mu          sync.RWMutex
	tenants     map[string]*Tenant
	costPrice   int
	productName string
	graphURL    string
	obsURL      string
	port        int
	baseURL     string    // external base URL for public-facing endpoints; empty = localhost
	presenterN  *Notifier // SSE for presenter dashboard

	// Competitive pricing configuration
	competitivePricing bool
	pricingStrategy    string
	minMargin          int
	beatByPercent      int
}

// NewArenaServer creates a new arena server.
func NewArenaServer(costPrice int, productName, graphURL, obsURL string, port int, baseURL string, competitivePricing bool, pricingStrategy string, minMargin, beatByPercent int) *ArenaServer {
	return &ArenaServer{
		tenants:            make(map[string]*Tenant),
		costPrice:          costPrice,
		productName:        productName,
		graphURL:           graphURL,
		obsURL:             obsURL,
		port:               port,
		baseURL:            strings.TrimRight(baseURL, "/"),
		presenterN:         NewNotifier(),
		competitivePricing: competitivePricing,
		pricingStrategy:    pricingStrategy,
		minMargin:          minMargin,
		beatByPercent:      beatByPercent,
	}
}

// GetTenant returns a tenant by UUID.
func (s *ArenaServer) GetTenant(uuid string) *Tenant {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tenants[uuid]
}

// HasTenantNamed returns true if a tenant with the given name (case-insensitive) already exists.
func (s *ArenaServer) HasTenantNamed(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	lower := strings.ToLower(name)
	for _, t := range s.tenants {
		if strings.ToLower(t.Name) == lower {
			return true
		}
	}
	return false
}

// ListTenants returns all tenants.
func (s *ArenaServer) ListTenants() []*Tenant {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Tenant, 0, len(s.tenants))
	for _, t := range s.tenants {
		result = append(result, t)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}

// RemoveTenant removes a tenant from the arena and deregisters it from the shopping graph.
func (s *ArenaServer) RemoveTenant(id string) bool {
	s.mu.Lock()
	tenant, ok := s.tenants[id]
	if !ok {
		s.mu.Unlock()
		return false
	}
	delete(s.tenants, id)
	s.mu.Unlock()

	// Deregister from shopping graph
	go s.deregisterFromGraph(id)

	// Notify presenter SSE
	s.presenterN.Send(SaleEvent{Type: "merchant_left", MerchantID: id, Buyer: tenant.Name})

	// Forward to obs-hub
	s.forwardToObsHub("arena", "merchant_left", "Merchant left: "+tenant.Name, nil)

	log.Printf("removed tenant %s (%s) from arena", tenant.Name, id)
	return true
}

// deregisterFromGraph sends a DELETE to the shopping graph to remove a merchant.
func (s *ArenaServer) deregisterFromGraph(id string) {
	if s.graphURL == "" {
		return
	}
	req, _ := http.NewRequest(http.MethodDelete, s.graphURL+"/merchants/"+id, nil)
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("deregister from graph: %v", err)
		return
	}
	resp.Body.Close()
	log.Printf("deregistered tenant %s from shopping graph", id)
}

// ServeHTTP implements http.Handler with top-level routing.
func (s *ArenaServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Limit request body size for POST/PUT
	if r.Method == http.MethodPost || r.Method == http.MethodPut {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	}

	path := r.URL.Path

	// Top-level routes
	switch {
	case path == "/":
		s.handleLanding(w, r)
		return
	case path == "/register" && r.Method == http.MethodPost:
		s.handleRegister(w, r)
		return
	case path == "/auto":
		s.handleAuto(w, r)
		return
	case path == "/merchants":
		s.handleListMerchants(w, r)
		return
	case path == "/config":
		s.handleConfig(w, r)
		return
	case path == "/events":
		s.presenterN.ServeHTTP(w, r)
		return
	case path == "/rankings":
		s.handleRankings(w, r)
		return
	case path == "/health":
		s.handleHealth(w, r)
		return
	}

	// Extract tenant UUID from first path segment
	trimmed := strings.TrimPrefix(path, "/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) == 0 {
		http.NotFound(w, r)
		return
	}

	uuid := parts[0]
	tenant := s.GetTenant(uuid)
	if tenant == nil {
		http.NotFound(w, r)
		return
	}

	http.StripPrefix("/"+uuid, tenant.Mux).ServeHTTP(w, r)
}

// forwardToObsHub sends an arena event to the obs-hub /event endpoint.
func (s *ArenaServer) forwardToObsHub(source, eventType, summary string, data any) {
	if s.obsURL == "" {
		return
	}
	go func() {
		body, _ := json.Marshal(map[string]any{
			"source":  source,
			"type":    eventType,
			"summary": summary,
			"data":    data,
		})
		resp, err := httpClient.Post(s.obsURL+"/event", "application/json", bytes.NewReader(body))
		if err != nil {
			log.Printf("forward to obs-hub: %v", err)
			return
		}
		resp.Body.Close()
	}()
}

// handleConfig returns arena configuration (product name and cost price).
func (s *ArenaServer) handleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"product_name": s.productName,
		"cost_price":   s.costPrice,
	})
}

// handleRankings fetches rankings from the shopping graph and returns merchant_id -> rank.
func (s *ArenaServer) handleRankings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if s.graphURL == "" {
		json.NewEncoder(w).Encode(map[string]any{"rankings": map[string]any{}})
		return
	}

	body, _ := json.Marshal(map[string]any{"query": s.productName, "limit": 300})
	resp, err := httpClient.Post(s.graphURL+"/search", "application/json", bytes.NewReader(body))
	if err != nil {
		json.NewEncoder(w).Encode(map[string]any{"rankings": map[string]any{}})
		return
	}
	defer resp.Body.Close()

	var result struct {
		Results []struct {
			Rank       int    `json:"rank"`
			MerchantID string `json:"merchant_id"`
			Price      int    `json:"price"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		json.NewEncoder(w).Encode(map[string]any{"rankings": map[string]any{}})
		return
	}

	rankings := make(map[string]any)
	for _, r := range result.Results {
		rankings[r.MerchantID] = map[string]int{"rank": r.Rank, "price": r.Price}
	}
	json.NewEncoder(w).Encode(map[string]any{"rankings": rankings})
}

// handleHealth returns a simple health check response.
func (s *ArenaServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	n := len(s.tenants)
	s.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":  "ok",
		"tenants": n,
		"product": s.productName,
	})
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, UCP-Agent, X-A2A-Extensions")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Type")
}

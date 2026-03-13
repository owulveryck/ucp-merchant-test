package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
)

// ArenaServer manages multiple tenant merchants for the conference demo.
type ArenaServer struct {
	mu          sync.RWMutex
	tenants     map[string]*Tenant
	costPrice   int
	productName string
	graphURL    string
	obsURL      string
	port        int
	presenterN  *Notifier // SSE for presenter dashboard
}

// NewArenaServer creates a new arena server.
func NewArenaServer(costPrice int, productName, graphURL, obsURL string, port int) *ArenaServer {
	return &ArenaServer{
		tenants:     make(map[string]*Tenant),
		costPrice:   costPrice,
		productName: productName,
		graphURL:    graphURL,
		obsURL:      obsURL,
		port:        port,
		presenterN:  NewNotifier(),
	}
}

// GetTenant returns a tenant by UUID.
func (s *ArenaServer) GetTenant(uuid string) *Tenant {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tenants[uuid]
}

// ListTenants returns all tenants.
func (s *ArenaServer) ListTenants() []*Tenant {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Tenant, 0, len(s.tenants))
	for _, t := range s.tenants {
		result = append(result, t)
	}
	return result
}

// ServeHTTP implements http.Handler with top-level routing.
func (s *ArenaServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
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
	case path == "/merchants":
		s.handleListMerchants(w, r)
		return
	case path == "/config":
		s.handleConfig(w, r)
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
		resp, err := http.Post(s.obsURL+"/event", "application/json", bytes.NewReader(body))
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

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, UCP-Agent, X-A2A-Extensions")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Type")
}

package shoppinggraph

import (
	"encoding/json"
	"net/http"
)

// Handler provides HTTP endpoints for the shopping graph.
type Handler struct {
	graph *ShoppingGraph
}

// NewHandler creates a new HTTP handler for the shopping graph.
func NewHandler(graph *ShoppingGraph) *Handler {
	return &Handler{graph: graph}
}

// Mux returns an http.ServeMux with all routes registered.
func (h *Handler) Mux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /search", h.handleSearch)
	mux.HandleFunc("GET /health", h.handleHealth)
	mux.HandleFunc("GET /ranking", h.handleGetRanking)
	mux.HandleFunc("PUT /ranking", h.handlePutRanking)
	return mux
}

func (h *Handler) handleSearch(w http.ResponseWriter, r *http.Request) {
	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"detail":"invalid request"}`, http.StatusBadRequest)
		return
	}

	results := h.graph.Search(req.Query, req.Limit)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"results": results,
		"total":   len(results),
	})
}

var availableAlgorithms = []RankingAlgorithm{RankJaccard, RankJaccardPrice, RankPriceOnly}

func (h *Handler) handleGetRanking(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"algorithm": h.graph.GetRankAlgo(),
		"available": availableAlgorithms,
	})
}

func (h *Handler) handlePutRanking(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Algorithm RankingAlgorithm `json:"algorithm"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"detail":"invalid request"}`, http.StatusBadRequest)
		return
	}
	valid := false
	for _, a := range availableAlgorithms {
		if req.Algorithm == a {
			valid = true
			break
		}
	}
	if !valid {
		http.Error(w, `{"detail":"unknown algorithm"}`, http.StatusBadRequest)
		return
	}
	h.graph.SetRankAlgo(req.Algorithm)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"algorithm": h.graph.GetRankAlgo(),
		"available": availableAlgorithms,
	})
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	h.graph.mu.RLock()
	defer h.graph.mu.RUnlock()

	online := 0
	for _, m := range h.graph.Merchants {
		if m.Online {
			online++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"merchants_total":  len(h.graph.Merchants),
		"merchants_online": online,
		"products_total":   len(h.graph.Products),
		"groups_total":     len(h.graph.Groups),
		"last_updated":     h.graph.LastUpdated,
	})
}

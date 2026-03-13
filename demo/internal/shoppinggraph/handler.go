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
	mux.HandleFunc("POST /merchants", h.handleAddMerchant)
	mux.HandleFunc("DELETE /merchants/{id}", h.handleRemoveMerchant)
	mux.HandleFunc("PUT /boost", h.handleSetBoost)
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

func (h *Handler) handleAddMerchant(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID            string   `json:"id"`
		Name          string   `json:"name"`
		Endpoint      string   `json:"endpoint"`
		Score         int      `json:"score"`
		DiscountHints []string `json:"discount_hints,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" {
		http.Error(w, `{"detail":"invalid request"}`, http.StatusBadRequest)
		return
	}
	if req.Score == 0 {
		req.Score = 50
	}
	h.graph.AddMerchant(&MerchantNode{
		ID:            req.ID,
		Name:          req.Name,
		Endpoint:      req.Endpoint,
		Score:         req.Score,
		DiscountHints: req.DiscountHints,
	})
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "added", "id": req.ID})
}

func (h *Handler) handleRemoveMerchant(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"detail":"missing id"}`, http.StatusBadRequest)
		return
	}
	h.graph.RemoveMerchant(id)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "removed", "id": id})
}

func (h *Handler) handleSetBoost(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MerchantID string `json:"merchant_id"`
		Amount     int    `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.MerchantID == "" {
		http.Error(w, `{"detail":"invalid request"}`, http.StatusBadRequest)
		return
	}
	h.graph.SetBoost(req.MerchantID, req.Amount)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
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

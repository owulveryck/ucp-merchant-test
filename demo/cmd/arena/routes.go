package main

import (
	"encoding/json"
	"net/http"
)

func (s *ArenaServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, `{"detail":"name is required"}`, http.StatusBadRequest)
		return
	}

	tenant := s.RegisterTenant(req.Name)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"id":        tenant.ID,
		"url":       tenant.Merchant.baseURL(),
		"dashboard": tenant.Merchant.baseURL() + "/dashboard",
	})
}

func (s *ArenaServer) handleListMerchants(w http.ResponseWriter, r *http.Request) {
	tenants := s.ListTenants()

	type merchantInfo struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Price       int    `json:"price"`
		Stock       int    `json:"stock"`
		Boost       int    `json:"boost"`
		Margin      int    `json:"margin"`
		BoostCost   int    `json:"boost_cost"`
		NetMargin   int    `json:"net_margin"`
		TotalProfit int    `json:"total_profit"`
		SalesCount  int    `json:"sales_count"`
	}

	result := make([]merchantInfo, 0, len(tenants))
	for _, t := range tenants {
		t.Config.mu.RLock()
		price := t.Config.SellingPrice
		stock := t.Config.Stock
		boost := t.Config.BoostScore
		t.Config.mu.RUnlock()

		margin := price - s.costPrice
		boostCost := boost * margin / 100
		netMargin := margin - boostCost

		t.Merchant.mu.Lock()
		totalProfit := t.Merchant.totalProfit
		salesCount := t.Merchant.salesCount
		t.Merchant.mu.Unlock()

		result = append(result, merchantInfo{
			ID:          t.ID,
			Name:        t.Name,
			Price:       price,
			Stock:       stock,
			Boost:       boost,
			Margin:      margin,
			BoostCost:   boostCost,
			NetMargin:   netMargin,
			TotalProfit: totalProfit,
			SalesCount:  salesCount,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"merchants": result,
		"total":     len(result),
	})
}

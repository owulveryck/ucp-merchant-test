package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

func handleGetConfig(w http.ResponseWriter, r *http.Request, cfg *MerchantConfig, costPrice int, m *arenaMerchant) {
	cfg.mu.RLock()
	sellingPrice := cfg.SellingPrice
	boostScore := cfg.BoostScore
	cfg.mu.RUnlock()

	margin := sellingPrice - costPrice
	boostCostPerSale := boostScore * margin / 100
	netMarginPerSale := margin - boostCostPerSale

	m.mu.Lock()
	totalProfit := m.totalProfit
	salesCount := m.salesCount
	m.mu.Unlock()

	cfg.mu.RLock()
	defer cfg.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"selling_price":       cfg.SellingPrice,
		"stock":               cfg.Stock,
		"discount_codes":      cfg.DiscountCodes,
		"boost_score":         cfg.BoostScore,
		"cost_price":          costPrice,
		"margin":              margin,
		"boost_cost_per_sale": boostCostPerSale,
		"net_margin_per_sale": netMarginPerSale,
		"total_profit":        totalProfit,
		"sales_count":         salesCount,
	})
}

func handlePutConfig(w http.ResponseWriter, r *http.Request, cfg *MerchantConfig, costPrice int, srv *ArenaServer, tenantID string, m *arenaMerchant) {
	var req struct {
		SellingPrice  *int           `json:"selling_price"`
		Stock         *int           `json:"stock"`
		DiscountCodes []DiscountCode `json:"discount_codes"`
		BoostScore    *int           `json:"boost_score"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"detail":"invalid request"}`, http.StatusBadRequest)
		return
	}

	cfg.mu.Lock()
	if req.SellingPrice != nil {
		if *req.SellingPrice < costPrice {
			cfg.mu.Unlock()
			http.Error(w, `{"detail":"price must be >= cost price"}`, http.StatusBadRequest)
			return
		}
		cfg.SellingPrice = *req.SellingPrice
	}
	if req.Stock != nil {
		if *req.Stock < 0 {
			cfg.mu.Unlock()
			http.Error(w, `{"detail":"stock must be >= 0"}`, http.StatusBadRequest)
			return
		}
		cfg.Stock = *req.Stock
	}
	if req.DiscountCodes != nil {
		cfg.DiscountCodes = req.DiscountCodes
	}
	if req.BoostScore != nil {
		score := *req.BoostScore
		if score < 0 {
			score = 0
		}
		if score > 100 {
			score = 100
		}
		cfg.BoostScore = score
	}
	cfg.mu.Unlock()

	// Update boost in shopping graph
	if req.BoostScore != nil && srv.graphURL != "" {
		go func() {
			body, _ := json.Marshal(map[string]any{
				"merchant_id": tenantID,
				"amount":      *req.BoostScore,
			})
			req, err := http.NewRequest(http.MethodPut, srv.graphURL+"/boost", bytes.NewReader(body))
			if err != nil {
				return
			}
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Printf("update boost: %v", err)
				return
			}
			resp.Body.Close()
		}()
	}

	// Notify presenter of config change
	srv.presenterN.Send(SaleEvent{Type: "config_update", OrderID: tenantID})

	// Forward to obs-hub
	srv.forwardToObsHub("arena", "config_update", "Config updated: "+tenantID, nil)

	handleGetConfig(w, r, cfg, costPrice, m)
}

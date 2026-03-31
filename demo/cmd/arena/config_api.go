package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

func handleGetConfig(w http.ResponseWriter, r *http.Request, cfg *MerchantConfig, costPrice int, m *arenaMerchant) {
	cfg.mu.RLock()
	maxCPCBid := cfg.MaxCPCBid
	cfg.mu.RUnlock()

	m.mu.Lock()
	totalRevenue := m.totalRevenue
	totalAdSpend := m.totalBoostSpend
	totalUnitsSold := m.totalUnitsSold
	consultationCount := m.consultationCount
	salesCount := m.salesCount
	actualCPC := m.lastActualCPC
	m.mu.Unlock()

	netProfit := totalRevenue - (costPrice * totalUnitsSold) - totalAdSpend

	cfg.mu.RLock()
	defer cfg.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"selling_price":      cfg.SellingPrice,
		"stock":              cfg.Stock,
		"discount_codes":     cfg.DiscountCodes,
		"max_cpc_bid":        maxCPCBid,
		"shipping_options":   cfg.ShippingOptions,
		"cost_price":         costPrice,
		"actual_cpc":         actualCPC,
		"total_ad_spend":     totalAdSpend,
		"total_revenue":      totalRevenue,
		"net_profit":         netProfit,
		"consultation_count": consultationCount,
		"sales_count":        salesCount,
		"pricing_algo":       cfg.PricingAlgo,
		"accent_color":       cfg.AccentColor,
		"emoji":              cfg.Emoji,
	})
}

func handlePutConfig(w http.ResponseWriter, r *http.Request, cfg *MerchantConfig, costPrice int, srv *ArenaServer, tenantID string, m *arenaMerchant) {
	var req struct {
		SellingPrice    *int             `json:"selling_price"`
		Stock           *int             `json:"stock"`
		DiscountCodes   []DiscountCode   `json:"discount_codes"`
		MaxCPCBid       *int             `json:"max_cpc_bid"`
		ShippingOptions []ShippingOption `json:"shipping_options"`
		PricingAlgo     *string          `json:"pricing_algo"`
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
	if req.MaxCPCBid != nil {
		bid := *req.MaxCPCBid
		if bid < 0 {
			bid = 0
		}
		if bid > 200 {
			bid = 200
		}
		cfg.MaxCPCBid = bid
	}
	if req.ShippingOptions != nil {
		cfg.ShippingOptions = req.ShippingOptions
	}
	if req.PricingAlgo != nil {
		cfg.PricingAlgo = *req.PricingAlgo
	}
	cfg.mu.Unlock()

	// Update bid in shopping graph
	if req.MaxCPCBid != nil && srv.graphURL != "" {
		go func() {
			body, _ := json.Marshal(map[string]any{
				"merchant_id": tenantID,
				"amount":      *req.MaxCPCBid,
			})
			req, err := http.NewRequest(http.MethodPut, srv.graphURL+"/bid", bytes.NewReader(body))
			if err != nil {
				return
			}
			req.Header.Set("Content-Type", "application/json")
			resp, err := httpClient.Do(req)
			if err != nil {
				log.Printf("update bid: %v", err)
				return
			}
			resp.Body.Close()
		}()
	}

	// Trigger shopping graph re-poll on price/stock change
	needsPoll := (req.SellingPrice != nil || req.Stock != nil) && srv.graphURL != ""
	if needsPoll {
		go func() {
			pollReq, err := http.NewRequest(http.MethodPost, srv.graphURL+"/poll/"+tenantID, nil)
			if err != nil {
				return
			}
			resp, err := httpClient.Do(pollReq)
			if err != nil {
				log.Printf("trigger re-poll: %v", err)
				return
			}
			resp.Body.Close()
			// Notify presenter AFTER re-poll completes
			srv.presenterN.Send(SaleEvent{Type: "config_update", OrderID: tenantID})
			srv.forwardToObsHub("arena", "config_update", "Config updated: "+tenantID, nil)
		}()
	} else {
		srv.presenterN.Send(SaleEvent{Type: "config_update", OrderID: tenantID})
		srv.forwardToObsHub("arena", "config_update", "Config updated: "+tenantID, nil)
	}

	handleGetConfig(w, r, cfg, costPrice, m)
}

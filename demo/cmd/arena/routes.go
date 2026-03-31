package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"strings"
	"unicode"
)

func (s *ArenaServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, `{"detail":"name is required"}`, http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		http.Error(w, `{"detail":"name is required"}`, http.StatusBadRequest)
		return
	}
	if len([]rune(name)) > 30 {
		http.Error(w, `{"detail":"30 caractères max"}`, http.StatusBadRequest)
		return
	}
	for _, c := range name {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != ' ' && c != '-' && c != '_' {
			http.Error(w, `{"detail":"caractères invalides"}`, http.StatusBadRequest)
			return
		}
	}
	if s.HasTenantNamed(name) {
		http.Error(w, `{"detail":"nom déjà pris"}`, http.StatusConflict)
		return
	}

	tenant := s.RegisterTenant(name)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"id":        tenant.ID,
		"url":       tenant.Merchant.baseURL(),
		"dashboard": tenant.Merchant.baseURL() + "/dashboard",
	})
}

var autoNameAdjectives = []string{"Cyber", "Mega", "Turbo", "Hyper", "Ultra", "Pixel", "Neon", "Star", "Zen", "Flash"}
var autoNameNouns = []string{"Shop", "Deals", "Bazar", "Market", "Store", "Corner", "Hub", "Express", "Zone", "Outlet"}

func (s *ArenaServer) handleAuto(w http.ResponseWriter, r *http.Request) {
	name := fmt.Sprintf("%s%s%d",
		autoNameAdjectives[rand.IntN(len(autoNameAdjectives))],
		autoNameNouns[rand.IntN(len(autoNameNouns))],
		rand.IntN(100),
	)
	tenant := s.RegisterTenant(name)
	http.Redirect(w, r, "/"+tenant.ID+"/dashboard", http.StatusSeeOther)
}

func (s *ArenaServer) handleListMerchants(w http.ResponseWriter, r *http.Request) {
	tenants := s.ListTenants()

	type merchantInfo struct {
		ID                string `json:"id"`
		Name              string `json:"name"`
		Price             int    `json:"price"`
		Stock             int    `json:"stock"`
		MaxCPCBid         int    `json:"max_cpc_bid"`
		Margin            int    `json:"margin"`
		ActualCPC         int    `json:"actual_cpc"`
		TotalAdSpend      int    `json:"total_ad_spend"`
		NetProfit         int    `json:"net_profit"`
		SalesCount        int    `json:"sales_count"`
		ConsultationCount int    `json:"consultation_count"`
		AccentColor       string `json:"accent_color"`
		Emoji             string `json:"emoji"`
	}

	result := make([]merchantInfo, 0, len(tenants))
	for _, t := range tenants {
		t.Config.mu.RLock()
		price := t.Config.SellingPrice
		stock := t.Config.Stock
		maxCPCBid := t.Config.MaxCPCBid
		accentColor := t.Config.AccentColor
		emoji := t.Config.Emoji
		t.Config.mu.RUnlock()

		margin := price - s.costPrice

		t.Merchant.mu.Lock()
		totalRevenue := t.Merchant.totalRevenue
		totalAdSpend := t.Merchant.totalBoostSpend
		totalUnitsSold := t.Merchant.totalUnitsSold
		consultationCount := t.Merchant.consultationCount
		salesCount := t.Merchant.salesCount
		actualCPC := t.Merchant.lastActualCPC
		t.Merchant.mu.Unlock()

		netProfit := totalRevenue - (s.costPrice * totalUnitsSold) - totalAdSpend

		result = append(result, merchantInfo{
			ID:                t.ID,
			Name:              t.Name,
			Price:             price,
			Stock:             stock,
			MaxCPCBid:         maxCPCBid,
			Margin:            margin,
			ActualCPC:         actualCPC,
			TotalAdSpend:      totalAdSpend,
			NetProfit:         netProfit,
			SalesCount:        salesCount,
			ConsultationCount: consultationCount,
			AccentColor:       accentColor,
			Emoji:             emoji,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"merchants": result,
		"total":     len(result),
	})
}

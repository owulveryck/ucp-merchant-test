package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive"
	"github.com/owulveryck/ucp-merchant-test/pkg/model"
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
		log.Printf("[handlePutConfig] ERROR decoding request: %v", err)
		http.Error(w, `{"detail":"invalid request"}`, http.StatusBadRequest)
		return
	}

	log.Printf("[handlePutConfig] Received: selling_price=%v, pricing_algo=%v", req.SellingPrice, req.PricingAlgo)

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

func handleTestAutoCompete(w http.ResponseWriter, r *http.Request, m *arenaMerchant, tenantID string) {
	w.Header().Set("Content-Type", "application/json")

	// Get pricing agent
	if m.pricingAgent == nil {
		log.Printf("[TestAutoCompete] No pricing agent configured")
		http.Error(w, `{"error":"competitive pricing not enabled"}`, http.StatusInternalServerError)
		return
	}

	// Get product price
	m.config.mu.RLock()
	ourPrice := m.config.SellingPrice
	m.config.mu.RUnlock()

	// Create a fake line item to trigger the agent
	lineItems := []model.LineItem{
		{
			Item: model.Item{
				ID:    m.product.ID,
				Price: ourPrice,
			},
			Quantity: 1,
		},
	}

	// Call the pricing agent directly
	log.Printf("[TestAutoCompete] Calling pricing agent with price %d", ourPrice)
	result := m.pricingAgent.ApplyDiscountsWithContext([]string{"AUTO_COMPETE"}, lineItems)

	// Get agent decisions (type assert to DiscountAdapter)
	var decisions *competitive.AgentDecisions
	if adapter, ok := m.pricingAgent.(*competitive.DiscountAdapter); ok {
		decisions = adapter.GetLastDecisions()
	}

	if result == nil || len(result.Applied) == 0 {
		log.Printf("[TestAutoCompete] No discount returned")

		// Format "no discount" messages
		agent1Msg := "Aucun concurrent trouvé"
		agent2Msg := "Position inconnue"
		agent3Msg := "Prix actuel optimal"
		agent4Msg := "✅ Prix maintenu"
		mainMessage := "Vous avez déjà le meilleur prix !"

		if decisions != nil {
			intel := decisions.Intel
			val := decisions.Validation

			// Agent 1: Show market position
			competitorCount := intel.TotalCount - 1
			agent1Msg = fmt.Sprintf("Trouvé %d concurrent(s). Le moins cher: $%.2f (rang %d/%d)",
				competitorCount, float64(intel.LowestPrice)/100, intel.OurRank, intel.TotalCount)

			// Agent 2: Market position
			if intel.OurRank == 1 {
				agent2Msg = "✅ Vous êtes le moins cher !"
			} else {
				agent2Msg = fmt.Sprintf("⚠️ Vous êtes en position %d/%d", intel.OurRank, intel.TotalCount)
			}

			// Agent 3: Strategy
			if intel.OurRank == 1 {
				agent3Msg = "Vous avez déjà le meilleur prix !"
				mainMessage = "Vous avez déjà le meilleur prix !"
			} else {
				agent3Msg = fmt.Sprintf("Il faut baisser à $%.2f pour gagner", float64(intel.LowestPrice-100)/100)
				mainMessage = "Baissez votre prix pour gagner !"
			}

			// Agent 4: Why no discount?
			if val.Rejected {
				// Validation rejected the discount
				agent4Msg = fmt.Sprintf("❌ %s", val.RejectionReason)
			} else if intel.OurRank == 1 {
				// Already cheapest
				agent4Msg = "✅ Vous êtes déjà le meilleur"
			} else {
				// Unknown reason
				agent4Msg = "⚠️ Impossible de calculer le prix optimal"
			}
		}

		json.NewEncoder(w).Encode(map[string]any{
			"success":         true,
			"no_discount":     true,
			"message":         mainMessage,
			"current_price":   ourPrice,
			"final_price":     ourPrice,
			"discount_amount": 0,
			"margin_percent":  0,
			"reasoning": map[string]string{
				"agent1": agent1Msg,
				"agent2": agent2Msg,
				"agent3": agent3Msg,
				"agent4": agent4Msg,
			},
		})
		return
	}

	// Calculate final price
	discountAmount := 0
	for _, disc := range result.Applied {
		discountAmount += disc.Amount
	}
	finalPrice := ourPrice - discountAmount
	marginPercent := 0
	if finalPrice > 0 {
		marginPercent = ((finalPrice - m.costPrice) * 100) / finalPrice
	}

	log.Printf("[TestAutoCompete] Success: final price %d, discount %d, margin %d%%", finalPrice, discountAmount, marginPercent)

	// Format agent decisions into simple French messages
	agent1Msg := "Analyse en cours..."
	agent2Msg := "Analyse en cours..."
	agent3Msg := "Calcul du prix..."
	agent4Msg := "Validation..."

	if decisions != nil {
		intel := decisions.Intel
		insight := decisions.Insight
		rec := decisions.Recommendation
		val := decisions.Validation

		// Agent 1: Price Intelligence
		competitorCount := intel.TotalCount - 1 // Exclude ourselves
		agent1Msg = fmt.Sprintf("Trouvé %d concurrent(s). Le moins cher: $%.2f (rang %d/%d)",
			competitorCount, float64(intel.LowestPrice)/100, intel.OurRank, intel.TotalCount)

		// Agent 2: Market Analysis
		positionMsg := "position moyenne"
		if insight.Position == "expensive" {
			positionMsg = "trop cher"
		} else if insight.Position == "cheap" {
			positionMsg = "bon prix"
		}
		agent2Msg = fmt.Sprintf("Vous êtes %s. %s", positionMsg, insight.Reasoning)

		// Agent 3: Strategy Recommender
		strategyName := "compétitive"
		if rec.Strategy == "aggressive" {
			strategyName = "agressive"
		}
		reasoningText := ""
		if len(rec.Reasoning) > 0 {
			reasoningText = rec.Reasoning[0]
		}
		agent3Msg = fmt.Sprintf("Stratégie %s<br>Prix cible: <strong>$%.2f</strong><br>%s",
			strategyName, float64(rec.TargetPrice)/100, reasoningText)

		// Agent 4: Margin Validator
		if val.Approved && !val.Rejected {
			if len(val.Warnings) > 0 {
				// Adjusted but approved
				agent4Msg = fmt.Sprintf("⚠️ Prix ajusté à $%.2f pour respecter %d%% de marge minimum",
					float64(val.FinalPrice)/100, val.Margin)
			} else {
				// Approved as-is
				agent4Msg = fmt.Sprintf("✅ Validé ! Vous gagnerez %d%% de marge",
					val.Margin)
			}
		} else if val.Rejected {
			agent4Msg = fmt.Sprintf("❌ Rejeté: %s", val.RejectionReason)
		}
	}

	// Return the calculated price with agent reasoning
	json.NewEncoder(w).Encode(map[string]any{
		"success":         true,
		"current_price":   ourPrice,
		"final_price":     finalPrice,
		"discount_amount": discountAmount,
		"margin_percent":  marginPercent,
		"reasoning": map[string]string{
			"agent1": agent1Msg,
			"agent2": agent2Msg,
			"agent3": agent3Msg,
			"agent4": agent4Msg,
		},
	})
}

package obs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/owulveryck/ucp-merchant-test/demo/internal/a2aclient"
)

// Handler provides HTTP endpoints for the observability hub.
type Handler struct {
	hub           *Hub
	catalogClient *a2aclient.Client
	graphURL      string
	arenaURL      string
}

// NewHandler creates a new HTTP handler.
func NewHandler(hub *Hub, graphURL, arenaURL string) *Handler {
	return &Handler{
		hub:           hub,
		catalogClient: a2aclient.NewClient("dashboard", "US", ""),
		graphURL:      graphURL,
		arenaURL:      arenaURL,
	}
}

// Mux returns an http.Handler with all routes registered and CORS middleware applied.
func (h *Handler) Mux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /event", h.handlePostEvent)
	mux.HandleFunc("GET /events", h.handleSSE)
	mux.HandleFunc("GET /report", h.handleReport)
	mux.HandleFunc("GET /report/json", h.handleReportJSON)
	mux.HandleFunc("POST /command", h.handlePostCommand)
	mux.HandleFunc("GET /commands", h.handleCommandsSSE)
	mux.HandleFunc("GET /catalog", h.handleCatalog)
	mux.HandleFunc("GET /graph/health", h.proxyGraph)
	mux.HandleFunc("GET /graph/ranking", h.proxyGraph)
	mux.HandleFunc("PUT /graph/ranking", h.proxyGraph)
	mux.HandleFunc("GET /arena/merchants", h.proxyArena)
	mux.HandleFunc("GET /arena/rankings", h.proxyArena)
	mux.HandleFunc("GET /arena/config", h.proxyArena)
	mux.HandleFunc("POST /arena/command", h.proxyArena)
	mux.HandleFunc("GET /status", h.handleStatus)
	mux.HandleFunc("GET /arena", h.handleArenaDashboard)
	mux.HandleFunc("GET /arena2", h.handleArena2Dashboard)
	mux.HandleFunc("GET /insights", h.handleInsightsDashboard)
	mux.HandleFunc("GET /", h.handleDashboard)
	return corsMiddleware(mux)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) handlePostEvent(w http.ResponseWriter, r *http.Request) {
	var e Event
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		http.Error(w, `{"detail":"invalid event"}`, http.StatusBadRequest)
		return
	}
	h.hub.Add(e)
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) handleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	ch := h.hub.Subscribe()
	defer h.hub.Unsubscribe(ch)

	// Send existing events first
	for _, e := range h.hub.Events() {
		data, _ := json.Marshal(e)
		fmt.Fprintf(w, "data: %s\n\n", data)
	}
	flusher.Flush()

	for {
		select {
		case e, ok := <-ch:
			if !ok {
				return
			}
			data, _ := json.Marshal(e)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (h *Handler) handlePostCommand(w http.ResponseWriter, r *http.Request) {
	var cmd Command
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil || cmd.Instruction == "" {
		http.Error(w, `{"detail":"invalid command"}`, http.StatusBadRequest)
		return
	}

	fmt.Printf("[DEBUG] Received command: %s\n", cmd.Instruction)

	// Execute buying flow in-process (no external Gemini agent)
	go h.executeBuyingFlow(cmd.Instruction)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]any{
		"status":    "accepted",
		"connected": true,
	})
}

func (h *Handler) handleCommandsSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	h.hub.IncrCommandConsumers()
	defer h.hub.DecrCommandConsumers()

	// Flush headers immediately so the client's Do() returns
	// and the SSE connection is fully established.
	fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	for {
		select {
		case cmd := <-h.hub.Commands():
			data, _ := json.Marshal(cmd)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (h *Handler) handleReportJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(h.hub.Report())
}

func (h *Handler) handleReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(reportHTML))
}

func (h *Handler) handleCatalog(w http.ResponseWriter, r *http.Request) {
	portStr := r.URL.Query().Get("port")
	if portStr == "" {
		http.Error(w, `{"detail":"missing port param"}`, http.StatusBadRequest)
		return
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		http.Error(w, `{"detail":"invalid port"}`, http.StatusBadRequest)
		return
	}

	result, err := h.catalogClient.SendAction("http://localhost:"+portStr, "list_products", map[string]any{"limit": float64(50)})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{"detail": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *Handler) proxyGraph(w http.ResponseWriter, r *http.Request) {
	// Strip "/graph" prefix to get the shopping graph path
	target := r.URL.Path[len("/graph"):]
	url := h.graphURL + target

	req, err := http.NewRequestWithContext(r.Context(), r.Method, url, r.Body)
	if err != nil {
		http.Error(w, `{"detail":"proxy error"}`, http.StatusBadGateway)
		return
	}
	req.Header.Set("Content-Type", r.Header.Get("Content-Type"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, `{"detail":"graph unreachable"}`, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (h *Handler) proxyArena(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Path[len("/arena"):]
	url := h.arenaURL + target

	req, err := http.NewRequestWithContext(r.Context(), r.Method, url, r.Body)
	if err != nil {
		http.Error(w, `{"detail":"proxy error"}`, http.StatusBadGateway)
		return
	}
	req.Header.Set("Content-Type", r.Header.Get("Content-Type"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, `{"detail":"arena unreachable"}`, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (h *Handler) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	// Always return connected=true since we handle buying in-process (no external Gemini agent)
	json.NewEncoder(w).Encode(map[string]any{
		"agent_connected": true,
	})
}

func (h *Handler) executeBuyingFlow(instruction string) {
	fmt.Printf("[DEBUG] executeBuyingFlow started for: %s\n", instruction)

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[DEBUG] executeBuyingFlow panic: %v\n", r)
			h.hub.Add(Event{
				Type:    "agent_error",
				Source:  "buying_agent",
				Summary: "Buying flow crashed",
				Data:    map[string]any{"error": fmt.Sprintf("%v", r)},
			})
		}
	}()

	fmt.Printf("[DEBUG] Adding first event\n")
	// Step 1: Search Shopping Graph
	h.hub.Add(Event{
		Type:    "agent_message",
		Source:  "buying_agent",
		Summary: "Searching for products",
		Data:    map[string]any{"message": "🔍 Recherche: " + instruction},
	})

	fmt.Printf("[DEBUG] Searching Shopping Graph at %s\n", h.graphURL)
	searchBody := fmt.Sprintf(`{"query": "casque", "limit": 10}`)
	searchResp, err := http.Post(h.graphURL+"/search", "application/json",
		strings.NewReader(searchBody))
	if err != nil {
		fmt.Printf("[DEBUG] Search error: %v\n", err)
		h.hub.Add(Event{
			Type:    "agent_message",
			Source:  "buying_agent",
			Summary: "Error searching",
			Data:    map[string]any{"message": "❌ Shopping Graph indisponible: " + err.Error()},
		})
		return
	}
	defer searchResp.Body.Close()

	fmt.Printf("[DEBUG] Parsing search results\n")
	var results struct {
		Results []struct {
			MerchantID   string `json:"merchant_id"`
			MerchantName string `json:"merchant_name"`
			Price        int    `json:"price"`
			InStock      bool   `json:"in_stock"`
		} `json:"results"`
	}
	json.NewDecoder(searchResp.Body).Decode(&results)
	fmt.Printf("[DEBUG] Found %d results\n", len(results.Results))

	// Step 2: Find cheapest and collect all prices for comparison
	var cheapest *struct {
		MerchantID   string
		MerchantName string
		Price        int
	}
	var allMerchants []struct {
		Name  string
		Price int
	}

	for _, r := range results.Results {
		if r.InStock {
			allMerchants = append(allMerchants, struct {
				Name  string
				Price int
			}{r.MerchantName, r.Price})

			if cheapest == nil || r.Price < cheapest.Price {
				cheapest = &struct {
					MerchantID   string
					MerchantName string
					Price        int
				}{r.MerchantID, r.MerchantName, r.Price}
			}
		}
	}

	fmt.Printf("[DEBUG] Found %d merchants, cheapest: %v\n", len(allMerchants), cheapest)

	if cheapest == nil {
		h.hub.Add(Event{
			Type:    "agent_message",
			Source:  "buying_agent",
			Summary: "No merchants available",
			Data:    map[string]any{"message": "❌ Aucun marchand disponible"},
		})
		return
	}

	fmt.Printf("[DEBUG] Building comparison message\n")
	// Build comparison message
	var comparison strings.Builder
	comparison.WriteString(fmt.Sprintf("📊 Comparaison des prix :\n"))
	for _, m := range allMerchants {
		if m.Name == cheapest.MerchantName {
			comparison.WriteString(fmt.Sprintf("   • %s: $%.2f ← ✅ LE MOINS CHER\n", m.Name, float64(m.Price)/100))
		} else {
			diff := float64(m.Price-cheapest.Price) / 100
			comparison.WriteString(fmt.Sprintf("   • %s: $%.2f (+$%.2f)\n", m.Name, float64(m.Price)/100, diff))
		}
	}

	fmt.Printf("[DEBUG] Sending comparison event\n")
	h.hub.Add(Event{
		Type:    "agent_message",
		Source:  "buying_agent",
		Summary: "Comparing prices",
		Data:    map[string]any{"message": comparison.String()},
	})
	fmt.Printf("[DEBUG] Comparison event sent\n")

	// Decision message
	savingsPercent := 0.0
	if len(allMerchants) > 1 {
		var secondCheapest int
		for _, m := range allMerchants {
			if m.Name != cheapest.MerchantName {
				if secondCheapest == 0 || m.Price < secondCheapest {
					secondCheapest = m.Price
				}
			}
		}
		if secondCheapest > 0 {
			savingsPercent = float64(secondCheapest-cheapest.Price) / float64(secondCheapest) * 100
		}
	}

	decisionMsg := fmt.Sprintf("🎯 DÉCISION : %s sélectionné !\n\n"+
		"Pourquoi ?\n"+
		"   • Prix le plus bas : $%.2f\n"+
		"   • %d concurrent(s) comparé(s)\n",
		cheapest.MerchantName,
		float64(cheapest.Price)/100,
		len(allMerchants)-1)

	if savingsPercent > 0 {
		decisionMsg += fmt.Sprintf("   • Économie : %.1f%% vs 2ème meilleur prix\n", savingsPercent)
	}

	h.hub.Add(Event{
		Type:    "agent_message",
		Source:  "buying_agent",
		Summary: "Decision made",
		Data:    map[string]any{"message": decisionMsg},
	})

	// Step 3: Create checkout with AUTO_COMPETE
	h.hub.Add(Event{
		Type:    "agent_message",
		Source:  "buying_agent",
		Summary: "Creating checkout",
		Data:    map[string]any{"message": "🛒 Création du panier..."},
	})

	checkoutBody := `{
		"items": [{"product_id": "casque_audio", "quantity": 1}],
		"customer": {"email": "agent@acheteur.com"},
		"discount_codes": ["AUTO_COMPETE"]
	}`
	checkoutResp, err := http.Post(h.arenaURL+"/"+cheapest.MerchantID+"/checkouts",
		"application/json", strings.NewReader(checkoutBody))
	if err == nil {
		defer checkoutResp.Body.Close()
		var checkout struct {
			Totals []struct {
				Type   string `json:"type"`
				Amount int    `json:"amount"`
			} `json:"totals"`
		}
		json.NewDecoder(checkoutResp.Body).Decode(&checkout)

		// Find final total
		var finalPrice int
		for _, t := range checkout.Totals {
			if t.Type == "total" {
				finalPrice = t.Amount
				break
			}
		}

		if finalPrice > 0 {
			h.hub.Add(Event{
				Type:    "agent_message",
				Source:  "buying_agent",
				Summary: "Purchase completed",
				Data: map[string]any{
					"message": fmt.Sprintf("✅ Achat confirmé ! Prix final: $%.2f (avec AUTO_COMPETE)", float64(finalPrice)/100),
				},
			})
		}
	}

	// Highlight the winner in the arena
	h.hub.Add(Event{
		Type:    "merchant_selected",
		Source:  "buying_agent",
		Summary: fmt.Sprintf("Selected %s as winner", cheapest.MerchantName),
		Data: map[string]any{
			"merchant_id":   cheapest.MerchantID,
			"merchant_name": cheapest.MerchantName,
			"price":         cheapest.Price,
		},
	})
}

const reportHTML = `<!DOCTYPE html>
<html><head><title>Demo Report</title>
<link href="https://fonts.googleapis.com/css2?family=Outfit:wght@400;600;700;800&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
<style>body{font-family:'Outfit',system-ui,sans-serif;max-width:800px;margin:2rem auto;padding:0 1rem;background:#FDF0EE;color:#1A1A2E}
h1{font-weight:800;color:#1A1A2E}
pre{font-family:'JetBrains Mono',monospace;background:#FFFFFF;padding:1.5rem;border-radius:16px;border:1px solid #2D2D2D;box-shadow:6px 6px 0px #E5004C;overflow-x:auto;color:#2D2D2D}</style>
</head><body><h1>Demo Report</h1><pre id="report">Loading...</pre>
<script>fetch('/report/json').then(r=>r.json()).then(d=>document.getElementById('report').textContent=JSON.stringify(d,null,2))</script>
</body></html>`

package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// HandleCallAgent proxies requests to agents.
func (h *DashboardHandler) HandleCallAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Agent  string                 `json:"agent"`
		Method string                 `json:"method"`
		Params map[string]interface{} `json:"params"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Determine agent URL
	var agentURL string
	switch req.Agent {
	case "customer-growth":
		agentURL = h.customerGrowthURL + "/a2a"
	case "competitiveness":
		agentURL = h.competitivenessURL + "/a2a"
	default:
		http.Error(w, "Unknown agent", http.StatusBadRequest)
		return
	}

	// Build JSON-RPC request
	jsonrpcReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  req.Method,
		"params":  req.Params,
		"id":      1,
	}

	// Marshal request
	reqBody, err := json.Marshal(jsonrpcReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Call agent
	resp, err := http.Post(agentURL, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.Write(respBody)
}

// HandleAgentsStatus checks the health of all agents.
func (h *DashboardHandler) HandleAgentsStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"customer_growth":  checkAgentHealth(h.customerGrowthURL),
		"competitiveness": checkAgentHealth(h.competitivenessURL),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func checkAgentHealth(baseURL string) map[string]interface{} {
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		return map[string]interface{}{
			"status": "offline",
			"error":  err.Error(),
		}
	}
	defer resp.Body.Close()

	var health map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}
	}

	return health
}

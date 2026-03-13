package obs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

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
	mux.HandleFunc("GET /arena/config", h.proxyArena)
	mux.HandleFunc("POST /arena/command", h.proxyArena)
	mux.HandleFunc("GET /status", h.handleStatus)
	mux.HandleFunc("GET /arena", h.handleArenaDashboard)
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
	h.hub.SendCommand(cmd)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]any{
		"status":    "accepted",
		"connected": h.hub.HasCommandConsumer(),
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
	json.NewEncoder(w).Encode(map[string]any{
		"agent_connected": h.hub.HasCommandConsumer(),
	})
}

const reportHTML = `<!DOCTYPE html>
<html><head><title>Demo Report</title>
<style>body{font-family:system-ui;max-width:800px;margin:2rem auto;padding:0 1rem}
pre{background:#f5f5f5;padding:1rem;border-radius:8px;overflow-x:auto}</style>
</head><body><h1>Demo Report</h1><pre id="report">Loading...</pre>
<script>fetch('/report/json').then(r=>r.json()).then(d=>document.getElementById('report').textContent=JSON.stringify(d,null,2))</script>
</body></html>`

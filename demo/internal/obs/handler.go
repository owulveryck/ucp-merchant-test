package obs

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Handler provides HTTP endpoints for the observability hub.
type Handler struct {
	hub *Hub
}

// NewHandler creates a new HTTP handler.
func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
}

// Mux returns an http.ServeMux with all routes registered.
func (h *Handler) Mux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /event", h.handlePostEvent)
	mux.HandleFunc("GET /events", h.handleSSE)
	mux.HandleFunc("GET /report", h.handleReport)
	mux.HandleFunc("GET /report/json", h.handleReportJSON)
	mux.HandleFunc("GET /", h.handleDashboard)
	return mux
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
	w.Header().Set("Access-Control-Allow-Origin", "*")

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

func (h *Handler) handleReportJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(h.hub.Report())
}

func (h *Handler) handleReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(reportHTML))
}

const reportHTML = `<!DOCTYPE html>
<html><head><title>Demo Report</title>
<style>body{font-family:system-ui;max-width:800px;margin:2rem auto;padding:0 1rem}
pre{background:#f5f5f5;padding:1rem;border-radius:8px;overflow-x:auto}</style>
</head><body><h1>Demo Report</h1><pre id="report">Loading...</pre>
<script>fetch('/report/json').then(r=>r.json()).then(d=>document.getElementById('report').textContent=JSON.stringify(d,null,2))</script>
</body></html>`

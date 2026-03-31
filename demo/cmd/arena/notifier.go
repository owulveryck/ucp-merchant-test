package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// SaleEvent represents a completed sale notification.
type SaleEvent struct {
	Type              string `json:"type"`
	MerchantID        string `json:"merchant_id,omitempty"`
	OrderID           string `json:"order_id"`
	Buyer             string `json:"buyer"`
	Total             int    `json:"total"`
	TotalRevenue      int    `json:"total_revenue,omitempty"`
	TotalAdSpend      int    `json:"total_ad_spend,omitempty"`
	ConsultationCount int    `json:"consultation_count,omitempty"`
	NetProfit         int    `json:"net_profit,omitempty"`
}

// Notifier broadcasts SSE events to connected clients.
type Notifier struct {
	mu          sync.Mutex
	subscribers map[chan []byte]struct{}
}

// NewNotifier creates a new SSE notifier.
func NewNotifier() *Notifier {
	return &Notifier{
		subscribers: make(map[chan []byte]struct{}),
	}
}

// Send broadcasts an event to all subscribers.
func (n *Notifier) Send(event SaleEvent) {
	data, _ := json.Marshal(event)
	n.mu.Lock()
	defer n.mu.Unlock()
	for ch := range n.subscribers {
		select {
		case ch <- data:
		default:
			log.Printf("notifier: dropped event (type=%s) for slow subscriber", event.Type)
		}
	}
}

// SendRaw broadcasts raw JSON bytes to all subscribers.
func (n *Notifier) SendRaw(data []byte) {
	n.mu.Lock()
	defer n.mu.Unlock()
	for ch := range n.subscribers {
		select {
		case ch <- data:
		default:
			log.Printf("notifier: dropped raw event for slow subscriber")
		}
	}
}

// ServeHTTP handles SSE connections.
func (n *Notifier) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	ch := make(chan []byte, 32)
	n.mu.Lock()
	n.subscribers[ch] = struct{}{}
	n.mu.Unlock()

	defer func() {
		n.mu.Lock()
		delete(n.subscribers, ch)
		n.mu.Unlock()
	}()

	// Send keepalive
	fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	for {
		select {
		case data := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

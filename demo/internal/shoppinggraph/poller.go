package shoppinggraph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/owulveryck/ucp-merchant-test/demo/internal/a2aclient"
)

// Poller periodically polls merchants to refresh the shopping graph.
type Poller struct {
	graph    *ShoppingGraph
	client   *a2aclient.Client
	interval time.Duration
	obsURL   string
	stop     chan struct{}
}

// NewPoller creates a new merchant poller.
func NewPoller(graph *ShoppingGraph, client *a2aclient.Client, interval time.Duration, obsURL string) *Poller {
	return &Poller{
		graph:    graph,
		client:   client,
		interval: interval,
		obsURL:   obsURL,
		stop:     make(chan struct{}),
	}
}

// Start begins polling in the background. It does an initial poll immediately.
func (p *Poller) Start() {
	p.pollAll()
	go func() {
		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				p.pollAll()
			case <-p.stop:
				return
			}
		}
	}()
}

// Stop stops the poller.
func (p *Poller) Stop() {
	close(p.stop)
}

func (p *Poller) pollAll() {
	p.graph.mu.RLock()
	merchants := make([]*MerchantNode, 0, len(p.graph.Merchants))
	for _, m := range p.graph.Merchants {
		merchants = append(merchants, m)
	}
	p.graph.mu.RUnlock()

	for _, m := range merchants {
		p.pollMerchant(m)
	}
}

func (p *Poller) emitEvent(eventType, summary string) {
	if p.obsURL == "" {
		return
	}
	event := map[string]any{
		"source":  "shopping-graph",
		"type":    eventType,
		"summary": summary,
	}
	data, _ := json.Marshal(event)
	resp, err := http.Post(p.obsURL+"/event", "application/json", bytes.NewReader(data))
	if err != nil {
		log.Printf("obs event error: %v", err)
		return
	}
	resp.Body.Close()
}

func (p *Poller) pollMerchant(m *MerchantNode) {
	result, err := p.client.SendAction(m.Endpoint, "list_products", map[string]any{
		"limit": float64(50),
	})
	if err != nil {
		log.Printf("poll %s (%s): %v", m.Name, m.Endpoint, err)
		p.emitEvent("tool_call", fmt.Sprintf("Poll failed: %s at %s", m.Name, m.Endpoint))
		p.graph.MarkOffline(m.ID)
		return
	}

	rawProducts, ok := result["products"].([]any)
	if !ok {
		log.Printf("poll %s: unexpected products format", m.Name)
		p.emitEvent("tool_call", fmt.Sprintf("Poll failed: %s at %s", m.Name, m.Endpoint))
		p.graph.MarkOffline(m.ID)
		return
	}

	var products []*ProductNode
	for _, raw := range rawProducts {
		pm, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		node := &ProductNode{
			MerchantID:   m.ID,
			MerchantName: m.Name,
			MerchantURL:  m.Endpoint,
		}
		if v, ok := pm["id"].(string); ok {
			node.ProductID = v
		}
		if v, ok := pm["title"].(string); ok {
			node.Title = v
		}
		if v, ok := pm["image_url"].(string); ok {
			node.ImageURL = v
		}
		if v, ok := pm["price"].(float64); ok {
			node.Price = int(v)
		}
		if v, ok := pm["quantity"].(float64); ok {
			node.Quantity = int(v)
		}
		products = append(products, node)
	}

	log.Printf("poll %s: %d products", m.Name, len(products))
	p.emitEvent("tool_call", fmt.Sprintf("Polled %s — %d products at %s", m.Name, len(products), m.Endpoint))
	p.graph.UpdateMerchantProducts(m.ID, products)
}

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/owulveryck/ucp-merchant-test/pkg/model"
	"github.com/owulveryck/ucp-merchant-test/pkg/ucp"
)

// EventHub broadcasts events to SSE subscribers.
type EventHub struct {
	mu          sync.Mutex
	subscribers []chan model.DashboardEvent
}

var hub = &EventHub{}

func (h *EventHub) Subscribe() chan model.DashboardEvent {
	ch := make(chan model.DashboardEvent, 64)
	h.mu.Lock()
	h.subscribers = append(h.subscribers, ch)
	h.mu.Unlock()
	return ch
}

func (h *EventHub) Unsubscribe(ch chan model.DashboardEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for i, s := range h.subscribers {
		if s == ch {
			h.subscribers = append(h.subscribers[:i], h.subscribers[i+1:]...)
			close(ch)
			return
		}
	}
}

func (h *EventHub) Publish(event model.DashboardEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, ch := range h.subscribers {
		select {
		case ch <- event:
		default:
		}
	}
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDashboardHTML())
}

func handleSSE(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Send initial snapshot
	snapshot := getSSESnapshot()

	data, _ := json.Marshal(snapshot)
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()

	ch := hub.Subscribe()
	defer hub.Unsubscribe(ch)

	for {
		select {
		case <-r.Context().Done():
			return
		case event := <-ch:
			data, _ := json.Marshal(event)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

func handleAPIProducts(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	switch r.Method {
	case http.MethodGet:
		handleAPIListProducts(w, r)
	case http.MethodPost:
		handleAPIAddProduct(w, r)
	case http.MethodPut:
		handleAPIUpdateProduct(w, r)
	case http.MethodDelete:
		handleAPIDeleteProduct(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleAPIListProducts(w http.ResponseWriter, r *http.Request) {
	catalogMu.Lock()
	products := make([]Product, len(catalog))
	copy(products, catalog)
	catalogMu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func handleAPIAddProduct(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title              string   `json:"title"`
		Category           string   `json:"category"`
		Brand              string   `json:"brand"`
		Price              int      `json:"price"`
		Quantity           int      `json:"quantity"`
		ImageURL           string   `json:"image_url"`
		Description        string   `json:"description"`
		AvailableCountries []string `json:"available_countries"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if input.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}
	catalogMu.Lock()
	productSeq++
	countries := make([]ucp.Country, len(input.AvailableCountries))
	for i, c := range input.AvailableCountries {
		countries[i] = ucp.NewCountry(c)
	}
	p := Product{
		ID:                 fmt.Sprintf("SKU-%03d", productSeq),
		Title:              input.Title,
		Category:           ucp.Category(input.Category),
		Brand:              input.Brand,
		Price:              input.Price,
		Quantity:           input.Quantity,
		ImageURL:           input.ImageURL,
		Description:        input.Description,
		AvailableCountries: countries,
	}
	catalog = append(catalog, p)
	catalogMu.Unlock()

	hub.Publish(model.DashboardEvent{
		Type:      "product_added",
		ID:        p.ID,
		Summary:   fmt.Sprintf("Product %s added: %s ($%.2f, stock: %d)", p.ID, p.Title, float64(p.Price)/100, p.Quantity),
		Timestamp: time.Now(),
		Data:      p,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func handleAPIUpdateProduct(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ID                 string   `json:"id"`
		Title              string   `json:"title"`
		Category           string   `json:"category"`
		Brand              string   `json:"brand"`
		Price              *int     `json:"price"`
		Quantity           *int     `json:"quantity"`
		ImageURL           string   `json:"image_url"`
		Description        string   `json:"description"`
		AvailableCountries []string `json:"available_countries"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	catalogMu.Lock()
	var found *Product
	for i := range catalog {
		if catalog[i].ID == input.ID {
			found = &catalog[i]
			break
		}
	}
	if found == nil {
		catalogMu.Unlock()
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}
	if input.Title != "" {
		found.Title = input.Title
	}
	if input.Category != "" {
		found.Category = ucp.Category(input.Category)
	}
	if input.Brand != "" {
		found.Brand = input.Brand
	}
	if input.Price != nil {
		found.Price = *input.Price
	}
	if input.Quantity != nil {
		found.Quantity = *input.Quantity
	}
	if input.ImageURL != "" {
		found.ImageURL = input.ImageURL
	}
	if input.Description != "" {
		found.Description = input.Description
	}
	if input.AvailableCountries != nil {
		c := make([]ucp.Country, len(input.AvailableCountries))
		for i, v := range input.AvailableCountries {
			c[i] = ucp.NewCountry(v)
		}
		found.AvailableCountries = c
	}
	updated := *found
	catalogMu.Unlock()

	hub.Publish(model.DashboardEvent{
		Type:      "product_updated",
		ID:        updated.ID,
		Summary:   fmt.Sprintf("Product %s updated: %s ($%.2f, stock: %d)", updated.ID, updated.Title, float64(updated.Price)/100, updated.Quantity),
		Timestamp: time.Now(),
		Data:      updated,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func handleAPIDeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id query param required", http.StatusBadRequest)
		return
	}

	catalogMu.Lock()
	idx := -1
	var removed Product
	for i := range catalog {
		if catalog[i].ID == id {
			idx = i
			removed = catalog[i]
			break
		}
	}
	if idx == -1 {
		catalogMu.Unlock()
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}
	catalog = append(catalog[:idx], catalog[idx+1:]...)
	catalogMu.Unlock()

	hub.Publish(model.DashboardEvent{
		Type:      "product_removed",
		ID:        removed.ID,
		Summary:   fmt.Sprintf("Product %s removed: %s", removed.ID, removed.Title),
		Timestamp: time.Now(),
		Data:      removed,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted", "id": id})
}

func countActiveCheckouts() int {
	if merchantInstance == nil {
		return 0
	}
	count := 0
	for _, co := range merchantInstance.checkouts {
		if co.Status != "completed" && co.Status != "canceled" {
			count++
		}
	}
	return count
}

func mapValues[K comparable, V any](m map[K]V) []V {
	vals := make([]V, 0, len(m))
	for _, v := range m {
		vals = append(vals, v)
	}
	return vals
}

func getSSESnapshot() map[string]interface{} {
	catalogMu.Lock()
	productsCopy := make([]Product, len(catalog))
	copy(productsCopy, catalog)
	catalogMu.Unlock()

	var cartCount, checkoutCount, orderCount int
	var cartsList, checkoutsList, ordersList interface{}
	if merchantInstance != nil {
		merchantInstance.mu.Lock()
		cartCount = len(merchantInstance.carts)
		checkoutCount = countActiveCheckoutsLocked()
		orderCount = len(merchantInstance.orders)
		cartsList = mapValues(merchantInstance.carts)
		checkoutsList = mapValues(merchantInstance.checkouts)
		ordersList = mapValues(merchantInstance.orders)
		merchantInstance.mu.Unlock()
	}

	return map[string]interface{}{
		"type":           "snapshot",
		"cart_count":     cartCount,
		"checkout_count": checkoutCount,
		"order_count":    orderCount,
		"product_count":  len(productsCopy),
		"products":       productsCopy,
		"carts":          cartsList,
		"checkouts":      checkoutsList,
		"orders":         ordersList,
	}
}

// countActiveCheckoutsLocked must be called with merchantInstance.mu held.
func countActiveCheckoutsLocked() int {
	count := 0
	for _, co := range merchantInstance.checkouts {
		if co.Status != "completed" && co.Status != "canceled" {
			count++
		}
	}
	return count
}

func getDashboardHTML() string {
	return strings.ReplaceAll(dashboardHTML, "{{MERCHANT_NAME}}", merchantName)
}

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{{MERCHANT_NAME}} - Dashboard</title>
<style>
:root{
  --bg:#0f1117;--bg-card:#1a1d27;--bg-hover:#1e2130;--border:#2a2d3a;--border-light:#353849;
  --text:#e0e0e0;--text-dim:#888;--text-bright:#fff;
  --accent:#8b5cf6;--accent-hover:#7c3aed;
  --blue:#3b82f6;--purple:#a78bfa;--cyan:#22d3ee;--amber:#f59e0b;--green:#22c55e;--red:#ef4444;--teal:#2dd4bf;--orange:#fb923c;
  --sidebar-w:220px;--header-h:56px;
  --radius:8px;--radius-sm:6px;
}
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:var(--bg);color:var(--text);min-height:100vh;overflow:hidden}

/* ── Header ── */
.header{position:fixed;top:0;left:0;right:0;height:var(--header-h);background:var(--bg-card);border-bottom:1px solid var(--border);display:flex;align-items:center;padding:0 20px;gap:12px;z-index:50}
.header-brand{display:flex;align-items:center;gap:10px;min-width:var(--sidebar-w)}
.header-brand h1{font-size:16px;font-weight:600;color:var(--text-bright)}
.live-dot{width:8px;height:8px;border-radius:50%;background:var(--green);animation:pulse 2s infinite}
.live-label{font-size:11px;color:var(--green);font-weight:500;text-transform:uppercase;letter-spacing:.5px}
@keyframes pulse{0%,100%{opacity:1}50%{opacity:.4}}
.header-spacer{flex:1}
.header-conn{display:flex;align-items:center;gap:6px;font-size:11px;color:var(--text-dim)}
.conn-dot{width:8px;height:8px;border-radius:50%;background:var(--text-dim)}
.conn-dot.ok{background:var(--green)}
.conn-dot.err{background:var(--red)}
.notif-bell{position:relative;cursor:pointer;padding:6px;border-radius:var(--radius-sm);border:none;background:transparent;color:var(--text-dim)}
.notif-bell:hover{background:var(--bg-hover);color:var(--text)}
.notif-bell svg{width:18px;height:18px;stroke:currentColor;fill:none;stroke-width:2;stroke-linecap:round;stroke-linejoin:round}
.notif-badge{position:absolute;top:2px;right:2px;min-width:14px;height:14px;border-radius:7px;background:var(--red);color:#fff;font-size:9px;font-weight:700;display:flex;align-items:center;justify-content:center;padding:0 3px}
.notif-badge:empty,.notif-badge[data-count="0"]{display:none}

/* ── Sidebar ── */
.sidebar{position:fixed;top:var(--header-h);left:0;bottom:0;width:var(--sidebar-w);background:var(--bg-card);border-right:1px solid var(--border);display:flex;flex-direction:column;padding:12px 0;z-index:40;transition:width .2s}
.nav-item{display:flex;align-items:center;gap:10px;padding:10px 20px;cursor:pointer;font-size:13px;font-weight:500;color:var(--text-dim);border-left:3px solid transparent;transition:all .15s}
.nav-item:hover{background:var(--bg-hover);color:var(--text)}
.nav-item.active{color:var(--accent);border-left-color:var(--accent);background:rgba(139,92,246,.08)}
.nav-item svg{width:18px;height:18px;stroke:currentColor;fill:none;stroke-width:2;stroke-linecap:round;stroke-linejoin:round;flex-shrink:0}
.nav-item .nav-label{white-space:nowrap;overflow:hidden}
.nav-item .nav-count{margin-left:auto;background:var(--bg);padding:1px 6px;border-radius:10px;font-size:10px;font-weight:600;min-width:20px;text-align:center}
.sidebar-footer{margin-top:auto;padding:12px 20px;border-top:1px solid var(--border);font-size:11px;color:var(--text-dim)}

/* ── Content ── */
.content{margin-left:var(--sidebar-w);margin-top:var(--header-h);height:calc(100vh - var(--header-h));overflow-y:auto;padding:20px}
.view{display:none}
.view.active{display:block}

/* ── Stat Cards ── */
.stat-row{display:grid;grid-template-columns:repeat(6,1fr);gap:12px;margin-bottom:20px}
.stat-card{background:var(--bg-card);border:1px solid var(--border);border-radius:var(--radius);padding:16px;display:flex;align-items:flex-start;gap:12px;border-left:3px solid var(--border)}
.stat-card.blue{border-left-color:var(--blue)}
.stat-card.purple{border-left-color:var(--purple)}
.stat-card.cyan{border-left-color:var(--cyan)}
.stat-card.amber{border-left-color:var(--amber)}
.stat-card.green{border-left-color:var(--green)}
.stat-card.accent{border-left-color:var(--accent)}
.stat-icon{width:36px;height:36px;border-radius:var(--radius-sm);display:flex;align-items:center;justify-content:center;flex-shrink:0}
.stat-icon svg{width:20px;height:20px;stroke:currentColor;fill:none;stroke-width:2;stroke-linecap:round;stroke-linejoin:round}
.stat-card.blue .stat-icon{background:rgba(59,130,246,.12);color:var(--blue)}
.stat-card.purple .stat-icon{background:rgba(167,139,250,.12);color:var(--purple)}
.stat-card.cyan .stat-icon{background:rgba(34,211,238,.12);color:var(--cyan)}
.stat-card.amber .stat-icon{background:rgba(245,158,11,.12);color:var(--amber)}
.stat-card.green .stat-icon{background:rgba(34,197,94,.12);color:var(--green)}
.stat-card.accent .stat-icon{background:rgba(139,92,246,.12);color:var(--accent)}
.stat-info .stat-val{font-size:28px;font-weight:700;color:var(--text-bright);font-variant-numeric:tabular-nums;line-height:1.1}
.stat-info .stat-label{font-size:11px;color:var(--text-dim);text-transform:uppercase;letter-spacing:.5px;margin-top:2px}
@keyframes valBump{0%{transform:scale(1)}50%{transform:scale(1.15)}100%{transform:scale(1)}}
.val-bump{animation:valBump .3s ease}

/* ── Panels ── */
.panel{background:var(--bg-card);border:1px solid var(--border);border-radius:var(--radius);overflow:hidden;margin-bottom:16px}
.panel-header{padding:12px 16px;border-bottom:1px solid var(--border);font-weight:600;font-size:14px;display:flex;align-items:center;gap:8px;justify-content:space-between}
.panel-header .left{display:flex;align-items:center;gap:8px}
.panel-header svg{width:16px;height:16px;stroke:currentColor;fill:none;stroke-width:2;stroke-linecap:round;stroke-linejoin:round}
.panel-body{padding:0;max-height:400px;overflow-y:auto}

/* ── Overview grid ── */
.overview-grid{display:grid;grid-template-columns:1fr 1fr;gap:16px}
.overview-grid .panel-body{max-height:300px}
.recent-orders-row{display:grid;grid-template-columns:repeat(auto-fill,minmax(200px,1fr));gap:12px;margin-bottom:16px}
.order-card{background:var(--bg-card);border:1px solid var(--border);border-radius:var(--radius);padding:14px;cursor:pointer;transition:border-color .15s}
.order-card:hover{border-color:var(--accent)}
.order-card .oc-id{font-size:12px;font-weight:600;color:var(--text-bright);margin-bottom:4px}
.order-card .oc-total{font-size:18px;font-weight:700;color:var(--text-bright);margin-bottom:6px;font-variant-numeric:tabular-nums}
.order-card .oc-meta{font-size:11px;color:var(--text-dim)}

/* ── Tables ── */
table{width:100%;border-collapse:collapse;font-size:13px}
th{text-align:left;padding:8px 12px;color:var(--text-dim);font-weight:500;font-size:11px;text-transform:uppercase;letter-spacing:.5px;border-bottom:1px solid var(--border);position:sticky;top:0;background:var(--bg-card)}
td{padding:8px 12px;border-bottom:1px solid var(--bg-hover)}
.clickable-row{cursor:pointer}
.clickable-row:hover td{background:var(--bg-hover)}
.cell-stack{display:flex;flex-direction:column;gap:1px}
.cell-stack .primary{font-weight:500;color:var(--text-bright)}
.cell-stack .secondary{font-size:11px;color:var(--text-dim)}
.thumb{width:40px;height:40px;border-radius:6px;object-fit:cover;background:var(--bg);border:1px solid var(--border)}
.thumb-placeholder{width:40px;height:40px;border-radius:6px;background:var(--bg-hover);display:flex;align-items:center;justify-content:center}
.thumb-placeholder svg{width:18px;height:18px;stroke:var(--text-dim);fill:none;stroke-width:1.5}
.stock-warn{color:var(--amber);font-weight:600}
@keyframes rowFlash{0%{background:rgba(139,92,246,.15)}100%{background:transparent}}
.row-flash{animation:rowFlash .8s ease}

/* ── Toolbar ── */
.toolbar{display:flex;align-items:center;gap:8px;padding:12px 16px;border-bottom:1px solid var(--border);flex-wrap:wrap}
.toolbar input[type="text"],.toolbar select{padding:6px 10px;border-radius:var(--radius-sm);border:1px solid var(--border);background:var(--bg);color:var(--text);font-size:12px}
.toolbar input[type="text"]{width:200px}
.toolbar input::placeholder{color:#555}

/* ── Buttons ── */
.btn{padding:6px 14px;border-radius:var(--radius-sm);border:1px solid var(--border);background:var(--bg-hover);color:var(--text);cursor:pointer;font-size:12px;font-weight:500;transition:all .15s;display:inline-flex;align-items:center;gap:4px}
.btn:hover{background:var(--border)}
.btn-primary{background:var(--accent);border-color:var(--accent);color:#fff}
.btn-primary:hover{background:var(--accent-hover)}
.btn-danger{background:var(--red);border-color:var(--red);color:#fff}
.btn-danger:hover{background:#b91c1c}
.btn-sm{padding:4px 10px;font-size:11px}
.btn-ghost{background:transparent;border-color:transparent}
.btn-ghost:hover{background:var(--bg-hover)}
.qty-btn{width:26px;height:26px;border-radius:4px;border:1px solid var(--border);background:var(--bg);color:var(--text);cursor:pointer;font-size:14px;display:inline-flex;align-items:center;justify-content:center;transition:all .15s}
.qty-btn:hover{background:var(--bg-hover);border-color:var(--accent)}

/* ── Sub-tabs ── */
.sub-tabs{display:flex;gap:0;border-bottom:1px solid var(--border);padding:0 16px}
.sub-tab{padding:10px 16px;font-size:13px;font-weight:500;color:var(--text-dim);cursor:pointer;border-bottom:2px solid transparent;transition:all .15s}
.sub-tab:hover{color:var(--text)}
.sub-tab.active{color:var(--accent);border-bottom-color:var(--accent)}
.sub-tab .count{font-size:11px;background:var(--bg);padding:1px 6px;border-radius:8px;margin-left:4px}
.sub-content{display:none}
.sub-content.active{display:block}

/* ── Status badges ── */
.status{display:inline-block;padding:2px 8px;border-radius:10px;font-size:11px;font-weight:500}
.status-active,.status-incomplete{background:rgba(59,130,246,.15);color:var(--blue)}
.status-ready_for_complete{background:rgba(34,197,94,.15);color:#4ade80}
.status-completed,.status-confirmed{background:rgba(34,197,94,.15);color:var(--green)}
.status-canceled{background:rgba(239,68,68,.15);color:var(--red)}
.status-processing{background:rgba(59,130,246,.15);color:var(--blue)}
.status-shipped{background:rgba(45,212,191,.15);color:var(--teal)}
.status-in_transit{background:rgba(245,158,11,.15);color:var(--amber)}
.status-out_for_delivery{background:rgba(251,146,60,.15);color:var(--orange)}
.status-delivered{background:rgba(34,197,94,.15);color:var(--green)}

/* ── Country badges ── */
.country-badge{display:inline-block;padding:2px 8px;border-radius:10px;font-size:11px;font-weight:500}
.country-worldwide{background:rgba(34,197,94,.15);color:var(--green)}
.country-restricted{background:rgba(245,158,11,.15);color:var(--amber)}

/* ── User badge ── */
.user-badge{display:inline-flex;align-items:center;gap:3px;padding:1px 6px;border-radius:10px;font-size:11px;font-weight:500}
.user-badge.authenticated{background:rgba(74,222,128,.12);color:#4ade80}
.user-badge.guest{background:var(--bg);color:var(--text-dim)}

/* ── Order timeline ── */
.timeline{display:flex;align-items:center;gap:0;padding:8px 0;justify-content:center}
.tl-step{display:flex;flex-direction:column;align-items:center;position:relative;flex:1;max-width:100px}
.tl-dot{width:24px;height:24px;border-radius:50%;border:2px solid var(--border);background:var(--bg);display:flex;align-items:center;justify-content:center;z-index:1}
.tl-dot svg{width:12px;height:12px;stroke:currentColor;fill:none;stroke-width:2.5}
.tl-dot.done{background:var(--green);border-color:var(--green);color:#fff}
.tl-dot.current{border-color:var(--accent);color:var(--accent);animation:pulse 2s infinite}
.tl-dot.canceled{background:var(--red);border-color:var(--red);color:#fff}
.tl-line{position:absolute;top:12px;left:calc(50% + 12px);right:calc(-50% + 12px);height:2px;background:var(--border)}
.tl-line.done{background:var(--green)}
.tl-label{font-size:9px;color:var(--text-dim);margin-top:4px;text-align:center;text-transform:uppercase;letter-spacing:.3px}
.tl-step.done .tl-label{color:var(--green)}
.tl-step.current .tl-label{color:var(--accent);font-weight:600}

/* ── Activity feed ── */
.activity-card{display:flex;gap:10px;padding:10px 16px;border-bottom:1px solid var(--bg-hover);align-items:flex-start;transition:background .15s}
.activity-card:hover{background:var(--bg-hover)}
.act-icon{width:32px;height:32px;border-radius:var(--radius-sm);display:flex;align-items:center;justify-content:center;flex-shrink:0}
.act-icon svg{width:16px;height:16px;stroke:currentColor;fill:none;stroke-width:2;stroke-linecap:round;stroke-linejoin:round}
.act-icon.cart{background:rgba(59,130,246,.12);color:var(--blue)}
.act-icon.checkout{background:rgba(167,139,250,.12);color:var(--purple)}
.act-icon.order{background:rgba(34,197,94,.12);color:var(--green)}
.act-icon.product{background:rgba(245,158,11,.12);color:var(--amber)}
.act-icon.error{background:rgba(239,68,68,.12);color:var(--red)}
.act-body{flex:1;min-width:0}
.act-summary{font-size:13px;color:var(--text);line-height:1.4}
.act-time{font-size:11px;color:var(--text-dim);margin-top:2px}
.filter-bar{display:flex;gap:4px;padding:8px 16px;border-bottom:1px solid var(--border);flex-wrap:wrap}
.filter-btn{padding:4px 12px;border-radius:14px;border:1px solid var(--border);background:transparent;color:var(--text-dim);cursor:pointer;font-size:11px;font-weight:500;transition:all .15s}
.filter-btn:hover{border-color:var(--text-dim)}
.filter-btn.active{background:var(--accent);border-color:var(--accent);color:#fff}

/* ── Modal ── */
.modal-overlay{position:fixed;top:0;left:0;right:0;bottom:0;background:rgba(0,0,0,.5);backdrop-filter:blur(8px);display:flex;align-items:center;justify-content:center;z-index:100;animation:fadeIn .15s ease}
@keyframes fadeIn{from{opacity:0}to{opacity:1}}
@keyframes modalIn{from{opacity:0;transform:scale(.95) translateY(10px)}to{opacity:1;transform:scale(1) translateY(0)}}
.modal-card{background:var(--bg-card);border:1px solid var(--border);border-radius:12px;max-width:560px;width:90%;max-height:80vh;overflow-y:auto;position:relative;animation:modalIn .2s ease}
.modal-header{padding:16px 20px;border-bottom:1px solid var(--border);display:flex;align-items:center;justify-content:space-between}
.modal-header h2{font-size:16px;font-weight:600;color:var(--text-bright)}
.modal-close{background:none;border:none;color:var(--text-dim);font-size:20px;cursor:pointer;padding:4px 8px;border-radius:4px}
.modal-close:hover{color:var(--text-bright);background:var(--bg-hover)}
.modal-body{padding:16px 20px}
.modal-body .section{margin-bottom:16px}
.modal-body .section-title{font-size:11px;text-transform:uppercase;letter-spacing:1px;color:var(--text-dim);margin-bottom:8px;font-weight:500}
.modal-body table{margin-bottom:8px}
.detail-row{display:flex;justify-content:space-between;padding:4px 0;font-size:13px}
.detail-row .label{color:var(--text-dim)}
.detail-row .val{color:var(--text);font-weight:500}
.modal-footer{padding:12px 20px;border-top:1px solid var(--border);display:flex;justify-content:flex-end;gap:8px}

/* ── Add/Edit modal form ── */
.form-grid{display:grid;grid-template-columns:1fr 1fr;gap:12px}
.form-group{display:flex;flex-direction:column;gap:4px}
.form-group.full{grid-column:1/-1}
.form-group label{font-size:11px;color:var(--text-dim);text-transform:uppercase;letter-spacing:.5px;font-weight:500}
.form-group input,.form-group select,.form-group textarea{padding:8px 10px;border-radius:var(--radius-sm);border:1px solid var(--border);background:var(--bg);color:var(--text);font-size:13px;font-family:inherit}
.form-group input:focus,.form-group select:focus,.form-group textarea:focus{outline:none;border-color:var(--accent)}
.form-group textarea{resize:vertical;min-height:60px}
.form-group .img-preview{width:60px;height:60px;border-radius:var(--radius-sm);object-fit:cover;border:1px solid var(--border);margin-top:4px}

/* ── Toast ── */
.toast-container{position:fixed;bottom:20px;right:20px;display:flex;flex-direction:column-reverse;gap:8px;z-index:200;pointer-events:none}
.toast{background:var(--bg-card);border:1px solid var(--border);border-radius:var(--radius);padding:12px 16px;min-width:280px;max-width:380px;box-shadow:0 8px 24px rgba(0,0,0,.4);pointer-events:auto;animation:toastIn .3s ease;display:flex;gap:10px;align-items:flex-start}
@keyframes toastIn{from{opacity:0;transform:translateX(40px)}to{opacity:1;transform:translateX(0)}}
@keyframes toastOut{from{opacity:1;transform:translateX(0)}to{opacity:0;transform:translateX(40px)}}
.toast.removing{animation:toastOut .3s ease forwards}
.toast-icon{flex-shrink:0;width:20px;height:20px}
.toast-icon svg{width:20px;height:20px;stroke:currentColor;fill:none;stroke-width:2}
.toast-body{flex:1;font-size:12px;color:var(--text);line-height:1.4}
.toast-close{background:none;border:none;color:var(--text-dim);cursor:pointer;padding:0;font-size:14px;line-height:1}

/* ── Empty state ── */
.empty{padding:32px;text-align:center;color:var(--text-dim);font-size:13px}

/* ── Responsive ── */
@media(max-width:1024px){
  .sidebar{width:56px}
  .sidebar .nav-label,.sidebar .nav-count,.sidebar-footer{display:none}
  .nav-item{padding:10px 0;justify-content:center}
  .content{margin-left:56px}
  .stat-row{grid-template-columns:repeat(3,1fr)}
  .header-brand{min-width:56px}
}
@media(max-width:768px){
  .stat-row{grid-template-columns:repeat(2,1fr)}
  .overview-grid{grid-template-columns:1fr}
}
</style>
</head>
<body>

<!-- Header -->
<div class="header">
  <div class="header-brand">
    <h1>{{MERCHANT_NAME}}</h1>
    <span class="live-dot"></span>
    <span class="live-label">Live</span>
  </div>
  <div class="header-spacer"></div>
  <div class="header-conn">
    <span class="conn-dot" id="conn-dot"></span>
    <span id="conn-status">Connecting...</span>
  </div>
  <button class="notif-bell" id="notif-bell" title="Notifications">
    <svg viewBox="0 0 24 24"><path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 0 1-3.46 0"/></svg>
    <span class="notif-badge" id="notif-badge"></span>
  </button>
</div>

<!-- Sidebar -->
<div class="sidebar" id="sidebar">
  <div class="nav-item active" data-view="overview" onclick="showView('overview')">
    <svg viewBox="0 0 24 24"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/></svg>
    <span class="nav-label">Overview</span>
  </div>
  <div class="nav-item" data-view="products" onclick="showView('products')">
    <svg viewBox="0 0 24 24"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg>
    <span class="nav-label">Products</span>
    <span class="nav-count" id="nav-products-count">0</span>
  </div>
  <div class="nav-item" data-view="orders" onclick="showView('orders')">
    <svg viewBox="0 0 24 24"><path d="M6 2L3 6v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6l-3-4z"/><line x1="3" y1="6" x2="21" y2="6"/><path d="M16 10a4 4 0 0 1-8 0"/></svg>
    <span class="nav-label">Orders</span>
    <span class="nav-count" id="nav-orders-count">0</span>
  </div>
  <div class="nav-item" data-view="activity" onclick="showView('activity')">
    <svg viewBox="0 0 24 24"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>
    <span class="nav-label">Activity</span>
    <span class="nav-count" id="nav-activity-count">0</span>
  </div>
  <div class="sidebar-footer">UCP Merchant</div>
</div>

<!-- Content -->
<div class="content" id="content">

  <!-- STAT ROW (always visible) -->
  <div class="stat-row">
    <div class="stat-card blue">
      <div class="stat-icon"><svg viewBox="0 0 24 24"><circle cx="9" cy="21" r="1"/><circle cx="20" cy="21" r="1"/><path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/></svg></div>
      <div class="stat-info"><div class="stat-val" id="stat-carts">0</div><div class="stat-label">Active Carts</div></div>
    </div>
    <div class="stat-card purple">
      <div class="stat-icon"><svg viewBox="0 0 24 24"><rect x="1" y="4" width="22" height="16" rx="2" ry="2"/><line x1="1" y1="10" x2="23" y2="10"/></svg></div>
      <div class="stat-info"><div class="stat-val" id="stat-checkouts">0</div><div class="stat-label">Active Checkouts</div></div>
    </div>
    <div class="stat-card cyan">
      <div class="stat-icon"><svg viewBox="0 0 24 24"><path d="M6 2L3 6v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6l-3-4z"/><line x1="3" y1="6" x2="21" y2="6"/><path d="M16 10a4 4 0 0 1-8 0"/></svg></div>
      <div class="stat-info"><div class="stat-val" id="stat-orders">0</div><div class="stat-label">Total Orders</div></div>
    </div>
    <div class="stat-card amber">
      <div class="stat-icon"><svg viewBox="0 0 24 24"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/></svg></div>
      <div class="stat-info"><div class="stat-val" id="stat-products">0</div><div class="stat-label">Products</div></div>
    </div>
    <div class="stat-card green">
      <div class="stat-icon"><svg viewBox="0 0 24 24"><line x1="12" y1="1" x2="12" y2="23"/><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg></div>
      <div class="stat-info"><div class="stat-val" id="stat-revenue">$0.00</div><div class="stat-label">Revenue</div></div>
    </div>
    <div class="stat-card accent">
      <div class="stat-icon"><svg viewBox="0 0 24 24"><polyline points="23 6 13.5 15.5 8.5 10.5 1 18"/><polyline points="17 6 23 6 23 12"/></svg></div>
      <div class="stat-info"><div class="stat-val" id="stat-conversion">0%</div><div class="stat-label">Conversion</div></div>
    </div>
  </div>

  <!-- VIEW: Overview -->
  <div class="view active" id="view-overview">
    <div class="recent-orders-row" id="recent-orders-row"></div>
    <div class="overview-grid">
      <div class="panel">
        <div class="panel-header"><div class="left"><svg viewBox="0 0 24 24"><circle cx="9" cy="21" r="1"/><circle cx="20" cy="21" r="1"/><path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/></svg> Active Carts</div></div>
        <div class="panel-body" id="overview-carts"><div class="empty">No active carts</div></div>
      </div>
      <div class="panel">
        <div class="panel-header"><div class="left"><svg viewBox="0 0 24 24"><rect x="1" y="4" width="22" height="16" rx="2" ry="2"/><line x1="1" y1="10" x2="23" y2="10"/></svg> Active Checkouts</div></div>
        <div class="panel-body" id="overview-checkouts"><div class="empty">No active checkouts</div></div>
      </div>
    </div>
    <div class="panel">
      <div class="panel-header"><div class="left"><svg viewBox="0 0 24 24"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg> Recent Activity</div></div>
      <div class="panel-body" id="overview-activity"><div class="empty">Waiting for activity...</div></div>
    </div>
  </div>

  <!-- VIEW: Products -->
  <div class="view" id="view-products">
    <div class="panel">
      <div class="toolbar" id="products-toolbar">
        <input type="text" id="product-search" placeholder="Search products...">
        <select id="product-category-filter"><option value="">All Categories</option></select>
        <div style="flex:1"></div>
        <button class="btn btn-primary btn-sm" id="add-product-btn" onclick="showProductModal()">+ Add Product</button>
      </div>
      <div class="panel-body" style="max-height:calc(100vh - 260px)">
        <table>
          <thead><tr><th style="width:50px"></th><th>Product</th><th>Brand</th><th>Price</th><th>Stock</th><th style="width:120px">Actions</th></tr></thead>
          <tbody id="products-table"></tbody>
        </table>
        <div class="empty" id="products-empty">No products</div>
      </div>
    </div>
  </div>

  <!-- VIEW: Orders -->
  <div class="view" id="view-orders">
    <div class="panel">
      <div class="sub-tabs">
        <div class="sub-tab active" data-sub="carts" onclick="showSubTab('carts')">Active Carts <span class="count" id="sub-carts-count">0</span></div>
        <div class="sub-tab" data-sub="checkouts" onclick="showSubTab('checkouts')">Checkouts <span class="count" id="sub-checkouts-count">0</span></div>
        <div class="sub-tab" data-sub="all-orders" onclick="showSubTab('all-orders')">Orders <span class="count" id="sub-orders-count">0</span></div>
      </div>
      <div class="sub-content active" id="sub-carts">
        <div class="panel-body" style="max-height:calc(100vh - 260px)">
          <table><thead><tr><th>ID</th><th>User</th><th>Items</th><th>Total</th></tr></thead><tbody id="carts-table"></tbody></table>
          <div class="empty" id="carts-empty">No active carts</div>
        </div>
      </div>
      <div class="sub-content" id="sub-checkouts">
        <div class="panel-body" style="max-height:calc(100vh - 260px)">
          <table><thead><tr><th>ID</th><th>User</th><th>Status</th><th>Total</th><th>Buyer</th></tr></thead><tbody id="checkouts-table"></tbody></table>
          <div class="empty" id="checkouts-empty">No active checkouts</div>
        </div>
      </div>
      <div class="sub-content" id="sub-all-orders">
        <div class="panel-body" style="max-height:calc(100vh - 260px)">
          <table><thead><tr><th>Order ID</th><th>User</th><th>Status</th><th>Tracking</th><th>Total</th></tr></thead><tbody id="orders-table"></tbody></table>
          <div class="empty" id="orders-empty">No orders yet</div>
        </div>
      </div>
    </div>
  </div>

  <!-- VIEW: Activity -->
  <div class="view" id="view-activity">
    <div class="panel">
      <div class="filter-bar">
        <button class="filter-btn active" data-filter="all" onclick="setActivityFilter('all')">All</button>
        <button class="filter-btn" data-filter="cart" onclick="setActivityFilter('cart')">Carts</button>
        <button class="filter-btn" data-filter="checkout" onclick="setActivityFilter('checkout')">Checkouts</button>
        <button class="filter-btn" data-filter="order" onclick="setActivityFilter('order')">Orders</button>
        <button class="filter-btn" data-filter="product" onclick="setActivityFilter('product')">Products</button>
      </div>
      <div class="panel-body" style="max-height:calc(100vh - 240px)" id="activity-feed"></div>
    </div>
  </div>

</div>

<!-- Toast container -->
<div class="toast-container" id="toast-container"></div>

<script>
const $=id=>document.getElementById(id);
let productsState=[], cartsState={}, checkoutsState={}, ordersState={};
let activityEvents=[], activityFilter='all', currentView='overview';
let searchTimer=null, notifCount=0;

// ── Helpers ──
function escapeHTML(s){const d=document.createElement('div');d.textContent=s;return d.innerHTML}
function formatCents(c){return '$'+(c/100).toFixed(2)}
function getTotal(totals){const t=totals?.find(t=>t.type==='total');return t?formatCents(t.amount):'--'}
function getTotalAmount(totals){const t=totals?.find(t=>t.type==='total');return t?t.amount:0}
function statusBadge(s){return '<span class="status status-'+s+'">'+s.replace(/_/g,' ')+'</span>'}

function countriesBadge(c){if(!c||!c.length)return '<span class="country-badge country-worldwide">Worldwide</span>';if(c.length<=3)return '<span class="country-badge country-restricted">'+c.join(', ')+'</span>';return '<span class="country-badge country-restricted">'+c.slice(0,3).join(', ')+' +'+String(c.length-3)+'</span>'}
function userBadge(ownerId){
  if(ownerId)return '<span class="user-badge authenticated"><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg> '+escapeHTML(ownerId)+'</span>';
  return '<span class="user-badge guest">Guest</span>';
}
function relativeTime(ts){
  const diff=Math.floor((Date.now()-new Date(ts).getTime())/1000);
  if(diff<60)return diff+'s ago';if(diff<3600)return Math.floor(diff/60)+'m ago';
  if(diff<86400)return Math.floor(diff/3600)+'h ago';return Math.floor(diff/86400)+'d ago';
}
function eventCategory(type){
  if(type.startsWith('cart'))return 'cart';
  if(type.startsWith('checkout'))return 'checkout';
  if(type.startsWith('order'))return 'order';
  if(type.startsWith('product'))return 'product';
  return 'other';
}
function bumpVal(el){el.classList.remove('val-bump');void el.offsetWidth;el.classList.add('val-bump')}

// ── SVG icons ──
const svgIcons={
  cart:'<svg viewBox="0 0 24 24"><circle cx="9" cy="21" r="1"/><circle cx="20" cy="21" r="1"/><path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/></svg>',
  checkout:'<svg viewBox="0 0 24 24"><rect x="1" y="4" width="22" height="16" rx="2" ry="2"/><line x1="1" y1="10" x2="23" y2="10"/></svg>',
  order:'<svg viewBox="0 0 24 24"><path d="M6 2L3 6v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6l-3-4z"/><line x1="3" y1="6" x2="21" y2="6"/><path d="M16 10a4 4 0 0 1-8 0"/></svg>',
  product:'<svg viewBox="0 0 24 24"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/></svg>',
  check:'<svg viewBox="0 0 24 24"><polyline points="20 6 9 17 4 12"/></svg>',
  x:'<svg viewBox="0 0 24 24"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>',
  edit:'<svg viewBox="0 0 24 24"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>',
  pkg:'<svg viewBox="0 0 24 24"><line x1="16.5" y1="9.4" x2="7.5" y2="4.21"/><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg>',
  truck:'<svg viewBox="0 0 24 24"><rect x="1" y="3" width="15" height="13"/><polygon points="16 8 20 8 23 11 23 16 16 16 16 8"/><circle cx="5.5" cy="18.5" r="2.5"/><circle cx="18.5" cy="18.5" r="2.5"/></svg>',
  activity:'<svg viewBox="0 0 24 24"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>'
};

// ── View switching ──
function showView(name){
  currentView=name;
  document.querySelectorAll('.view').forEach(v=>v.classList.remove('active'));
  document.querySelectorAll('.nav-item').forEach(n=>n.classList.remove('active'));
  $('view-'+name)?.classList.add('active');
  document.querySelector('.nav-item[data-view="'+name+'"]')?.classList.add('active');
  renderCurrentView();
}
function showSubTab(name){
  document.querySelectorAll('.sub-tab').forEach(t=>t.classList.remove('active'));
  document.querySelectorAll('.sub-content').forEach(c=>c.classList.remove('active'));
  document.querySelector('.sub-tab[data-sub="'+name+'"]')?.classList.add('active');
  $('sub-'+name)?.classList.add('active');
}

// ── Stats ──
function updateStats(){
  const cartCount=Object.keys(cartsState).length;
  let activeCoCount=0;
  Object.values(checkoutsState).forEach(co=>{if(co.status!=='completed'&&co.status!=='canceled')activeCoCount++});
  const orderCount=Object.keys(ordersState).length;
  const productCount=productsState.length;

  let revenue=0;
  Object.values(ordersState).forEach(o=>{revenue+=getTotalAmount(o.totals)});

  const totalCheckouts=Object.keys(checkoutsState).length;
  const completedCheckouts=Object.values(checkoutsState).filter(co=>co.status==='completed').length;
  const convRate=totalCheckouts>0?Math.round(completedCheckouts/totalCheckouts*100):0;

  const updates=[
    ['stat-carts',cartCount],['stat-checkouts',activeCoCount],['stat-orders',orderCount],
    ['stat-products',productCount],['stat-revenue',formatCents(revenue)],['stat-conversion',convRate+'%']
  ];
  updates.forEach(([id,val])=>{
    const el=$(id);
    const sv=String(val);
    if(el.textContent!==sv){el.textContent=sv;bumpVal(el)}
  });

  $('nav-products-count').textContent=productCount;
  $('nav-orders-count').textContent=orderCount;
  $('nav-activity-count').textContent=activityEvents.length;
  $('sub-carts-count').textContent=cartCount;
  $('sub-checkouts-count').textContent=activeCoCount;
  $('sub-orders-count').textContent=orderCount;
}

// ── Render dispatch ──
function renderCurrentView(){
  switch(currentView){
    case 'overview':renderOverview();break;
    case 'products':renderProducts();break;
    case 'orders':renderCarts();renderCheckouts();renderOrders();break;
    case 'activity':renderActivityFeed();break;
  }
}
function renderAll(){
  updateStats();
  renderCurrentView();
}

// ── Overview ──
function renderOverview(){
  // Recent orders (last 5)
  const row=$('recent-orders-row');
  const oArr=Object.values(ordersState).slice(-5).reverse();
  if(oArr.length){
    row.innerHTML=oArr.map(o=>'<div class="order-card" onclick=\'showDetailModal("order",ordersState["'+o.id+'"])\'>'
      +'<div class="oc-id">'+escapeHTML(o.id)+'</div>'
      +'<div class="oc-total">'+getTotal(o.totals)+'</div>'
      +'<div>'+statusBadge(o.status)+'</div>'
      +'<div class="oc-meta">'+userBadge(o.owner_id)+'</div></div>').join('');
  } else { row.innerHTML=''; }

  // Active carts summary
  const cartsDiv=$('overview-carts');
  const cArr=Object.values(cartsState).slice(0,5);
  if(cArr.length){
    cartsDiv.innerHTML='<table><thead><tr><th>ID</th><th>User</th><th>Items</th><th>Total</th></tr></thead><tbody>'
      +cArr.map(c=>'<tr class="clickable-row" onclick=\'showDetailModal("cart",cartsState["'+c.id+'"])\'><td>'+escapeHTML(c.id)+'</td><td>'+userBadge(c.owner_id)+'</td><td>'+c.line_items.length+'</td><td>'+getTotal(c.totals)+'</td></tr>').join('')
      +'</tbody></table>';
  } else { cartsDiv.innerHTML='<div class="empty">No active carts</div>'; }

  // Active checkouts summary
  const coDiv=$('overview-checkouts');
  const coArr=Object.values(checkoutsState).filter(co=>co.status!=='completed'&&co.status!=='canceled').slice(0,5);
  if(coArr.length){
    coDiv.innerHTML='<table><thead><tr><th>ID</th><th>Status</th><th>Total</th></tr></thead><tbody>'
      +coArr.map(co=>'<tr class="clickable-row" onclick=\'showDetailModal("checkout",checkoutsState["'+co.id+'"])\'><td>'+escapeHTML(co.id)+'</td><td>'+statusBadge(co.status)+'</td><td>'+getTotal(co.totals)+'</td></tr>').join('')
      +'</tbody></table>';
  } else { coDiv.innerHTML='<div class="empty">No active checkouts</div>'; }

  // Recent activity (last 15)
  const actDiv=$('overview-activity');
  const recent=activityEvents.slice(0,15);
  if(recent.length){
    actDiv.innerHTML=recent.map(e=>renderActivityCard(e)).join('');
  } else { actDiv.innerHTML='<div class="empty">Waiting for activity...</div>'; }
}

// ── Products ──
function getFilteredProducts(){
  const search=($('product-search')?.value||'').toLowerCase();
  const cat=$('product-category-filter')?.value||'';
  return productsState.filter(p=>{
    if(search&&!p.title.toLowerCase().includes(search)&&!p.id.toLowerCase().includes(search)&&!(p.brand||'').toLowerCase().includes(search))return false;
    if(cat&&p.category!==cat)return false;
    return true;
  });
}
function updateCategoryFilter(){
  const sel=$('product-category-filter');
  const cats=[...new Set(productsState.map(p=>p.category).filter(Boolean))].sort();
  const cur=sel.value;
  sel.innerHTML='<option value="">All Categories</option>'+cats.map(c=>'<option value="'+escapeHTML(c)+'">'+escapeHTML(c)+'</option>').join('');
  sel.value=cur;
}
function renderProducts(){
  updateCategoryFilter();
  const filtered=getFilteredProducts();
  const tb=$('products-table');const empty=$('products-empty');
  tb.innerHTML='';
  empty.style.display=filtered.length?'none':'';
  filtered.forEach(p=>{
    const tr=document.createElement('tr');
    tr.dataset.id=p.id;
    tr.className='clickable-row';
    const thumbHTML=p.image_url
      ?'<img class="thumb" src="'+escapeHTML(p.image_url)+'" onerror="this.outerHTML=\'<div class=thumb-placeholder>'+svgIcons.pkg.replace(/"/g,"&quot;")+'</div>\'">'
      :'<div class="thumb-placeholder">'+svgIcons.pkg+'</div>';
    const stockClass=p.quantity<10?'stock-warn':'';
    tr.innerHTML='<td>'+thumbHTML+'</td>'
      +'<td><div class="cell-stack"><span class="primary">'+escapeHTML(p.title)+'</span><span class="secondary">'+escapeHTML(p.category||'')+'</span></div></td>'
      +'<td>'+escapeHTML(p.brand||'')+'</td>'
      +'<td>'+formatCents(p.price)+'</td>'
      +'<td class="'+stockClass+'"><button class="qty-btn" onclick="event.stopPropagation();changeQty(\''+p.id+'\',-1)">-</button> '+p.quantity+' <button class="qty-btn" onclick="event.stopPropagation();changeQty(\''+p.id+'\',1)">+</button></td>'
      +'<td><button class="btn btn-sm" onclick="event.stopPropagation();showProductModal(\''+p.id+'\')">Edit</button> <button class="btn btn-danger btn-sm" onclick="event.stopPropagation();deleteProduct(\''+p.id+'\')">Del</button></td>';
    tr.onclick=()=>showDetailModal('product',p);
    tb.appendChild(tr);
  });
}

// Product search debounce
$('product-search')?.addEventListener('input',()=>{clearTimeout(searchTimer);searchTimer=setTimeout(renderProducts,300)});
$('product-category-filter')?.addEventListener('change',renderProducts);


// Product add/edit modal
function showProductModal(editId){
  const p=editId?productsState.find(x=>x.id===editId):null;
  const isEdit=!!p;
  const overlay=document.createElement('div');
  overlay.className='modal-overlay';
  overlay.onclick=e=>{if(e.target===overlay)overlay.remove()};
  overlay.innerHTML='<div class="modal-card"><div class="modal-header"><h2>'+(isEdit?'Edit Product':'Add Product')+'</h2><button class="modal-close" onclick="this.closest(\'.modal-overlay\').remove()">&#10005;</button></div>'
    +'<div class="modal-body"><div class="form-grid">'
    +'<div class="form-group"><label>Title</label><input type="text" id="pm-title" value="'+escapeHTML(p?.title||'')+'"></div>'
    +'<div class="form-group"><label>Category</label><input type="text" id="pm-category" value="'+escapeHTML(p?.category||'')+'"></div>'
    +'<div class="form-group"><label>Brand</label><input type="text" id="pm-brand" value="'+escapeHTML(p?.brand||'')+'"></div>'
    +'<div class="form-group"><label>Price (cents)</label><input type="number" id="pm-price" value="'+(p?.price||0)+'"></div>'
    +'<div class="form-group"><label>Stock</label><input type="number" id="pm-qty" value="'+(p?.quantity||0)+'"></div>'
    +'<div class="form-group"><label>Countries</label><input type="text" id="pm-countries" value="'+escapeHTML((p?.available_countries||[]).join(', '))+'" placeholder="US, FR, DE (empty=worldwide)"></div>'
    +'<div class="form-group full"><label>Image URL</label><input type="text" id="pm-image" value="'+escapeHTML(p?.image_url||'')+'">'+(p?.image_url?'<img class="img-preview" src="'+escapeHTML(p.image_url)+'" onerror="this.style.display=\'none\'">':'')+'</div>'
    +'<div class="form-group full"><label>Description</label><textarea id="pm-desc">'+escapeHTML(p?.description||'')+'</textarea></div>'
    +'</div></div>'
    +'<div class="modal-footer"><button class="btn" onclick="this.closest(\'.modal-overlay\').remove()">Cancel</button><button class="btn btn-primary" onclick="saveProductModal('+(isEdit?"'"+p.id+"'":"null")+')">Save</button></div></div>';
  document.body.appendChild(overlay);
  overlay.querySelector('#pm-title').focus();
}

async function saveProductModal(editId){
  const title=$('pm-title').value.trim();
  const category=$('pm-category').value.trim();
  const brand=$('pm-brand').value.trim();
  const price=parseInt($('pm-price').value)||0;
  const qty=parseInt($('pm-qty').value)||0;
  const countriesRaw=$('pm-countries').value.trim();
  const countries=countriesRaw?countriesRaw.split(',').map(s=>s.trim().toUpperCase()).filter(Boolean):[];
  const image=$('pm-image').value.trim();
  const desc=$('pm-desc').value.trim();
  if(!title){alert('Title is required');return}
  const body={title,category,brand,price,quantity:qty,image_url:image,description:desc,available_countries:countries.length?countries:null};
  if(editId)body.id=editId;
  const res=await fetch('/api/products',{method:editId?'PUT':'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});
  if(res.ok){document.querySelector('.modal-overlay')?.remove()}
  else{alert(await res.text())}
}

async function changeQty(id,delta){
  const p=productsState.find(x=>x.id===id);if(!p)return;
  await fetch('/api/products',{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify({id,quantity:Math.max(0,p.quantity+delta)})});
}
async function deleteProduct(id){
  if(!confirm('Delete product '+id+'?'))return;
  await fetch('/api/products?id='+encodeURIComponent(id),{method:'DELETE'});
}

// ── Carts ──
function renderCarts(){
  const tb=$('carts-table');const empty=$('carts-empty');
  tb.innerHTML='';
  const arr=Object.values(cartsState);
  empty.style.display=arr.length?'none':'';
  arr.forEach(c=>{
    const tr=document.createElement('tr');tr.className='clickable-row';tr.id='cart-'+c.id;
    tr.innerHTML='<td>'+escapeHTML(c.id)+'</td><td>'+userBadge(c.owner_id)+'</td><td>'+c.line_items.length+' item(s)</td><td>'+getTotal(c.totals)+'</td>';
    tr.onclick=()=>showDetailModal('cart',c);
    tb.appendChild(tr);
  });
}

// ── Checkouts ──
function renderCheckouts(){
  const tb=$('checkouts-table');const empty=$('checkouts-empty');
  tb.innerHTML='';
  const arr=Object.values(checkoutsState).filter(co=>co.status!=='completed'&&co.status!=='canceled');
  empty.style.display=arr.length?'none':'';
  arr.forEach(co=>{
    const tr=document.createElement('tr');tr.className='clickable-row';tr.id='checkout-'+co.id;
    tr.innerHTML='<td>'+escapeHTML(co.id)+'</td><td>'+userBadge(co.owner_id)+'</td><td>'+statusBadge(co.status)+'</td><td>'+getTotal(co.totals)+'</td><td>'+(co.buyer?.name||'--')+'</td>';
    tr.onclick=()=>showDetailModal('checkout',co);
    tb.appendChild(tr);
  });
}

// ── Orders ──
function renderOrders(){
  const tb=$('orders-table');const empty=$('orders-empty');
  tb.innerHTML='';
  const arr=Object.values(ordersState);
  empty.style.display=arr.length?'none':'';
  arr.forEach(o=>{
    const tr=document.createElement('tr');tr.className='clickable-row';
    tr.innerHTML='<td>'+escapeHTML(o.id)+'</td><td>'+userBadge(o.owner_id)+'</td><td>'+statusBadge(o.status)+'</td><td>'+(o.shipment?.tracking_number||'--')+'</td><td>'+getTotal(o.totals)+'</td>';
    tr.onclick=()=>showDetailModal('order',o);
    tb.appendChild(tr);
  });
}

// ── Activity ──
function renderActivityCard(e){
  const cat=eventCategory(e.type);
  const iconCat=cat==='other'?'product':cat;
  const isError=e.type.includes('canceled')||e.type.includes('removed');
  return '<div class="activity-card" data-cat="'+cat+'">'
    +'<div class="act-icon '+(isError?'error':iconCat)+'">'+svgIcons[iconCat]+'</div>'
    +'<div class="act-body"><div class="act-summary">'+escapeHTML(e.summary)+'</div><div class="act-time">'+relativeTime(e.timestamp)+'</div></div></div>';
}
function renderActivityFeed(){
  const feed=$('activity-feed');
  const filtered=activityFilter==='all'?activityEvents:activityEvents.filter(e=>eventCategory(e.type)===activityFilter);
  if(filtered.length){
    feed.innerHTML=filtered.map(e=>renderActivityCard(e)).join('');
  } else {
    feed.innerHTML='<div class="empty">No '+(activityFilter==='all'?'':'matching ')+' activity</div>';
  }
}
function setActivityFilter(f){
  activityFilter=f;
  document.querySelectorAll('.filter-btn').forEach(b=>b.classList.toggle('active',b.dataset.filter===f));
  renderActivityFeed();
}
function addActivity(event){
  activityEvents.unshift(event);
  if(activityEvents.length>200)activityEvents.length=200;
  // Show toast if not on activity view
  if(currentView!=='activity'){
    notifCount++;
    $('notif-badge').textContent=notifCount;
    $('notif-badge').dataset.count=notifCount;
    const important=['checkout_completed','order_confirmed','order_delivered','cart_created','checkout_created'];
    if(important.includes(event.type))showToast(event);
  }
}

// ── Toast ──
function showToast(event){
  const container=$('toast-container');
  if(container.children.length>=3)container.lastChild.remove();
  const cat=eventCategory(event.type);
  const toast=document.createElement('div');
  toast.className='toast';
  toast.innerHTML='<div class="toast-icon" style="color:var(--'+cat+')">'+(svgIcons[cat]||svgIcons.activity)+'</div><div class="toast-body">'+escapeHTML(event.summary)+'</div><button class="toast-close" onclick="this.parentElement.remove()">&#10005;</button>';
  container.prepend(toast);
  setTimeout(()=>{toast.classList.add('removing');setTimeout(()=>toast.remove(),300)},5000);
}

// Notification bell clears count and goes to activity
$('notif-bell').onclick=()=>{notifCount=0;$('notif-badge').textContent='';$('notif-badge').dataset.count='0';showView('activity')};

// ── Detail Modals ──
function showDetailModal(type, data){
  const overlay=document.createElement('div');
  overlay.className='modal-overlay';
  overlay.onclick=e=>{if(e.target===overlay)overlay.remove()};
  const card=document.createElement('div');
  card.className='modal-card';
  let title='',body='';

  if(type==='cart'){
    title='Cart '+escapeHTML(data.id);
    body='<div class="section"><div class="section-title">Owner</div><div>'+userBadge(data.owner_id)+'</div></div>';
    body+=renderLineItemsSection(data.line_items)+renderTotalsSection(data.totals);
    if(data.messages?.length){
      body+='<div class="section"><div class="section-title">Messages</div>';
      data.messages.forEach(m=>{body+='<div class="detail-row"><span class="label">'+escapeHTML(m.type)+'</span><span class="val">'+escapeHTML(m.text)+'</span></div>'});
      body+='</div>';
    }
  } else if(type==='order'){
    title='Order '+escapeHTML(data.id);
    body='<div class="section"><div class="section-title">Owner</div><div>'+userBadge(data.owner_id)+'</div></div>';
    body+='<div class="section"><div class="section-title">Status</div>'+renderOrderTimeline(data.status)+'</div>';
    if(data.confirmation_number){
      body+='<div class="section"><div class="detail-row"><span class="label">Confirmation</span><span class="val">'+escapeHTML(data.confirmation_number)+'</span></div>';
      if(data.created_at)body+='<div class="detail-row"><span class="label">Created</span><span class="val">'+new Date(data.created_at).toLocaleString()+'</span></div>';
      body+='</div>';
    }
    body+=renderLineItemsSection(data.line_items)+renderTotalsSection(data.totals);
    if(data.shipment){
      body+='<div class="section"><div class="section-title">Shipment</div>';
      body+='<div class="detail-row"><span class="label">Tracking #</span><span class="val" style="cursor:pointer" onclick="navigator.clipboard.writeText(\''+escapeHTML(data.shipment.tracking_number)+'\');this.textContent=\'Copied!\'">'+escapeHTML(data.shipment.tracking_number)+' &#128203;</span></div>';
      body+='<div class="detail-row"><span class="label">Carrier</span><span class="val">'+escapeHTML(data.shipment.carrier)+'</span></div>';
      if(data.shipment.estimated_delivery)body+='<div class="detail-row"><span class="label">Est. Delivery</span><span class="val">'+escapeHTML(data.shipment.estimated_delivery)+'</span></div>';
      if(data.shipment.shipped_at)body+='<div class="detail-row"><span class="label">Shipped</span><span class="val">'+new Date(data.shipment.shipped_at).toLocaleString()+'</span></div>';
      if(data.shipment.delivered_at&&data.shipment.delivered_at!=='0001-01-01T00:00:00Z')body+='<div class="detail-row"><span class="label">Delivered</span><span class="val">'+new Date(data.shipment.delivered_at).toLocaleString()+'</span></div>';
      body+='</div>';
    }
    if(data.buyer){
      body+='<div class="section"><div class="section-title">Buyer</div>';
      if(data.buyer.name)body+='<div class="detail-row"><span class="label">Name</span><span class="val">'+escapeHTML(data.buyer.name)+'</span></div>';
      if(data.buyer.email)body+='<div class="detail-row"><span class="label">Email</span><span class="val">'+escapeHTML(data.buyer.email)+'</span></div>';
      if(data.buyer.address){const a=data.buyer.address;const parts=[a.street,a.city,a.state,a.zip,a.country].filter(Boolean);if(parts.length)body+='<div class="detail-row"><span class="label">Address</span><span class="val">'+escapeHTML(parts.join(', '))+'</span></div>';}
      body+='</div>';
    }
  } else if(type==='product'){
    title=escapeHTML(data.title)+' <span style="color:var(--text-dim);font-size:13px">'+escapeHTML(data.id)+'</span>';
    body='<div class="section">';
    body+='<div class="detail-row"><span class="label">Category</span><span class="val">'+escapeHTML(data.category||'')+'</span></div>';
    body+='<div class="detail-row"><span class="label">Brand</span><span class="val">'+escapeHTML(data.brand||'')+'</span></div>';
    body+='<div class="detail-row"><span class="label">Price</span><span class="val">'+formatCents(data.price)+'</span></div>';
    body+='<div class="detail-row"><span class="label">Stock</span><span class="val">'+data.quantity+'</span></div>';
    body+='<div class="detail-row"><span class="label">Availability</span><span class="val">'+countriesBadge(data.available_countries)+'</span></div></div>';
    if(data.description)body+='<div class="section"><div class="section-title">Description</div><div style="font-size:13px;line-height:1.5;color:#ccc">'+escapeHTML(data.description)+'</div></div>';
    if(data.image_url)body+='<div class="section"><div class="section-title">Image</div><img src="'+escapeHTML(data.image_url)+'" style="max-width:200px;border-radius:8px" onerror="this.style.display=\'none\'"></div>';
  } else if(type==='checkout'){
    title='Checkout '+escapeHTML(data.id);
    body='<div class="section"><div class="section-title">Owner</div><div>'+userBadge(data.owner_id)+'</div></div>';
    body+='<div class="section"><div class="section-title">Status</div><div>'+statusBadge(data.status)+'</div></div>';
    body+=renderLineItemsSection(data.line_items)+renderTotalsSection(data.totals);
    if(data.buyer){
      body+='<div class="section"><div class="section-title">Buyer</div>';
      if(data.buyer.name)body+='<div class="detail-row"><span class="label">Name</span><span class="val">'+escapeHTML(data.buyer.name)+'</span></div>';
      if(data.buyer.email)body+='<div class="detail-row"><span class="label">Email</span><span class="val">'+escapeHTML(data.buyer.email)+'</span></div>';
      if(data.buyer.address){const a=data.buyer.address;const parts=[a.street,a.city,a.state,a.zip,a.country].filter(Boolean);if(parts.length)body+='<div class="detail-row"><span class="label">Address</span><span class="val">'+escapeHTML(parts.join(', '))+'</span></div>';}
      body+='</div>';
    }
    if(data.links?.length){
      body+='<div class="section"><div class="section-title">Links</div>';
      data.links.forEach(l=>{body+='<div class="detail-row"><span class="label">'+escapeHTML(l.rel)+'</span><span class="val">'+escapeHTML(l.url)+'</span></div>'});
      body+='</div>';
    }
    if(data.order){
      body+='<div class="section"><div class="section-title">Order</div>';
      body+='<div class="detail-row"><span class="label">Order ID</span><span class="val">'+escapeHTML(data.order.id)+'</span></div>';
      body+='<div class="detail-row"><span class="label">Confirmation</span><span class="val">'+escapeHTML(data.order.confirmation_number)+'</span></div>';
      body+='<div class="detail-row"><span class="label">Status</span><span class="val">'+statusBadge(data.order.status)+'</span></div></div>';
    }
  }
  card.innerHTML='<div class="modal-header"><h2>'+title+'</h2><button class="modal-close" onclick="this.closest(\'.modal-overlay\').remove()">&#10005;</button></div><div class="modal-body">'+body+'</div>';
  overlay.appendChild(card);
  document.body.appendChild(overlay);
}

// ── Order Timeline ──
function renderOrderTimeline(status){
  const steps=['confirmed','processing','shipped','in_transit','out_for_delivery','delivered'];
  const isCanceled=status==='canceled';
  const currentIdx=steps.indexOf(status);
  return '<div class="timeline">'+steps.map((s,i)=>{
    let dotClass='tl-dot';
    let stepClass='tl-step';
    if(isCanceled){dotClass+=' canceled';stepClass+=' canceled'}
    else if(i<currentIdx){dotClass+=' done';stepClass+=' done'}
    else if(i===currentIdx){dotClass+=' current';stepClass+=' current'}
    const icon=i<currentIdx?svgIcons.check:(isCanceled?svgIcons.x:'');
    const line=i<steps.length-1?'<div class="tl-line'+(i<currentIdx?' done':'')+'"></div>':'';
    return '<div class="'+stepClass+'">'+line+'<div class="'+dotClass+'">'+icon+'</div><div class="tl-label">'+s.replace(/_/g,' ')+'</div></div>';
  }).join('')+'</div>';
}

function renderLineItemsSection(items){
  if(!items?.length)return '';
  let h='<div class="section"><div class="section-title">Line Items</div><table><thead><tr><th style="width:40px"></th><th>Product</th><th>Qty</th><th>Subtotal</th></tr></thead><tbody>';
  items.forEach(li=>{
    const sub=li.totals?.find(t=>t.type==='subtotal');
    const imgHTML=li.item?.image_url?'<img class="thumb" src="'+escapeHTML(li.item.image_url)+'" onerror="this.style.display=\'none\'">':'<div class="thumb-placeholder">'+svgIcons.pkg+'</div>';
    h+='<tr><td>'+imgHTML+'</td><td>'+escapeHTML(li.item?.title||li.id)+'</td><td>'+li.quantity+'</td><td>'+(sub?formatCents(sub.amount):'--')+'</td></tr>';
  });
  h+='</tbody></table></div>';
  return h;
}
function renderTotalsSection(totals){
  if(!totals?.length)return '';
  let h='<div class="section"><div class="section-title">Totals</div>';
  totals.forEach(t=>{h+='<div class="detail-row"><span class="label">'+escapeHTML(t.type)+'</span><span class="val">'+formatCents(t.amount)+'</span></div>'});
  return h+'</div>';
}

document.addEventListener('keydown',e=>{if(e.key==='Escape'){const m=document.querySelector('.modal-overlay');if(m)m.remove()}});

// ── SSE Connection ──
function connect(){
  const dot=$('conn-dot');const status=$('conn-status');
  dot.className='conn-dot';status.textContent='Connecting...';
  const es=new EventSource('/events');
  es.onopen=()=>{dot.className='conn-dot ok';status.textContent='Connected'};
  es.onerror=()=>{dot.className='conn-dot err';status.textContent='Reconnecting...'};
  es.onmessage=e=>{
    const d=JSON.parse(e.data);
    if(d.type==='snapshot'){
      productsState=d.products||[];
      cartsState={};(d.carts||[]).forEach(c=>cartsState[c.id]=c);
      checkoutsState={};(d.checkouts||[]).forEach(c=>checkoutsState[c.id]=c);
      ordersState={};(d.orders||[]).forEach(o=>ordersState[o.id]=o);
      renderAll();
      return;
    }
    // Update state regardless of active view
    addActivity(d);
    if(d.data){
      switch(d.type){
        case 'cart_created':case 'cart_updated':cartsState[d.id]=d.data;break;
        case 'cart_canceled':delete cartsState[d.id];break;
        case 'checkout_created':case 'checkout_updated':case 'checkout_completed':case 'checkout_canceled':
          checkoutsState[d.id]=d.data;break;
        case 'product_added':productsState.push(d.data);break;
        case 'product_updated':productsState=productsState.map(p=>p.id===d.data.id?d.data:p);break;
        case 'product_removed':productsState=productsState.filter(p=>p.id!==d.id);break;
        case 'order_confirmed':case 'order_processing':case 'order_shipped':
        case 'order_in_transit':case 'order_out_for_delivery':case 'order_delivered':case 'order_canceled':
          ordersState[d.id]=d.data;break;
      }
    }
    renderAll();
    // Row flash effect
    if(d.id){
      setTimeout(()=>{
        const row=document.querySelector('tr[data-id="'+d.id+'"]')||document.getElementById('cart-'+d.id)||document.getElementById('checkout-'+d.id);
        if(row){row.classList.remove('row-flash');void row.offsetWidth;row.classList.add('row-flash')}
      },50);
    }
  };
}
connect();
</script>
</body>
</html>`

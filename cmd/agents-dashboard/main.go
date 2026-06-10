package main

import (
	"embed"
	"flag"
	"html/template"
	"log"
	"net/http"
)

//go:embed templates/*
var templates embed.FS

func main() {
	port := flag.String("port", "8080", "Port to listen on")
	customerGrowthURL := flag.String("customer-growth-url", "http://localhost:9001", "Customer Growth agent URL")
	competitivenessURL := flag.String("competitiveness-url", "http://localhost:9002", "Competitiveness agent URL")
	flag.Parse()

	handler := NewDashboardHandler(*customerGrowthURL, *competitivenessURL)

	http.HandleFunc("/", handler.HandleIndex)
	http.HandleFunc("/api/call-agent", handler.HandleCallAgent)
	http.HandleFunc("/api/agents/status", handler.HandleAgentsStatus)

	addr := ":" + *port
	log.Printf("Agents Dashboard starting on http://localhost%s", addr)
	log.Printf("Customer Growth agent: %s", *customerGrowthURL)
	log.Printf("Competitiveness agent: %s", *competitivenessURL)
	log.Fatal(http.ListenAndServe(addr, nil))
}

type DashboardHandler struct {
	customerGrowthURL  string
	competitivenessURL string
	tmpl               *template.Template
}

func NewDashboardHandler(customerGrowthURL, competitivenessURL string) *DashboardHandler {
	tmpl := template.Must(template.ParseFS(templates, "templates/*.html"))
	return &DashboardHandler{
		customerGrowthURL:  customerGrowthURL,
		competitivenessURL: competitivenessURL,
		tmpl:               tmpl,
	}
}

func (h *DashboardHandler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"CustomerGrowthURL":  h.customerGrowthURL,
		"CompetitivenessURL": h.competitivenessURL,
	}
	h.tmpl.ExecuteTemplate(w, "index.html", data)
}

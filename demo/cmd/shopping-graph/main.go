package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/owulveryck/ucp-merchant-test/demo/internal/a2aclient"
	"github.com/owulveryck/ucp-merchant-test/demo/internal/shoppinggraph"
)

func main() {
	port := flag.Int("port", 9000, "port to listen on")
	configFile := flag.String("config", "config/shopping_graph.yaml", "config file path")
	obsURL := flag.String("obs-url", "", "observability hub URL")
	pollInterval := flag.Duration("poll-interval", 30*time.Second, "merchant poll interval")
	flag.Parse()

	data, err := os.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("read config: %v", err)
	}

	var cfg shoppinggraph.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("parse config: %v", err)
	}

	graph := shoppinggraph.NewShoppingGraph()

	for _, mc := range cfg.Merchants {
		graph.Merchants[mc.ID] = &shoppinggraph.MerchantNode{
			ID:            mc.ID,
			Name:          mc.Name,
			Endpoint:      mc.Endpoint,
			Score:         mc.Score,
			DiscountHints: mc.DiscountHints,
		}
	}

	client := a2aclient.NewClient("john.doe@example.com", "US", *obsURL)
	poller := shoppinggraph.NewPoller(graph, client, *pollInterval)
	poller.Start()
	defer poller.Stop()

	handler := shoppinggraph.NewHandler(graph)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Shopping Graph starting on http://localhost:%d", *port)
	log.Printf("Search endpoint: POST http://localhost:%d/search", *port)
	if err := http.ListenAndServe(addr, handler.Mux()); err != nil {
		log.Fatal(err)
	}
}

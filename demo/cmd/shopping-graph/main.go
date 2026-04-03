package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
	dynamic := flag.Bool("dynamic", false, "accept dynamic merchant registration via POST /merchants (skip config file requirement)")
	flag.Parse()

	graph := shoppinggraph.NewShoppingGraph()

	if !*dynamic {
		data, err := os.ReadFile(*configFile)
		if err != nil {
			log.Fatalf("read config: %v", err)
		}

		var cfg shoppinggraph.Config
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			log.Fatalf("parse config: %v", err)
		}

		for _, mc := range cfg.Merchants {
			graph.Merchants[mc.ID] = &shoppinggraph.MerchantNode{
				ID:            mc.ID,
				Name:          mc.Name,
				Endpoint:      mc.Endpoint,
				MaxCPCBid:     mc.MaxCPCBid,
				DiscountHints: mc.DiscountHints,
			}
		}
	} else {
		log.Println("Dynamic mode: merchants can be registered via POST /merchants")
	}

	client := a2aclient.NewClient("john.doe@example.com", "US", *obsURL)
	poller := shoppinggraph.NewPoller(graph, client, *pollInterval, client.ObsURL())
	poller.Start()

	handler := shoppinggraph.NewHandler(graph, poller)

	addr := fmt.Sprintf(":%d", *port)
	srv := &http.Server{
		Addr:    addr,
		Handler: handler.Mux(),
	}

	go func() {
		log.Printf("Shopping Graph starting on http://localhost:%d", *port)
		log.Printf("Search endpoint: POST http://localhost:%d/search", *port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down shopping graph...")
	poller.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Shopping graph shutdown error: ", err)
	}
	log.Println("Shopping graph stopped")
}

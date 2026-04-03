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
)

func main() {
	port := flag.Int("port", 8888, "port to listen on")
	costPrice := flag.Int("cost-price", 5000, "cost price in cents (minimum selling price)")
	productName := flag.String("product-name", "Casque Audio", "product name")
	graphURL := flag.String("graph-url", "http://localhost:9000", "shopping graph URL")
	obsURL := flag.String("obs-url", "", "observability hub URL")
	baseURL := flag.String("base-url", "", "external base URL (e.g. https://demo.example.com); if empty, uses http://localhost:PORT")
	flag.Parse()

	arena := NewArenaServer(*costPrice, *productName, *graphURL, *obsURL, *port, *baseURL)

	addr := fmt.Sprintf(":%d", *port)
	srv := &http.Server{
		Addr:    addr,
		Handler: arena,
	}

	go func() {
		log.Printf("Arena starting on http://localhost:%d", *port)
		log.Printf("Landing:   http://localhost:%d/", *port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down arena server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Arena shutdown error: ", err)
	}
	log.Println("Arena server stopped")
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

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

	"github.com/owulveryck/ucp-merchant-test/demo/internal/obs"
)

func main() {
	port := flag.Int("port", 9002, "port to listen on")
	graphURL := flag.String("graph-url", "http://localhost:9000", "shopping graph base URL")
	arenaURL := flag.String("arena-url", "http://localhost:8888", "arena server base URL")
	flag.Parse()

	hub := obs.NewHub()
	handler := obs.NewHandler(hub, *graphURL, *arenaURL)

	addr := fmt.Sprintf(":%d", *port)
	srv := &http.Server{
		Addr:    addr,
		Handler: handler.Mux(),
	}

	go func() {
		log.Printf("Observability Hub starting on http://localhost:%d", *port)
		log.Printf("Dashboard: http://localhost:%d/", *port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down obs-hub...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Obs-hub shutdown error: ", err)
	}
	log.Println("Obs-hub stopped")
}

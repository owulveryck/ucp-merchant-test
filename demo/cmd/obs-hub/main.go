package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/owulveryck/ucp-merchant-test/demo/internal/obs"
)

func main() {
	port := flag.Int("port", 9002, "port to listen on")
	flag.Parse()

	hub := obs.NewHub()
	handler := obs.NewHandler(hub)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Observability Hub starting on http://localhost:%d", *port)
	log.Printf("Dashboard: http://localhost:%d/", *port)
	if err := http.ListenAndServe(addr, handler.Mux()); err != nil {
		log.Fatal(err)
	}
}

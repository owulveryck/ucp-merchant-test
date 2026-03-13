package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := flag.Int("port", 8888, "port to listen on")
	costPrice := flag.Int("cost-price", 5000, "cost price in cents (minimum selling price)")
	productName := flag.String("product-name", "Casque Audio", "product name")
	graphURL := flag.String("graph-url", "http://localhost:9000", "shopping graph URL")
	obsURL := flag.String("obs-url", "", "observability hub URL")
	flag.Parse()

	srv := NewArenaServer(*costPrice, *productName, *graphURL, *obsURL, *port)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Arena starting on http://localhost:%d", *port)
	log.Printf("Landing:   http://localhost:%d/", *port)
	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatal(err)
	}
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

package main

import (
	"flag"
	"log"

	"github.com/owulveryck/ucp-merchant-test/pkg/a2a"
)

func main() {
	port := flag.String("port", "9001", "Port to listen on")
	flag.Parse()

	// Create the agent
	agent := NewCustomerGrowthA2AAgent()

	// Start the A2A server
	addr := ":" + *port
	log.Fatal(a2a.Serve(agent, addr))
}

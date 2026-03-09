package main

import "sync"

// Unified in-memory stores used by both REST and MCP transports.
var (
	checkouts   = map[string]*Checkout{}
	orders      = map[string]*Order{}
	carts       = map[string]*Cart{}
	checkoutSeq int
	orderSeq    int
	cartSeq     int
	storeMu     sync.Mutex

	// Checkout metadata for order creation.
	checkoutWebhooks     = map[string]string{}
	checkoutDestinations = map[string]*FulfillmentDestination{}
	checkoutOptionTitles = map[string]string{}

	// MCP order cancel channels (keyed by order ID).
	orderCancelChs   = map[string]chan struct{}{}
	orderCancelChsMu sync.Mutex

	// Address sequence counter for dynamic addresses.
	addrSeqCounter int
	addrSeqMu      sync.Mutex
)

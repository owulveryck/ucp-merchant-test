package main

import (
	"sync"

	"github.com/owulveryck/ucp-merchant-test/internal/idempotency"
	"github.com/owulveryck/ucp-merchant-test/internal/model"
)

// Unified in-memory stores used by both REST and MCP transports.
var (
	checkouts   = map[string]*model.Checkout{}
	orders      = map[string]*model.Order{}
	carts       = map[string]*model.Cart{}
	checkoutSeq int
	orderSeq    int
	cartSeq     int
	storeMu     sync.Mutex

	// Checkout metadata for order creation.
	checkoutWebhooks     = map[string]string{}
	checkoutDestinations = map[string]*model.FulfillmentDestination{}
	checkoutOptionTitles = map[string]string{}

	// MCP order cancel channels (keyed by order ID).
	orderCancelChs   = map[string]chan struct{}{}
	orderCancelChsMu sync.Mutex

	// Address sequence counter for dynamic addresses.
	addrSeqCounter int
	addrSeqMu      sync.Mutex

	// Global idempotency store instance.
	idempotencyStoreInstance = idempotency.NewStore()
)

func resetStores() {
	storeMu.Lock()
	checkouts = map[string]*model.Checkout{}
	orders = map[string]*model.Order{}
	carts = map[string]*model.Cart{}
	checkoutSeq = 0
	orderSeq = 0
	cartSeq = 0
	checkoutWebhooks = map[string]string{}
	checkoutDestinations = map[string]*model.FulfillmentDestination{}
	checkoutOptionTitles = map[string]string{}
	storeMu.Unlock()

	orderCancelChsMu.Lock()
	for _, ch := range orderCancelChs {
		select {
		case <-ch:
		default:
			close(ch)
		}
	}
	orderCancelChs = map[string]chan struct{}{}
	orderCancelChsMu.Unlock()

	addrSeqMu.Lock()
	addrSeqCounter = 0
	addrSeqMu.Unlock()

	// MCP-specific state
	mcpCheckoutStates = map[string]*model.MCPCheckoutState{}
	mcpOrderShipments = map[string]*model.Shipment{}
	mcpOrderOwners = map[string]string{}

	// Idempotency store
	idempotencyStoreInstance.Reset()

	// Dynamic addresses
	shopData.Mu.Lock()
	shopData.DynamicAddresses = make(map[string][]CSVAddress)
	shopData.Mu.Unlock()

	// Session counter
	sessionMu.Lock()
	sessionCounter = 0
	sessionMu.Unlock()

	// OAuth token state
	oauthServer.Reset()
}

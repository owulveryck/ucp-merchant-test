package store

import (
	"fmt"
	"sync"

	"github.com/owulveryck/ucp-merchant-test/internal/model"
)

// Store encapsulates all in-memory data maps and sequences.
type Store struct {
	mu sync.Mutex

	Checkouts map[string]*model.Checkout
	Orders    map[string]*model.Order
	Carts     map[string]*model.Cart

	CheckoutSeq int
	OrderSeq    int
	CartSeq     int

	// Checkout metadata for order creation.
	CheckoutWebhooks     map[string]string
	CheckoutDestinations map[string]*model.FulfillmentDestination
	CheckoutOptionTitles map[string]string

	// MCP-specific state
	MCPCheckoutStates map[string]*model.MCPCheckoutState
	MCPOrderShipments map[string]*model.Shipment
	MCPOrderOwners    map[string]string

	// MCP order cancel channels (keyed by order ID).
	OrderCancelChs   map[string]chan struct{}
	OrderCancelChsMu sync.Mutex

	// Address sequence counter for dynamic addresses.
	AddrSeqCounter int
	AddrSeqMu      sync.Mutex

	// Session tracking
	SessionCounter int
	SessionMu      sync.Mutex
}

// New creates a new empty store.
func New() *Store {
	return &Store{
		Checkouts:            map[string]*model.Checkout{},
		Orders:               map[string]*model.Order{},
		Carts:                map[string]*model.Cart{},
		CheckoutWebhooks:     map[string]string{},
		CheckoutDestinations: map[string]*model.FulfillmentDestination{},
		CheckoutOptionTitles: map[string]string{},
		MCPCheckoutStates:    map[string]*model.MCPCheckoutState{},
		MCPOrderShipments:    map[string]*model.Shipment{},
		MCPOrderOwners:       map[string]string{},
		OrderCancelChs:       map[string]chan struct{}{},
	}
}

// Lock locks the main store mutex.
func (s *Store) Lock() { s.mu.Lock() }

// Unlock unlocks the main store mutex.
func (s *Store) Unlock() { s.mu.Unlock() }

// NewSessionID generates a new session ID.
func (s *Store) NewSessionID() string {
	s.SessionMu.Lock()
	defer s.SessionMu.Unlock()
	s.SessionCounter++
	return fmt.Sprintf("session-%04d", s.SessionCounter)
}

// Reset clears all store state.
func (s *Store) Reset() {
	s.mu.Lock()
	s.Checkouts = map[string]*model.Checkout{}
	s.Orders = map[string]*model.Order{}
	s.Carts = map[string]*model.Cart{}
	s.CheckoutSeq = 0
	s.OrderSeq = 0
	s.CartSeq = 0
	s.CheckoutWebhooks = map[string]string{}
	s.CheckoutDestinations = map[string]*model.FulfillmentDestination{}
	s.CheckoutOptionTitles = map[string]string{}
	s.mu.Unlock()

	s.OrderCancelChsMu.Lock()
	for _, ch := range s.OrderCancelChs {
		select {
		case <-ch:
		default:
			close(ch)
		}
	}
	s.OrderCancelChs = map[string]chan struct{}{}
	s.OrderCancelChsMu.Unlock()

	s.AddrSeqMu.Lock()
	s.AddrSeqCounter = 0
	s.AddrSeqMu.Unlock()

	s.MCPCheckoutStates = map[string]*model.MCPCheckoutState{}
	s.MCPOrderShipments = map[string]*model.Shipment{}
	s.MCPOrderOwners = map[string]string{}

	s.SessionMu.Lock()
	s.SessionCounter = 0
	s.SessionMu.Unlock()
}

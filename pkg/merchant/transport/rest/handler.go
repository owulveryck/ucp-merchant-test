// Package rest implements the UCP Shopping Service REST transport.
//
// It exposes a UCP-compliant HTTP API (version 2026-01-11) for checkout
// session management, order retrieval, and test simulation. All business
// logic is delegated to the merchant.Merchant interface; this package
// handles only HTTP request parsing, response formatting, CORS,
// idempotency, UCP version negotiation, and webhook delivery.
package rest

import (
	"net/http"
	"sync"

	"github.com/owulveryck/ucp-merchant-test/pkg/auth"
	"github.com/owulveryck/ucp-merchant-test/pkg/idempotency"
	"github.com/owulveryck/ucp-merchant-test/pkg/merchant"
)

// Server is the REST transport for a UCP-compliant merchant.
// It translates HTTP requests into merchant.Merchant method calls
// and formats the results as JSON responses per the UCP Shopping
// Service REST schema.
type Server struct {
	merchant         merchant.Merchant
	auth             *auth.OAuthServer
	idempotency      *idempotency.Store
	simulationSecret string
	scheme           func() string
	listenPort       func() int

	// Transport-specific state
	mu          sync.Mutex
	webhookURLs map[string]string // checkoutID -> webhook URL
}

// Option configures a Server.
type Option func(*Server)

// WithIdempotency sets the idempotency store for request deduplication.
func WithIdempotency(store *idempotency.Store) Option {
	return func(s *Server) { s.idempotency = store }
}

// WithSimulationSecret sets the secret required for test simulation endpoints.
func WithSimulationSecret(secret string) Option {
	return func(s *Server) { s.simulationSecret = secret }
}

// WithScheme sets the URL scheme function (http or https).
func WithScheme(fn func() string) Option {
	return func(s *Server) { s.scheme = fn }
}

// WithListenPort sets the function returning the server's listen port.
func WithListenPort(fn func() int) Option {
	return func(s *Server) { s.listenPort = fn }
}

// New creates a new REST transport server.
func New(m merchant.Merchant, authServer *auth.OAuthServer, opts ...Option) *Server {
	s := &Server{
		merchant:    m,
		auth:        authServer,
		webhookURLs: map[string]string{},
		scheme:      func() string { return "http" },
		listenPort:  func() int { return 8081 },
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Handler returns an http.Handler with all REST routes registered.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/shopping-api/checkout-sessions/", s.handleCheckoutSessions)
	mux.HandleFunc("/shopping-api/checkout-sessions", s.handleCheckoutSessions)
	mux.HandleFunc("/orders/", s.handleOrders)
	mux.HandleFunc("/testing/simulate-shipping/", s.simulateShipping)
	return mux
}

// Reset clears transport-specific state (webhook URLs).
func (s *Server) Reset() {
	s.mu.Lock()
	s.webhookURLs = map[string]string{}
	s.mu.Unlock()
	if s.idempotency != nil {
		s.idempotency.Reset()
	}
}

// SetWebhookURL stores a webhook URL for a checkout ID.
func (s *Server) SetWebhookURL(checkoutID, url string) {
	s.mu.Lock()
	s.webhookURLs[checkoutID] = url
	s.mu.Unlock()
}

// GetWebhookURL retrieves the webhook URL for a checkout ID.
func (s *Server) GetWebhookURL(checkoutID string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.webhookURLs[checkoutID]
}

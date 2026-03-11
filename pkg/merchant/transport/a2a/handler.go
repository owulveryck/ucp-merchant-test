package a2a

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	a2alib "github.com/a2aproject/a2a-go/a2a"
	"github.com/a2aproject/a2a-go/a2asrv"

	"github.com/owulveryck/ucp-merchant-test/pkg/auth"
	"github.com/owulveryck/ucp-merchant-test/pkg/merchant"
	"github.com/owulveryck/ucp-merchant-test/pkg/ucp"
)

// Server is the A2A transport for a UCP-compliant merchant.
//
// It wraps the a2a-go library to provide a JSON-RPC 2.0 endpoint
// that delegates to the merchant.Merchant interface. The server
// manages per-context session state (active checkout IDs) to support
// multi-turn A2A conversations.
type Server struct {
	merchant       merchant.Merchant
	auth           *auth.OAuthServer
	jsonrpcHandler http.Handler
	merchantName   string
	listenPort     func() int
	schemeFn       func() string

	mu       sync.Mutex
	sessions map[string]*sessionState
}

// sessionState tracks per-context state for an A2A conversation.
type sessionState struct {
	ownerID    string
	country    ucp.Country
	checkoutID string
}

// Option configures a [Server].
type Option func(*Server)

// WithMerchantName sets the merchant name returned in the Agent Card.
func WithMerchantName(name string) Option {
	return func(s *Server) { s.merchantName = name }
}

// WithListenPort sets the function returning the server's listen port.
// The function is called on each request to support dynamic port assignment.
func WithListenPort(fn func() int) Option {
	return func(s *Server) { s.listenPort = fn }
}

// WithScheme sets the function returning the URL scheme ("http" or "https").
func WithScheme(fn func() string) Option {
	return func(s *Server) { s.schemeFn = fn }
}

// New creates a new A2A transport server.
//
// The server registers an [a2asrv.AgentExecutor] backed by the merchant
// implementation and creates a JSON-RPC handler via a2a-go.
func New(m merchant.Merchant, authServer *auth.OAuthServer, opts ...Option) *Server {
	s := &Server{
		merchant:   m,
		auth:       authServer,
		listenPort: func() int { return 8081 },
		schemeFn:   func() string { return "http" },
		sessions:   make(map[string]*sessionState),
	}
	for _, opt := range opts {
		opt(s)
	}

	exec := &executor{server: s}
	reqHandler := a2asrv.NewHandler(exec)
	s.jsonrpcHandler = a2asrv.NewJSONRPCHandler(reqHandler)

	return s
}

// ServeHTTP implements [http.Handler] for the A2A JSON-RPC endpoint.
//
// It sets CORS headers, validates bearer tokens, injects auth context,
// and delegates to the a2a-go JSON-RPC handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if s.auth.IsTokenExpired(r) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "token_expired"})
		return
	}

	ctx := r.Context()
	userID := s.auth.ExtractUserFromToken(r)
	userCountry := s.auth.ExtractUserCountry(r)
	ctx = context.WithValue(ctx, ctxUserID, userID)
	ctx = context.WithValue(ctx, ctxUserCountry, userCountry)

	s.jsonrpcHandler.ServeHTTP(w, r.WithContext(ctx))
}

// HandleAgentCard serves GET /.well-known/agent-card.json.
//
// It returns an A2A Agent Card advertising the UCP extension and
// shopping capabilities. The card is built dynamically to reflect
// the current server URL.
func (s *Server) HandleAgentCard(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	card := s.buildAgentCard()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(card)
}

// Reset clears all A2A session state.
func (s *Server) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions = make(map[string]*sessionState)
}

// buildAgentCard constructs the A2A Agent Card for the current server.
func (s *Server) buildAgentCard() *a2alib.AgentCard {
	name := s.merchantName
	if name == "" {
		name = "UCP Merchant"
	}
	url := fmt.Sprintf("%s://localhost:%d/a2a", s.schemeFn(), s.listenPort())

	return &a2alib.AgentCard{
		Name:               name,
		Description:        "UCP Shopping Service A2A Agent",
		Version:            "1.0.0",
		URL:                url,
		DefaultInputModes:  []string{"application/json"},
		DefaultOutputModes: []string{"application/json"},
		Capabilities: a2alib.AgentCapabilities{
			Extensions: []a2alib.AgentExtension{
				{
					URI:         ucpExtensionURI,
					Description: "Business agent supporting UCP",
					Params: map[string]any{
						"capabilities": map[string]any{
							"dev.ucp.shopping.checkout": []any{
								map[string]any{"version": "2026-01-11"},
							},
							"dev.ucp.shopping.fulfillment": []any{
								map[string]any{
									"version": "2026-01-11",
									"extends": "dev.ucp.shopping.checkout",
								},
							},
							"dev.ucp.shopping.discount": []any{
								map[string]any{
									"version": "2026-01-11",
									"extends": "dev.ucp.shopping.checkout",
								},
							},
							"dev.ucp.shopping.buyer_consent": []any{
								map[string]any{
									"version": "2026-01-11",
									"extends": "dev.ucp.shopping.checkout",
								},
							},
						},
					},
				},
			},
		},
		Skills: []a2alib.AgentSkill{
			{
				ID:          "catalog",
				Name:        "Product Catalog",
				Description: "Browse, search, and look up products in the merchant catalog",
			},
			{
				ID:          "cart",
				Name:        "Shopping Cart",
				Description: "Create, view, update, and cancel shopping carts",
			},
			{
				ID:          "checkout",
				Name:        "Checkout",
				Description: "Create, update, complete, and cancel checkout sessions",
			},
			{
				ID:          "orders",
				Name:        "Order Management",
				Description: "View, list, and cancel orders",
			},
		},
	}
}

const ucpExtensionURI = "https://ucp.dev/specification/reference?v=2026-01-11"

// setSessionCheckout stores the active checkout ID for a context.
func (s *Server) setSessionCheckout(contextID, checkoutID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[contextID]
	if !ok {
		sess = &sessionState{}
		s.sessions[contextID] = sess
	}
	sess.checkoutID = checkoutID
}

// resolveCheckoutID returns the checkout ID from the action data or
// falls back to the session's active checkout.
func (s *Server) resolveCheckoutID(ac *actionContext) string {
	if id, ok := ac.data["id"].(string); ok && id != "" {
		return id
	}
	if id, ok := ac.data["checkout_id"].(string); ok && id != "" {
		return id
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess, ok := s.sessions[ac.contextID]; ok {
		return sess.checkoutID
	}
	return ""
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, UCP-Agent, X-A2A-Extensions")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Type")
}

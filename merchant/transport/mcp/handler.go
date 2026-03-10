// Package mcp implements the UCP Shopping Service MCP transport.
//
// It exposes a JSON-RPC 2.0 endpoint (MCP protocol version 2025-03-26)
// that provides tool-based access to UCP shopping capabilities. The
// transport handles protocol negotiation (initialize, tools/list),
// session management (Mcp-Session-Id), argument parsing, checkout
// hash management for buyer approval flows, and product image
// encoding. All business logic is delegated to the merchant.Merchant
// interface.
package mcp

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/owulveryck/ucp-merchant-test/internal/auth"
	"github.com/owulveryck/ucp-merchant-test/internal/model"
	"github.com/owulveryck/ucp-merchant-test/merchant"
)

// Server is the MCP (Model Context Protocol) transport for a
// UCP-compliant merchant. It implements http.Handler and serves
// JSON-RPC 2.0 requests at a single HTTP POST endpoint, translating
// tool calls into merchant.Merchant method invocations.
type Server struct {
	merchant     merchant.Merchant
	auth         *auth.OAuthServer
	merchantName string
	listenPort   func() int

	mu             sync.Mutex
	sessionCounter int
}

// Option configures a Server.
type Option func(*Server)

// WithMerchantName sets the merchant name returned in initialize responses.
func WithMerchantName(name string) Option {
	return func(s *Server) { s.merchantName = name }
}

// WithListenPort sets the function returning the server's listen port.
func WithListenPort(fn func() int) Option {
	return func(s *Server) { s.listenPort = fn }
}

// New creates a new MCP transport server.
func New(m merchant.Merchant, authServer *auth.OAuthServer, opts ...Option) *Server {
	s := &Server{
		merchant:   m,
		auth:       authServer,
		listenPort: func() int { return 8081 },
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Reset clears transport-specific state.
func (s *Server) Reset() {
	s.mu.Lock()
	s.sessionCounter = 0
	s.mu.Unlock()
}

func (s *Server) newSessionID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessionCounter++
	return fmt.Sprintf("session-%04d", s.sessionCounter)
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Mcp-Session-Id, Authorization")
	w.Header().Set("Access-Control-Expose-Headers", "Mcp-Session-Id")
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

// ServeHTTP implements http.Handler for the MCP endpoint.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check for expired Bearer token
	if s.auth.IsTokenExpired(r) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "token_expired"})
		return
	}

	userID := s.auth.ExtractUserFromToken(r)
	userCountry := s.auth.ExtractUserCountry(r)

	var req model.JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, model.JSONRPCResponse{
			JSONRPC: "2.0",
			Error:   &model.RPCError{Code: -32700, Message: "Parse error"},
		})
		return
	}

	// Assign session ID on initialize
	if req.Method == "initialize" {
		sid := s.newSessionID()
		w.Header().Set("Mcp-Session-Id", sid)
	} else if sid := r.Header.Get("Mcp-Session-Id"); sid != "" {
		w.Header().Set("Mcp-Session-Id", sid)
	}

	switch req.Method {
	case "initialize":
		writeJSON(w, model.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: model.MCPInitializeResult{
				ProtocolVersion: "2025-03-26",
				Capabilities:    model.MCPCapabilities{Tools: map[string]any{}},
				ServerInfo:      model.MCPServerInfo{Name: s.merchantName, Version: "1.0.0"},
			},
		})

	case "notifications/initialized":
		w.WriteHeader(http.StatusNoContent)

	case "tools/list":
		writeJSON(w, model.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: model.MCPToolsListResult{
				Tools: getToolDefinitions(),
			},
		})

	case "tools/call":
		s.handleToolCall(w, req, userID, userCountry)

	default:
		writeJSON(w, model.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &model.RPCError{Code: -32601, Message: fmt.Sprintf("Method not found: %s", req.Method)},
		})
	}
}

func (s *Server) handleToolCall(w http.ResponseWriter, req model.JSONRPCRequest, userID, userCountry string) {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		writeJSON(w, model.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &model.RPCError{Code: -32602, Message: "Invalid params"},
		})
		return
	}

	handlers := map[string]func(map[string]interface{}, string, string) (interface{}, error){
		"list_products":       s.handleListProducts,
		"get_product_details": s.handleGetProductDetails,
		"search_catalog":      s.handleSearchCatalog,
		"lookup_product":      s.handleLookupProduct,
		"create_cart":         s.handleCreateCart,
		"get_cart":            s.handleGetCart,
		"update_cart":         s.handleUpdateCart,
		"cancel_cart":         s.handleCancelCart,
		"create_checkout":     s.handleCreateCheckout,
		"get_checkout":        s.handleGetCheckout,
		"update_checkout":     s.handleUpdateCheckout,
		"complete_checkout":   s.handleCompleteCheckout,
		"cancel_checkout":     s.handleCancelCheckout,
		"get_order":           s.handleGetOrder,
		"list_orders":         s.handleListOrders,
		"cancel_order":        s.handleCancelOrder,
	}

	handler, ok := handlers[params.Name]
	if !ok {
		writeJSON(w, model.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &model.RPCError{Code: -32602, Message: fmt.Sprintf("Unknown tool: %s", params.Name)},
		})
		return
	}

	result, err := handler(params.Arguments, userID, userCountry)
	if err != nil {
		writeJSON(w, model.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: model.MCPToolResult{
				Content: []model.MCPContentBlock{
					{Type: "text", Text: fmt.Sprintf("Error: %s", err.Error())},
				},
				IsError: true,
			},
		})
		return
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")

	content := []model.MCPContentBlock{
		{Type: "text", Text: string(resultJSON)},
	}

	// Extract image URLs and add as image content blocks (cap at 5)
	imageURLs := extractImageURLs(result)
	if len(imageURLs) > 5 {
		imageURLs = imageURLs[:5]
	}
	for _, imgURL := range imageURLs {
		data, mime, err := fetchAndEncodeImage(imgURL)
		if err != nil {
			log.Printf("Failed to fetch image %s: %v", imgURL, err)
			continue
		}
		content = append(content, model.MCPContentBlock{
			Type:     "image",
			Data:     data,
			MimeType: mime,
		})
	}

	writeJSON(w, model.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: model.MCPToolResult{
			Content: content,
		},
	})
}

// Package mcp implements the UCP Shopping Service MCP transport.
//
// It wraps the mcp-go library to provide a JSON-RPC 2.0 / Streamable HTTP
// endpoint with tool-based access to UCP shopping capabilities. All business
// logic is delegated to the merchant.Merchant interface.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"

	"github.com/owulveryck/ucp-merchant-test/pkg/auth"
	"github.com/owulveryck/ucp-merchant-test/pkg/merchant"
)

// permissiveSessionIDManager generates session IDs on initialize but
// accepts any (or no) session ID on subsequent requests — matching the
// previous hand-rolled implementation's behaviour.
type permissiveSessionIDManager struct {
	counter int64
}

func (m *permissiveSessionIDManager) Generate() string {
	n := atomic.AddInt64(&m.counter, 1)
	return fmt.Sprintf("session-%04d", n)
}

func (m *permissiveSessionIDManager) Validate(string) (bool, error)  { return false, nil }
func (m *permissiveSessionIDManager) Terminate(string) (bool, error) { return false, nil }

// contextKey is used for storing auth data in request context.
type contextKey int

const (
	ctxUserID contextKey = iota
	ctxUserCountry
)

// Server is the MCP transport for a UCP-compliant merchant.
type Server struct {
	mcpServer    *mcpserver.MCPServer
	httpServer   *mcpserver.StreamableHTTPServer
	merchant     merchant.Merchant
	auth         *auth.OAuthServer
	merchantName string
	listenPort   func() int
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

	name := s.merchantName
	if name == "" {
		name = "UCP Merchant"
	}

	s.mcpServer = mcpserver.NewMCPServer(name, "1.0.0")
	registerTools(s.mcpServer, s)

	s.httpServer = mcpserver.NewStreamableHTTPServer(s.mcpServer,
		mcpserver.WithHTTPContextFunc(s.contextFunc),
		mcpserver.WithSessionIdManager(&permissiveSessionIDManager{}),
	)

	return s
}

// contextFunc extracts auth info from the HTTP request and stores it in the context.
func (s *Server) contextFunc(ctx context.Context, r *http.Request) context.Context {
	userID := s.auth.ExtractUserFromToken(r)
	userCountry := s.auth.ExtractUserCountry(r)
	ctx = context.WithValue(ctx, ctxUserID, userID)
	ctx = context.WithValue(ctx, ctxUserCountry, userCountry)
	return ctx
}

// Reset clears transport-specific state.
func (s *Server) Reset() {
	// mcp-go manages its own session state; nothing to reset here.
}

// ServeHTTP implements http.Handler for the MCP endpoint.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Check for expired Bearer token
	if s.auth.IsTokenExpired(r) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "token_expired"})
		return
	}

	s.httpServer.ServeHTTP(w, r)
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Mcp-Session-Id, Authorization")
	w.Header().Set("Access-Control-Expose-Headers", "Mcp-Session-Id")
}

// toolResultFromError creates an error tool result for tool-level errors.
func toolResultFromError(err error) *mcp.CallToolResult {
	return mcp.NewToolResultError("Error: " + err.Error())
}

// toolResultFromJSON creates a text tool result from a JSON-marshaled value,
// optionally appending image content blocks.
func toolResultFromJSON(result interface{}, imageURLs []string) *mcp.CallToolResult {
	resultJSON, _ := json.MarshalIndent(result, "", "  ")

	content := []mcp.Content{
		mcp.TextContent{
			Type: "text",
			Text: string(resultJSON),
		},
	}

	if len(imageURLs) > 5 {
		imageURLs = imageURLs[:5]
	}
	for _, imgURL := range imageURLs {
		data, mime, err := fetchAndEncodeImage(imgURL)
		if err != nil {
			continue
		}
		content = append(content, mcp.ImageContent{
			Type:     "image",
			Data:     data,
			MIMEType: mime,
		})
	}

	return &mcp.CallToolResult{
		Content: content,
	}
}

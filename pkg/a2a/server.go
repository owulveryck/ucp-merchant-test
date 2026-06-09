// Package a2a provides Agent-to-Agent communication primitives.
package a2a

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Server is an HTTP server that exposes an Agent via JSON-RPC 2.0.
type Server struct {
	agent Agent
	mux   *http.ServeMux
}

// NewServer creates a new A2A server for the given agent.
func NewServer(agent Agent) *Server {
	s := &Server{
		agent: agent,
		mux:   http.NewServeMux(),
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.mux.HandleFunc("/a2a", s.handleJSONRPC)
	s.mux.HandleFunc("/identity", s.handleIdentity)
	s.mux.HandleFunc("/methods", s.handleMethods)
	s.mux.HandleFunc("/health", s.handleHealth)
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// handleJSONRPC handles JSON-RPC 2.0 requests.
func (s *Server) handleJSONRPC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, ParseError, "Parse error", nil, nil)
		return
	}

	// Validate JSON-RPC version
	if req.JSONRPC != "2.0" {
		s.writeError(w, InvalidRequest, "Invalid JSON-RPC version", nil, req.ID)
		return
	}

	// Call agent
	ctx := context.Background()
	result, err := s.agent.HandleRequest(ctx, req.Method, req.Params)
	if err != nil {
		s.writeError(w, InternalError, err.Error(), nil, req.ID)
		return
	}

	// Write success response
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      req.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleIdentity returns the agent's identity.
func (s *Server) handleIdentity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.agent.Identity())
}

// handleMethods returns the list of supported methods.
func (s *Server) handleMethods(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"methods": s.agent.SupportedMethods(),
	})
}

// handleHealth returns health status.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"agent":  s.agent.Identity().Name,
	})
}

// writeError writes a JSON-RPC error response.
func (s *Server) writeError(w http.ResponseWriter, code int, message string, data interface{}, id interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		Error: &RPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
		ID: id,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // JSON-RPC errors use 200 status
	json.NewEncoder(w).Encode(resp)
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe(addr string) error {
	identity := s.agent.Identity()
	log.Printf("[%s] Starting A2A server on %s", identity.Name, addr)
	log.Printf("[%s] Department: %s", identity.Name, identity.Department)
	log.Printf("[%s] Role: %s", identity.Name, identity.Role)
	log.Printf("[%s] Endpoints:", identity.Name)
	log.Printf("  - POST   %s/a2a       (JSON-RPC 2.0)", addr)
	log.Printf("  - GET    %s/identity  (Agent identity)", addr)
	log.Printf("  - GET    %s/methods   (Supported methods)", addr)
	log.Printf("  - GET    %s/health    (Health check)", addr)

	return http.ListenAndServe(addr, s)
}

// Serve is an alias for ListenAndServe.
func Serve(agent Agent, addr string) error {
	server := NewServer(agent)
	return server.ListenAndServe(addr)
}

// FormatMessage formats a conversational message from an agent.
func FormatMessage(agentName, department string, content string) string {
	return fmt.Sprintf("Bonjour, je suis %s du département %s. %s", agentName, department, content)
}

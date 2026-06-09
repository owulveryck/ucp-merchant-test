// Package a2a provides Agent-to-Agent communication primitives using JSON-RPC 2.0.
package a2a

// JSONRPCRequest represents a JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
	ID      interface{}            `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

// RPCError represents a JSON-RPC 2.0 error.
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// AgentIdentity represents an agent's identity card.
type AgentIdentity struct {
	Name       string `json:"name"`
	Department string `json:"department"`
	Role       string `json:"role"`
	Version    string `json:"version,omitempty"`
}

// AgentResponse is the standard response format from an agent.
type AgentResponse struct {
	Agent    AgentIdentity `json:"agent"`
	Message  string        `json:"message"`
	Decision interface{}   `json:"decision"`
}

// Standard JSON-RPC error codes.
const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

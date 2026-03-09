package model

import "encoding/json"

// JSONRPCRequest is an incoming JSON-RPC 2.0 request for the MCP transport.
// MCP uses JSON-RPC 2.0 over HTTP, where each UCP operation maps to a named
// tool method (e.g., "tools/call" for create_checkout, search_catalog).
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse is an outgoing JSON-RPC 2.0 response, carrying either a
// Result on success or an Error on failure.
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError is a JSON-RPC 2.0 error object with a numeric code, human-readable
// message, and optional structured data.
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ToolDef defines an MCP tool with its name, human-readable description, and
// JSON Schema for input validation. Tools map to UCP operations (e.g.,
// create_checkout, search_catalog).
type ToolDef struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

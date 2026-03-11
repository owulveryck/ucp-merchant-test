package a2aclient

import "encoding/json"

// JSONRPCRequest is a JSON-RPC 2.0 request for A2A message/send.
type JSONRPCRequest struct {
	JSONRPC string     `json:"jsonrpc"`
	Method  string     `json:"method"`
	ID      string     `json:"id"`
	Params  SendParams `json:"params"`
}

// SendParams wraps the message payload for message/send.
type SendParams struct {
	Message Message `json:"message"`
}

// Message is an A2A message with role and parts.
type Message struct {
	Role  string `json:"role"`
	Parts []Part `json:"parts"`
}

// Part is a content part in an A2A message.
// Use either Text (for TextPart) or Data (for DataPart).
type Part struct {
	Type string         `json:"type"`
	Text string         `json:"text,omitempty"`
	Data map[string]any `json:"data,omitempty"`
}

// JSONRPCResponse is a JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      string          `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError is a JSON-RPC 2.0 error object.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MessageResponse is the parsed result of a message/send call.
// The A2A server returns a Message directly (not a Task).
type MessageResponse struct {
	Role  string          `json:"role"`
	Parts json.RawMessage `json:"parts"`
}

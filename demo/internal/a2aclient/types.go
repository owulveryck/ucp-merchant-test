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
	MessageID string `json:"messageId"`
	Role      string `json:"role"`
	Parts     []Part `json:"parts"`
}

// Part is a content part in an A2A message.
// Use either Text (for TextPart) or Data (for DataPart).
type Part struct {
	Kind string         `json:"kind"`
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

// TaskResponse is the a2a-go response envelope (a Task, not a bare Message).
type TaskResponse struct {
	ID        string          `json:"id"`
	ContextID string          `json:"contextId"`
	Status    TaskStatus      `json:"status"`
	History   json.RawMessage `json:"history,omitempty"`
}

// TaskStatus holds the state and optional message of a task.
type TaskStatus struct {
	State   string           `json:"state"`
	Message *MessageResponse `json:"message,omitempty"`
}

// MessageResponse is the parsed result message within a task.
type MessageResponse struct {
	Role  string          `json:"role"`
	Parts json.RawMessage `json:"parts"`
}

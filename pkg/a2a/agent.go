// Package a2a provides Agent-to-Agent communication primitives.
package a2a

import "context"

// Agent represents an autonomous agent that can handle requests.
type Agent interface {
	// Identity returns the agent's identity card.
	Identity() AgentIdentity

	// HandleRequest processes a method call and returns a response.
	HandleRequest(ctx context.Context, method string, params map[string]interface{}) (interface{}, error)

	// SupportedMethods returns the list of methods this agent supports.
	SupportedMethods() []string
}

package client

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

type permissiveSessionIDManager struct {
	counter int64
}

func (m *permissiveSessionIDManager) Generate() string {
	n := atomic.AddInt64(&m.counter, 1)
	return fmt.Sprintf("session-%04d", n)
}

func (m *permissiveSessionIDManager) Validate(string) (bool, error)  { return false, nil }
func (m *permissiveSessionIDManager) Terminate(string) (bool, error) { return false, nil }

// MCPServer wraps the shopping Agent as an MCP StreamableHTTP server.
type MCPServer struct {
	agent      *Agent
	httpServer *mcpserver.StreamableHTTPServer
}

// NewMCPServer creates an MCP server exposing the agent as a "shopping" tool.
func NewMCPServer(agent *Agent) *MCPServer {
	srv := mcpserver.NewMCPServer("Shopping Assistant", "1.0.0")

	tool := mcp.NewTool("shopping",
		mcp.WithDescription("Execute a shopping task: search products, compare prices, checkout across merchants"),
		mcp.WithString("instruction", mcp.Required(), mcp.Description("The shopping task description")),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		instruction, ok := req.GetArguments()["instruction"].(string)
		if !ok || instruction == "" {
			return mcp.NewToolResultError("instruction parameter is required"), nil
		}

		result, err := agent.Run(ctx, instruction)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("agent error: %v", err)), nil
		}

		return mcp.NewToolResultText(result), nil
	})

	httpServer := mcpserver.NewStreamableHTTPServer(srv,
		mcpserver.WithSessionIdManager(&permissiveSessionIDManager{}),
	)

	return &MCPServer{
		agent:      agent,
		httpServer: httpServer,
	}
}

// ServeHTTP implements http.Handler with CORS headers.
func (s *MCPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Mcp-Session-Id")
	w.Header().Set("Access-Control-Expose-Headers", "Mcp-Session-Id")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	s.httpServer.ServeHTTP(w, r)
}

package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestMCP_Initialize(t *testing.T) {
	ts := newTestServer(t)
	resp, result := ts.mcpRequest("initialize", nil, "")

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	sid := resp.Header.Get("Mcp-Session-Id")
	if sid == "" {
		t.Fatal("expected Mcp-Session-Id header")
	}

	res := result["result"].(map[string]interface{})
	if res["protocolVersion"] != "2025-03-26" {
		t.Errorf("expected protocolVersion 2025-03-26, got %v", res["protocolVersion"])
	}
	if res["capabilities"] == nil {
		t.Error("expected capabilities in response")
	}
	serverInfo := res["serverInfo"].(map[string]interface{})
	if serverInfo["name"] == nil || serverInfo["name"] == "" {
		t.Error("expected serverInfo.name")
	}
}

func TestMCP_NotificationsInitialized(t *testing.T) {
	ts := newTestServer(t)

	// Notifications must not include an "id" field per JSON-RPC 2.0 spec.
	// mcp-go returns 202 Accepted for notifications.
	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	}
	b, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", ts.URL+"/mcp", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 202 {
		t.Fatalf("expected 202, got %d", resp.StatusCode)
	}
}

func TestMCP_ToolsList(t *testing.T) {
	ts := newTestServer(t)
	_, result := ts.mcpRequest("tools/list", nil, "")

	res := result["result"].(map[string]interface{})
	tools, ok := res["tools"].([]interface{})
	if !ok {
		t.Fatal("expected tools array in response")
	}
	if len(tools) == 0 {
		t.Fatal("expected at least one tool definition")
	}

	// Verify each tool has name, description, inputSchema
	for _, tool := range tools {
		td := tool.(map[string]interface{})
		if td["name"] == nil || td["name"] == "" {
			t.Error("tool missing name")
		}
		if td["description"] == nil || td["description"] == "" {
			t.Errorf("tool %v missing description", td["name"])
		}
		if td["inputSchema"] == nil {
			t.Errorf("tool %v missing inputSchema", td["name"])
		}
	}
}

func TestMCP_MethodNotFound(t *testing.T) {
	ts := newTestServer(t)
	_, result := ts.mcpRequest("nonexistent/method", nil, "")

	errObj, ok := result["error"].(map[string]interface{})
	if !ok {
		t.Fatal("expected error in response")
	}
	code := errObj["code"].(float64)
	if code != -32601 {
		t.Errorf("expected error code -32601, got %v", code)
	}
}

func TestMCP_ParseError(t *testing.T) {
	ts := newTestServer(t)

	req, _ := http.NewRequest("POST", ts.URL+"/mcp", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	errObj, ok := result["error"].(map[string]interface{})
	if !ok {
		t.Fatal("expected error in response")
	}
	code := errObj["code"].(float64)
	if code != -32700 {
		t.Errorf("expected error code -32700, got %v", code)
	}
}

func TestMCP_InvalidToolParams(t *testing.T) {
	ts := newTestServer(t)
	// tools/call with no params (missing name)
	_, result := ts.mcpRequest("tools/call", nil, "")

	errObj, ok := result["error"].(map[string]interface{})
	if !ok {
		t.Fatal("expected error in response")
	}
	code := errObj["code"].(float64)
	if code != -32602 {
		t.Errorf("expected error code -32602, got %v", code)
	}
}

func TestMCP_UnknownTool(t *testing.T) {
	ts := newTestServer(t)
	_, result := ts.mcpRequest("tools/call", map[string]interface{}{
		"name":      "nonexistent_tool",
		"arguments": map[string]interface{}{},
	}, "")

	errObj, ok := result["error"].(map[string]interface{})
	if !ok {
		t.Fatal("expected error in response")
	}
	code := errObj["code"].(float64)
	if code != -32602 {
		t.Errorf("expected error code -32602, got %v", code)
	}
}

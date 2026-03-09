package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync/atomic"
	"time"
)

var mcpIDCounter int64

// injectToken creates a Bearer token directly in the accessTokens map and returns the token string.
func (ts *testServer) injectToken(userID, country string) string {
	ts.t.Helper()
	token := randomHex(16)
	oauthMu.Lock()
	accessTokens[token] = &tokenEntry{
		Token:     token,
		UserID:    userID,
		Country:   country,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	oauthMu.Unlock()
	return token
}

// injectExpiredToken creates an already-expired Bearer token and returns the token string.
func (ts *testServer) injectExpiredToken(userID, country string) string {
	ts.t.Helper()
	token := randomHex(16)
	oauthMu.Lock()
	accessTokens[token] = &tokenEntry{
		Token:     token,
		UserID:    userID,
		Country:   country,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	oauthMu.Unlock()
	return token
}

// mcpRequest sends a JSON-RPC 2.0 POST to /mcp and returns the parsed response.
func (ts *testServer) mcpRequest(method string, params interface{}, token string) (*http.Response, map[string]interface{}) {
	ts.t.Helper()
	id := atomic.AddInt64(&mcpIDCounter, 1)

	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
	}
	if params != nil {
		reqBody["params"] = params
	}

	b, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", ts.URL+"/mcp", bytes.NewReader(b))
	if err != nil {
		ts.t.Fatalf("Failed to create MCP request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		ts.t.Fatalf("MCP request failed: %v", err)
	}

	respBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	var result map[string]interface{}
	json.Unmarshal(respBody, &result)
	resp.Body = io.NopCloser(bytes.NewReader(respBody))

	return resp, result
}

// mcpInitialize sends an initialize request and returns the Mcp-Session-Id.
func (ts *testServer) mcpInitialize(token string) string {
	ts.t.Helper()
	resp, _ := ts.mcpRequest("initialize", nil, token)
	return resp.Header.Get("Mcp-Session-Id")
}

// mcpToolCall calls a tool and extracts the result content text, returning the unmarshaled map and isError flag.
func (ts *testServer) mcpToolCall(toolName string, args map[string]interface{}, token string) (map[string]interface{}, bool) {
	ts.t.Helper()
	params := map[string]interface{}{
		"name":      toolName,
		"arguments": args,
	}
	_, rpcResp := ts.mcpRequest("tools/call", params, token)

	// Check for JSON-RPC error
	if rpcResp["error"] != nil {
		errObj := rpcResp["error"].(map[string]interface{})
		return map[string]interface{}{
			"rpc_error_code":    errObj["code"],
			"rpc_error_message": errObj["message"],
		}, true
	}

	result, ok := rpcResp["result"].(map[string]interface{})
	if !ok {
		ts.t.Fatalf("mcpToolCall: no result in response")
	}

	isError, _ := result["isError"].(bool)

	content, ok := result["content"].([]interface{})
	if !ok || len(content) == 0 {
		return nil, isError
	}

	first := content[0].(map[string]interface{})
	text, _ := first["text"].(string)

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		// Return raw text as a map entry
		return map[string]interface{}{"_raw": text}, isError
	}
	return parsed, isError
}

// mcpToolCallRaw returns the raw JSON-RPC response for inspecting error codes.
func (ts *testServer) mcpToolCallRaw(toolName string, args map[string]interface{}, token string) (*http.Response, map[string]interface{}) {
	ts.t.Helper()
	params := map[string]interface{}{
		"name":      toolName,
		"arguments": args,
	}
	return ts.mcpRequest("tools/call", params, token)
}

// mcpCreateAndCompleteCheckout creates a checkout, selects shipping, and completes it.
// Returns (checkoutID, orderID).
func (ts *testServer) mcpCreateAndCompleteCheckout(token string) (string, string) {
	ts.t.Helper()

	// Create checkout
	createResult, isErr := ts.mcpToolCall("create_checkout", map[string]interface{}{
		"checkout": map[string]interface{}{
			"line_items": []interface{}{
				map[string]interface{}{
					"item":     map[string]interface{}{"id": "bouquet_roses", "title": "Bouquet of Roses"},
					"quantity": 1,
				},
			},
		},
	}, token)
	if isErr {
		ts.t.Fatalf("mcpCreateAndCompleteCheckout: create_checkout failed: %v", createResult)
	}
	checkoutID := createResult["id"].(string)

	// Get shipping options
	_, isErr = ts.mcpToolCall("get_shipping_options", map[string]interface{}{
		"checkout_id": checkoutID,
	}, token)
	if isErr {
		ts.t.Fatalf("mcpCreateAndCompleteCheckout: get_shipping_options failed")
	}

	// Update with shipping
	updateResult, isErr := ts.mcpToolCall("update_checkout", map[string]interface{}{
		"id": checkoutID,
		"checkout": map[string]interface{}{
			"shipping_option_id": "standard",
		},
	}, token)
	if isErr {
		ts.t.Fatalf("mcpCreateAndCompleteCheckout: update_checkout failed: %v", updateResult)
	}
	checkoutHash := updateResult["checkout_hash"].(string)

	// Complete with hash
	completeResult, isErr := ts.mcpToolCall("complete_checkout", map[string]interface{}{
		"id": checkoutID,
		"approval": map[string]interface{}{
			"checkout_hash": checkoutHash,
		},
	}, token)
	if isErr {
		ts.t.Fatalf("mcpCreateAndCompleteCheckout: complete_checkout failed: %v", completeResult)
	}

	orderMap := completeResult["order"].(map[string]interface{})
	orderID := orderMap["id"].(string)

	return checkoutID, orderID
}

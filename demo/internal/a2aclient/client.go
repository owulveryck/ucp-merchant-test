package a2aclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
)

// Client sends A2A JSON-RPC requests to merchant servers.
type Client struct {
	httpClient *http.Client
	tokens     *TokenCache
	username   string
	country    string
	obsURL     string
	reqID      atomic.Int64
}

// NewClient creates a new A2A client.
func NewClient(username, country, obsURL string) *Client {
	return &Client{
		httpClient: &http.Client{},
		tokens:     NewTokenCache(),
		username:   username,
		country:    country,
		obsURL:     obsURL,
	}
}

// ObsURL returns the observability hub URL.
func (c *Client) ObsURL() string {
	return c.obsURL
}

// ensureToken returns a cached or newly obtained token for the base URL.
func (c *Client) ensureToken(baseURL string) (string, error) {
	tok := c.tokens.Get(baseURL)
	if tok != "" {
		return tok, nil
	}
	tok, expiresIn, err := ObtainToken(c.httpClient, baseURL, c.username, c.country)
	if err != nil {
		return "", err
	}
	c.tokens.Set(baseURL, tok, expiresIn)
	return tok, nil
}

// SendAction sends an A2A message/send request with the given action and data.
// It returns the parsed data from the first DataPart in the response.
func (c *Client) SendAction(baseURL, action string, data map[string]any) (map[string]any, error) {
	token, err := c.ensureToken(baseURL)
	if err != nil {
		return nil, fmt.Errorf("auth: %w", err)
	}

	if data == nil {
		data = make(map[string]any)
	}
	data["action"] = action

	seq := c.reqID.Add(1)
	id := fmt.Sprintf("req-%d", seq)
	msgID := fmt.Sprintf("msg-%d", seq)
	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "message/send",
		ID:      id,
		Params: SendParams{
			Message: Message{
				MessageID: msgID,
				Role:      "user",
				Parts: []Part{
					{Kind: "data", Data: data},
				},
			},
		},
	}

	respBody, err := c.doRequest(baseURL, token, reqBody)
	if err != nil {
		return nil, err
	}

	var rpcResp JSONRPCResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return nil, fmt.Errorf("decode rpc: %w (body: %s)", err, string(respBody))
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("rpc error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	return extractDataFromResult(rpcResp.Result)
}

// SendActionWithPayment sends an A2A message/send with both an action DataPart
// and a payment DataPart (for complete_checkout).
func (c *Client) SendActionWithPayment(baseURL, action string, data map[string]any, payment map[string]any) (map[string]any, error) {
	token, err := c.ensureToken(baseURL)
	if err != nil {
		return nil, fmt.Errorf("auth: %w", err)
	}

	if data == nil {
		data = make(map[string]any)
	}
	data["action"] = action

	seq := c.reqID.Add(1)
	id := fmt.Sprintf("req-%d", seq)
	msgID := fmt.Sprintf("msg-%d", seq)
	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "message/send",
		ID:      id,
		Params: SendParams{
			Message: Message{
				MessageID: msgID,
				Role:      "user",
				Parts: []Part{
					{Kind: "data", Data: data},
					{Kind: "data", Data: map[string]any{
						"a2a.ucp.checkout.payment": payment,
					}},
				},
			},
		},
	}

	respBody, err := c.doRequest(baseURL, token, reqBody)
	if err != nil {
		return nil, err
	}

	var rpcResp JSONRPCResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return nil, fmt.Errorf("decode rpc: %w (body: %s)", err, string(respBody))
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("rpc error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	return extractDataFromResult(rpcResp.Result)
}

// doRequest marshals the request, sends it, and returns the response body.
func (c *Client) doRequest(baseURL, token string, reqBody JSONRPCRequest) ([]byte, error) {
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequest("POST", baseURL+"/a2a", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	return respBody, nil
}

// extractDataFromResult parses the JSON-RPC result, trying TaskResponse first
// (a2a-go wraps responses in a Task), then falling back to bare MessageResponse.
func extractDataFromResult(raw json.RawMessage) (map[string]any, error) {
	var partsRaw json.RawMessage

	// Try parsing as a Task (a2a-go execution manager wraps in Task envelope)
	var task TaskResponse
	if err := json.Unmarshal(raw, &task); err == nil && task.Status.Message != nil {
		partsRaw = task.Status.Message.Parts
	} else {
		// Fall back to bare MessageResponse
		var msg MessageResponse
		if err := json.Unmarshal(raw, &msg); err != nil {
			return nil, fmt.Errorf("decode result: %w (raw: %s)", err, string(raw))
		}
		partsRaw = msg.Parts
	}

	var parts []json.RawMessage
	if err := json.Unmarshal(partsRaw, &parts); err != nil {
		return nil, fmt.Errorf("decode parts: %w", err)
	}

	for _, rawPart := range parts {
		var part struct {
			Kind string         `json:"kind"`
			Data map[string]any `json:"data,omitempty"`
			Text string         `json:"text,omitempty"`
		}
		if err := json.Unmarshal(rawPart, &part); err != nil {
			continue
		}
		if part.Data != nil {
			return part.Data, nil
		}
		if part.Text != "" && len(part.Text) > 6 && part.Text[:6] == "Error:" {
			return nil, fmt.Errorf("%s", part.Text)
		}
	}

	return nil, fmt.Errorf("no data part in response")
}

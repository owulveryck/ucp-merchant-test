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

	id := fmt.Sprintf("req-%d", c.reqID.Add(1))
	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "message/send",
		ID:      id,
		Params: SendParams{
			Message: Message{
				Role: "user",
				Parts: []Part{
					{Type: "data", Data: data},
				},
			},
		},
	}

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

	var rpcResp JSONRPCResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return nil, fmt.Errorf("decode rpc: %w (body: %s)", err, string(respBody))
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("rpc error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	// Parse the result — it should be a Message with parts
	var msgResp MessageResponse
	if err := json.Unmarshal(rpcResp.Result, &msgResp); err != nil {
		return nil, fmt.Errorf("decode message: %w (result: %s)", err, string(rpcResp.Result))
	}

	// Parse parts to find data
	var parts []json.RawMessage
	if err := json.Unmarshal(msgResp.Parts, &parts); err != nil {
		return nil, fmt.Errorf("decode parts: %w", err)
	}

	for _, rawPart := range parts {
		var part struct {
			Type string         `json:"type"`
			Data map[string]any `json:"data,omitempty"`
			Text string         `json:"text,omitempty"`
		}
		if err := json.Unmarshal(rawPart, &part); err != nil {
			continue
		}
		if part.Data != nil {
			return part.Data, nil
		}
		// If it's a text part with an error, return as error
		if part.Text != "" && len(part.Text) > 6 && part.Text[:6] == "Error:" {
			return nil, fmt.Errorf("%s", part.Text)
		}
	}

	return nil, fmt.Errorf("no data part in response")
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

	id := fmt.Sprintf("req-%d", c.reqID.Add(1))
	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "message/send",
		ID:      id,
		Params: SendParams{
			Message: Message{
				Role: "user",
				Parts: []Part{
					{Type: "data", Data: data},
					{Type: "data", Data: map[string]any{
						"a2a.ucp.checkout.payment": payment,
					}},
				},
			},
		},
	}

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

	var rpcResp JSONRPCResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return nil, fmt.Errorf("decode rpc: %w (body: %s)", err, string(respBody))
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("rpc error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	var msgResp MessageResponse
	if err := json.Unmarshal(rpcResp.Result, &msgResp); err != nil {
		return nil, fmt.Errorf("decode message: %w", err)
	}

	var parts []json.RawMessage
	if err := json.Unmarshal(msgResp.Parts, &parts); err != nil {
		return nil, fmt.Errorf("decode parts: %w", err)
	}

	for _, rawPart := range parts {
		var part struct {
			Type string         `json:"type"`
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

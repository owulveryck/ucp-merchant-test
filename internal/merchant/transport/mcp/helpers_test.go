package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	gomcp "github.com/mark3labs/mcp-go/mcp"

	"github.com/owulveryck/ucp-merchant-test/internal/ucp"
)

func TestUserIDFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), ctxUserID, "user1")
	if got := userIDFromContext(ctx); got != "user1" {
		t.Errorf("expected user1, got %s", got)
	}

	if got := userIDFromContext(context.Background()); got != "" {
		t.Errorf("expected empty, got %s", got)
	}
}

func TestUserCountryFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), ctxUserCountry, ucp.Country("US"))
	if got := userCountryFromContext(ctx); got != "US" {
		t.Errorf("expected US, got %s", got)
	}

	if got := userCountryFromContext(context.Background()); got != "" {
		t.Errorf("expected empty, got %s", got)
	}
}

func TestParseLineItemRequests_Valid(t *testing.T) {
	data := map[string]interface{}{
		"line_items": []interface{}{
			map[string]interface{}{
				"product_id": "SKU-001",
				"quantity":   float64(2),
			},
			map[string]interface{}{
				"id": "li_1",
				"item": map[string]interface{}{
					"id": "SKU-002",
				},
				"quantity": float64(1),
			},
		},
	}

	items := parseLineItemRequests(data)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].ProductID != "SKU-001" {
		t.Errorf("expected product_id SKU-001, got %s", items[0].ProductID)
	}
	if items[0].Quantity != 2 {
		t.Errorf("expected quantity 2, got %d", items[0].Quantity)
	}
	if items[1].ID != "li_1" {
		t.Errorf("expected id li_1, got %s", items[1].ID)
	}
	if items[1].Item == nil || items[1].Item.ID != "SKU-002" {
		t.Error("expected item.id SKU-002")
	}
}

func TestParseLineItemRequests_Empty(t *testing.T) {
	data := map[string]interface{}{}
	items := parseLineItemRequests(data)
	if items != nil {
		t.Errorf("expected nil, got %v", items)
	}
}

func TestParseLineItemRequests_InvalidTypes(t *testing.T) {
	data := map[string]interface{}{
		"line_items": []interface{}{
			"not a map",
			42,
			map[string]interface{}{"product_id": "SKU-001", "quantity": float64(1)},
		},
	}
	items := parseLineItemRequests(data)
	if len(items) != 1 {
		t.Errorf("expected 1 valid item, got %d", len(items))
	}
}

func TestParseBuyerRequest_Valid(t *testing.T) {
	data := map[string]interface{}{
		"first_name": "John",
		"last_name":  "Doe",
		"email":      "john@example.com",
		"name":       "John Doe",
	}
	b := parseBuyerRequest(data)
	if b == nil {
		t.Fatal("expected non-nil buyer")
	}
	if b.FirstName != "John" {
		t.Errorf("expected FirstName John, got %s", b.FirstName)
	}
	if b.LastName != "Doe" {
		t.Errorf("expected LastName Doe, got %s", b.LastName)
	}
	if b.Email != "john@example.com" {
		t.Errorf("expected email john@example.com, got %s", b.Email)
	}
	if b.Name != "John Doe" {
		t.Errorf("expected Name John Doe, got %s", b.Name)
	}
}

func TestParseBuyerRequest_Nil(t *testing.T) {
	b := parseBuyerRequest(nil)
	if b != nil {
		t.Error("expected nil for nil input")
	}
}

func TestToolResultFromError(t *testing.T) {
	result := toolResultFromError(errors.New("something broke"))
	if !result.IsError {
		t.Error("expected IsError=true")
	}
	text := extractTextContent(t, result)
	if !strings.HasPrefix(text, "Error: ") {
		t.Errorf("expected 'Error: ...' text, got %q", text)
	}
}

func TestToolResultFromJSON(t *testing.T) {
	data := map[string]string{"id": "123"}
	result := toolResultFromJSON(data, nil)

	if result.IsError {
		t.Error("expected IsError=false")
	}
	text := extractTextContent(t, result)
	var parsed map[string]string
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("content is not valid JSON: %v", err)
	}
	if parsed["id"] != "123" {
		t.Errorf("expected id=123, got %s", parsed["id"])
	}
}

// extractTextContent marshals the first content element to extract the text field.
func extractTextContent(t *testing.T, result *gomcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("no content in result")
	}
	b, _ := json.Marshal(result.Content[0])
	var tc struct {
		Text string `json:"text"`
	}
	json.Unmarshal(b, &tc)
	return tc.Text
}

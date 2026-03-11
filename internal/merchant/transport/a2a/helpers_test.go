package a2a

import (
	"context"
	"testing"

	a2alib "github.com/a2aproject/a2a-go/a2a"

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
	data := map[string]any{
		"line_items": []any{
			map[string]any{
				"product_id": "SKU-001",
				"quantity":   float64(2),
			},
			map[string]any{
				"id": "li_1",
				"item": map[string]any{
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
	data := map[string]any{}
	items := parseLineItemRequests(data)
	if items != nil {
		t.Errorf("expected nil, got %v", items)
	}
}

func TestParseBuyerRequest_Valid(t *testing.T) {
	data := map[string]any{
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
	if b.Email != "john@example.com" {
		t.Errorf("expected email john@example.com, got %s", b.Email)
	}
}

func TestParseBuyerRequest_Nil(t *testing.T) {
	b := parseBuyerRequest(nil)
	if b != nil {
		t.Error("expected nil for nil input")
	}
}

func TestAsDataPart_Value(t *testing.T) {
	dp := a2alib.DataPart{Data: map[string]any{"key": "val"}}
	data, ok := asDataPart(dp)
	if !ok {
		t.Fatal("expected ok=true for DataPart value")
	}
	if data["key"] != "val" {
		t.Errorf("expected key=val, got %v", data["key"])
	}
}

func TestAsDataPart_Pointer(t *testing.T) {
	dp := &a2alib.DataPart{Data: map[string]any{"key": "val"}}
	data, ok := asDataPart(dp)
	if !ok {
		t.Fatal("expected ok=true for *DataPart")
	}
	if data["key"] != "val" {
		t.Errorf("expected key=val, got %v", data["key"])
	}
}

func TestAsDataPart_TextPart(t *testing.T) {
	tp := a2alib.TextPart{Text: "hello"}
	_, ok := asDataPart(tp)
	if ok {
		t.Error("expected ok=false for TextPart")
	}
}

func TestExtractPaymentFromParts(t *testing.T) {
	parts := a2alib.ContentParts{
		a2alib.DataPart{Data: map[string]any{"action": "complete_checkout"}},
		a2alib.DataPart{Data: map[string]any{
			"a2a.ucp.checkout.payment": map[string]any{
				"handler_id": "mock",
				"credential": map[string]any{"token": "success_token"},
			},
		}},
	}

	payment := extractPaymentFromParts(parts)
	if payment == nil {
		t.Fatal("expected non-nil payment")
	}
	if payment["handler_id"] != "mock" {
		t.Errorf("expected handler_id=mock, got %v", payment["handler_id"])
	}
}

func TestExtractPaymentFromParts_NoParts(t *testing.T) {
	parts := a2alib.ContentParts{
		a2alib.DataPart{Data: map[string]any{"action": "complete_checkout"}},
	}

	payment := extractPaymentFromParts(parts)
	if payment != nil {
		t.Error("expected nil payment when no payment DataPart")
	}
}

func TestExtractAction(t *testing.T) {
	msg := a2alib.NewMessage(a2alib.MessageRoleUser,
		a2alib.DataPart{Data: map[string]any{"action": "list_products", "category": "flowers"}},
	)
	action, data := extractAction(msg)
	if action != "list_products" {
		t.Errorf("expected action=list_products, got %s", action)
	}
	if data["category"] != "flowers" {
		t.Errorf("expected category=flowers, got %v", data["category"])
	}
}

func TestExtractAction_NoAction(t *testing.T) {
	msg := a2alib.NewMessage(a2alib.MessageRoleUser,
		a2alib.TextPart{Text: "hello"},
	)
	action, _ := extractAction(msg)
	if action != "" {
		t.Errorf("expected empty action, got %s", action)
	}
}

func TestExtractAction_NilMessage(t *testing.T) {
	action, _ := extractAction(nil)
	if action != "" {
		t.Errorf("expected empty action, got %s", action)
	}
}

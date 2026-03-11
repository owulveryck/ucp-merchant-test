package mcp

import (
	"context"
	"strings"
	"testing"

	"github.com/owulveryck/ucp-merchant-test/pkg/merchant"
	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/merchanttest"
	"github.com/owulveryck/ucp-merchant-test/pkg/model"
	"github.com/owulveryck/ucp-merchant-test/pkg/ucp"
)

func ctxWithUser(userID string, country ucp.Country) context.Context {
	ctx := context.WithValue(context.Background(), ctxUserID, userID)
	return context.WithValue(ctx, ctxUserCountry, country)
}

func TestHandleCreateCheckout_WithLineItems(t *testing.T) {
	mock := merchanttest.NewMock()
	mock.CreateCheckoutFunc = func(ownerID string, country ucp.Country, req *model.CheckoutRequest) (*model.Checkout, string, error) {
		if ownerID != "user1" {
			t.Errorf("expected ownerID user1, got %s", ownerID)
		}
		if len(req.LineItems) != 1 {
			t.Errorf("expected 1 line item, got %d", len(req.LineItems))
		}
		return &model.Checkout{ID: "co_1", Status: "incomplete"}, "hash123", nil
	}

	s := newTestMCPServer(mock)
	result, err := s.handleCreateCheckout(ctxWithUser("user1", "US"), makeToolRequest(map[string]interface{}{
		"checkout": map[string]interface{}{
			"line_items": []interface{}{
				map[string]interface{}{"product_id": "SKU-001", "quantity": float64(1)},
			},
		},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Error("expected no error")
	}
	text := extractText(t, result)
	if !strings.Contains(text, "co_1") {
		t.Error("expected checkout ID in response")
	}
	if !strings.Contains(text, "hash123") {
		t.Error("expected checkout_hash in response")
	}
}

func TestHandleCreateCheckout_WithCartID(t *testing.T) {
	mock := merchanttest.NewMock()
	mock.GetCartFunc = func(id, ownerID string) (*model.Cart, error) {
		return &model.Cart{
			ID: id,
			LineItems: []model.LineItem{
				{ID: "li_1", Item: model.Item{ID: "SKU-001"}, Quantity: 2},
			},
		}, nil
	}
	mock.CreateCheckoutFunc = func(ownerID string, country ucp.Country, req *model.CheckoutRequest) (*model.Checkout, string, error) {
		if len(req.LineItems) != 1 {
			t.Errorf("expected 1 line item from cart, got %d", len(req.LineItems))
		}
		return &model.Checkout{ID: "co_1", Status: "incomplete"}, "", nil
	}

	s := newTestMCPServer(mock)
	result, err := s.handleCreateCheckout(ctxWithUser("user1", "US"), makeToolRequest(map[string]interface{}{
		"checkout": map[string]interface{}{
			"cart_id": "cart_1",
		},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Error("expected no error")
	}
}

func TestHandleGetCheckout_Found(t *testing.T) {
	mock := merchanttest.NewMock()
	mock.GetCheckoutFunc = func(id, ownerID string) (*model.Checkout, string, error) {
		return &model.Checkout{ID: id, Status: "incomplete"}, "hash456", nil
	}

	s := newTestMCPServer(mock)
	result, err := s.handleGetCheckout(ctxWithUser("user1", "US"), makeToolRequest(map[string]interface{}{
		"id": "co_1",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Error("expected no error")
	}
	text := extractText(t, result)
	if !strings.Contains(text, "hash456") {
		t.Error("expected checkout_hash in response")
	}
}

func TestHandleGetCheckout_NotFound(t *testing.T) {
	mock := merchanttest.NewMock()
	mock.GetCheckoutFunc = func(id, ownerID string) (*model.Checkout, string, error) {
		return nil, "", merchant.ErrNotFound
	}

	s := newTestMCPServer(mock)
	result, err := s.handleGetCheckout(ctxWithUser("user1", "US"), makeToolRequest(map[string]interface{}{
		"id": "nonexistent",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for not found")
	}
}

func TestHandleUpdateCheckout(t *testing.T) {
	mock := merchanttest.NewMock()
	mock.UpdateCheckoutFunc = func(id, ownerID string, req *model.CheckoutRequest) (*model.Checkout, string, error) {
		return &model.Checkout{ID: id, Status: "incomplete"}, "", nil
	}

	s := newTestMCPServer(mock)
	result, err := s.handleUpdateCheckout(ctxWithUser("user1", "US"), makeToolRequest(map[string]interface{}{
		"id": "co_1",
		"checkout": map[string]interface{}{
			"line_items": []interface{}{
				map[string]interface{}{"product_id": "SKU-002", "quantity": float64(3)},
			},
		},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Error("expected no error")
	}
}

func TestHandleCompleteCheckout_WithApprovalHash(t *testing.T) {
	mock := merchanttest.NewMock()
	var gotHash string
	mock.CompleteCheckoutFunc = func(id, ownerID string, country ucp.Country, approvalHash string, req *model.CheckoutRequest) (*model.Checkout, *model.Order, string, error) {
		gotHash = approvalHash
		return &model.Checkout{ID: id, Status: "completed"}, &model.Order{ID: "ord_1"}, "", nil
	}

	s := newTestMCPServer(mock)
	result, err := s.handleCompleteCheckout(ctxWithUser("user1", "US"), makeToolRequest(map[string]interface{}{
		"id": "co_1",
		"approval": map[string]interface{}{
			"checkout_hash": "abc123",
		},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Error("expected no error")
	}
	if gotHash != "abc123" {
		t.Errorf("expected approval hash abc123, got %s", gotHash)
	}
}

func TestHandleCompleteCheckout_MissingApproval(t *testing.T) {
	mock := merchanttest.NewMock()
	s := newTestMCPServer(mock)

	result, err := s.handleCompleteCheckout(ctxWithUser("user1", "US"), makeToolRequest(map[string]interface{}{
		"id": "co_1",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for missing approval")
	}
}

func TestHandleCancelCheckout(t *testing.T) {
	mock := merchanttest.NewMock()
	mock.CancelCheckoutFunc = func(id, ownerID string) (*model.Checkout, string, error) {
		return &model.Checkout{ID: id, Status: "canceled"}, "", nil
	}

	s := newTestMCPServer(mock)
	result, err := s.handleCancelCheckout(ctxWithUser("user1", "US"), makeToolRequest(map[string]interface{}{
		"id": "co_1",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Error("expected no error")
	}
	text := extractText(t, result)
	if !strings.Contains(text, "canceled") {
		t.Error("expected canceled status in response")
	}
}

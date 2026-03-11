package a2a

import (
	"context"
	"testing"

	a2alib "github.com/a2aproject/a2a-go/a2a"

	"github.com/owulveryck/ucp-merchant-test/internal/merchant/merchanttest"
	"github.com/owulveryck/ucp-merchant-test/internal/model"
	"github.com/owulveryck/ucp-merchant-test/internal/ucp"
)

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

	s := newTestServer(mock)
	ac := &actionContext{
		userID:    "user1",
		country:   "US",
		contextID: "ctx-1",
		data: map[string]any{
			"checkout": map[string]any{
				"line_items": []any{
					map[string]any{"product_id": "SKU-001", "quantity": float64(1)},
				},
			},
		},
	}
	result, err := s.handleCreateCheckout(context.Background(), ac)
	if err != nil {
		t.Fatal(err)
	}

	co, ok := result["a2a.ucp.checkout"].(map[string]any)
	if !ok {
		t.Fatal("expected a2a.ucp.checkout in result")
	}
	if co["id"] != "co_1" {
		t.Errorf("expected id=co_1, got %v", co["id"])
	}
	if co["checkout_hash"] != "hash123" {
		t.Errorf("expected checkout_hash=hash123, got %v", co["checkout_hash"])
	}
}

func TestHandleCreateCheckout_SessionTracking(t *testing.T) {
	mock := merchanttest.NewMock()
	mock.CreateCheckoutFunc = func(ownerID string, country ucp.Country, req *model.CheckoutRequest) (*model.Checkout, string, error) {
		return &model.Checkout{ID: "co_42", Status: "incomplete"}, "", nil
	}

	s := newTestServer(mock)
	ac := &actionContext{
		userID:    "user1",
		contextID: "ctx-1",
		data: map[string]any{
			"line_items": []any{
				map[string]any{"product_id": "SKU-001", "quantity": float64(1)},
			},
		},
	}
	_, err := s.handleCreateCheckout(context.Background(), ac)
	if err != nil {
		t.Fatal(err)
	}

	// Verify checkout ID was stored in session.
	id := s.resolveCheckoutID(&actionContext{contextID: "ctx-1", data: map[string]any{}})
	if id != "co_42" {
		t.Errorf("expected session checkout co_42, got %s", id)
	}
}

func TestHandleGetCheckout(t *testing.T) {
	mock := merchanttest.NewMock()
	mock.GetCheckoutFunc = func(id, ownerID string) (*model.Checkout, string, error) {
		return &model.Checkout{ID: id, Status: "incomplete"}, "hash456", nil
	}

	s := newTestServer(mock)
	ac := &actionContext{
		userID: "user1",
		data:   map[string]any{"id": "co_1"},
	}
	result, err := s.handleGetCheckout(context.Background(), ac)
	if err != nil {
		t.Fatal(err)
	}

	co := result["a2a.ucp.checkout"].(map[string]any)
	if co["checkout_hash"] != "hash456" {
		t.Errorf("expected checkout_hash=hash456, got %v", co["checkout_hash"])
	}
}

func TestHandleUpdateCheckout(t *testing.T) {
	mock := merchanttest.NewMock()
	mock.UpdateCheckoutFunc = func(id, ownerID string, req *model.CheckoutRequest) (*model.Checkout, string, error) {
		return &model.Checkout{ID: id, Status: "incomplete"}, "", nil
	}

	s := newTestServer(mock)
	ac := &actionContext{
		userID: "user1",
		data: map[string]any{
			"id": "co_1",
			"checkout": map[string]any{
				"line_items": []any{
					map[string]any{"product_id": "SKU-002", "quantity": float64(3)},
				},
			},
		},
	}
	result, err := s.handleUpdateCheckout(context.Background(), ac)
	if err != nil {
		t.Fatal(err)
	}

	co := result["a2a.ucp.checkout"].(map[string]any)
	if co["id"] != "co_1" {
		t.Errorf("expected id=co_1, got %v", co["id"])
	}
}

func TestHandleCompleteCheckout(t *testing.T) {
	mock := merchanttest.NewMock()
	var gotHash string
	mock.CompleteCheckoutFunc = func(id, ownerID string, country ucp.Country, approvalHash string, req *model.CheckoutRequest) (*model.Checkout, *model.Order, string, error) {
		gotHash = approvalHash
		return &model.Checkout{
			ID:     id,
			Status: "completed",
			Order:  &model.OrderRef{ID: "ord_1", PermalinkURL: "https://example.com/orders/ord_1"},
		}, &model.Order{ID: "ord_1"}, "", nil
	}

	s := newTestServer(mock)
	s.setSessionCheckout("ctx-1", "co_1")
	ac := &actionContext{
		userID:    "user1",
		country:   "US",
		contextID: "ctx-1",
		data:      map[string]any{"checkout_hash": "abc123"},
		parts: a2alib.ContentParts{
			a2alib.DataPart{Data: map[string]any{"action": "complete_checkout", "checkout_hash": "abc123"}},
			a2alib.DataPart{Data: map[string]any{
				"a2a.ucp.checkout.payment": map[string]any{
					"handler_id": "mock",
					"credential": map[string]any{"token": "success_token"},
				},
			}},
		},
	}
	result, err := s.handleCompleteCheckout(context.Background(), ac)
	if err != nil {
		t.Fatal(err)
	}

	if gotHash != "abc123" {
		t.Errorf("expected approval hash abc123, got %s", gotHash)
	}

	co := result["a2a.ucp.checkout"].(map[string]any)
	if co["status"] != "completed" {
		t.Errorf("expected status=completed, got %v", co["status"])
	}

	order, ok := co["order"].(map[string]any)
	if !ok {
		t.Fatal("expected order in completed checkout")
	}
	if order["id"] != "ord_1" {
		t.Errorf("expected order id=ord_1, got %v", order["id"])
	}
}

func TestHandleCancelCheckout(t *testing.T) {
	mock := merchanttest.NewMock()
	mock.CancelCheckoutFunc = func(id, ownerID string) (*model.Checkout, string, error) {
		return &model.Checkout{ID: id, Status: "canceled"}, "", nil
	}

	s := newTestServer(mock)
	ac := &actionContext{
		userID: "user1",
		data:   map[string]any{"id": "co_1"},
	}
	result, err := s.handleCancelCheckout(context.Background(), ac)
	if err != nil {
		t.Fatal(err)
	}

	co := result["a2a.ucp.checkout"].(map[string]any)
	if co["status"] != "canceled" {
		t.Errorf("expected status=canceled, got %v", co["status"])
	}
}

func TestParsePaymentData(t *testing.T) {
	data := map[string]any{
		"handler_id": "mock_payment_handler",
		"credential": map[string]any{
			"token": "success_token",
		},
	}

	pd := parsePaymentData(data)
	if pd.HandlerID != "mock_payment_handler" {
		t.Errorf("expected handler_id=mock_payment_handler, got %s", pd.HandlerID)
	}
	if pd.Credential == nil || pd.Credential.Token != "success_token" {
		t.Error("expected credential.token=success_token")
	}
}

func TestParsePaymentData_WithInstruments(t *testing.T) {
	data := map[string]any{
		"instruments": []any{
			map[string]any{
				"handler_id": "gpay_1234",
				"credential": map[string]any{
					"token": "payment_token",
				},
			},
		},
	}

	pd := parsePaymentData(data)
	if pd.Credential == nil || pd.Credential.Token != "payment_token" {
		t.Error("expected credential.token=payment_token from instruments")
	}
}

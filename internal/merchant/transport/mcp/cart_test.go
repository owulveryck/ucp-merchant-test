package mcp

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/owulveryck/ucp-merchant-test/internal/merchant/merchanttest"
	"github.com/owulveryck/ucp-merchant-test/internal/model"
)

func TestHandleCreateCart_Success(t *testing.T) {
	mock := merchanttest.NewMock()
	mock.CreateCartFunc = func(ownerID string, items []model.LineItemRequest) (*model.Cart, error) {
		if len(items) != 1 {
			t.Errorf("expected 1 item, got %d", len(items))
		}
		return &model.Cart{ID: "cart_1"}, nil
	}

	s := newTestMCPServer(mock)
	ctx := context.WithValue(context.Background(), ctxUserID, "user1")
	result, err := s.handleCreateCart(ctx, makeToolRequest(map[string]interface{}{
		"cart": map[string]interface{}{
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
	if !strings.Contains(text, "cart_1") {
		t.Error("expected cart ID in response")
	}
}

func TestHandleCreateCart_MissingItems(t *testing.T) {
	mock := merchanttest.NewMock()
	s := newTestMCPServer(mock)

	result, err := s.handleCreateCart(context.Background(), makeToolRequest(map[string]interface{}{
		"cart": map[string]interface{}{},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for missing items")
	}
}

func TestHandleCreateCart_MerchantError(t *testing.T) {
	mock := merchanttest.NewMock()
	mock.CreateCartFunc = func(ownerID string, items []model.LineItemRequest) (*model.Cart, error) {
		return nil, errors.New("out of stock")
	}

	s := newTestMCPServer(mock)
	result, err := s.handleCreateCart(context.Background(), makeToolRequest(map[string]interface{}{
		"cart": map[string]interface{}{
			"line_items": []interface{}{
				map[string]interface{}{"product_id": "SKU-001", "quantity": float64(1)},
			},
		},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error result")
	}
}

func TestHandleGetCart_Found(t *testing.T) {
	mock := merchanttest.NewMock()
	mock.GetCartFunc = func(id, ownerID string) (*model.Cart, error) {
		return &model.Cart{ID: id}, nil
	}

	s := newTestMCPServer(mock)
	result, err := s.handleGetCart(context.Background(), makeToolRequest(map[string]interface{}{
		"id": "cart_1",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Error("expected no error")
	}
}

func TestHandleGetCart_NotFound(t *testing.T) {
	mock := merchanttest.NewMock()
	mock.GetCartFunc = func(id, ownerID string) (*model.Cart, error) {
		return nil, errors.New("not found")
	}

	s := newTestMCPServer(mock)
	result, err := s.handleGetCart(context.Background(), makeToolRequest(map[string]interface{}{
		"id": "nonexistent",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for not found cart")
	}
}

func TestHandleUpdateCart(t *testing.T) {
	mock := merchanttest.NewMock()
	mock.UpdateCartFunc = func(id, ownerID string, items []model.LineItemRequest) (*model.Cart, error) {
		return &model.Cart{ID: id}, nil
	}

	s := newTestMCPServer(mock)
	result, err := s.handleUpdateCart(context.Background(), makeToolRequest(map[string]interface{}{
		"id": "cart_1",
		"cart": map[string]interface{}{
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

func TestHandleCancelCart(t *testing.T) {
	mock := merchanttest.NewMock()
	mock.CancelCartFunc = func(id, ownerID string) (*model.Cart, error) {
		return &model.Cart{ID: id}, nil
	}

	s := newTestMCPServer(mock)
	result, err := s.handleCancelCart(context.Background(), makeToolRequest(map[string]interface{}{
		"id": "cart_1",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Error("expected no error")
	}
}

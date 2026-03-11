package mcp

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/merchanttest"
	"github.com/owulveryck/ucp-merchant-test/pkg/model"
)

func TestHandleGetOrder_Found(t *testing.T) {
	mock := merchanttest.NewMock()
	mock.GetOrderFunc = func(id, ownerID string) (*model.Order, error) {
		return &model.Order{ID: id, CheckoutID: "co_1"}, nil
	}

	s := newTestMCPServer(mock)
	result, err := s.handleGetOrder(ctxWithUser("user1", "US"), makeToolRequest(map[string]interface{}{
		"id": "ord_1",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Error("expected no error")
	}
	text := extractText(t, result)
	if !strings.Contains(text, "ord_1") {
		t.Error("expected order ID in response")
	}
}

func TestHandleGetOrder_NotFound(t *testing.T) {
	mock := merchanttest.NewMock()
	mock.GetOrderFunc = func(id, ownerID string) (*model.Order, error) {
		return nil, errors.New("not found")
	}

	s := newTestMCPServer(mock)
	result, err := s.handleGetOrder(ctxWithUser("user1", "US"), makeToolRequest(map[string]interface{}{
		"id": "nonexistent",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for not found order")
	}
}

func TestHandleListOrders_WithResults(t *testing.T) {
	mock := merchanttest.NewMock()
	mock.ListOrdersFunc = func(ownerID string) ([]*model.Order, error) {
		return []*model.Order{
			{ID: "ord_1", CheckoutID: "co_1", Totals: []model.Total{{Type: "total", DisplayText: "$29.99"}}},
			{ID: "ord_2", CheckoutID: "co_2"},
		}, nil
	}

	s := newTestMCPServer(mock)
	result, err := s.handleListOrders(ctxWithUser("user1", "US"), makeToolRequest(map[string]interface{}{}))
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Error("expected no error")
	}
	text := extractText(t, result)
	if !strings.Contains(text, "ord_1") || !strings.Contains(text, "ord_2") {
		t.Error("expected both orders in response")
	}
}

func TestHandleListOrders_Empty(t *testing.T) {
	mock := merchanttest.NewMock()
	mock.ListOrdersFunc = func(ownerID string) ([]*model.Order, error) {
		return nil, nil
	}

	s := newTestMCPServer(mock)
	result, err := s.handleListOrders(ctxWithUser("user1", "US"), makeToolRequest(map[string]interface{}{}))
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Error("expected no error")
	}
}

func TestHandleCancelOrder(t *testing.T) {
	mock := merchanttest.NewMock()
	mock.CancelOrderFunc = func(id, ownerID string) error {
		return nil
	}

	s := newTestMCPServer(mock)
	result, err := s.handleCancelOrder(context.WithValue(context.Background(), ctxUserID, "user1"), makeToolRequest(map[string]interface{}{
		"id": "ord_1",
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

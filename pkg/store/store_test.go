package store

import (
	"testing"

	"github.com/owulveryck/ucp-merchant-test/pkg/model"
)

func TestNewStore(t *testing.T) {
	s := New()
	if s.Checkouts == nil || s.Orders == nil || s.Carts == nil {
		t.Fatal("expected non-nil maps")
	}
}

func TestStoreReset(t *testing.T) {
	s := New()
	s.Checkouts["co_1"] = &model.Checkout{ID: "co_1"}
	s.Orders["ord_1"] = &model.Order{ID: "ord_1"}
	s.Carts["cart_1"] = &model.Cart{ID: "cart_1"}
	s.CheckoutSeq = 5
	s.OrderSeq = 3
	s.CartSeq = 2

	s.Reset()

	if len(s.Checkouts) != 0 {
		t.Error("expected empty checkouts after reset")
	}
	if len(s.Orders) != 0 {
		t.Error("expected empty orders after reset")
	}
	if len(s.Carts) != 0 {
		t.Error("expected empty carts after reset")
	}
	if s.CheckoutSeq != 0 || s.OrderSeq != 0 || s.CartSeq != 0 {
		t.Error("expected zero sequences after reset")
	}
}

func TestNewSessionID(t *testing.T) {
	s := New()
	id1 := s.NewSessionID()
	id2 := s.NewSessionID()
	if id1 == id2 {
		t.Error("expected unique session IDs")
	}
	if id1 != "session-0001" {
		t.Errorf("expected session-0001, got %s", id1)
	}
}

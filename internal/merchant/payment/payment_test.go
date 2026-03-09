package payment

import (
	"testing"

	"github.com/owulveryck/ucp-merchant-test/internal/model"
)

func TestParsePaymentDefault(t *testing.T) {
	p := ParsePayment(nil)
	if p.SelectedInstrumentID != "instr_1" {
		t.Errorf("expected instr_1, got %s", p.SelectedInstrumentID)
	}
	if len(p.Handlers) != 3 {
		t.Errorf("expected 3 default handlers, got %d", len(p.Handlers))
	}
}

func TestParsePaymentWithData(t *testing.T) {
	req := &model.PaymentRequest{
		SelectedInstrumentID: "custom_instr",
		Instruments:          []map[string]interface{}{{"id": "i1"}},
		Handlers:             []map[string]interface{}{{"id": "h1"}},
	}
	p := ParsePayment(req)
	if p.SelectedInstrumentID != "custom_instr" {
		t.Errorf("expected custom_instr, got %s", p.SelectedInstrumentID)
	}
	if len(p.Instruments) != 1 {
		t.Errorf("expected 1 instrument, got %d", len(p.Instruments))
	}
	if len(p.Handlers) != 1 {
		t.Errorf("expected 1 handler, got %d", len(p.Handlers))
	}
}

func TestParseBuyer(t *testing.T) {
	req := &model.BuyerRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
	}
	b := ParseBuyer(req)
	if b == nil {
		t.Fatal("expected buyer")
	}
	if b.FirstName != "John" {
		t.Errorf("expected John, got %s", b.FirstName)
	}
	if b.Email != "john@example.com" {
		t.Errorf("expected john@example.com, got %s", b.Email)
	}
}

func TestParseBuyerNil(t *testing.T) {
	b := ParseBuyer(nil)
	if b != nil {
		t.Error("expected nil buyer when not provided")
	}
}

package discount

import (
	"testing"

	"github.com/owulveryck/ucp-merchant-test/internal/data"
	"github.com/owulveryck/ucp-merchant-test/internal/model"
)

func TestApplyDiscountsPercentage(t *testing.T) {
	ds := data.New()
	ds.Discounts = []data.CSVDiscount{
		{Code: "10OFF", Type: "percentage", Value: 10, Description: "10% off"},
	}

	items := []model.LineItem{
		{Totals: []model.Total{{Type: "subtotal", Amount: 10000}}},
	}

	raw := map[string]interface{}{
		"codes": []interface{}{"10OFF"},
	}

	result := ApplyDiscounts(raw, items, ds)
	if result == nil {
		t.Fatal("expected non-nil discounts")
	}
	if len(result.Applied) != 1 {
		t.Fatalf("expected 1 applied discount, got %d", len(result.Applied))
	}
	if result.Applied[0].Amount != 1000 {
		t.Errorf("expected 1000 (10%% of 10000), got %d", result.Applied[0].Amount)
	}
}

func TestApplyDiscountsFixed(t *testing.T) {
	ds := data.New()
	ds.Discounts = []data.CSVDiscount{
		{Code: "FIXED500", Type: "fixed_amount", Value: 500, Description: "$5 off"},
	}

	items := []model.LineItem{
		{Totals: []model.Total{{Type: "subtotal", Amount: 10000}}},
	}

	raw := map[string]interface{}{
		"codes": []interface{}{"FIXED500"},
	}

	result := ApplyDiscounts(raw, items, ds)
	if result == nil {
		t.Fatal("expected non-nil discounts")
	}
	if result.Applied[0].Amount != 500 {
		t.Errorf("expected 500, got %d", result.Applied[0].Amount)
	}
}

func TestApplyDiscountsNil(t *testing.T) {
	ds := data.New()
	result := ApplyDiscounts(nil, nil, ds)
	if result != nil {
		t.Error("expected nil for nil input")
	}
}

func TestApplyDiscountsUnknownCode(t *testing.T) {
	ds := data.New()

	items := []model.LineItem{
		{Totals: []model.Total{{Type: "subtotal", Amount: 10000}}},
	}

	raw := map[string]interface{}{
		"codes": []interface{}{"UNKNOWN"},
	}

	result := ApplyDiscounts(raw, items, ds)
	if result == nil {
		t.Fatal("expected non-nil result (codes are tracked even if invalid)")
	}
	if len(result.Codes) != 1 {
		t.Errorf("expected 1 code, got %d", len(result.Codes))
	}
	if len(result.Applied) != 0 {
		t.Error("expected no applied discounts for unknown code")
	}
}

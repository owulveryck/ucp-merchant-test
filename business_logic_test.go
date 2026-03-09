package main

import (
	"testing"
)

func TestTotalsCalculationOnCreate(t *testing.T) {
	ts := newTestServer(t)
	expectedPrice := 3500

	data := ts.createCheckoutSession("", 0, nil, false)

	lineItems := data["line_items"].([]interface{})
	li := lineItems[0].(map[string]interface{})
	liTotals := li["totals"].([]interface{})
	for _, tot := range liTotals {
		totMap := tot.(map[string]interface{})
		if totMap["type"] == "subtotal" {
			if int(totMap["amount"].(float64)) != expectedPrice {
				t.Fatalf("Line item subtotal should match DB price %d, got %v", expectedPrice, totMap["amount"])
			}
		}
		if totMap["type"] == "total" {
			if int(totMap["amount"].(float64)) != expectedPrice {
				t.Fatalf("Line item total should match DB price %d, got %v", expectedPrice, totMap["amount"])
			}
		}
	}

	totals := data["totals"].([]interface{})
	for _, tot := range totals {
		totMap := tot.(map[string]interface{})
		if totMap["type"] == "subtotal" {
			if int(totMap["amount"].(float64)) != expectedPrice {
				t.Fatalf("Subtotal should match DB price %d, got %v", expectedPrice, totMap["amount"])
			}
		}
		if totMap["type"] == "total" {
			if int(totMap["amount"].(float64)) != expectedPrice {
				t.Fatalf("Total should match DB price %d, got %v", expectedPrice, totMap["amount"])
			}
		}
	}
}

func TestTotalsRecalculationOnUpdate(t *testing.T) {
	ts := newTestServer(t)
	expectedPrice := 3500

	data := ts.createCheckoutSession("", 0, nil, false)
	checkoutID := data["id"].(string)

	updatePayload := buildUpdateFromCheckout(data)
	updatePayload["line_items"].([]map[string]interface{})[0]["quantity"] = 2

	resp, updated := ts.updateCheckout(checkoutID, updatePayload, nil)
	if resp.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}

	expectedTotal := expectedPrice * 2
	totals := updated["totals"].([]interface{})
	for _, tot := range totals {
		totMap := tot.(map[string]interface{})
		if totMap["type"] == "total" {
			actual := int(totMap["amount"].(float64))
			if actual != expectedTotal {
				t.Fatalf("Expected total %d, got %d", expectedTotal, actual)
			}
		}
	}
}

func TestDiscountFlow(t *testing.T) {
	ts := newTestServer(t)
	expectedPrice := 3500

	data := ts.createCheckoutSession("", 0, nil, false)
	checkoutID := data["id"].(string)

	updatePayload := buildUpdateFromCheckout(data)
	updatePayload["discounts"] = map[string]interface{}{"codes": []string{"10OFF"}}

	resp, updated := ts.updateCheckout(checkoutID, updatePayload, nil)
	if resp.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}

	expectedTotal := int(float64(expectedPrice) * 0.9)
	totals := updated["totals"].([]interface{})
	for _, tot := range totals {
		totMap := tot.(map[string]interface{})
		if totMap["type"] == "total" {
			actual := int(totMap["amount"].(float64))
			if actual != expectedTotal {
				t.Fatalf("Discount not applied correctly. Expected %d, got %d", expectedTotal, actual)
			}
		}
	}

	discounts, ok := updated["discounts"].(map[string]interface{})
	if !ok {
		t.Fatal("Applied discounts field missing")
	}
	applied, ok := discounts["applied"].([]interface{})
	if !ok || len(applied) == 0 {
		t.Fatal("Applied discounts list empty")
	}
	if applied[0].(map[string]interface{})["code"] != "10OFF" {
		t.Fatalf("Expected discount code '10OFF', got '%v'", applied[0].(map[string]interface{})["code"])
	}
}

func TestMultipleDiscountsAccepted(t *testing.T) {
	ts := newTestServer(t)
	expectedPrice := 3500

	data := ts.createCheckoutSession("", 0, nil, false)
	checkoutID := data["id"].(string)

	updatePayload := buildUpdateFromCheckout(data)
	updatePayload["discounts"] = map[string]interface{}{"codes": []string{"10OFF", "WELCOME20"}}

	_, updated := ts.updateCheckout(checkoutID, updatePayload, nil)

	// 3500 * 0.9 = 3150, 3150 * 0.8 = 2520
	expectedTotal := int(float64(int(float64(expectedPrice)*0.9)) * 0.8)

	totals := updated["totals"].([]interface{})
	for _, tot := range totals {
		totMap := tot.(map[string]interface{})
		if totMap["type"] == "total" {
			actual := int(totMap["amount"].(float64))
			if actual != expectedTotal {
				t.Fatalf("Multiple discounts failed. Expected %d, got %d", expectedTotal, actual)
			}
		}
	}

	discounts := updated["discounts"].(map[string]interface{})
	applied := discounts["applied"].([]interface{})
	if len(applied) != 2 {
		t.Fatalf("Expected 2 applied discounts, got %d", len(applied))
	}

	codes := map[string]bool{}
	for _, a := range applied {
		codes[a.(map[string]interface{})["code"].(string)] = true
	}
	if !codes["10OFF"] || !codes["WELCOME20"] {
		t.Fatalf("Expected 10OFF and WELCOME20, got %v", codes)
	}
}

func TestMultipleDiscountsOneRejected(t *testing.T) {
	ts := newTestServer(t)
	expectedPrice := 3500

	data := ts.createCheckoutSession("", 0, nil, false)
	checkoutID := data["id"].(string)

	updatePayload := buildUpdateFromCheckout(data)
	updatePayload["discounts"] = map[string]interface{}{"codes": []string{"10OFF", "INVALID_CODE"}}

	_, updated := ts.updateCheckout(checkoutID, updatePayload, nil)

	expectedTotal := int(float64(expectedPrice) * 0.9)
	totals := updated["totals"].([]interface{})
	for _, tot := range totals {
		totMap := tot.(map[string]interface{})
		if totMap["type"] == "total" {
			actual := int(totMap["amount"].(float64))
			if actual != expectedTotal {
				t.Fatalf("Expected %d, got %d", expectedTotal, actual)
			}
		}
	}

	discounts := updated["discounts"].(map[string]interface{})
	applied := discounts["applied"].([]interface{})
	if len(applied) != 1 {
		t.Fatalf("Expected 1 applied discount, got %d", len(applied))
	}
	if applied[0].(map[string]interface{})["code"] != "10OFF" {
		t.Fatal("Expected only 10OFF applied")
	}
}

func TestFixedAmountDiscount(t *testing.T) {
	ts := newTestServer(t)
	expectedPrice := 3500

	data := ts.createCheckoutSession("", 0, nil, false)
	checkoutID := data["id"].(string)

	updatePayload := buildUpdateFromCheckout(data)
	updatePayload["discounts"] = map[string]interface{}{"codes": []string{"FIXED500"}}

	_, updated := ts.updateCheckout(checkoutID, updatePayload, nil)

	expectedTotal := expectedPrice - 500
	totals := updated["totals"].([]interface{})
	for _, tot := range totals {
		totMap := tot.(map[string]interface{})
		if totMap["type"] == "total" {
			actual := int(totMap["amount"].(float64))
			if actual != expectedTotal {
				t.Fatalf("Fixed discount failed. Expected %d, got %d", expectedTotal, actual)
			}
		}
	}

	discounts := updated["discounts"].(map[string]interface{})
	applied := discounts["applied"].([]interface{})
	if len(applied) == 0 {
		t.Fatal("Applied discounts missing")
	}
	if applied[0].(map[string]interface{})["code"] != "FIXED500" {
		t.Fatal("Expected FIXED500")
	}
	discountAmount := int(applied[0].(map[string]interface{})["amount"].(float64))
	if discountAmount != 500 {
		t.Fatalf("Expected discount amount 500, got %d", discountAmount)
	}
}

func TestBuyerConsent(t *testing.T) {
	ts := newTestServer(t)

	payload := ts.createCheckoutPayload("", 0)
	payload["buyer"] = map[string]interface{}{
		"first_name": "Consent",
		"last_name":  "Tester",
		"email":      "consent@example.com",
		"consent": map[string]interface{}{
			"marketing":    true,
			"analytics":    false,
			"sale_of_data": false,
		},
	}

	resp, data := ts.createCheckout(payload)
	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		t.Fatalf("Expected 201, got %d", resp.StatusCode)
	}
	checkoutID := data["id"].(string)

	_, getData := ts.getCheckout(checkoutID)

	buyer, ok := getData["buyer"].(map[string]interface{})
	if !ok {
		t.Fatal("Buyer info missing")
	}
	consent, ok := buyer["consent"].(map[string]interface{})
	if !ok {
		t.Fatal("Consent info missing")
	}
	if consent["marketing"] != true {
		t.Fatalf("Marketing consent not persisted: %v", consent["marketing"])
	}
	if consent["analytics"] != false {
		t.Fatalf("Analytics consent not persisted: %v", consent["analytics"])
	}
}

func TestBuyerInfoPersistence(t *testing.T) {
	ts := newTestServer(t)

	data := ts.createCheckoutSession("", 0, nil, false)
	checkoutID := data["id"].(string)

	updatePayload := buildUpdateFromCheckout(data)
	updatePayload["buyer"] = map[string]interface{}{
		"email":      "test@example.com",
		"first_name": "Test",
		"last_name":  "User",
	}

	resp, _ := ts.updateCheckout(checkoutID, updatePayload, nil)
	if resp.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}

	_, getData := ts.getCheckout(checkoutID)
	buyer := getData["buyer"].(map[string]interface{})
	if buyer["email"] != "test@example.com" {
		t.Fatalf("Email mismatch: %v", buyer["email"])
	}
	if buyer["first_name"] != "Test" {
		t.Fatalf("First name mismatch: %v", buyer["first_name"])
	}
}

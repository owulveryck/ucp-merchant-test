package main

import (
	"fmt"
	"testing"
	"time"
)

func TestFulfillmentFlow(t *testing.T) {
	ts := newTestServer(t)
	data := ts.createCheckoutSession("", 0, nil, false)
	checkoutID := data["id"].(string)

	// 1. Update with fulfillment address
	updatePayload := buildUpdateFromCheckout(data)
	updatePayload["fulfillment"] = map[string]interface{}{
		"methods": []map[string]interface{}{
			{
				"type": "shipping",
				"destinations": []map[string]interface{}{
					{
						"id":               "dest_1",
						"full_name":        "John Doe",
						"street_address":   "123 Main St",
						"address_locality": "Springfield",
						"address_region":   "IL",
						"postal_code":      "62704",
						"address_country":  "US",
					},
				},
				"selected_destination_id": "dest_1",
			},
		},
	}

	_, updated := ts.updateCheckout(checkoutID, updatePayload, nil)

	// Verify options are generated
	f := updated["fulfillment"].(map[string]interface{})
	methods := f["methods"].([]interface{})
	method := methods[0].(map[string]interface{})
	groups := method["groups"].([]interface{})
	group := groups[0].(map[string]interface{})
	options := group["options"].([]interface{})
	if len(options) == 0 {
		t.Fatal("Fulfillment options not generated")
	}

	// 2. Select first option
	optionID := options[0].(map[string]interface{})["id"].(string)
	optionTotals := options[0].(map[string]interface{})["totals"].([]interface{})
	optionCost := 0
	for _, tot := range optionTotals {
		totMap := tot.(map[string]interface{})
		if totMap["type"] == "total" {
			optionCost = int(totMap["amount"].(float64))
		}
	}

	updatePayload["fulfillment"] = map[string]interface{}{
		"methods": []map[string]interface{}{
			{
				"type":                    "shipping",
				"selected_destination_id": "dest_1",
				"groups":                  []map[string]interface{}{{"selected_option_id": optionID}},
			},
		},
	}

	_, final := ts.updateCheckout(checkoutID, updatePayload, nil)

	expectedTotal := 3500 + optionCost
	totals := final["totals"].([]interface{})
	for _, tot := range totals {
		totMap := tot.(map[string]interface{})
		if totMap["type"] == "total" {
			actual := int(totMap["amount"].(float64))
			if actual != expectedTotal {
				t.Fatalf("Total not updated correctly. Expected %d, got %d", expectedTotal, actual)
			}
		}
	}
}

func TestDynamicFulfillment(t *testing.T) {
	ts := newTestServer(t)
	data := ts.createCheckoutSession("", 0, nil, false)
	checkoutID := data["id"].(string)

	// US address
	updatePayload := buildUpdateFromCheckout(data)
	updatePayload["fulfillment"] = map[string]interface{}{
		"methods": []map[string]interface{}{
			{
				"type": "shipping",
				"destinations": []map[string]interface{}{
					{"id": "dest_us", "address_country": "US", "postal_code": "62704"},
				},
				"selected_destination_id": "dest_us",
			},
		},
	}

	_, usData := ts.updateCheckout(checkoutID, updatePayload, nil)
	usOptions := getOptions(t, usData)
	foundUS := false
	for _, o := range usOptions {
		if o["id"] == "exp-ship-us" {
			foundUS = true
		}
	}
	if !foundUS {
		t.Fatalf("Expected US express option, got %v", usOptions)
	}

	// CA address
	updatePayload["fulfillment"] = map[string]interface{}{
		"methods": []map[string]interface{}{
			{
				"type": "shipping",
				"destinations": []map[string]interface{}{
					{"id": "dest_ca", "address_country": "CA", "postal_code": "M5V 2H1"},
				},
				"selected_destination_id": "dest_ca",
			},
		},
	}

	_, caData := ts.updateCheckout(checkoutID, updatePayload, nil)
	caOptions := getOptions(t, caData)
	foundIntl := false
	for _, o := range caOptions {
		if o["id"] == "exp-ship-intl" {
			foundIntl = true
		}
	}
	if !foundIntl {
		t.Fatalf("Expected Intl express option, got %v", caOptions)
	}
}

func TestUnknownCustomerNoAddress(t *testing.T) {
	ts := newTestServer(t)
	buyer := map[string]interface{}{"fullName": "Unknown Person", "email": "unknown@example.com"}
	data := ts.createCheckoutSession("", 0, buyer, false)
	checkoutID := data["id"].(string)

	updatePayload := buildUpdateFromCheckout(data)
	updatePayload["fulfillment"] = map[string]interface{}{
		"methods": []map[string]interface{}{{"type": "shipping"}},
	}

	_, updated := ts.updateCheckout(checkoutID, updatePayload, nil)

	f := updated["fulfillment"].(map[string]interface{})
	methods := f["methods"].([]interface{})
	method := methods[0].(map[string]interface{})
	if method["destinations"] != nil {
		dests, ok := method["destinations"].([]interface{})
		if ok && len(dests) > 0 {
			t.Fatal("Expected no destinations for unknown customer")
		}
	}
}

func TestKnownCustomerNoAddress(t *testing.T) {
	ts := newTestServer(t)
	// Jane Doe (cust_3) has no address
	buyer := map[string]interface{}{"fullName": "Jane Doe", "email": "jane.doe@example.com"}
	data := ts.createCheckoutSession("", 0, buyer, false)
	checkoutID := data["id"].(string)

	updatePayload := buildUpdateFromCheckout(data)
	updatePayload["fulfillment"] = map[string]interface{}{
		"methods": []map[string]interface{}{{"type": "shipping"}},
	}

	_, updated := ts.updateCheckout(checkoutID, updatePayload, nil)

	f := updated["fulfillment"].(map[string]interface{})
	methods := f["methods"].([]interface{})
	method := methods[0].(map[string]interface{})
	if method["destinations"] != nil {
		dests, ok := method["destinations"].([]interface{})
		if ok && len(dests) > 0 {
			t.Fatal("Expected no destinations for known customer with no addresses")
		}
	}
}

func TestKnownCustomerOneAddress(t *testing.T) {
	ts := newTestServer(t)
	// John Doe (cust_1) has at least 2 addresses
	buyer := map[string]interface{}{"fullName": "John Doe", "email": "john.doe@example.com"}
	data := ts.createCheckoutSession("", 0, buyer, false)
	checkoutID := data["id"].(string)

	updatePayload := buildUpdateFromCheckout(data)
	updatePayload["fulfillment"] = map[string]interface{}{
		"methods": []map[string]interface{}{{"type": "shipping"}},
	}

	_, updated := ts.updateCheckout(checkoutID, updatePayload, nil)

	f := updated["fulfillment"].(map[string]interface{})
	methods := f["methods"].([]interface{})
	method := methods[0].(map[string]interface{})
	dests, ok := method["destinations"].([]interface{})
	if !ok || len(dests) < 2 {
		t.Fatalf("Expected at least 2 destinations, got %d", len(dests))
	}
	firstDest := dests[0].(map[string]interface{})
	if firstDest["address_country"] != "US" {
		t.Fatalf("Expected US address, got %v", firstDest["address_country"])
	}
}

func TestKnownCustomerMultipleAddressesSelection(t *testing.T) {
	ts := newTestServer(t)
	buyer := map[string]interface{}{"fullName": "John Doe", "email": "john.doe@example.com"}
	data := ts.createCheckoutSession("", 0, buyer, false)
	checkoutID := data["id"].(string)

	// Trigger injection
	updatePayload := buildUpdateFromCheckout(data)
	updatePayload["fulfillment"] = map[string]interface{}{
		"methods": []map[string]interface{}{{"type": "shipping"}},
	}
	_, updated := ts.updateCheckout(checkoutID, updatePayload, nil)

	f := updated["fulfillment"].(map[string]interface{})
	methods := f["methods"].([]interface{})
	method := methods[0].(map[string]interface{})
	dests := method["destinations"].([]interface{})
	if len(dests) < 2 {
		t.Fatalf("Expected at least 2 destinations, got %d", len(dests))
	}

	destIDs := make(map[string]bool)
	for _, d := range dests {
		dm := d.(map[string]interface{})
		destIDs[dm["id"].(string)] = true
	}
	if !destIDs["addr_1"] || !destIDs["addr_2"] {
		t.Fatalf("Expected addr_1 and addr_2, got %v", destIDs)
	}

	// Select addr_2
	updatePayload["fulfillment"] = map[string]interface{}{
		"methods": []map[string]interface{}{
			{"type": "shipping", "selected_destination_id": "addr_2"},
		},
	}
	_, final := ts.updateCheckout(checkoutID, updatePayload, nil)

	fFinal := final["fulfillment"].(map[string]interface{})
	mFinal := fFinal["methods"].([]interface{})[0].(map[string]interface{})
	if mFinal["selected_destination_id"] != "addr_2" {
		t.Fatalf("Expected selected_destination_id=addr_2, got %v", mFinal["selected_destination_id"])
	}
}

func TestKnownCustomerNewAddress(t *testing.T) {
	ts := newTestServer(t)
	buyer := map[string]interface{}{"fullName": "John Doe", "email": "john.doe@example.com"}
	data := ts.createCheckoutSession("", 0, buyer, true)
	checkoutID := data["id"].(string)

	updatePayload := buildUpdateFromCheckout(data)
	updatePayload["fulfillment"] = map[string]interface{}{
		"methods": []map[string]interface{}{
			{
				"type": "shipping",
				"destinations": []map[string]interface{}{
					{
						"id":               "dest_new",
						"address_country":  "CA",
						"postal_code":      "M5V 2H1",
						"street_address":   "123 New St",
					},
				},
				"selected_destination_id": "dest_new",
			},
		},
	}

	_, updated := ts.updateCheckout(checkoutID, updatePayload, nil)

	f := updated["fulfillment"].(map[string]interface{})
	methods := f["methods"].([]interface{})
	method := methods[0].(map[string]interface{})
	dests := method["destinations"].([]interface{})
	if len(dests) != 1 {
		t.Fatalf("Expected 1 destination, got %d", len(dests))
	}
	if dests[0].(map[string]interface{})["id"] != "dest_new" {
		t.Fatal("Expected dest_new")
	}

	// Verify intl options
	groups := method["groups"].([]interface{})
	group := groups[0].(map[string]interface{})
	options := group["options"].([]interface{})
	foundIntl := false
	for _, o := range options {
		if o.(map[string]interface{})["id"] == "exp-ship-intl" {
			foundIntl = true
		}
	}
	if !foundIntl {
		t.Fatal("Expected international shipping option for CA address")
	}
}

func TestNewUserNewAddressPersistence(t *testing.T) {
	ts := newTestServer(t)
	email := fmt.Sprintf("new.user.%d@example.com", time.Now().UnixNano())
	buyer := map[string]interface{}{"fullName": "New User", "email": email}
	data := ts.createCheckoutSession("", 0, buyer, false)
	checkoutID := data["id"].(string)

	// New address without ID
	updatePayload := buildUpdateFromCheckout(data)
	updatePayload["fulfillment"] = map[string]interface{}{
		"methods": []map[string]interface{}{
			{
				"type": "shipping",
				"destinations": []map[string]interface{}{
					{
						"street_address":   "789 Pine St",
						"address_locality": "Villagetown",
						"address_region":   "NY",
						"postal_code":      "10001",
						"address_country":  "US",
					},
				},
			},
		},
	}

	_, updated := ts.updateCheckout(checkoutID, updatePayload, nil)

	f := updated["fulfillment"].(map[string]interface{})
	methods := f["methods"].([]interface{})
	method := methods[0].(map[string]interface{})
	dests := method["destinations"].([]interface{})
	if len(dests) != 1 {
		t.Fatalf("Expected 1 destination, got %d", len(dests))
	}
	generatedID := dests[0].(map[string]interface{})["id"].(string)
	if generatedID == "" {
		t.Fatal("ID should be generated for new address")
	}

	// Create another checkout for same user, verify address is injected
	data2 := ts.createCheckoutSession("", 0, buyer, false)
	checkoutID2 := data2["id"].(string)

	updatePayload2 := buildUpdateFromCheckout(data2)
	updatePayload2["fulfillment"] = map[string]interface{}{
		"methods": []map[string]interface{}{{"type": "shipping"}},
	}
	_, updated2 := ts.updateCheckout(checkoutID2, updatePayload2, nil)

	f2 := updated2["fulfillment"].(map[string]interface{})
	methods2 := f2["methods"].([]interface{})
	method2 := methods2[0].(map[string]interface{})
	dests2 := method2["destinations"].([]interface{})
	if dests2 == nil || len(dests2) == 0 {
		t.Fatal("Expected persisted address to be injected")
	}
	foundID := false
	for _, d := range dests2 {
		if d.(map[string]interface{})["id"] == generatedID {
			foundID = true
		}
	}
	if !foundID {
		t.Fatalf("Expected generated ID %s in injected addresses", generatedID)
	}
}

func TestKnownUserExistingAddressReuse(t *testing.T) {
	ts := newTestServer(t)
	buyer := map[string]interface{}{"fullName": "John Doe", "email": "john.doe@example.com"}
	data := ts.createCheckoutSession("", 0, buyer, false)
	checkoutID := data["id"].(string)

	// Send address matching addr_1 without ID
	updatePayload := buildUpdateFromCheckout(data)
	updatePayload["fulfillment"] = map[string]interface{}{
		"methods": []map[string]interface{}{
			{
				"type": "shipping",
				"destinations": []map[string]interface{}{
					{
						"street_address":   "123 Main St",
						"address_locality": "Springfield",
						"address_region":   "IL",
						"postal_code":      "62704",
						"address_country":  "US",
					},
				},
			},
		},
	}

	_, updated := ts.updateCheckout(checkoutID, updatePayload, nil)

	f := updated["fulfillment"].(map[string]interface{})
	methods := f["methods"].([]interface{})
	method := methods[0].(map[string]interface{})
	dests := method["destinations"].([]interface{})
	if len(dests) != 1 {
		t.Fatalf("Expected 1 destination, got %d", len(dests))
	}
	if dests[0].(map[string]interface{})["id"] != "addr_1" {
		t.Fatalf("Expected reuse of addr_1, got %v", dests[0].(map[string]interface{})["id"])
	}
}

func TestFreeShippingOnExpensiveOrder(t *testing.T) {
	ts := newTestServer(t)
	// bouquet_roses is 3500. Qty 3 = 10500 > 10000 threshold
	data := ts.createCheckoutSession("", 3, nil, true)
	checkoutID := data["id"].(string)

	updatePayload := buildUpdateFromCheckout(data)
	updatePayload["fulfillment"] = map[string]interface{}{
		"methods": []map[string]interface{}{
			{
				"type": "shipping",
				"destinations": []map[string]interface{}{
					{"id": "dest_us", "address_country": "US", "postal_code": "62704"},
				},
				"selected_destination_id": "dest_us",
			},
		},
	}

	_, updated := ts.updateCheckout(checkoutID, updatePayload, nil)
	options := getOptions(t, updated)

	var freeShipping map[string]interface{}
	for _, o := range options {
		if o["id"] == "std-ship" {
			freeShipping = o
			break
		}
	}
	if freeShipping == nil {
		t.Fatal("std-ship option not found")
	}

	optTotals := freeShipping["totals"].([]interface{})
	for _, tot := range optTotals {
		totMap := tot.(map[string]interface{})
		if totMap["type"] == "total" {
			if int(totMap["amount"].(float64)) != 0 {
				t.Fatalf("Expected free shipping (0), got %v", totMap["amount"])
			}
		}
	}

	title := freeShipping["title"].(string)
	if title == "" || (title != "Free Standard Shipping" && title[0:4] != "Free") {
		t.Fatalf("Expected 'Free' in title, got '%s'", title)
	}
}

func TestFreeShippingForSpecificItem(t *testing.T) {
	ts := newTestServer(t)
	// bouquet_roses is eligible for free shipping
	data := ts.createCheckoutSession("bouquet_roses", 1, nil, true)
	checkoutID := data["id"].(string)

	updatePayload := buildUpdateFromCheckout(data)
	updatePayload["fulfillment"] = map[string]interface{}{
		"methods": []map[string]interface{}{
			{
				"type": "shipping",
				"destinations": []map[string]interface{}{
					{"id": "dest_us", "address_country": "US", "postal_code": "62704"},
				},
				"selected_destination_id": "dest_us",
			},
		},
	}

	_, updated := ts.updateCheckout(checkoutID, updatePayload, nil)
	options := getOptions(t, updated)

	var freeShipping map[string]interface{}
	for _, o := range options {
		if o["id"] == "std-ship" {
			freeShipping = o
			break
		}
	}
	if freeShipping == nil {
		t.Fatal("std-ship option not found")
	}

	optTotals := freeShipping["totals"].([]interface{})
	for _, tot := range optTotals {
		totMap := tot.(map[string]interface{})
		if totMap["type"] == "total" {
			if int(totMap["amount"].(float64)) != 0 {
				t.Fatalf("Expected free shipping (0), got %v", totMap["amount"])
			}
		}
	}
}

func getOptions(t *testing.T, data map[string]interface{}) []map[string]interface{} {
	t.Helper()
	f := data["fulfillment"].(map[string]interface{})
	methods := f["methods"].([]interface{})
	method := methods[0].(map[string]interface{})
	groups := method["groups"].([]interface{})
	group := groups[0].(map[string]interface{})
	rawOptions := group["options"].([]interface{})
	var options []map[string]interface{}
	for _, o := range rawOptions {
		options = append(options, o.(map[string]interface{}))
	}
	return options
}

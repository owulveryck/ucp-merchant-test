package data

import "testing"

func TestNew(t *testing.T) {
	ds := New()
	if ds.DynamicAddresses == nil {
		t.Fatal("expected non-nil DynamicAddresses")
	}
}

func TestFindCustomerByEmail(t *testing.T) {
	ds := New()
	ds.Customers = []CSVCustomer{
		{ID: "cust_1", Name: "John Doe", Email: "john@example.com"},
		{ID: "cust_2", Name: "Jane Doe", Email: "jane@example.com"},
	}

	c := ds.FindCustomerByEmail("john@example.com")
	if c == nil || c.ID != "cust_1" {
		t.Errorf("expected cust_1, got %v", c)
	}

	c = ds.FindCustomerByEmail("JOHN@EXAMPLE.COM")
	if c == nil || c.ID != "cust_1" {
		t.Error("expected case-insensitive match")
	}

	c = ds.FindCustomerByEmail("nobody@example.com")
	if c != nil {
		t.Error("expected nil for unknown email")
	}
}

func TestFindAddressesByCustomerID(t *testing.T) {
	ds := New()
	ds.Addresses = []CSVAddress{
		{ID: "addr_1", CustomerID: "cust_1", City: "NYC"},
		{ID: "addr_2", CustomerID: "cust_1", City: "LA"},
		{ID: "addr_3", CustomerID: "cust_2", City: "Chicago"},
	}

	addrs := ds.FindAddressesByCustomerID("cust_1")
	if len(addrs) != 2 {
		t.Errorf("expected 2 addresses, got %d", len(addrs))
	}
}

func TestFindDiscountByCode(t *testing.T) {
	ds := New()
	ds.Discounts = []CSVDiscount{
		{Code: "10OFF", Type: "percentage", Value: 10, Description: "10% off"},
	}

	d := ds.FindDiscountByCode("10off")
	if d == nil || d.Value != 10 {
		t.Error("expected case-insensitive discount lookup")
	}

	d = ds.FindDiscountByCode("NOPE")
	if d != nil {
		t.Error("expected nil for unknown code")
	}
}

func TestFindPaymentInstrumentByID(t *testing.T) {
	ds := New()
	ds.PaymentInstruments = []CSVPaymentInstrument{
		{ID: "pi_1", Type: "card", Token: "tok_1"},
	}

	pi := ds.FindPaymentInstrumentByID("pi_1")
	if pi == nil || pi.Token != "tok_1" {
		t.Error("expected instrument pi_1")
	}

	pi = ds.FindPaymentInstrumentByID("pi_999")
	if pi != nil {
		t.Error("expected nil for unknown ID")
	}
}

func TestFindPaymentInstrumentByToken(t *testing.T) {
	ds := New()
	ds.PaymentInstruments = []CSVPaymentInstrument{
		{ID: "pi_1", Type: "card", Token: "tok_1"},
	}

	pi := ds.FindPaymentInstrumentByToken("tok_1")
	if pi == nil || pi.ID != "pi_1" {
		t.Error("expected instrument by token")
	}
}

func TestGetShippingRatesForCountry(t *testing.T) {
	ds := New()
	ds.ShippingRates = []CSVShippingRate{
		{ID: "r1", CountryCode: "US", ServiceLevel: "standard", Price: 500, Title: "Standard US"},
		{ID: "r2", CountryCode: "default", ServiceLevel: "standard", Price: 800, Title: "Standard Default"},
		{ID: "r3", CountryCode: "default", ServiceLevel: "express", Price: 1500, Title: "Express Default"},
	}

	rates := ds.GetShippingRatesForCountry("US")
	if len(rates) != 2 {
		t.Fatalf("expected 2 rates (US standard + default express), got %d", len(rates))
	}
	// US standard should override default standard
	for _, r := range rates {
		if r.ServiceLevel == "standard" && r.CountryCode != "US" {
			t.Error("expected US-specific standard rate, not default")
		}
	}
}

func TestMatchExistingAddress(t *testing.T) {
	addrs := []CSVAddress{
		{ID: "addr_1", StreetAddress: "123 Main St", City: "NYC", State: "NY", PostalCode: "10001", Country: "US"},
	}

	matched := MatchExistingAddress(addrs, "123 main st", "nyc", "ny", "10001", "us")
	if matched == nil || matched.ID != "addr_1" {
		t.Error("expected case-insensitive address match")
	}

	matched = MatchExistingAddress(addrs, "456 Oak Ave", "LA", "CA", "90001", "US")
	if matched != nil {
		t.Error("expected nil for non-matching address")
	}
}

func TestSaveDynamicAddress(t *testing.T) {
	ds := New()
	addr := CSVAddress{ID: "addr_dyn_1", StreetAddress: "456 Oak", City: "LA"}
	ds.SaveDynamicAddress("john@example.com", addr)

	addrs := ds.FindAddressesForEmail("john@example.com")
	if len(addrs) != 1 || addrs[0].ID != "addr_dyn_1" {
		t.Error("expected saved dynamic address")
	}
}

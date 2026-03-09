package payment

import "github.com/owulveryck/ucp-merchant-test/internal/model"

// ParsePayment extracts payment info from a request map.
func ParsePayment(req map[string]interface{}) model.Payment {
	paymentRaw, ok := req["payment"]
	if !ok || paymentRaw == nil {
		return DefaultPayment()
	}

	paymentMap, ok := paymentRaw.(map[string]interface{})
	if !ok {
		return DefaultPayment()
	}

	p := &model.Payment{}

	if sid, ok := paymentMap["selected_instrument_id"].(string); ok {
		p.SelectedInstrumentID = sid
	}

	if instRaw, ok := paymentMap["instruments"].([]interface{}); ok {
		for _, inst := range instRaw {
			if m, ok := inst.(map[string]interface{}); ok {
				p.Instruments = append(p.Instruments, m)
			}
		}
	}
	if p.Instruments == nil {
		p.Instruments = []map[string]interface{}{}
	}

	if hRaw, ok := paymentMap["handlers"].([]interface{}); ok {
		for _, h := range hRaw {
			if m, ok := h.(map[string]interface{}); ok {
				p.Handlers = append(p.Handlers, m)
			}
		}
	}
	if p.Handlers == nil {
		p.Handlers = DefaultPaymentHandlers()
	}

	return *p
}

// DefaultPayment returns the default payment configuration.
func DefaultPayment() model.Payment {
	return model.Payment{
		SelectedInstrumentID: "instr_1",
		Instruments:          []map[string]interface{}{},
		Handlers:             DefaultPaymentHandlers(),
	}
}

// DefaultPaymentHandlers returns the default payment handlers.
func DefaultPaymentHandlers() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"id":                 "google_pay",
			"name":               "google.pay",
			"version":            "2026-01-11",
			"spec":               "https://ucp.dev/specs/payment/google_pay",
			"config_schema":      "https://ucp.dev/schemas/payment/google_pay.json",
			"instrument_schemas": []string{"https://ucp.dev/schemas/payment/google_pay_instrument.json"},
			"config":             map[string]interface{}{},
		},
		{
			"id":                 "mock_payment_handler",
			"name":               "mock_payment_handler",
			"version":            "2026-01-11",
			"spec":               "https://ucp.dev/specs/payment/mock",
			"config_schema":      "https://ucp.dev/schemas/payment/mock.json",
			"instrument_schemas": []string{"https://ucp.dev/schemas/payment/mock_instrument.json"},
			"config":             map[string]interface{}{},
		},
		{
			"id":                 "shop_pay",
			"name":               "com.shopify.shop_pay",
			"version":            "2026-01-11",
			"spec":               "https://ucp.dev/specs/payment/shop_pay",
			"config_schema":      "https://ucp.dev/schemas/payment/shop_pay.json",
			"instrument_schemas": []string{"https://ucp.dev/schemas/payment/shop_pay_instrument.json"},
			"config":             map[string]interface{}{"shop_id": "merchant_1"},
		},
	}
}

// ParseBuyer extracts buyer info from a request map.
func ParseBuyer(req map[string]interface{}) *model.Buyer {
	buyerRaw, ok := req["buyer"]
	if !ok || buyerRaw == nil {
		return nil
	}
	buyerMap, ok := buyerRaw.(map[string]interface{})
	if !ok {
		return nil
	}

	b := &model.Buyer{}
	if v, ok := buyerMap["first_name"].(string); ok {
		b.FirstName = v
	}
	if v, ok := buyerMap["last_name"].(string); ok {
		b.LastName = v
	}
	if v, ok := buyerMap["fullName"].(string); ok {
		b.FullName = v
	}
	if v, ok := buyerMap["name"].(string); ok && b.FullName == "" && b.FirstName == "" {
		b.FullName = v
	}
	if v, ok := buyerMap["email"].(string); ok {
		b.Email = v
	}

	if consentRaw, ok := buyerMap["consent"].(map[string]interface{}); ok {
		c := &model.Consent{}
		if v, ok := consentRaw["marketing"].(bool); ok {
			c.Marketing = &v
		}
		if v, ok := consentRaw["analytics"].(bool); ok {
			c.Analytics = &v
		}
		if v, ok := consentRaw["sale_of_data"].(bool); ok {
			c.SaleOfData = &v
		}
		b.Consent = c
	}

	return b
}

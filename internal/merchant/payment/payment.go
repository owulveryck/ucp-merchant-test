package payment

import "github.com/owulveryck/ucp-merchant-test/internal/model"

// ParsePayment extracts payment info from a typed request.
func ParsePayment(req *model.PaymentRequest) model.Payment {
	if req == nil {
		return DefaultPayment()
	}

	p := &model.Payment{}
	p.SelectedInstrumentID = req.SelectedInstrumentID

	p.Instruments = req.Instruments
	if p.Instruments == nil {
		p.Instruments = []map[string]interface{}{}
	}

	p.Handlers = req.Handlers
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

// ParseBuyer extracts buyer info from a typed request.
func ParseBuyer(req *model.BuyerRequest) *model.Buyer {
	if req == nil {
		return nil
	}

	b := &model.Buyer{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		FullName:  req.FullName,
		Email:     req.Email,
	}
	if req.Name != "" && b.FullName == "" && b.FirstName == "" {
		b.FullName = req.Name
	}

	if req.Consent != nil {
		b.Consent = &model.Consent{
			Marketing:  req.Consent.Marketing,
			Analytics:  req.Consent.Analytics,
			SaleOfData: req.Consent.SaleOfData,
		}
	}

	return b
}

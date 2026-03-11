package payment

import "github.com/owulveryck/ucp-merchant-test/pkg/model"

// ParsePayment constructs a [model.Payment] from the platform's payment
// configuration submitted during a UCP checkout create or update operation.
//
// In the UCP payment model, the payment object is required in every checkout
// response (even when the platform does not explicitly provide one). When req
// is nil (no payment section in the request), [DefaultPayment] is returned
// with pre-configured handlers and a default selected instrument.
//
// When req is provided, its fields are used directly:
//   - SelectedInstrumentID: which instrument the buyer has chosen
//   - Instruments: payment instruments collected by the platform (opaque blobs)
//   - Handlers: payment handler configurations (defaults to [DefaultPaymentHandlers]
//     when not provided, ensuring the checkout always advertises available payment methods)
//
// Instruments and Handlers remain as []map[string]interface{} because their
// schemas are handler-specific per the UCP specification — each payment handler
// (e.g., Google Pay, Shop Pay) defines its own instrument and configuration format.
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

// DefaultPayment returns the default payment configuration used when the
// platform does not provide payment information in the checkout request.
//
// The default includes a pre-selected instrument ID ("instr_1"), an empty
// instruments array (no instruments collected yet), and the full set of
// payment handlers from [DefaultPaymentHandlers]. This ensures every checkout
// response includes the payment object as required by UCP, allowing platforms
// to discover available payment methods from the first response.
func DefaultPayment() model.Payment {
	return model.Payment{
		SelectedInstrumentID: "instr_1",
		Instruments:          []map[string]interface{}{},
		Handlers:             DefaultPaymentHandlers(),
	}
}

// DefaultPaymentHandlers returns the payment handler configurations advertised
// by this merchant in checkout responses. Payment handlers enable the "N-to-N"
// interoperability model in UCP: platforms discover handlers from the business's
// UCP profile (/.well-known/ucp) and use them to collect payment instruments
// from the buyer.
//
// This test implementation advertises three handlers:
//   - google_pay (com.google.pay): Google Pay integration
//   - mock_payment_handler: a simulated handler for conformance testing
//   - shop_pay (com.shopify.shop_pay): Shopify Shop Pay integration
//
// Each handler includes its specification URL, configuration schema URL,
// instrument schema URLs, and handler-specific configuration. The version
// field matches the UCP protocol version ("2026-01-11").
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

// ParseBuyer converts a [model.BuyerRequest] into a [model.Buyer] for inclusion
// in the UCP checkout response. Returns nil when req is nil, indicating no buyer
// information was provided.
//
// Buyer identity is foundational in UCP commerce:
//   - Email enables address lookup for fulfillment destination pre-population
//     (the business looks up known addresses by buyer email)
//   - Name fields are used in fulfillment expectations and order confirmation
//   - Consent preferences (marketing, analytics, sale_of_data) are forwarded
//     through the checkout to the order for privacy compliance
//
// Name resolution follows a fallback chain: FirstName/LastName are used when
// provided; FullName (from the "fullName" JSON field) is used as-is; the Name
// field (from the "name" JSON field) populates FullName only when neither
// FullName nor FirstName is set, preventing accidental overwrites.
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

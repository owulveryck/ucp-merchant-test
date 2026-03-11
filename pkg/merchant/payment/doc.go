// Package payment implements payment parsing and defaults for the UCP Shopping
// Service Checkout Capability.
//
// In the Universal Commerce Protocol, the payment object is part of the checkout
// session and manages how the buyer pays for their purchase. Payment handlers
// are discovered from the business's UCP profile at /.well-known/ucp and define
// the processing specifications for collecting payment instruments.
//
// # Payment Handlers
//
// Payment handlers enable "N-to-N" interoperability between platforms, businesses,
// and payment providers. Each handler specification answers:
//
//   - Who participates and what are their roles?
//   - What prerequisites (onboarding/setup) are required?
//   - How is the handler configured and advertised?
//   - How are instruments acquired and processed?
//
// This test implementation provides default handlers for common payment methods:
// card payments (com.google.pay) and a simulated tokenized handler. The
// DefaultPaymentHandlers function returns the handler configuration advertised
// in checkout responses.
//
// # Payment Instruments
//
// When the buyer submits payment, the platform populates payment.instruments
// with the collected instrument data. Each instrument includes an ID, type
// (e.g., "card"), brand, last digits, and a processing token. The
// selected_instrument_id field indicates which instrument the buyer has chosen.
//
// # Buyer
//
// The ParseBuyer function extracts buyer information (first name, last name,
// email, phone) from the request payload. Buyer identity is foundational for
// UCP commerce experiences: it enables address lookup for fulfillment
// destination population, payment instrument retrieval, and personalized
// pricing through the Identity Linking capability.
//
// # Payment in Checkout
//
// The payment object is required in checkout responses (even though it's optional
// on checkout creation). ParsePayment extracts payment configuration from the
// request or returns a DefaultPayment with pre-configured handlers and a
// default selected instrument.
package payment

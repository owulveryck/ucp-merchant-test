// Package merchant defines interfaces for UCP (Universal Commerce Protocol)
// compliant merchant implementations.
//
// The Universal Commerce Protocol (version 2026-01-11) specifies a set of
// capabilities that a merchant must support for interoperable commerce.
// This package maps each UCP capability to a Go interface:
//
//   - Cataloger:   read-only product catalog (catalog browsing, search, lookup)
//   - Carter:      shopping cart lifecycle (create, get, update, cancel)
//   - Checkouter:  checkout session management (dev.ucp.shopping.checkout)
//   - Orderer:     order management (dev.ucp.shopping.order)
//
// The top-level Merchant interface composes all of the above.
// A concrete implementation (e.g. simple_merchant) must satisfy Merchant
// to be considered UCP-compliant.
//
// # Fulfillment and Discounts
//
// The UCP capabilities dev.ucp.shopping.fulfillment and
// dev.ucp.shopping.discount do not appear as sub-interfaces of Merchant.
// In UCP, fulfillment configuration (address selection, shipping option
// selection) and discount code application are performed as part of
// checkout update operations — they are fields on the CheckoutRequest,
// not standalone API endpoints. Therefore, the Checkouter interface
// already covers these capabilities through its UpdateCheckout method.
//
// Implementations that need fulfillment or discount data sources should
// accept them as constructor dependencies:
//
//   - fulfillment.FulfillmentDataSource (internal/merchant/fulfillment)
//     provides address lookup, shipping rates, and promotions
//   - discount.DiscountLookup (internal/merchant/discount)
//     provides discount code validation
//
// These are data-layer interfaces, not UCP operation interfaces. They
// feed into the checkout update flow but are not exposed to the platform.
//
// # Buyer Consent
//
// The dev.ucp.shopping.buyer_consent capability is handled via the
// Buyer.Consent field on CheckoutRequest, processed within
// Checkouter.CreateCheckout and Checkouter.UpdateCheckout. No separate
// interface is needed.
package merchant

import (
	"github.com/owulveryck/ucp-merchant-test/internal/catalog"
	"github.com/owulveryck/ucp-merchant-test/internal/model"
)

// Cataloger exposes read-only product catalog operations.
//
// UCP merchants must provide catalog browsing for platforms and agents
// to discover products. This interface matches catalog.Catalog and is
// defined as a type alias to avoid duplicating the method set.
type Cataloger = catalog.Catalog

// Carter handles the shopping cart lifecycle.
//
// Carts are optional pre-checkout containers that let buyers accumulate
// items before starting a checkout session. A cart can be converted to
// a checkout via CheckoutRequest.CartID.
//
// All methods accept an ownerID for access control scoping. An empty
// ownerID represents a guest (unauthenticated) session.
type Carter interface {
	CreateCart(ownerID string, items []model.LineItemRequest) (*model.Cart, error)
	GetCart(id, ownerID string) (*model.Cart, error)
	UpdateCart(id, ownerID string, items []model.LineItemRequest) (*model.Cart, error)
	CancelCart(id, ownerID string) (*model.Cart, error)
}

// Checkouter handles the checkout session lifecycle as defined by the
// UCP Shopping Checkout capability (dev.ucp.shopping.checkout, version
// 2026-01-11).
//
// A checkout session transitions through statuses: incomplete →
// ready_for_complete → completed (or canceled at any non-completed
// stage). The UpdateCheckout method handles line item changes, buyer
// identity, fulfillment configuration (dev.ucp.shopping.fulfillment),
// discount code application (dev.ucp.shopping.discount), and buyer
// consent (dev.ucp.shopping.buyer_consent).
//
// All methods return a checkout hash (string) alongside the checkout.
// The hash is a SHA-256 digest of the material checkout fields used by
// the MCP transport for the buyer approval flow. REST transports may
// ignore this value.
//
// CompleteCheckout validates the approval hash, processes payment, and
// creates the resulting Order.
type Checkouter interface {
	CreateCheckout(ownerID, country string, req *model.CheckoutRequest) (*model.Checkout, string, error)
	GetCheckout(id, ownerID string) (*model.Checkout, string, error)
	UpdateCheckout(id, ownerID string, req *model.CheckoutRequest) (*model.Checkout, string, error)
	CompleteCheckout(id, ownerID, country, approvalHash string, req *model.CheckoutRequest) (*model.Checkout, *model.Order, string, error)
	CancelCheckout(id, ownerID string) (*model.Checkout, string, error)
}

// Orderer handles order management as defined by the UCP Shopping Order
// capability (dev.ucp.shopping.order, version 2026-01-11).
//
// Orders are created by completing a checkout session. Once created,
// orders support retrieval, listing, cancellation, and updates
// (fulfillment events, adjustments).
type Orderer interface {
	GetOrder(id, ownerID string) (*model.Order, error)
	ListOrders(ownerID string) ([]*model.Order, error)
	CancelOrder(id, ownerID string) error
	UpdateOrder(id string, req model.OrderUpdateRequest) (*model.Order, error)
}

// Merchant composes all UCP capability interfaces into a single facade.
//
// Any UCP-compliant merchant implementation must satisfy this interface.
// The interface covers the full UCP Shopping Service surface:
//   - Product catalog (Cataloger)
//   - Shopping carts (Carter)
//   - Checkout sessions with fulfillment, discounts, and consent (Checkouter)
//   - Order management (Orderer)
//
// Reset clears all transient state (checkouts, orders, carts, sessions).
// It is required by the UCP conformance test harness to isolate test
// runs.
type Merchant interface {
	Cataloger
	Carter
	Checkouter
	Orderer
	Reset()
}

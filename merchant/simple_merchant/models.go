package main

import "github.com/owulveryck/ucp-merchant-test/internal/model"

// Type aliases re-export model types for backward compatibility with tests.

type UCPEnvelope = model.UCPEnvelope
type Capability = model.Capability
type Checkout = model.Checkout
type Link = model.Link
type LineItem = model.LineItem
type Item = model.Item
type Total = model.Total
type OrderRef = model.OrderRef
type Fulfillment = model.Fulfillment
type FulfillmentMethod = model.FulfillmentMethod
type FulfillmentGroup = model.FulfillmentGroup
type FulfillmentOption = model.FulfillmentOption
type FulfillmentDestination = model.FulfillmentDestination
type Payment = model.Payment
type Buyer = model.Buyer
type Consent = model.Consent
type Discounts = model.Discounts
type AppliedDiscount = model.AppliedDiscount
type Order = model.Order
type OrderLineItem = model.OrderLineItem
type OrderQuantity = model.OrderQuantity
type OrderFulfillment = model.OrderFulfillment
type Expectation = model.Expectation
type FulfillmentEvent = model.FulfillmentEvent
type EventLineItem = model.EventLineItem
type Adjustment = model.Adjustment
type Cart = model.Cart
type Message = model.Message

// MCP types
type MCPCheckoutState = model.MCPCheckoutState
type Shipment = model.Shipment
type ShippingOption = model.ShippingOption

// JSON-RPC types
type jsonRPCRequest = model.JSONRPCRequest
type jsonRPCResponse = model.JSONRPCResponse
type rpcError = model.RPCError
type toolDef = model.ToolDef

// Event types
type DashboardEvent = model.DashboardEvent

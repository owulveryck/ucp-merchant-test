// Package webhook implements order event delivery for the UCP Shopping Service.
//
// In the Universal Commerce Protocol, businesses send order lifecycle events
// to platforms via webhook (dev.ucp.shopping.order). The platform provides its
// webhook URL in the order capability's config field during capability
// negotiation. The business discovers this URL from the platform's UCP-Agent
// profile header and uses it to POST order status changes.
//
// # Webhook URL Resolution
//
// The UCP-Agent header (RFC 8941 Dictionary format) contains a profile parameter
// pointing to the platform's UCP discovery profile. ResolveWebhookURL fetches
// this profile, parses the JSON structure, and extracts the webhook_url from
// the platform's capabilities configuration.
//
// Example UCP-Agent header:
//
//	agent_name=test; profile="https://platform.example/.well-known/ucp"
//
// # Event Delivery
//
// SendWebhookEvent asynchronously POSTs a JSON event payload to the resolved
// webhook URL. Events include an event_type field indicating the order lifecycle
// transition (e.g., order_created, order_updated). Per UCP guidelines, businesses
// must send the full order entity on updates (not incremental deltas) and must
// retry failed webhook deliveries.
//
// Webhook payloads should be signed per the UCP Message Signatures specification
// (RFC 9421) to ensure authenticity and integrity. This test implementation
// sends unsigned events for simplicity.
package webhook

// Package event provides a publish-subscribe hub for broadcasting real-time
// events from the UCP merchant server.
//
// In the Universal Commerce Protocol, order status changes are communicated
// to platforms via webhook events (dev.ucp.shopping.order). This package
// implements an in-process event hub that distributes DashboardEvent messages
// to subscribed clients via server-sent events (SSE). Subscribers receive
// events on a buffered channel and can unsubscribe at any time, at which point
// the channel is closed.
//
// The Hub is used by the merchant dashboard to provide live visibility into
// checkout lifecycle transitions, order placements, cart operations, catalog
// queries, and other UCP operations as they occur.
package event

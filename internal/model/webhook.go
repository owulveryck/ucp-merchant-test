package model

// WebhookEvent represents a webhook notification sent to the platform.
type WebhookEvent struct {
	EventType  string `json:"event_type"`
	CheckoutID string `json:"checkout_id"`
	Order      *Order `json:"order,omitempty"`
}

package rest

import (
	"fmt"
	"net/http"
	"time"

	"github.com/owulveryck/ucp-merchant-test/pkg/model"
	"github.com/owulveryck/ucp-merchant-test/pkg/webhook"
)

func (s *Server) simulateShipping(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	secret := r.Header.Get("Simulation-Secret")
	if secret != s.simulationSecret {
		writeError(w, http.StatusForbidden, "Invalid or missing simulation secret")
		return
	}

	id := extractPathParam(r.URL.Path, "/testing/simulate-shipping/")

	event := model.FulfillmentEvent{
		ID:             fmt.Sprintf("evt_ship_%s", id),
		OccurredAt:     time.Now().UTC().Format(time.RFC3339),
		Type:           "shipped",
		TrackingNumber: fmt.Sprintf("TRK-%s", id),
		Description:    "Order shipped",
	}

	// Get current order to build event line items and preserve existing events.
	// Use empty ownerID — UpdateOrder doesn't check ownership; we just need
	// the current state. Try with empty owner first.
	order, err := s.merchant.UpdateOrder(id, model.OrderUpdateRequest{})
	if mapError(w, err) {
		return
	}

	for _, li := range order.LineItems {
		event.LineItems = append(event.LineItems, model.EventLineItem{
			ID:       li.ID,
			Quantity: li.Quantity.Total,
		})
	}

	// Append the new event to existing events
	allEvents := append(order.Fulfillment.Events, event)
	order, err = s.merchant.UpdateOrder(id, model.OrderUpdateRequest{
		Fulfillment: &model.OrderFulfillmentUpdate{
			Events: allEvents,
		},
	})
	if mapError(w, err) {
		return
	}

	// Send webhook
	if order.CheckoutID != "" {
		if webhookURL := s.GetWebhookURL(order.CheckoutID); webhookURL != "" {
			webhook.SendWebhookEvent(webhookURL, model.WebhookEvent{
				EventType:  "order_shipped",
				CheckoutID: order.CheckoutID,
				Order:      order,
			})
		}
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{"status": "shipped"})
}

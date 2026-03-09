package event

import (
	"sync"

	"github.com/owulveryck/ucp-merchant-test/internal/model"
)

// Hub broadcasts events to SSE subscribers.
type Hub struct {
	mu          sync.Mutex
	subscribers []chan model.DashboardEvent
}

// NewHub creates a new event hub.
func NewHub() *Hub {
	return &Hub{}
}

// Subscribe creates a buffered subscriber channel for receiving dashboard events.
// The channel has capacity for 64 events to avoid blocking the publisher.
func (h *Hub) Subscribe() chan model.DashboardEvent {
	ch := make(chan model.DashboardEvent, 64)
	h.mu.Lock()
	h.subscribers = append(h.subscribers, ch)
	h.mu.Unlock()
	return ch
}

// Unsubscribe removes a subscriber and closes its channel.
func (h *Hub) Unsubscribe(ch chan model.DashboardEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for i, s := range h.subscribers {
		if s == ch {
			h.subscribers = append(h.subscribers[:i], h.subscribers[i+1:]...)
			close(ch)
			return
		}
	}
}

// Publish broadcasts an event to all subscribers. If a subscriber's channel
// is full, the event is dropped for that subscriber to avoid blocking.
func (h *Hub) Publish(event model.DashboardEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, ch := range h.subscribers {
		select {
		case ch <- event:
		default:
		}
	}
}

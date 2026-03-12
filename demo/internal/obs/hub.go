package obs

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Event is an observability event emitted by demo components.
type Event struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
	Type      string    `json:"type"`
	Summary   string    `json:"summary"`
	Data      any       `json:"data,omitempty"`
	Duration  int64     `json:"duration_ms,omitempty"`
}

// Command is an instruction sent from the dashboard to the client agent.
type Command struct {
	Instruction string `json:"instruction"`
}

// Hub collects events and broadcasts them to SSE subscribers.
type Hub struct {
	mu          sync.RWMutex
	events      []Event
	subscribers map[chan Event]struct{}
	commands    chan Command
}

// NewHub creates a new observability hub.
func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[chan Event]struct{}),
		commands:    make(chan Command, 8),
	}
}

// Add records a new event and broadcasts it.
func (h *Hub) Add(e Event) {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now()
	}

	h.mu.Lock()
	h.events = append(h.events, e)
	subs := make([]chan Event, 0, len(h.subscribers))
	for ch := range h.subscribers {
		subs = append(subs, ch)
	}
	h.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- e:
		default:
		}
	}
}

// Events returns all recorded events.
func (h *Hub) Events() []Event {
	h.mu.RLock()
	defer h.mu.RUnlock()
	cp := make([]Event, len(h.events))
	copy(cp, h.events)
	return cp
}

// Subscribe returns a channel that receives new events.
func (h *Hub) Subscribe() chan Event {
	ch := make(chan Event, 64)
	h.mu.Lock()
	h.subscribers[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

// Unsubscribe removes a subscriber channel.
func (h *Hub) Unsubscribe(ch chan Event) {
	h.mu.Lock()
	delete(h.subscribers, ch)
	h.mu.Unlock()
	close(ch)
}

// SendCommand sends a command to the commands channel (non-blocking).
func (h *Hub) SendCommand(cmd Command) {
	select {
	case h.commands <- cmd:
	default:
	}
}

// Commands returns a read-only channel for consuming commands.
func (h *Hub) Commands() <-chan Command {
	return h.commands
}

// Report generates a summary report as JSON.
func (h *Hub) Report() json.RawMessage {
	events := h.Events()
	sources := make(map[string]int)
	types := make(map[string]int)
	for _, e := range events {
		sources[e.Source]++
		types[e.Type]++
	}
	report := map[string]any{
		"total_events": len(events),
		"by_source":    sources,
		"by_type":      types,
	}
	data, _ := json.Marshal(report)
	return data
}

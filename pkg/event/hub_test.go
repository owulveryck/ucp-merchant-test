package event

import (
	"testing"
	"time"

	"github.com/owulveryck/ucp-merchant-test/pkg/model"
)

func TestPubSub(t *testing.T) {
	h := NewHub()
	ch := h.Subscribe()

	evt := model.DashboardEvent{Type: "test", Summary: "hello", Timestamp: time.Now()}
	h.Publish(evt)

	select {
	case received := <-ch:
		if received.Type != "test" {
			t.Errorf("expected type 'test', got %s", received.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event")
	}

	h.Unsubscribe(ch)
}

func TestUnsubscribe(t *testing.T) {
	h := NewHub()
	ch := h.Subscribe()
	h.Unsubscribe(ch)

	// Channel should be closed
	_, ok := <-ch
	if ok {
		t.Error("expected channel to be closed")
	}
}

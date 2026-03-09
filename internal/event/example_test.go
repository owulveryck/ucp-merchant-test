package event_test

import (
	"fmt"
	"time"

	"github.com/owulveryck/ucp-merchant-test/internal/event"
	"github.com/owulveryck/ucp-merchant-test/internal/model"
)

func ExampleHub() {
	hub := event.NewHub()

	ch := hub.Subscribe()

	hub.Publish(model.DashboardEvent{
		Type:      "checkout_created",
		ID:        "co_001",
		Summary:   "New checkout session",
		Timestamp: time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
	})

	evt := <-ch
	fmt.Println(evt.Type)
	fmt.Println(evt.ID)

	hub.Unsubscribe(ch)
	// Output:
	// checkout_created
	// co_001
}

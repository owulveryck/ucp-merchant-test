package idempotency_test

import (
	"fmt"

	"github.com/owulveryck/ucp-merchant-test/pkg/idempotency"
)

func ExampleStore() {
	s := idempotency.NewStore()

	// First request: not yet seen
	_, exists := s.Check("req-abc")
	fmt.Println("exists:", exists)

	// Store the result
	hash := idempotency.HashPayload([]byte(`{"items":["roses"]}`))
	s.Store("req-abc", hash, 200, []byte(`{"id":"co_001"}`))

	// Second request: already processed
	entry, exists := s.Check("req-abc")
	fmt.Println("exists:", exists)
	fmt.Println("status:", entry.StatusCode)
	// Output:
	// exists: false
	// exists: true
	// status: 200
}

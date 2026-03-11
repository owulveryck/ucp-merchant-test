package store_test

import (
	"fmt"

	"github.com/owulveryck/ucp-merchant-test/pkg/store"
)

func ExampleStore() {
	s := store.New()

	id1 := s.NewSessionID()
	id2 := s.NewSessionID()

	fmt.Println(id1)
	fmt.Println(id2)
	// Output:
	// session-0001
	// session-0002
}

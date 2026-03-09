package payment_test

import (
	"fmt"

	"github.com/owulveryck/ucp-merchant-test/internal/merchant/payment"
)

func ExampleParsePayment() {
	req := map[string]interface{}{
		"payment": map[string]interface{}{
			"selected_instrument_id": "instr_1",
			"instruments":            []interface{}{},
		},
	}

	p := payment.ParsePayment(req)
	fmt.Println(p.SelectedInstrumentID)
	fmt.Println(len(p.Handlers) > 0)
	// Output:
	// instr_1
	// true
}

func ExampleParseBuyer() {
	req := map[string]interface{}{
		"buyer": map[string]interface{}{
			"first_name": "John",
			"last_name":  "Doe",
			"email":      "john@example.com",
		},
	}

	b := payment.ParseBuyer(req)
	fmt.Println(b.FirstName)
	fmt.Println(b.LastName)
	fmt.Println(b.Email)
	// Output:
	// John
	// Doe
	// john@example.com
}

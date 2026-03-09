package payment_test

import (
	"fmt"

	"github.com/owulveryck/ucp-merchant-test/internal/merchant/payment"
	"github.com/owulveryck/ucp-merchant-test/internal/model"
)

func ExampleParsePayment() {
	req := &model.PaymentRequest{
		SelectedInstrumentID: "instr_1",
		Instruments:          []map[string]interface{}{},
	}

	p := payment.ParsePayment(req)
	fmt.Println(p.SelectedInstrumentID)
	fmt.Println(len(p.Handlers) > 0)
	// Output:
	// instr_1
	// true
}

func ExampleParseBuyer() {
	req := &model.BuyerRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
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

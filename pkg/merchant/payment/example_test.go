package payment_test

import (
	"fmt"

	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/payment"
	"github.com/owulveryck/ucp-merchant-test/pkg/model"
)

func ExampleParsePayment() {
	req := &model.PaymentRequest{
		SelectedInstrumentID: "instr_1",
		Instruments:          []map[string]any{},
	}

	p := payment.ParsePayment(req)
	fmt.Println(p.SelectedInstrumentID)
	fmt.Println(len(p.Handlers) > 0)
	// Output:
	// instr_1
	// true
}

func ExampleDefaultPayment() {
	p := payment.DefaultPayment()
	fmt.Println(p.SelectedInstrumentID)
	fmt.Println(len(p.Handlers))
	// Output:
	// instr_1
	// 3
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

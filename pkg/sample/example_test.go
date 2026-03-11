package sample_test

import (
	"fmt"

	"github.com/owulveryck/ucp-merchant-test/pkg/sample"
)

func ExampleDataSource() {
	ds := sample.New()
	ds.Customers = []sample.CSVCustomer{
		{ID: "cust_1", Name: "John Doe", Email: "john@example.com"},
	}

	cust := ds.FindCustomerByEmail("john@example.com")
	fmt.Println(cust.Name)

	d := ds.FindDiscountByCode("10OFF")
	fmt.Println(d)
	// Output:
	// John Doe
	// <nil>
}

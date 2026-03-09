package data_test

import (
	"fmt"

	"github.com/owulveryck/ucp-merchant-test/internal/data"
)

func ExampleDataSource() {
	ds := data.New()
	ds.Customers = []data.CSVCustomer{
		{ID: "cust_1", Name: "John Doe", Email: "john@example.com"},
	}
	ds.Addresses = []data.CSVAddress{
		{ID: "addr_1", CustomerID: "cust_1", StreetAddress: "123 Main St", City: "Springfield", State: "IL", PostalCode: "62701", Country: "US"},
	}
	ds.Discounts = []data.CSVDiscount{
		{Code: "10OFF", Type: "percentage", Value: 10, Description: "10% Off"},
	}

	cust := ds.FindCustomerByEmail("john@example.com")
	fmt.Println(cust.Name)

	addrs := ds.FindAddressesForEmail("john@example.com")
	fmt.Println(addrs[0].City)

	d := ds.FindDiscountByCode("10OFF")
	fmt.Println(d.Description)
	// Output:
	// John Doe
	// Springfield
	// 10% Off
}

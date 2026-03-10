package catalog_test

import (
	"encoding/json"
	"fmt"

	"github.com/owulveryck/ucp-merchant-test/internal/catalog"
	"github.com/owulveryck/ucp-merchant-test/internal/ucp"
)

func ExampleContainsCountry() {
	countries := []ucp.Country{ucp.NewCountry("US"), ucp.NewCountry("CA"), ucp.NewCountry("GB")}

	fmt.Println(ucp.ContainsCountry(countries, ucp.NewCountry("us")))
	fmt.Println(ucp.ContainsCountry(countries, ucp.NewCountry("CA")))
	fmt.Println(ucp.ContainsCountry(countries, ucp.NewCountry("FR")))
	// Output:
	// true
	// true
	// false
}

func ExampleCategoryStat() {
	stat := catalog.CategoryStat{Name: "bouquets", Count: 3}

	b, _ := json.Marshal(stat)
	fmt.Println(string(b))
	// Output:
	// {"name":"bouquets","count":3}
}

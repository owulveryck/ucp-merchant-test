package catalog_test

import (
	"encoding/json"
	"fmt"

	"github.com/owulveryck/ucp-merchant-test/internal/catalog"
)

func ExampleContainsCountry() {
	countries := []string{"US", "CA", "GB"}

	fmt.Println(catalog.ContainsCountry(countries, "us"))
	fmt.Println(catalog.ContainsCountry(countries, "CA"))
	fmt.Println(catalog.ContainsCountry(countries, "FR"))
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

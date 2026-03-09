package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	dataDir := flag.String("data-dir", "", "directory containing CSV test data files")
	outputDir := flag.String("output-dir", "", "directory to write JSON files (defaults to data-dir)")
	flag.Parse()

	if *dataDir == "" {
		fmt.Fprintln(os.Stderr, "usage: csvtojson --data-dir <dir> [--output-dir <dir>]")
		os.Exit(1)
	}
	if *outputDir == "" {
		*outputDir = *dataDir
	}

	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "error creating output dir: %v\n", err)
		os.Exit(1)
	}

	if err := convertProducts(*dataDir, *outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "products: %v\n", err)
		os.Exit(1)
	}
	if err := convertCustomers(*dataDir, *outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "customers: %v\n", err)
		os.Exit(1)
	}
	if err := convertAddresses(*dataDir, *outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "addresses: %v\n", err)
		os.Exit(1)
	}
	if err := convertDiscounts(*dataDir, *outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "discounts: %v\n", err)
		os.Exit(1)
	}
	if err := convertShippingRates(*dataDir, *outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "shipping_rates: %v\n", err)
		os.Exit(1)
	}
	if err := convertPromotions(*dataDir, *outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "promotions: %v\n", err)
		os.Exit(1)
	}
	if err := convertPaymentInstruments(*dataDir, *outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "payment_instruments: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("JSON files written to", *outputDir)
}

func readCSV(path string) ([][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.LazyQuotes = true
	return r.ReadAll()
}

func writeJSON(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

type product struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Price    int    `json:"price"`
	ImageURL string `json:"image_url"`
	Quantity int    `json:"quantity"`
}

func convertProducts(dataDir, outputDir string) error {
	prodRows, err := readCSV(filepath.Join(dataDir, "products.csv"))
	if err != nil {
		return err
	}
	invRows, err := readCSV(filepath.Join(dataDir, "inventory.csv"))
	if err != nil {
		return err
	}
	inv := map[string]int{}
	for i, row := range invRows {
		if i == 0 || len(row) < 2 {
			continue
		}
		qty, _ := strconv.Atoi(row[1])
		inv[row[0]] = qty
	}

	var products []product
	for i, row := range prodRows {
		if i == 0 || len(row) < 4 {
			continue
		}
		price, _ := strconv.Atoi(row[2])
		products = append(products, product{
			ID:       row[0],
			Title:    row[1],
			Price:    price,
			ImageURL: row[3],
			Quantity: inv[row[0]],
		})
	}
	return writeJSON(filepath.Join(outputDir, "products.json"), products)
}

type customer struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func convertCustomers(dataDir, outputDir string) error {
	rows, err := readCSV(filepath.Join(dataDir, "customers.csv"))
	if err != nil {
		return err
	}
	var customers []customer
	for i, row := range rows {
		if i == 0 || len(row) < 3 {
			continue
		}
		customers = append(customers, customer{ID: row[0], Name: row[1], Email: row[2]})
	}
	return writeJSON(filepath.Join(outputDir, "customers.json"), customers)
}

type address struct {
	ID            string `json:"id"`
	CustomerID    string `json:"customer_id"`
	StreetAddress string `json:"street_address"`
	City          string `json:"city"`
	State         string `json:"state"`
	PostalCode    string `json:"postal_code"`
	Country       string `json:"country"`
}

func convertAddresses(dataDir, outputDir string) error {
	rows, err := readCSV(filepath.Join(dataDir, "addresses.csv"))
	if err != nil {
		return err
	}
	var addresses []address
	for i, row := range rows {
		if i == 0 || len(row) < 7 {
			continue
		}
		addresses = append(addresses, address{
			ID: row[0], CustomerID: row[1], StreetAddress: row[2],
			City: row[3], State: row[4], PostalCode: row[5], Country: row[6],
		})
	}
	return writeJSON(filepath.Join(outputDir, "addresses.json"), addresses)
}

type discountEntry struct {
	Code        string `json:"code"`
	Type        string `json:"type"`
	Value       int    `json:"value"`
	Description string `json:"description"`
}

func convertDiscounts(dataDir, outputDir string) error {
	rows, err := readCSV(filepath.Join(dataDir, "discounts.csv"))
	if err != nil {
		return err
	}
	var discounts []discountEntry
	for i, row := range rows {
		if i == 0 || len(row) < 4 {
			continue
		}
		val, _ := strconv.Atoi(row[2])
		discounts = append(discounts, discountEntry{Code: row[0], Type: row[1], Value: val, Description: row[3]})
	}
	return writeJSON(filepath.Join(outputDir, "discounts.json"), discounts)
}

type shippingRate struct {
	ID           string `json:"id"`
	CountryCode  string `json:"country_code"`
	ServiceLevel string `json:"service_level"`
	Price        int    `json:"price"`
	Title        string `json:"title"`
}

func convertShippingRates(dataDir, outputDir string) error {
	rows, err := readCSV(filepath.Join(dataDir, "shipping_rates.csv"))
	if err != nil {
		return err
	}
	var rates []shippingRate
	for i, row := range rows {
		if i == 0 || len(row) < 5 {
			continue
		}
		price, _ := strconv.Atoi(row[3])
		rates = append(rates, shippingRate{ID: row[0], CountryCode: row[1], ServiceLevel: row[2], Price: price, Title: row[4]})
	}
	return writeJSON(filepath.Join(outputDir, "shipping_rates.json"), rates)
}

type promotion struct {
	ID              string   `json:"id"`
	Type            string   `json:"type"`
	MinSubtotal     int      `json:"min_subtotal"`
	EligibleItemIDs []string `json:"eligible_item_ids"`
	Description     string   `json:"description"`
}

func convertPromotions(dataDir, outputDir string) error {
	rows, err := readCSV(filepath.Join(dataDir, "promotions.csv"))
	if err != nil {
		return err
	}
	var promotions []promotion
	for i, row := range rows {
		if i == 0 || len(row) < 5 {
			continue
		}
		minSub, _ := strconv.Atoi(row[2])
		var eligible []string
		if row[3] != "" {
			json.Unmarshal([]byte(row[3]), &eligible)
		}
		if eligible == nil {
			eligible = []string{}
		}
		promotions = append(promotions, promotion{
			ID: row[0], Type: row[1], MinSubtotal: minSub,
			EligibleItemIDs: eligible, Description: row[4],
		})
	}
	return writeJSON(filepath.Join(outputDir, "promotions.json"), promotions)
}

type paymentInstrument struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Brand      string `json:"brand"`
	LastDigits string `json:"last_digits"`
	Token      string `json:"token"`
	HandlerID  string `json:"handler_id"`
}

func convertPaymentInstruments(dataDir, outputDir string) error {
	rows, err := readCSV(filepath.Join(dataDir, "payment_instruments.csv"))
	if err != nil {
		return err
	}
	var instruments []paymentInstrument
	for i, row := range rows {
		if i == 0 || len(row) < 6 {
			continue
		}
		instruments = append(instruments, paymentInstrument{
			ID: row[0], Type: row[1], Brand: row[2],
			LastDigits: row[3], Token: row[4], HandlerID: row[5],
		})
	}
	return writeJSON(filepath.Join(outputDir, "payment_instruments.json"), instruments)
}

package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// CSV-loaded data stores for the flower shop dataset.

type CSVCustomer struct {
	ID    string
	Name  string
	Email string
}

type CSVAddress struct {
	ID            string
	CustomerID    string
	StreetAddress string
	City          string
	State         string
	PostalCode    string
	Country       string
}

type CSVPaymentInstrument struct {
	ID         string
	Type       string
	Brand      string
	LastDigits string
	Token      string
	HandlerID  string
}

type CSVDiscount struct {
	Code        string
	Type        string // "percentage" or "fixed_amount"
	Value       int    // percentage value or fixed amount in cents
	Description string
}

type CSVShippingRate struct {
	ID           string
	CountryCode  string
	ServiceLevel string
	Price        int
	Title        string
}

type CSVPromotion struct {
	ID              string
	Type            string
	MinSubtotal     int
	EligibleItemIDs []string
	Description     string
}

type ConformanceInput struct {
	Currency string `json:"currency"`
	Items    []struct {
		ID    string `json:"id"`
		Title string `json:"title"`
		Price int    `json:"price"`
	} `json:"items"`
	OutOfStockItem struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	} `json:"out_of_stock_item"`
	NonExistentItem struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	} `json:"non_existent_item"`
}

// FlowerShopData holds all CSV-loaded data.
type FlowerShopData struct {
	mu                 sync.RWMutex
	Products           []Product
	Customers          []CSVCustomer
	Addresses          []CSVAddress
	PaymentInstruments []CSVPaymentInstrument
	Discounts          []CSVDiscount
	ShippingRates      []CSVShippingRate
	Promotions         []CSVPromotion
	ConformanceInput   *ConformanceInput

	// Dynamic address store: email -> []CSVAddress (for new addresses added during sessions)
	DynamicAddresses map[string][]CSVAddress
}

var shopData = &FlowerShopData{
	DynamicAddresses: make(map[string][]CSVAddress),
}

func loadFlowerShopData(dataDir string) error {
	if err := loadProducts(dataDir); err != nil {
		return fmt.Errorf("products: %w", err)
	}
	if err := loadInventory(dataDir); err != nil {
		return fmt.Errorf("inventory: %w", err)
	}
	if err := loadCustomers(dataDir); err != nil {
		return fmt.Errorf("customers: %w", err)
	}
	if err := loadAddresses(dataDir); err != nil {
		return fmt.Errorf("addresses: %w", err)
	}
	if err := loadPaymentInstruments(dataDir); err != nil {
		return fmt.Errorf("payment_instruments: %w", err)
	}
	if err := loadDiscounts(dataDir); err != nil {
		return fmt.Errorf("discounts: %w", err)
	}
	if err := loadShippingRates(dataDir); err != nil {
		return fmt.Errorf("shipping_rates: %w", err)
	}
	if err := loadPromotions(dataDir); err != nil {
		return fmt.Errorf("promotions: %w", err)
	}
	if err := loadConformanceInput(dataDir); err != nil {
		return fmt.Errorf("conformance_input: %w", err)
	}

	// Set catalog from loaded products
	catalog = shopData.Products
	productSeq = len(catalog)

	return nil
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

func loadProducts(dataDir string) error {
	rows, err := readCSV(filepath.Join(dataDir, "products.csv"))
	if err != nil {
		return err
	}
	// header: id,title,price,image_url
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 4 {
			continue
		}
		price, _ := strconv.Atoi(row[2])
		shopData.Products = append(shopData.Products, Product{
			ID:       row[0],
			Title:    row[1],
			Price:    price,
			ImageURL: row[3],
			Quantity: 0, // will be set by inventory
			Rank:     100,
		})
	}
	return nil
}

func loadInventory(dataDir string) error {
	rows, err := readCSV(filepath.Join(dataDir, "inventory.csv"))
	if err != nil {
		return err
	}
	// header: product_id,quantity
	inv := map[string]int{}
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 2 {
			continue
		}
		qty, _ := strconv.Atoi(row[1])
		inv[row[0]] = qty
	}
	for j := range shopData.Products {
		if q, ok := inv[shopData.Products[j].ID]; ok {
			shopData.Products[j].Quantity = q
		}
	}
	return nil
}

func loadCustomers(dataDir string) error {
	rows, err := readCSV(filepath.Join(dataDir, "customers.csv"))
	if err != nil {
		return err
	}
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 3 {
			continue
		}
		shopData.Customers = append(shopData.Customers, CSVCustomer{
			ID:    row[0],
			Name:  row[1],
			Email: row[2],
		})
	}
	return nil
}

func loadAddresses(dataDir string) error {
	rows, err := readCSV(filepath.Join(dataDir, "addresses.csv"))
	if err != nil {
		return err
	}
	// header: id,customer_id,street_address,city,state,postal_code,country
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 7 {
			continue
		}
		shopData.Addresses = append(shopData.Addresses, CSVAddress{
			ID:            row[0],
			CustomerID:    row[1],
			StreetAddress: row[2],
			City:          row[3],
			State:         row[4],
			PostalCode:    row[5],
			Country:       row[6],
		})
	}
	return nil
}

func loadPaymentInstruments(dataDir string) error {
	rows, err := readCSV(filepath.Join(dataDir, "payment_instruments.csv"))
	if err != nil {
		return err
	}
	// header: id,type,brand,last_digits,token,handler_id
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 6 {
			continue
		}
		shopData.PaymentInstruments = append(shopData.PaymentInstruments, CSVPaymentInstrument{
			ID:         row[0],
			Type:       row[1],
			Brand:      row[2],
			LastDigits: row[3],
			Token:      row[4],
			HandlerID:  row[5],
		})
	}
	return nil
}

func loadDiscounts(dataDir string) error {
	rows, err := readCSV(filepath.Join(dataDir, "discounts.csv"))
	if err != nil {
		return err
	}
	// header: code,type,value,description
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 4 {
			continue
		}
		val, _ := strconv.Atoi(row[2])
		shopData.Discounts = append(shopData.Discounts, CSVDiscount{
			Code:        row[0],
			Type:        row[1],
			Value:       val,
			Description: row[3],
		})
	}
	return nil
}

func loadShippingRates(dataDir string) error {
	rows, err := readCSV(filepath.Join(dataDir, "shipping_rates.csv"))
	if err != nil {
		return err
	}
	// header: id,country_code,service_level,price,title
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 5 {
			continue
		}
		price, _ := strconv.Atoi(row[3])
		shopData.ShippingRates = append(shopData.ShippingRates, CSVShippingRate{
			ID:           row[0],
			CountryCode:  row[1],
			ServiceLevel: row[2],
			Price:        price,
			Title:        row[4],
		})
	}
	return nil
}

func loadPromotions(dataDir string) error {
	rows, err := readCSV(filepath.Join(dataDir, "promotions.csv"))
	if err != nil {
		return err
	}
	// header: id,type,min_subtotal,eligible_item_ids,description
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 5 {
			continue
		}
		minSub, _ := strconv.Atoi(row[2])
		var eligible []string
		if row[3] != "" {
			// Parse JSON array like ["bouquet_roses"]
			var items []string
			if err := json.Unmarshal([]byte(row[3]), &items); err == nil {
				eligible = items
			}
		}
		shopData.Promotions = append(shopData.Promotions, CSVPromotion{
			ID:              row[0],
			Type:            row[1],
			MinSubtotal:     minSub,
			EligibleItemIDs: eligible,
			Description:     row[4],
		})
	}
	return nil
}

func loadConformanceInput(dataDir string) error {
	path := filepath.Join(dataDir, "conformance_input.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var ci ConformanceInput
	if err := json.Unmarshal(data, &ci); err != nil {
		return err
	}
	shopData.ConformanceInput = &ci
	return nil
}

// findCustomerByEmail finds a customer by email.
func findCustomerByEmail(email string) *CSVCustomer {
	for i := range shopData.Customers {
		if strings.EqualFold(shopData.Customers[i].Email, email) {
			return &shopData.Customers[i]
		}
	}
	return nil
}

// findAddressesByCustomerID returns all addresses for a customer.
func findAddressesByCustomerID(customerID string) []CSVAddress {
	var result []CSVAddress
	for _, a := range shopData.Addresses {
		if a.CustomerID == customerID {
			result = append(result, a)
		}
	}
	return result
}

// findAddressesForEmail returns stored addresses for an email (CSV + dynamic).
func findAddressesForEmail(email string) []CSVAddress {
	var result []CSVAddress
	cust := findCustomerByEmail(email)
	if cust != nil {
		result = append(result, findAddressesByCustomerID(cust.ID)...)
	}
	shopData.mu.RLock()
	if addrs, ok := shopData.DynamicAddresses[strings.ToLower(email)]; ok {
		result = append(result, addrs...)
	}
	shopData.mu.RUnlock()
	return result
}

// findDiscount looks up a discount code.
func findDiscountByCode(code string) *CSVDiscount {
	for i := range shopData.Discounts {
		if strings.EqualFold(shopData.Discounts[i].Code, code) {
			return &shopData.Discounts[i]
		}
	}
	return nil
}

// findPaymentInstrument looks up a payment instrument by ID.
func findPaymentInstrumentByID(id string) *CSVPaymentInstrument {
	for i := range shopData.PaymentInstruments {
		if shopData.PaymentInstruments[i].ID == id {
			return &shopData.PaymentInstruments[i]
		}
	}
	return nil
}

// findPaymentInstrumentByToken looks up instrument by token.
func findPaymentInstrumentByToken(token string) *CSVPaymentInstrument {
	for i := range shopData.PaymentInstruments {
		if shopData.PaymentInstruments[i].Token == token {
			return &shopData.PaymentInstruments[i]
		}
	}
	return nil
}

// getShippingRatesForCountry returns shipping rates applicable to a country.
func getShippingRatesForCountry(country string) []CSVShippingRate {
	var result []CSVShippingRate
	for _, r := range shopData.ShippingRates {
		if strings.EqualFold(r.CountryCode, country) || r.CountryCode == "default" {
			result = append(result, r)
		}
	}
	// Deduplicate by service_level: prefer country-specific over default.
	seen := map[string]bool{}
	var deduped []CSVShippingRate
	// First pass: country-specific
	for _, r := range result {
		if !strings.EqualFold(r.CountryCode, "default") {
			seen[r.ServiceLevel] = true
			deduped = append(deduped, r)
		}
	}
	// Second pass: defaults for unseen levels
	for _, r := range result {
		if strings.EqualFold(r.CountryCode, "default") && !seen[r.ServiceLevel] {
			deduped = append(deduped, r)
		}
	}
	return deduped
}

// matchExistingAddress checks if a submitted address matches an existing one by content.
func matchExistingAddress(addrs []CSVAddress, street, locality, region, postal, country string) *CSVAddress {
	for i := range addrs {
		a := &addrs[i]
		if strings.EqualFold(a.StreetAddress, street) &&
			strings.EqualFold(a.City, locality) &&
			strings.EqualFold(a.State, region) &&
			strings.EqualFold(a.PostalCode, postal) &&
			strings.EqualFold(a.Country, country) {
			return a
		}
	}
	return nil
}

// saveDynamicAddress stores a new address for a user email and returns its ID.
func saveDynamicAddress(email string, addr CSVAddress) string {
	shopData.mu.Lock()
	defer shopData.mu.Unlock()
	key := strings.ToLower(email)
	shopData.DynamicAddresses[key] = append(shopData.DynamicAddresses[key], addr)
	return addr.ID
}

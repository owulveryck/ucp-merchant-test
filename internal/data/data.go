package data

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/owulveryck/ucp-merchant-test/internal/catalog"
)

// CSVCustomer represents a test buyer identity loaded from customers.csv.
type CSVCustomer struct {
	ID    string
	Name  string
	Email string
}

// CSVAddress represents a shipping destination loaded from addresses.csv,
// linked to a customer by CustomerID.
type CSVAddress struct {
	ID            string
	CustomerID    string
	StreetAddress string
	City          string
	State         string
	PostalCode    string
	Country       string
}

// CSVPaymentInstrument represents a test payment method loaded from payment_instruments.csv.
type CSVPaymentInstrument struct {
	ID         string
	Type       string
	Brand      string
	LastDigits string
	Token      string
	HandlerID  string
}

// CSVDiscount represents a discount code loaded from discounts.csv, with type
// ("percentage" or "fixed_amount") and value.
type CSVDiscount struct {
	Code        string
	Type        string
	Value       int
	Description string
}

// CSVShippingRate represents a fulfillment cost loaded from shipping_rates.csv,
// keyed by country code and service level.
type CSVShippingRate struct {
	ID           string
	CountryCode  string
	ServiceLevel string
	Price        int
	Title        string
}

// CSVPromotion represents an automatic discount rule loaded from promotions.csv,
// such as free shipping for orders above a minimum subtotal.
type CSVPromotion struct {
	ID              string
	Type            string
	MinSubtotal     int
	EligibleItemIDs []string
	Description     string
}

// ConformanceInput holds test expectations loaded from conformance_input.json,
// defining currency, available items, out-of-stock items, and non-existent item IDs.
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

// DataSource holds all CSV-loaded data.
type DataSource struct {
	Mu                 sync.RWMutex
	Products           []catalog.Product
	Customers          []CSVCustomer
	Addresses          []CSVAddress
	PaymentInstruments []CSVPaymentInstrument
	Discounts          []CSVDiscount
	ShippingRates      []CSVShippingRate
	Promotions         []CSVPromotion
	ConformanceInput   *ConformanceInput

	DynamicAddresses map[string][]CSVAddress
}

// New creates a new empty DataSource.
func New() *DataSource {
	return &DataSource{
		DynamicAddresses: make(map[string][]CSVAddress),
	}
}

// Load loads all flower shop data from a directory.
func (ds *DataSource) Load(dataDir string) error {
	if err := ds.loadProducts(dataDir); err != nil {
		return fmt.Errorf("products: %w", err)
	}
	if err := ds.loadInventory(dataDir); err != nil {
		return fmt.Errorf("inventory: %w", err)
	}
	if err := ds.loadCustomers(dataDir); err != nil {
		return fmt.Errorf("customers: %w", err)
	}
	if err := ds.loadAddresses(dataDir); err != nil {
		return fmt.Errorf("addresses: %w", err)
	}
	if err := ds.loadPaymentInstruments(dataDir); err != nil {
		return fmt.Errorf("payment_instruments: %w", err)
	}
	if err := ds.loadDiscounts(dataDir); err != nil {
		return fmt.Errorf("discounts: %w", err)
	}
	if err := ds.loadShippingRates(dataDir); err != nil {
		return fmt.Errorf("shipping_rates: %w", err)
	}
	if err := ds.loadPromotions(dataDir); err != nil {
		return fmt.Errorf("promotions: %w", err)
	}
	if err := ds.loadConformanceInput(dataDir); err != nil {
		return fmt.Errorf("conformance_input: %w", err)
	}
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

func (ds *DataSource) loadProducts(dataDir string) error {
	rows, err := readCSV(filepath.Join(dataDir, "products.csv"))
	if err != nil {
		return err
	}
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 4 {
			continue
		}
		price, _ := strconv.Atoi(row[2])
		ds.Products = append(ds.Products, catalog.Product{
			ID:       row[0],
			Title:    row[1],
			Price:    price,
			ImageURL: row[3],
			Quantity: 0,
			Rank:     100,
		})
	}
	return nil
}

func (ds *DataSource) loadInventory(dataDir string) error {
	rows, err := readCSV(filepath.Join(dataDir, "inventory.csv"))
	if err != nil {
		return err
	}
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
	for j := range ds.Products {
		if q, ok := inv[ds.Products[j].ID]; ok {
			ds.Products[j].Quantity = q
		}
	}
	return nil
}

func (ds *DataSource) loadCustomers(dataDir string) error {
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
		ds.Customers = append(ds.Customers, CSVCustomer{
			ID:    row[0],
			Name:  row[1],
			Email: row[2],
		})
	}
	return nil
}

func (ds *DataSource) loadAddresses(dataDir string) error {
	rows, err := readCSV(filepath.Join(dataDir, "addresses.csv"))
	if err != nil {
		return err
	}
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 7 {
			continue
		}
		ds.Addresses = append(ds.Addresses, CSVAddress{
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

func (ds *DataSource) loadPaymentInstruments(dataDir string) error {
	rows, err := readCSV(filepath.Join(dataDir, "payment_instruments.csv"))
	if err != nil {
		return err
	}
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 6 {
			continue
		}
		ds.PaymentInstruments = append(ds.PaymentInstruments, CSVPaymentInstrument{
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

func (ds *DataSource) loadDiscounts(dataDir string) error {
	rows, err := readCSV(filepath.Join(dataDir, "discounts.csv"))
	if err != nil {
		return err
	}
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 4 {
			continue
		}
		val, _ := strconv.Atoi(row[2])
		ds.Discounts = append(ds.Discounts, CSVDiscount{
			Code:        row[0],
			Type:        row[1],
			Value:       val,
			Description: row[3],
		})
	}
	return nil
}

func (ds *DataSource) loadShippingRates(dataDir string) error {
	rows, err := readCSV(filepath.Join(dataDir, "shipping_rates.csv"))
	if err != nil {
		return err
	}
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 5 {
			continue
		}
		price, _ := strconv.Atoi(row[3])
		ds.ShippingRates = append(ds.ShippingRates, CSVShippingRate{
			ID:           row[0],
			CountryCode:  row[1],
			ServiceLevel: row[2],
			Price:        price,
			Title:        row[4],
		})
	}
	return nil
}

func (ds *DataSource) loadPromotions(dataDir string) error {
	rows, err := readCSV(filepath.Join(dataDir, "promotions.csv"))
	if err != nil {
		return err
	}
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
			var items []string
			if err := json.Unmarshal([]byte(row[3]), &items); err == nil {
				eligible = items
			}
		}
		ds.Promotions = append(ds.Promotions, CSVPromotion{
			ID:              row[0],
			Type:            row[1],
			MinSubtotal:     minSub,
			EligibleItemIDs: eligible,
			Description:     row[4],
		})
	}
	return nil
}

func (ds *DataSource) loadConformanceInput(dataDir string) error {
	path := filepath.Join(dataDir, "conformance_input.json")
	d, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var ci ConformanceInput
	if err := json.Unmarshal(d, &ci); err != nil {
		return err
	}
	ds.ConformanceInput = &ci
	return nil
}

// FindCustomerByEmail finds a customer by email.
func (ds *DataSource) FindCustomerByEmail(email string) *CSVCustomer {
	for i := range ds.Customers {
		if strings.EqualFold(ds.Customers[i].Email, email) {
			return &ds.Customers[i]
		}
	}
	return nil
}

// FindAddressesByCustomerID returns all addresses for a customer.
func (ds *DataSource) FindAddressesByCustomerID(customerID string) []CSVAddress {
	var result []CSVAddress
	for _, a := range ds.Addresses {
		if a.CustomerID == customerID {
			result = append(result, a)
		}
	}
	return result
}

// FindAddressesForEmail returns stored addresses for an email (CSV + dynamic).
func (ds *DataSource) FindAddressesForEmail(email string) []CSVAddress {
	var result []CSVAddress
	cust := ds.FindCustomerByEmail(email)
	if cust != nil {
		result = append(result, ds.FindAddressesByCustomerID(cust.ID)...)
	}
	ds.Mu.RLock()
	if addrs, ok := ds.DynamicAddresses[strings.ToLower(email)]; ok {
		result = append(result, addrs...)
	}
	ds.Mu.RUnlock()
	return result
}

// FindDiscountByCode looks up a discount code.
func (ds *DataSource) FindDiscountByCode(code string) *CSVDiscount {
	for i := range ds.Discounts {
		if strings.EqualFold(ds.Discounts[i].Code, code) {
			return &ds.Discounts[i]
		}
	}
	return nil
}

// FindPaymentInstrumentByID looks up a payment instrument by ID.
func (ds *DataSource) FindPaymentInstrumentByID(id string) *CSVPaymentInstrument {
	for i := range ds.PaymentInstruments {
		if ds.PaymentInstruments[i].ID == id {
			return &ds.PaymentInstruments[i]
		}
	}
	return nil
}

// FindPaymentInstrumentByToken looks up instrument by token.
func (ds *DataSource) FindPaymentInstrumentByToken(token string) *CSVPaymentInstrument {
	for i := range ds.PaymentInstruments {
		if ds.PaymentInstruments[i].Token == token {
			return &ds.PaymentInstruments[i]
		}
	}
	return nil
}

// GetShippingRatesForCountry returns shipping rates applicable to a country.
func (ds *DataSource) GetShippingRatesForCountry(country string) []CSVShippingRate {
	var result []CSVShippingRate
	for _, r := range ds.ShippingRates {
		if strings.EqualFold(r.CountryCode, country) || r.CountryCode == "default" {
			result = append(result, r)
		}
	}
	seen := map[string]bool{}
	var deduped []CSVShippingRate
	for _, r := range result {
		if !strings.EqualFold(r.CountryCode, "default") {
			seen[r.ServiceLevel] = true
			deduped = append(deduped, r)
		}
	}
	for _, r := range result {
		if strings.EqualFold(r.CountryCode, "default") && !seen[r.ServiceLevel] {
			deduped = append(deduped, r)
		}
	}
	return deduped
}

// MatchExistingAddress checks if a submitted address matches an existing one.
func MatchExistingAddress(addrs []CSVAddress, street, locality, region, postal, country string) *CSVAddress {
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

// GetPromotions returns the loaded promotions.
func (ds *DataSource) GetPromotions() []CSVPromotion {
	return ds.Promotions
}

// SaveDynamicAddress stores a new address for a user email.
func (ds *DataSource) SaveDynamicAddress(email string, addr CSVAddress) string {
	ds.Mu.Lock()
	defer ds.Mu.Unlock()
	key := strings.ToLower(email)
	ds.DynamicAddresses[key] = append(ds.DynamicAddresses[key], addr)
	return addr.ID
}

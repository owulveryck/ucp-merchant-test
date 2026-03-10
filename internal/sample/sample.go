package sample

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
	"github.com/owulveryck/ucp-merchant-test/internal/merchant/discount"
	"github.com/owulveryck/ucp-merchant-test/internal/merchant/fulfillment"
)

// CSVCustomer represents a test buyer identity loaded from customers.csv.
type CSVCustomer struct {
	ID    string
	Name  string
	Email string
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

// csvDiscount is an internal CSV representation of a discount code.
type csvDiscount struct {
	Code        string
	Type        string
	Value       int
	Description string
}

// csvShippingRate is an internal CSV representation of a shipping rate.
type csvShippingRate struct {
	ID           string
	CountryCode  string
	ServiceLevel string
	Price        int
	Title        string
}

// csvPromotion is an internal CSV representation of a promotion.
type csvPromotion struct {
	ID              string
	Type            string
	MinSubtotal     int
	EligibleItemIDs []string
	Description     string
}

// csvAddress is an internal CSV representation of an address.
type csvAddress struct {
	ID            string
	CustomerID    string
	StreetAddress string
	City          string
	State         string
	PostalCode    string
	Country       string
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
	addresses          []csvAddress
	PaymentInstruments []CSVPaymentInstrument
	discounts          []csvDiscount
	shippingRates      []csvShippingRate
	promotions         []csvPromotion
	ConformanceInput   *ConformanceInput

	dynamicAddresses map[string][]fulfillment.Address
}

// New creates a new empty DataSource.
func New() *DataSource {
	return &DataSource{
		dynamicAddresses: make(map[string][]fulfillment.Address),
	}
}

// ResetDynamicAddresses clears the dynamic addresses map.
func (ds *DataSource) ResetDynamicAddresses() {
	ds.Mu.Lock()
	ds.dynamicAddresses = make(map[string][]fulfillment.Address)
	ds.Mu.Unlock()
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
		ds.addresses = append(ds.addresses, csvAddress{
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
		ds.discounts = append(ds.discounts, csvDiscount{
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
		ds.shippingRates = append(ds.shippingRates, csvShippingRate{
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
		ds.promotions = append(ds.promotions, csvPromotion{
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

func (ds *DataSource) findAddressesByCustomerID(customerID string) []csvAddress {
	var result []csvAddress
	for _, a := range ds.addresses {
		if a.CustomerID == customerID {
			result = append(result, a)
		}
	}
	return result
}

func csvAddressToAddress(a csvAddress) fulfillment.Address {
	return fulfillment.Address{
		ID:            a.ID,
		CustomerID:    a.CustomerID,
		StreetAddress: a.StreetAddress,
		City:          a.City,
		State:         a.State,
		PostalCode:    a.PostalCode,
		Country:       a.Country,
	}
}

// FindAddressesForEmail returns stored addresses for an email (CSV + dynamic).
func (ds *DataSource) FindAddressesForEmail(email string) []fulfillment.Address {
	var result []fulfillment.Address
	cust := ds.FindCustomerByEmail(email)
	if cust != nil {
		for _, a := range ds.findAddressesByCustomerID(cust.ID) {
			result = append(result, csvAddressToAddress(a))
		}
	}
	ds.Mu.RLock()
	if addrs, ok := ds.dynamicAddresses[strings.ToLower(email)]; ok {
		result = append(result, addrs...)
	}
	ds.Mu.RUnlock()
	return result
}

// FindDiscountByCode looks up a discount code.
func (ds *DataSource) FindDiscountByCode(code string) *discount.Discount {
	for i := range ds.discounts {
		if strings.EqualFold(ds.discounts[i].Code, code) {
			d := ds.discounts[i]
			return &discount.Discount{
				Code:        d.Code,
				Type:        d.Type,
				Value:       d.Value,
				Description: d.Description,
			}
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
func (ds *DataSource) GetShippingRatesForCountry(country string) []fulfillment.ShippingRate {
	var result []csvShippingRate
	for _, r := range ds.shippingRates {
		if strings.EqualFold(r.CountryCode, country) || r.CountryCode == "default" {
			result = append(result, r)
		}
	}
	seen := map[string]bool{}
	var deduped []fulfillment.ShippingRate
	for _, r := range result {
		if !strings.EqualFold(r.CountryCode, "default") {
			seen[r.ServiceLevel] = true
			deduped = append(deduped, fulfillment.ShippingRate{
				ID:           r.ID,
				CountryCode:  r.CountryCode,
				ServiceLevel: r.ServiceLevel,
				Price:        r.Price,
				Title:        r.Title,
			})
		}
	}
	for _, r := range result {
		if strings.EqualFold(r.CountryCode, "default") && !seen[r.ServiceLevel] {
			deduped = append(deduped, fulfillment.ShippingRate{
				ID:           r.ID,
				CountryCode:  r.CountryCode,
				ServiceLevel: r.ServiceLevel,
				Price:        r.Price,
				Title:        r.Title,
			})
		}
	}
	return deduped
}

// GetPromotions returns the loaded promotions.
func (ds *DataSource) GetPromotions() []fulfillment.Promotion {
	result := make([]fulfillment.Promotion, len(ds.promotions))
	for i, p := range ds.promotions {
		result[i] = fulfillment.Promotion{
			ID:              p.ID,
			Type:            p.Type,
			MinSubtotal:     p.MinSubtotal,
			EligibleItemIDs: p.EligibleItemIDs,
			Description:     p.Description,
		}
	}
	return result
}

// SaveDynamicAddress stores a new address for a user email.
func (ds *DataSource) SaveDynamicAddress(email string, addr fulfillment.Address) string {
	ds.Mu.Lock()
	defer ds.Mu.Unlock()
	key := strings.ToLower(email)
	ds.dynamicAddresses[key] = append(ds.dynamicAddresses[key], addr)
	return addr.ID
}

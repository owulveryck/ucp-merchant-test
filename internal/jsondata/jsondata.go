package jsondata

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/owulveryck/ucp-merchant-test/internal/catalog"
	"github.com/owulveryck/ucp-merchant-test/internal/merchant/discount"
	"github.com/owulveryck/ucp-merchant-test/internal/merchant/fulfillment"
)

// Customer represents a test buyer identity.
type Customer struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// PaymentInstrument represents a test payment method.
type PaymentInstrument struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Brand      string `json:"brand"`
	LastDigits string `json:"last_digits"`
	Token      string `json:"token"`
	HandlerID  string `json:"handler_id"`
}

// jsonProduct is the JSON representation of a product with inventory.
type jsonProduct struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Price    int    `json:"price"`
	ImageURL string `json:"image_url"`
	Quantity int    `json:"quantity"`
}

// jsonDiscount is the JSON representation of a discount code.
type jsonDiscount struct {
	Code        string `json:"code"`
	Type        string `json:"type"`
	Value       int    `json:"value"`
	Description string `json:"description"`
}

// jsonAddress is the JSON representation of an address.
type jsonAddress struct {
	ID            string `json:"id"`
	CustomerID    string `json:"customer_id"`
	StreetAddress string `json:"street_address"`
	City          string `json:"city"`
	State         string `json:"state"`
	PostalCode    string `json:"postal_code"`
	Country       string `json:"country"`
}

// jsonShippingRate is the JSON representation of a shipping rate.
type jsonShippingRate struct {
	ID           string `json:"id"`
	CountryCode  string `json:"country_code"`
	ServiceLevel string `json:"service_level"`
	Price        int    `json:"price"`
	Title        string `json:"title"`
}

// jsonPromotion is the JSON representation of a promotion.
type jsonPromotion struct {
	ID              string   `json:"id"`
	Type            string   `json:"type"`
	MinSubtotal     int      `json:"min_subtotal"`
	EligibleItemIDs []string `json:"eligible_item_ids"`
	Description     string   `json:"description"`
}

// ConformanceInput holds test expectations loaded from conformance_input.json.
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

// DataSource holds all JSON-loaded data.
type DataSource struct {
	mu                 sync.RWMutex
	Products           []catalog.Product
	Customers          []Customer
	addresses          []jsonAddress
	PaymentInstruments []PaymentInstrument
	discounts          []jsonDiscount
	shippingRates      []jsonShippingRate
	promotions         []jsonPromotion
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
	ds.mu.Lock()
	ds.dynamicAddresses = make(map[string][]fulfillment.Address)
	ds.mu.Unlock()
}

func loadJSON(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// Load loads all data from a directory of JSON files.
func (ds *DataSource) Load(dataDir string) error {
	var products []jsonProduct
	if err := loadJSON(filepath.Join(dataDir, "products.json"), &products); err != nil {
		return fmt.Errorf("products: %w", err)
	}
	for _, p := range products {
		ds.Products = append(ds.Products, catalog.Product{
			ID:       p.ID,
			Title:    p.Title,
			Price:    p.Price,
			ImageURL: p.ImageURL,
			Quantity: p.Quantity,
			Rank:     100,
		})
	}

	if err := loadJSON(filepath.Join(dataDir, "customers.json"), &ds.Customers); err != nil {
		return fmt.Errorf("customers: %w", err)
	}
	if err := loadJSON(filepath.Join(dataDir, "addresses.json"), &ds.addresses); err != nil {
		return fmt.Errorf("addresses: %w", err)
	}
	if err := loadJSON(filepath.Join(dataDir, "payment_instruments.json"), &ds.PaymentInstruments); err != nil {
		return fmt.Errorf("payment_instruments: %w", err)
	}
	if err := loadJSON(filepath.Join(dataDir, "discounts.json"), &ds.discounts); err != nil {
		return fmt.Errorf("discounts: %w", err)
	}
	if err := loadJSON(filepath.Join(dataDir, "shipping_rates.json"), &ds.shippingRates); err != nil {
		return fmt.Errorf("shipping_rates: %w", err)
	}
	if err := loadJSON(filepath.Join(dataDir, "promotions.json"), &ds.promotions); err != nil {
		return fmt.Errorf("promotions: %w", err)
	}

	var ci ConformanceInput
	if err := loadJSON(filepath.Join(dataDir, "conformance_input.json"), &ci); err != nil {
		return fmt.Errorf("conformance_input: %w", err)
	}
	ds.ConformanceInput = &ci
	return nil
}

// FindCustomerByEmail finds a customer by email.
func (ds *DataSource) FindCustomerByEmail(email string) *Customer {
	for i := range ds.Customers {
		if strings.EqualFold(ds.Customers[i].Email, email) {
			return &ds.Customers[i]
		}
	}
	return nil
}

func (ds *DataSource) findAddressesByCustomerID(customerID string) []jsonAddress {
	var result []jsonAddress
	for _, a := range ds.addresses {
		if a.CustomerID == customerID {
			result = append(result, a)
		}
	}
	return result
}

func jsonAddressToAddress(a jsonAddress) fulfillment.Address {
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

// FindAddressesForEmail returns stored addresses for an email.
func (ds *DataSource) FindAddressesForEmail(email string) []fulfillment.Address {
	var result []fulfillment.Address
	cust := ds.FindCustomerByEmail(email)
	if cust != nil {
		for _, a := range ds.findAddressesByCustomerID(cust.ID) {
			result = append(result, jsonAddressToAddress(a))
		}
	}
	ds.mu.RLock()
	if addrs, ok := ds.dynamicAddresses[strings.ToLower(email)]; ok {
		result = append(result, addrs...)
	}
	ds.mu.RUnlock()
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
func (ds *DataSource) FindPaymentInstrumentByID(id string) *PaymentInstrument {
	for i := range ds.PaymentInstruments {
		if ds.PaymentInstruments[i].ID == id {
			return &ds.PaymentInstruments[i]
		}
	}
	return nil
}

// FindPaymentInstrumentByToken looks up instrument by token.
func (ds *DataSource) FindPaymentInstrumentByToken(token string) *PaymentInstrument {
	for i := range ds.PaymentInstruments {
		if ds.PaymentInstruments[i].Token == token {
			return &ds.PaymentInstruments[i]
		}
	}
	return nil
}

// GetShippingRatesForCountry returns shipping rates applicable to a country.
func (ds *DataSource) GetShippingRatesForCountry(country string) []fulfillment.ShippingRate {
	var result []jsonShippingRate
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
	ds.mu.Lock()
	defer ds.mu.Unlock()
	key := strings.ToLower(email)
	ds.dynamicAddresses[key] = append(ds.dynamicAddresses[key], addr)
	return addr.ID
}

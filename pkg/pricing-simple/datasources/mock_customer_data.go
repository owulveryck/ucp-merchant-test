// Package datasources provides data access for the pricing system.
package datasources

import (
	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-simple/models"
)

// MockCustomerDataSource provides mock customer data for testing.
type MockCustomerDataSource struct {
	customers map[string]models.CustomerProfile
}

// NewMockCustomerData creates a mock customer data source.
func NewMockCustomerData() *MockCustomerDataSource {
	return &MockCustomerDataSource{
		customers: map[string]models.CustomerProfile{
			"customer_premium": {
				CustomerID:      "customer_premium",
				TotalSpent:      150000, // $1500
				PurchaseCount:   15,
				LastPurchaseDays: 10,
			},
			"customer_gold": {
				CustomerID:      "customer_gold",
				TotalSpent:      75000, // $750
				PurchaseCount:   8,
				LastPurchaseDays: 25,
			},
			"customer_silver": {
				CustomerID:      "customer_silver",
				TotalSpent:      25000, // $250
				PurchaseCount:   3,
				LastPurchaseDays: 45,
			},
			"customer_standard": {
				CustomerID:      "customer_standard",
				TotalSpent:      10000, // $100
				PurchaseCount:   1,
				LastPurchaseDays: 90,
			},
			"customer_new": {
				CustomerID:      "customer_new",
				TotalSpent:      0,
				PurchaseCount:   0,
				LastPurchaseDays: 999,
			},
		},
	}
}

// GetCustomerProfile retrieves a customer profile.
func (m *MockCustomerDataSource) GetCustomerProfile(customerID string) (models.CustomerProfile, error) {
	if profile, ok := m.customers[customerID]; ok {
		return profile, nil
	}

	// Return new customer profile for unknown IDs
	return models.CustomerProfile{
		CustomerID:      customerID,
		TotalSpent:      0,
		PurchaseCount:   0,
		LastPurchaseDays: 999,
	}, nil
}

// AddCustomer adds or updates a customer profile.
func (m *MockCustomerDataSource) AddCustomer(profile models.CustomerProfile) {
	m.customers[profile.CustomerID] = profile
}

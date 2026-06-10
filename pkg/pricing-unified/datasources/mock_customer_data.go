// Package datasources provides data sources for the pricing system.
package datasources

import (
	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-unified/agents"
)

// MockCustomerDataSource provides mock customer data for testing.
type MockCustomerDataSource struct {
	customers map[string]agents.CustomerProfile
}

// NewMockCustomerDataSource creates a new mock customer data source.
func NewMockCustomerDataSource() *MockCustomerDataSource {
	return &MockCustomerDataSource{
		customers: map[string]agents.CustomerProfile{
			"default_customer": {
				CustomerID:       "default_customer",
				TotalSpent:       50000,  // $500 - Gold tier
				PurchaseCount:    5,
				LastPurchaseDays: 15,
			},
			"premium_customer": {
				CustomerID:       "premium_customer",
				TotalSpent:       150000, // $1500 - Premium tier
				PurchaseCount:    20,
				LastPurchaseDays: 5,
			},
			"standard_customer": {
				CustomerID:       "standard_customer",
				TotalSpent:       10000, // $100 - Standard tier
				PurchaseCount:    1,
				LastPurchaseDays: 90,
			},
			// Additional demo customers
			"elsi": {
				CustomerID:       "elsi",
				TotalSpent:       85000,  // $850 - Gold tier
				PurchaseCount:    8,
				LastPurchaseDays: 10,
			},
			"alice": {
				CustomerID:       "alice",
				TotalSpent:       120000, // $1200 - Premium tier
				PurchaseCount:    15,
				LastPurchaseDays: 7,
			},
			"bob": {
				CustomerID:       "bob",
				TotalSpent:       5000,   // $50 - Standard tier
				PurchaseCount:    1,
				LastPurchaseDays: 120,
			},
			"john": {
				CustomerID:       "john",
				TotalSpent:       35000,  // $350 - Silver tier
				PurchaseCount:    4,
				LastPurchaseDays: 20,
			},
		},
	}
}

// GetCustomerProfile retrieves a customer's profile.
func (m *MockCustomerDataSource) GetCustomerProfile(customerID string) (agents.CustomerProfile, error) {
	if profile, ok := m.customers[customerID]; ok {
		return profile, nil
	}

	// Return default profile for unknown customers
	return agents.CustomerProfile{
		CustomerID:       customerID,
		TotalSpent:       0,
		PurchaseCount:    0,
		LastPurchaseDays: 999,
	}, nil
}

// AddCustomer adds or updates a customer profile.
func (m *MockCustomerDataSource) AddCustomer(profile agents.CustomerProfile) {
	m.customers[profile.CustomerID] = profile
}

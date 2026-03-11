package merchanttest

import (
	"github.com/owulveryck/ucp-merchant-test/internal/catalog"
	"github.com/owulveryck/ucp-merchant-test/internal/model"
	"github.com/owulveryck/ucp-merchant-test/internal/ucp"
)

// Mock is a function-field mock of merchant.Merchant.
// Each interface method delegates to its corresponding XxxFunc field
// if non-nil, otherwise returns zero/nil defaults.
type Mock struct {
	FindFunc          func(id string) *catalog.Product
	FilterFunc        func(category ucp.Category, brand, query string, country ucp.Country, currency ucp.Currency, language ucp.Language) []catalog.Product
	CategoryCountFunc func() []catalog.CategoryStat
	LookupFunc        func(id string, shipsTo ucp.Country) *catalog.Product
	SearchFunc        func(params catalog.SearchParams) []catalog.SearchResult

	CreateCartFunc func(ownerID string, items []model.LineItemRequest) (*model.Cart, error)
	GetCartFunc    func(id, ownerID string) (*model.Cart, error)
	UpdateCartFunc func(id, ownerID string, items []model.LineItemRequest) (*model.Cart, error)
	CancelCartFunc func(id, ownerID string) (*model.Cart, error)

	CreateCheckoutFunc   func(ownerID string, country ucp.Country, req *model.CheckoutRequest) (*model.Checkout, string, error)
	GetCheckoutFunc      func(id, ownerID string) (*model.Checkout, string, error)
	UpdateCheckoutFunc   func(id, ownerID string, req *model.CheckoutRequest) (*model.Checkout, string, error)
	CompleteCheckoutFunc func(id, ownerID string, country ucp.Country, approvalHash string, req *model.CheckoutRequest) (*model.Checkout, *model.Order, string, error)
	CancelCheckoutFunc   func(id, ownerID string) (*model.Checkout, string, error)

	GetOrderFunc    func(id, ownerID string) (*model.Order, error)
	ListOrdersFunc  func(ownerID string) ([]*model.Order, error)
	CancelOrderFunc func(id, ownerID string) error
	UpdateOrderFunc func(id string, req model.OrderUpdateRequest) (*model.Order, error)

	ResetFunc func()
}

// NewMock returns a Mock with no function fields set.
func NewMock() *Mock {
	return &Mock{}
}

func (m *Mock) Find(id string) *catalog.Product {
	if m.FindFunc != nil {
		return m.FindFunc(id)
	}
	return nil
}

func (m *Mock) Filter(category ucp.Category, brand, query string, country ucp.Country, currency ucp.Currency, language ucp.Language) []catalog.Product {
	if m.FilterFunc != nil {
		return m.FilterFunc(category, brand, query, country, currency, language)
	}
	return nil
}

func (m *Mock) CategoryCount() []catalog.CategoryStat {
	if m.CategoryCountFunc != nil {
		return m.CategoryCountFunc()
	}
	return nil
}

func (m *Mock) Lookup(id string, shipsTo ucp.Country) *catalog.Product {
	if m.LookupFunc != nil {
		return m.LookupFunc(id, shipsTo)
	}
	return nil
}

func (m *Mock) Search(params catalog.SearchParams) []catalog.SearchResult {
	if m.SearchFunc != nil {
		return m.SearchFunc(params)
	}
	return nil
}

func (m *Mock) CreateCart(ownerID string, items []model.LineItemRequest) (*model.Cart, error) {
	if m.CreateCartFunc != nil {
		return m.CreateCartFunc(ownerID, items)
	}
	return nil, nil
}

func (m *Mock) GetCart(id, ownerID string) (*model.Cart, error) {
	if m.GetCartFunc != nil {
		return m.GetCartFunc(id, ownerID)
	}
	return nil, nil
}

func (m *Mock) UpdateCart(id, ownerID string, items []model.LineItemRequest) (*model.Cart, error) {
	if m.UpdateCartFunc != nil {
		return m.UpdateCartFunc(id, ownerID, items)
	}
	return nil, nil
}

func (m *Mock) CancelCart(id, ownerID string) (*model.Cart, error) {
	if m.CancelCartFunc != nil {
		return m.CancelCartFunc(id, ownerID)
	}
	return nil, nil
}

func (m *Mock) CreateCheckout(ownerID string, country ucp.Country, req *model.CheckoutRequest) (*model.Checkout, string, error) {
	if m.CreateCheckoutFunc != nil {
		return m.CreateCheckoutFunc(ownerID, country, req)
	}
	return nil, "", nil
}

func (m *Mock) GetCheckout(id, ownerID string) (*model.Checkout, string, error) {
	if m.GetCheckoutFunc != nil {
		return m.GetCheckoutFunc(id, ownerID)
	}
	return nil, "", nil
}

func (m *Mock) UpdateCheckout(id, ownerID string, req *model.CheckoutRequest) (*model.Checkout, string, error) {
	if m.UpdateCheckoutFunc != nil {
		return m.UpdateCheckoutFunc(id, ownerID, req)
	}
	return nil, "", nil
}

func (m *Mock) CompleteCheckout(id, ownerID string, country ucp.Country, approvalHash string, req *model.CheckoutRequest) (*model.Checkout, *model.Order, string, error) {
	if m.CompleteCheckoutFunc != nil {
		return m.CompleteCheckoutFunc(id, ownerID, country, approvalHash, req)
	}
	return nil, nil, "", nil
}

func (m *Mock) CancelCheckout(id, ownerID string) (*model.Checkout, string, error) {
	if m.CancelCheckoutFunc != nil {
		return m.CancelCheckoutFunc(id, ownerID)
	}
	return nil, "", nil
}

func (m *Mock) GetOrder(id, ownerID string) (*model.Order, error) {
	if m.GetOrderFunc != nil {
		return m.GetOrderFunc(id, ownerID)
	}
	return nil, nil
}

func (m *Mock) ListOrders(ownerID string) ([]*model.Order, error) {
	if m.ListOrdersFunc != nil {
		return m.ListOrdersFunc(ownerID)
	}
	return nil, nil
}

func (m *Mock) CancelOrder(id, ownerID string) error {
	if m.CancelOrderFunc != nil {
		return m.CancelOrderFunc(id, ownerID)
	}
	return nil
}

func (m *Mock) UpdateOrder(id string, req model.OrderUpdateRequest) (*model.Order, error) {
	if m.UpdateOrderFunc != nil {
		return m.UpdateOrderFunc(id, req)
	}
	return nil, nil
}

func (m *Mock) Reset() {
	if m.ResetFunc != nil {
		m.ResetFunc()
	}
}

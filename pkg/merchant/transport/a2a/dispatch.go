package a2a

import (
	"context"

	"github.com/owulveryck/ucp-merchant-test/pkg/merchant"
)

// actionHandler processes a single UCP action and returns the response
// data to be included in the A2A DataPart.
type actionHandler func(ctx context.Context, ac *actionContext) (map[string]any, error)

// actionHandlers returns the dispatch map from action names to handlers.
// Action names match the UCP Shopping Service tool names used in the
// MCP transport.
func (s *Server) actionHandlers() map[string]actionHandler {
	handlers := map[string]actionHandler{
		// Catalog
		"list_products":       s.handleListProducts,
		"get_product_details": s.handleGetProductDetails,
		"search_catalog":      s.handleSearchCatalog,
		"lookup_product":      s.handleLookupProduct,
		// Cart
		"create_cart": s.handleCreateCart,
		"get_cart":    s.handleGetCart,
		"update_cart": s.handleUpdateCart,
		"cancel_cart": s.handleCancelCart,
		// Checkout
		"create_checkout":   s.handleCreateCheckout,
		"get_checkout":      s.handleGetCheckout,
		"update_checkout":   s.handleUpdateCheckout,
		"complete_checkout": s.handleCompleteCheckout,
		"cancel_checkout":   s.handleCancelCheckout,
		// Orders
		"get_order":    s.handleGetOrder,
		"list_orders":  s.handleListOrders,
		"update_order": s.handleUpdateOrder,
		"cancel_order": s.handleCancelOrder,
	}
	if _, ok := s.merchant.(merchant.Promoter); ok {
		handlers["list_promotions"] = s.handleListPromotions
	}
	return handlers
}

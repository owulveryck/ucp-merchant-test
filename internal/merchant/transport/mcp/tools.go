package mcp

import (
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerTools registers all UCP tool definitions on the MCP server.
func registerTools(srv *server.MCPServer, s *Server) {
	metaSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"ucp-agent":       map[string]any{"type": "string", "description": "URI identifying the calling agent"},
			"idempotency-key": map[string]any{"type": "string", "description": "UUID for idempotent operations"},
		},
	}
	lineItemsSchema := map[string]any{
		"type": "array",
		"items": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"product_id": map[string]any{"type": "string", "description": "Product SKU ID from the catalog"},
				"quantity":   map[string]any{"type": "integer", "description": "Quantity to add", "minimum": 1},
			},
			"required": []string{"product_id"},
		},
	}

	// list_products
	srv.AddTool(mcp.NewTool("list_products",
		mcp.WithDescription("Browse the product catalog. Call with no arguments to see featured products and available categories. Filter by category, brand, or text query. Returns paginated results with category counts."),
		mcp.WithString("category", mcp.Description("Filter by category (case-insensitive)")),
		mcp.WithString("brand", mcp.Description("Filter by brand (case-insensitive)")),
		mcp.WithString("query", mcp.Description("Text search on product title")),
		mcp.WithNumber("limit", mcp.Description("Max results per page (default 20, max 50)")),
		mcp.WithNumber("offset", mcp.Description("Skip N products for pagination (default 0)")),
	), s.handleListProducts)

	// get_product_details
	srv.AddTool(mcp.NewTool("get_product_details",
		mcp.WithDescription("Get the full product sheet for a specific product, including description, price, stock availability, and image. Use after list_products to review details before recommending."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Product SKU ID (e.g. SKU-001)")),
	), s.handleGetProductDetails)

	// search_catalog
	srv.AddTool(mcp.NewTool("search_catalog",
		mcp.WithDescription("Search the product catalog by keyword. Matches against title, description, and category. Returns matching products with availability."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query string (matches title, description, category)")),
		mcp.WithNumber("limit", mcp.Description("Max results to return (1-300, default 10)")),
	), s.handleSearchCatalog)

	// lookup_product
	srv.AddTool(mcp.NewTool("lookup_product",
		mcp.WithDescription("Look up a single product by its ID. Returns full product details including description, price, stock, and available shipping countries."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Product ID (e.g. SKU-001)")),
	), s.handleLookupProduct)

	// create_cart
	srv.AddTool(mcp.NewToolWithRawSchema("create_cart",
		"Create a new shopping cart with line items. Each line item requires a product_id from the catalog and a quantity (minimum 1).",
		mustMarshal(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"meta": metaSchema,
				"cart": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"line_items": lineItemsSchema,
					},
					"required": []string{"line_items"},
				},
			},
			"required": []string{"cart"},
		}),
	), s.handleCreateCart)

	// get_cart
	srv.AddTool(mcp.NewTool("get_cart",
		mcp.WithDescription("Retrieve an existing cart by its ID. Returns the cart with current line items and calculated totals."),
		mcp.WithObject("meta", mcp.Properties(metaSchema["properties"].(map[string]any))),
		mcp.WithString("id", mcp.Required(), mcp.Description("Cart ID")),
	), s.handleGetCart)

	// update_cart
	srv.AddTool(mcp.NewToolWithRawSchema("update_cart",
		"Replace the line items of an existing cart. Provide the full set of line items desired.",
		mustMarshal(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"meta": metaSchema,
				"id":   map[string]any{"type": "string", "description": "Cart ID"},
				"cart": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"line_items": lineItemsSchema,
					},
					"required": []string{"line_items"},
				},
			},
			"required": []string{"id", "cart"},
		}),
	), s.handleUpdateCart)

	// cancel_cart
	srv.AddTool(mcp.NewTool("cancel_cart",
		mcp.WithDescription("Cancel and remove a cart. The cart ID becomes invalid after cancellation."),
		mcp.WithObject("meta", mcp.Properties(metaSchema["properties"].(map[string]any))),
		mcp.WithString("id", mcp.Required(), mcp.Description("Cart ID")),
	), s.handleCancelCart)

	// create_checkout
	srv.AddTool(mcp.NewToolWithRawSchema("create_checkout",
		"Create a checkout session. Provide either line_items directly or a cart_id to create from an existing cart. Optionally include buyer information (name, email, address).",
		mustMarshal(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"meta": metaSchema,
				"checkout": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"line_items": lineItemsSchema,
						"cart_id":    map[string]any{"type": "string", "description": "Create checkout from existing cart"},
						"buyer":      buyerSchema(),
					},
				},
			},
			"required": []string{"checkout"},
		}),
	), s.handleCreateCheckout)

	// get_checkout
	srv.AddTool(mcp.NewTool("get_checkout",
		mcp.WithDescription("Retrieve a checkout session by its ID. Returns current status, line items, totals, fulfillment options, and a checkout_hash for the approval flow."),
		mcp.WithObject("meta", mcp.Properties(metaSchema["properties"].(map[string]any))),
		mcp.WithString("id", mcp.Required(), mcp.Description("Checkout ID")),
	), s.handleGetCheckout)

	// update_checkout
	srv.AddTool(mcp.NewToolWithRawSchema("update_checkout",
		"Update a checkout session's line items or buyer information. When a valid shipping address is provided, the checkout transitions to ready_for_complete status with fulfillment options.",
		mustMarshal(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"meta": metaSchema,
				"id":   map[string]any{"type": "string", "description": "Checkout ID"},
				"checkout": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"line_items": lineItemsSchema,
						"buyer":      buyerSchema(),
					},
				},
			},
			"required": []string{"id", "checkout"},
		}),
	), s.handleUpdateCheckout)

	// complete_checkout
	srv.AddTool(mcp.NewToolWithRawSchema("complete_checkout",
		"Complete a checkout and place the order. Requires the checkout to be in ready_for_complete status. The approval object with checkout_hash (from get_checkout) proves the buyer reviewed and approved the exact checkout state before purchase.",
		mustMarshal(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"meta": metaSchema,
				"id":   map[string]any{"type": "string", "description": "Checkout ID"},
				"approval": map[string]any{
					"type":        "object",
					"description": "User approval with checkout hash to verify the user reviewed the exact checkout state",
					"properties": map[string]any{
						"checkout_hash": map[string]any{"type": "string", "description": "The checkout_hash from get_checkout, proving the user approved this exact state"},
					},
					"required": []string{"checkout_hash"},
				},
				"checkout": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"payment": map[string]any{
							"type":        "object",
							"description": "Payment information (any object accepted for testing)",
						},
					},
				},
			},
			"required": []string{"id", "approval"},
		}),
	), s.handleCompleteCheckout)

	// cancel_checkout
	srv.AddTool(mcp.NewTool("cancel_checkout",
		mcp.WithDescription("Cancel a checkout session. Only incomplete or ready_for_complete checkouts can be canceled. Completed checkouts cannot be canceled."),
		mcp.WithObject("meta", mcp.Properties(metaSchema["properties"].(map[string]any))),
		mcp.WithString("id", mcp.Required(), mcp.Description("Checkout ID")),
	), s.handleCancelCheckout)

	// get_order
	srv.AddTool(mcp.NewTool("get_order",
		mcp.WithDescription("Retrieve a placed order by its ID. Returns order details including line items, totals, buyer info, fulfillment status, and shipment tracking."),
		mcp.WithObject("meta", mcp.Properties(metaSchema["properties"].(map[string]any))),
		mcp.WithString("id", mcp.Required(), mcp.Description("Order ID")),
	), s.handleGetOrder)

	// list_orders
	srv.AddTool(mcp.NewTool("list_orders",
		mcp.WithDescription("List all placed orders for the authenticated user. Returns order summaries with status, total, and creation date."),
		mcp.WithObject("meta", mcp.Properties(metaSchema["properties"].(map[string]any))),
	), s.handleListOrders)

	// cancel_order
	srv.AddTool(mcp.NewTool("cancel_order",
		mcp.WithDescription("Cancel an order. Only possible when the order status is confirmed or processing (before shipment). Shipped orders cannot be canceled."),
		mcp.WithObject("meta", mcp.Properties(metaSchema["properties"].(map[string]any))),
		mcp.WithString("id", mcp.Required(), mcp.Description("Order ID")),
	), s.handleCancelOrder)
}

func buyerSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name":  map[string]any{"type": "string"},
			"email": map[string]any{"type": "string"},
			"address": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"street":  map[string]any{"type": "string"},
					"city":    map[string]any{"type": "string"},
					"state":   map[string]any{"type": "string"},
					"zip":     map[string]any{"type": "string"},
					"country": map[string]any{"type": "string"},
				},
			},
		},
	}
}

func mustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

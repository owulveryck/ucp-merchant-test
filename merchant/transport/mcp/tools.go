package mcp

import "github.com/owulveryck/ucp-merchant-test/internal/model"

// getToolDefinitions returns the MCP tool definitions for the UCP Shopping Service.
func getToolDefinitions() []model.ToolDef {
	metaSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"ucp-agent":       map[string]interface{}{"type": "string", "description": "URI identifying the calling agent"},
			"idempotency-key": map[string]interface{}{"type": "string", "description": "UUID for idempotent operations"},
		},
	}
	lineItemsSchema := map[string]interface{}{
		"type": "array",
		"items": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"product_id": map[string]interface{}{"type": "string", "description": "Product SKU ID from the catalog"},
				"quantity":   map[string]interface{}{"type": "integer", "description": "Quantity to add", "minimum": 1},
			},
			"required": []string{"product_id"},
		},
	}

	return []model.ToolDef{
		{
			Name:        "list_products",
			Description: "List products from the catalog. Call with no arguments first to see featured products and available categories. Use category, brand, or query filters to narrow results. Use get_product_details to get the full product sheet with description before recommending.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"category": map[string]interface{}{"type": "string", "description": "Filter by category (case-insensitive)"},
					"brand":    map[string]interface{}{"type": "string", "description": "Filter by brand (case-insensitive)"},
					"query":    map[string]interface{}{"type": "string", "description": "Text search on product title (case-insensitive partial match)"},
					"limit":    map[string]interface{}{"type": "integer", "description": "Max results per page (default 20, max 50)"},
					"offset":   map[string]interface{}{"type": "integer", "description": "Skip N products for pagination (default 0)"},
				},
			},
		},
		{
			Name:        "get_product_details",
			Description: "Get the full product sheet (fiche produit) for a specific product, including description and availability. Use this after list_products to get details before recommending a product.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{"type": "string", "description": "Product SKU ID (e.g. SKU-001)"},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "search_catalog",
			Description: "Search the product catalog by keyword. Returns matching products with availability info.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{"type": "string", "description": "Search query string (matches title, description, category)"},
					"limit": map[string]interface{}{"type": "integer", "description": "Max results to return (1-300, default 10)"},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "lookup_product",
			Description: "Look up a single product by its ID. Returns full product details including description, price, stock, and available countries.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{"type": "string", "description": "Product ID (e.g. SKU-001)"},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "create_cart",
			Description: "Create a new shopping cart with line items. Each line item needs a product_id (from list_products) and quantity.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"meta": metaSchema,
					"cart": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"line_items": lineItemsSchema,
						},
						"required": []string{"line_items"},
					},
				},
				"required": []string{"cart"},
			},
		},
		{
			Name:        "get_cart",
			Description: "Retrieve a cart by its ID",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"meta": metaSchema,
					"id":   map[string]interface{}{"type": "string", "description": "Cart ID"},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "update_cart",
			Description: "Update the line items of an existing cart",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"meta": metaSchema,
					"id":   map[string]interface{}{"type": "string", "description": "Cart ID"},
					"cart": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"line_items": lineItemsSchema,
						},
						"required": []string{"line_items"},
					},
				},
				"required": []string{"id", "cart"},
			},
		},
		{
			Name:        "cancel_cart",
			Description: "Cancel and remove a cart",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"meta": metaSchema,
					"id":   map[string]interface{}{"type": "string", "description": "Cart ID"},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "create_checkout",
			Description: "Create a checkout session. Provide either line_items directly or a cart_id to create from an existing cart. Optionally include buyer information.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"meta": metaSchema,
					"checkout": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"line_items": lineItemsSchema,
							"cart_id":    map[string]interface{}{"type": "string", "description": "Create checkout from existing cart"},
							"buyer": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"name":  map[string]interface{}{"type": "string"},
									"email": map[string]interface{}{"type": "string"},
									"address": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"street":  map[string]interface{}{"type": "string"},
											"city":    map[string]interface{}{"type": "string"},
											"state":   map[string]interface{}{"type": "string"},
											"zip":     map[string]interface{}{"type": "string"},
											"country": map[string]interface{}{"type": "string"},
										},
									},
								},
							},
						},
					},
				},
				"required": []string{"checkout"},
			},
		},
		{
			Name:        "get_checkout",
			Description: "Retrieve a checkout by its ID",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"meta": metaSchema,
					"id":   map[string]interface{}{"type": "string", "description": "Checkout ID"},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "update_checkout",
			Description: "Update a checkout's line items or buyer information. When buyer address is provided, status transitions to ready_for_complete.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"meta": metaSchema,
					"id":   map[string]interface{}{"type": "string", "description": "Checkout ID"},
					"checkout": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"line_items": lineItemsSchema,
							"buyer": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"name":  map[string]interface{}{"type": "string"},
									"email": map[string]interface{}{"type": "string"},
									"address": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"street":  map[string]interface{}{"type": "string"},
											"city":    map[string]interface{}{"type": "string"},
											"state":   map[string]interface{}{"type": "string"},
											"zip":     map[string]interface{}{"type": "string"},
											"country": map[string]interface{}{"type": "string"},
										},
									},
								},
							},
						},
					},
				},
				"required": []string{"id", "checkout"},
			},
		},
		{
			Name:        "complete_checkout",
			Description: "Complete a checkout and place the order. Checkout must be in ready_for_complete status. Requires an approval object with checkout_hash obtained from get_checkout, confirming the user has reviewed and approved the purchase.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"meta": metaSchema,
					"id":   map[string]interface{}{"type": "string", "description": "Checkout ID"},
					"approval": map[string]interface{}{
						"type":        "object",
						"description": "User approval with checkout hash to verify the user reviewed the exact checkout state",
						"properties": map[string]interface{}{
							"checkout_hash": map[string]interface{}{"type": "string", "description": "The checkout_hash from get_checkout, proving the user approved this exact state"},
						},
						"required": []string{"checkout_hash"},
					},
					"checkout": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"payment": map[string]interface{}{
								"type":        "object",
								"description": "Payment information (any object accepted for testing)",
							},
						},
					},
				},
				"required": []string{"id", "approval"},
			},
		},
		{
			Name:        "cancel_checkout",
			Description: "Cancel a checkout session. Cannot cancel completed checkouts.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"meta": metaSchema,
					"id":   map[string]interface{}{"type": "string", "description": "Checkout ID"},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "get_order",
			Description: "Retrieve a placed order by its ID, including line items, totals, buyer info, and shipment tracking details.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"meta": metaSchema,
					"id":   map[string]interface{}{"type": "string", "description": "Order ID"},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "list_orders",
			Description: "List all placed orders with their current status, confirmation number, total, and creation date.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"meta": metaSchema,
				},
			},
		},
		{
			Name:        "cancel_order",
			Description: "Cancel an order. Only possible when order status is 'confirmed' or 'processing' (before it has shipped). Once shipped, the order cannot be canceled.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"meta": metaSchema,
					"id":   map[string]interface{}{"type": "string", "description": "Order ID"},
				},
				"required": []string{"id"},
			},
		},
	}
}

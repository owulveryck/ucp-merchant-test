package main

import "fmt"

// Shared business logic used by both REST and MCP transports.

func buildLineItems(req map[string]interface{}) ([]LineItem, error) {
	rawItems, _ := req["line_items"].([]interface{})
	if len(rawItems) == 0 {
		return nil, fmt.Errorf("line_items is required")
	}

	var items []LineItem
	for i, raw := range rawItems {
		rawMap, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract item ID (support both item.id and product_id)
		itemID := ""
		if itemMap, ok := rawMap["item"].(map[string]interface{}); ok {
			itemID, _ = itemMap["id"].(string)
		}
		if itemID == "" {
			itemID, _ = rawMap["product_id"].(string)
		}
		if itemID == "" {
			return nil, fmt.Errorf("line item %d: missing item.id", i)
		}

		product := findProduct(itemID)
		if product == nil {
			return nil, fmt.Errorf("Product not found: %s", itemID)
		}

		qty := 1
		if q, ok := rawMap["quantity"].(float64); ok {
			qty = int(q)
		}
		if qty < 1 {
			qty = 1
		}

		// Check stock
		if product.Quantity <= 0 {
			return nil, fmt.Errorf("Insufficient stock for product %s", itemID)
		}
		if qty > product.Quantity {
			return nil, fmt.Errorf("Insufficient stock for product %s: requested %d, available %d", itemID, qty, product.Quantity)
		}

		lineTotal := product.Price * qty

		liID := fmt.Sprintf("li_%03d", i+1)
		if existingID, ok := rawMap["id"].(string); ok && existingID != "" {
			liID = existingID
		}

		items = append(items, LineItem{
			ID: liID,
			Item: Item{
				ID:       product.ID,
				Title:    product.Title,
				Price:    product.Price,
				ImageURL: product.ImageURL,
			},
			Quantity: qty,
			Totals: []Total{
				{Type: "subtotal", Amount: lineTotal},
				{Type: "total", Amount: lineTotal},
			},
		})
	}
	return items, nil
}

func calculateTotals(items []LineItem, shippingCost int, discounts *Discounts) []Total {
	subtotal := 0
	for _, li := range items {
		for _, t := range li.Totals {
			if t.Type == "subtotal" {
				subtotal += t.Amount
			}
		}
	}

	total := subtotal

	var totals []Total
	totals = append(totals, Total{
		Type:        "subtotal",
		DisplayText: fmt.Sprintf("$%.2f", float64(subtotal)/100),
		Amount:      subtotal,
	})

	// Apply discounts
	if discounts != nil {
		discountAmount := 0
		for _, d := range discounts.Applied {
			discountAmount += d.Amount
		}
		if discountAmount > 0 {
			total -= discountAmount
			totals = append(totals, Total{
				Type:        "discount",
				DisplayText: fmt.Sprintf("-$%.2f", float64(discountAmount)/100),
				Amount:      discountAmount,
			})
		}
	}

	if shippingCost > 0 {
		total += shippingCost
		totals = append(totals, Total{
			Type:        "fulfillment",
			DisplayText: fmt.Sprintf("$%.2f", float64(shippingCost)/100),
			Amount:      shippingCost,
		})
	}

	totals = append(totals, Total{
		Type:        "total",
		DisplayText: fmt.Sprintf("$%.2f", float64(total)/100),
		Amount:      total,
	})

	return totals
}

func parsePayment(req map[string]interface{}) Payment {
	paymentRaw, ok := req["payment"]
	if !ok || paymentRaw == nil {
		return defaultPayment()
	}

	paymentMap, ok := paymentRaw.(map[string]interface{})
	if !ok {
		return defaultPayment()
	}

	p := &Payment{}

	if sid, ok := paymentMap["selected_instrument_id"].(string); ok {
		p.SelectedInstrumentID = sid
	}

	// Parse instruments
	if instRaw, ok := paymentMap["instruments"].([]interface{}); ok {
		for _, inst := range instRaw {
			if m, ok := inst.(map[string]interface{}); ok {
				p.Instruments = append(p.Instruments, m)
			}
		}
	}
	if p.Instruments == nil {
		p.Instruments = []map[string]interface{}{}
	}

	// Parse handlers
	if hRaw, ok := paymentMap["handlers"].([]interface{}); ok {
		for _, h := range hRaw {
			if m, ok := h.(map[string]interface{}); ok {
				p.Handlers = append(p.Handlers, m)
			}
		}
	}
	if p.Handlers == nil {
		p.Handlers = defaultPaymentHandlers()
	}

	return *p
}

func defaultPayment() Payment {
	return Payment{
		SelectedInstrumentID: "instr_1",
		Instruments:          []map[string]interface{}{},
		Handlers:             defaultPaymentHandlers(),
	}
}

func defaultPaymentHandlers() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"id":                 "google_pay",
			"name":               "google.pay",
			"version":            "2026-01-11",
			"spec":               "https://ucp.dev/specs/payment/google_pay",
			"config_schema":      "https://ucp.dev/schemas/payment/google_pay.json",
			"instrument_schemas": []string{"https://ucp.dev/schemas/payment/google_pay_instrument.json"},
			"config":             map[string]interface{}{},
		},
		{
			"id":                 "mock_payment_handler",
			"name":               "mock_payment_handler",
			"version":            "2026-01-11",
			"spec":               "https://ucp.dev/specs/payment/mock",
			"config_schema":      "https://ucp.dev/schemas/payment/mock.json",
			"instrument_schemas": []string{"https://ucp.dev/schemas/payment/mock_instrument.json"},
			"config":             map[string]interface{}{},
		},
		{
			"id":                 "shop_pay",
			"name":               "com.shopify.shop_pay",
			"version":            "2026-01-11",
			"spec":               "https://ucp.dev/specs/payment/shop_pay",
			"config_schema":      "https://ucp.dev/schemas/payment/shop_pay.json",
			"instrument_schemas": []string{"https://ucp.dev/schemas/payment/shop_pay_instrument.json"},
			"config":             map[string]interface{}{"shop_id": "merchant_1"},
		},
	}
}

func parseBuyer(req map[string]interface{}) *Buyer {
	buyerRaw, ok := req["buyer"]
	if !ok || buyerRaw == nil {
		return nil
	}
	buyerMap, ok := buyerRaw.(map[string]interface{})
	if !ok {
		return nil
	}

	b := &Buyer{}
	if v, ok := buyerMap["first_name"].(string); ok {
		b.FirstName = v
	}
	if v, ok := buyerMap["last_name"].(string); ok {
		b.LastName = v
	}
	if v, ok := buyerMap["fullName"].(string); ok {
		b.FullName = v
	}
	// Support MCP "name" field -> map to FullName
	if v, ok := buyerMap["name"].(string); ok && b.FullName == "" && b.FirstName == "" {
		b.FullName = v
	}
	if v, ok := buyerMap["email"].(string); ok {
		b.Email = v
	}

	// Parse consent
	if consentRaw, ok := buyerMap["consent"].(map[string]interface{}); ok {
		c := &Consent{}
		if v, ok := consentRaw["marketing"].(bool); ok {
			c.Marketing = &v
		}
		if v, ok := consentRaw["analytics"].(bool); ok {
			c.Analytics = &v
		}
		if v, ok := consentRaw["sale_of_data"].(bool); ok {
			c.SaleOfData = &v
		}
		b.Consent = c
	}

	return b
}

func parseFulfillment(req map[string]interface{}, buyer *Buyer, co *Checkout) *Fulfillment {
	fulfillmentRaw, ok := req["fulfillment"]
	if !ok || fulfillmentRaw == nil {
		return nil
	}
	fMap, ok := fulfillmentRaw.(map[string]interface{})
	if !ok {
		return nil
	}

	methodsRaw, _ := fMap["methods"].([]interface{})
	if len(methodsRaw) == 0 {
		return nil
	}

	f := &Fulfillment{}
	for _, mRaw := range methodsRaw {
		mData, ok := mRaw.(map[string]interface{})
		if !ok {
			continue
		}
		method := FulfillmentMethod{}
		if v, ok := mData["id"].(string); ok {
			method.ID = v
		} else {
			method.ID = "method_shipping"
		}
		if v, ok := mData["type"].(string); ok {
			method.Type = v
		}
		// Collect line item IDs
		if co != nil {
			for _, li := range co.LineItems {
				method.LineItemIDs = append(method.LineItemIDs, li.ID)
			}
		}
		if method.LineItemIDs == nil {
			method.LineItemIDs = []string{}
		}

		// Parse destinations
		destsRaw, _ := mData["destinations"].([]interface{})
		if len(destsRaw) > 0 {
			for _, dRaw := range destsRaw {
				dMap, ok := dRaw.(map[string]interface{})
				if !ok {
					continue
				}
				dest := parseDestination(dMap, buyer)
				method.Destinations = append(method.Destinations, dest)
			}
		} else if method.Type == "shipping" {
			// Address injection: look up known customer addresses
			email := ""
			if buyer != nil {
				email = buyer.Email
			} else if co != nil && co.Buyer != nil {
				email = co.Buyer.Email
			}
			if email != "" {
				addresses := findAddressesForEmail(email)
				for _, addr := range addresses {
					method.Destinations = append(method.Destinations, FulfillmentDestination{
						ID:              addr.ID,
						StreetAddress:   addr.StreetAddress,
						AddressLocality: addr.City,
						AddressRegion:   addr.State,
						PostalCode:      addr.PostalCode,
						AddressCountry:  addr.Country,
					})
				}
			}
			if len(method.Destinations) == 0 {
				method.Destinations = nil
			}
		}

		// Selected destination
		if v, ok := mData["selected_destination_id"].(string); ok {
			method.SelectedDestinationID = v

			// Preserve existing destinations if we have them from a previous update
			if len(method.Destinations) == 0 && co != nil && co.Fulfillment != nil && len(co.Fulfillment.Methods) > 0 {
				method.Destinations = co.Fulfillment.Methods[0].Destinations
			}

			// Store the selected destination for order creation
			if co != nil {
				for _, d := range method.Destinations {
					if d.ID == v {
						dest := d
						checkoutDestinations[co.ID] = &dest
						break
					}
				}
			}

			// Generate shipping options based on destination country
			destCountry := ""
			for _, d := range method.Destinations {
				if d.ID == v {
					destCountry = d.AddressCountry
					break
				}
			}
			if destCountry != "" {
				options := generateShippingOptions(destCountry, co)
				groupLineItemIDs := method.LineItemIDs
				if groupLineItemIDs == nil {
					groupLineItemIDs = []string{}
				}
				method.Groups = []FulfillmentGroup{
					{ID: "group_1", LineItemIDs: groupLineItemIDs, Options: options},
				}
			}
		}

		// Parse groups (for selected_option_id)
		if groupsRaw, ok := mData["groups"].([]interface{}); ok && len(groupsRaw) > 0 {
			// Preserve existing options if we have them
			existingOptions := []FulfillmentOption{}
			if co != nil && co.Fulfillment != nil && len(co.Fulfillment.Methods) > 0 {
				existingMethod := co.Fulfillment.Methods[0]
				if len(existingMethod.Groups) > 0 {
					existingOptions = existingMethod.Groups[0].Options
				}
				// Also preserve destinations and selection
				if len(method.Destinations) == 0 {
					method.Destinations = existingMethod.Destinations
				}
				if method.SelectedDestinationID == "" {
					method.SelectedDestinationID = existingMethod.SelectedDestinationID
				}
			}

			for gi, gRaw := range groupsRaw {
				gMap, ok := gRaw.(map[string]interface{})
				if !ok {
					continue
				}
				groupLineItemIDs := method.LineItemIDs
				if groupLineItemIDs == nil {
					groupLineItemIDs = []string{}
				}
				group := FulfillmentGroup{
					ID:          fmt.Sprintf("group_%d", gi+1),
					LineItemIDs: groupLineItemIDs,
				}
				if v, ok := gMap["selected_option_id"].(string); ok {
					group.SelectedOptionID = v
					// Store the selected option title for order expectations
					if co != nil {
						for _, opt := range existingOptions {
							if opt.ID == v {
								checkoutOptionTitles[co.ID] = opt.Title
								break
							}
						}
					}
				}
				group.Options = existingOptions
				method.Groups = append(method.Groups, group)
			}
		}

		f.Methods = append(f.Methods, method)
	}

	return f
}

func parseDestination(dMap map[string]interface{}, buyer *Buyer) FulfillmentDestination {
	dest := FulfillmentDestination{}
	if v, ok := dMap["id"].(string); ok {
		dest.ID = v
	}
	if v, ok := dMap["full_name"].(string); ok {
		dest.FullName = v
	}
	if v, ok := dMap["street_address"].(string); ok {
		dest.StreetAddress = v
	}
	if v, ok := dMap["address_locality"].(string); ok {
		dest.AddressLocality = v
	}
	if v, ok := dMap["address_region"].(string); ok {
		dest.AddressRegion = v
	}
	if v, ok := dMap["postal_code"].(string); ok {
		dest.PostalCode = v
	}
	if v, ok := dMap["address_country"].(string); ok {
		dest.AddressCountry = v
	}

	// If no ID provided, try to match existing or generate one
	if dest.ID == "" {
		email := ""
		if buyer != nil {
			email = buyer.Email
		}
		if email != "" {
			existingAddrs := findAddressesForEmail(email)
			matched := matchExistingAddress(existingAddrs, dest.StreetAddress, dest.AddressLocality, dest.AddressRegion, dest.PostalCode, dest.AddressCountry)
			if matched != nil {
				dest.ID = matched.ID
			} else {
				// Generate new ID and save
				addrSeqMu.Lock()
				addrSeqCounter++
				dest.ID = fmt.Sprintf("addr_dyn_%d", addrSeqCounter)
				addrSeqMu.Unlock()
				saveDynamicAddress(email, CSVAddress{
					ID:            dest.ID,
					StreetAddress: dest.StreetAddress,
					City:          dest.AddressLocality,
					State:         dest.AddressRegion,
					PostalCode:    dest.PostalCode,
					Country:       dest.AddressCountry,
				})
			}
		} else {
			addrSeqMu.Lock()
			addrSeqCounter++
			dest.ID = fmt.Sprintf("addr_dyn_%d", addrSeqCounter)
			addrSeqMu.Unlock()
		}
	}

	return dest
}

func generateShippingOptions(country string, co *Checkout) []FulfillmentOption {
	rates := getShippingRatesForCountry(country)
	var options []FulfillmentOption

	// Check promotions for free shipping
	freeShipping := false
	if co != nil {
		subtotal := 0
		var itemIDs []string
		for _, li := range co.LineItems {
			for _, t := range li.Totals {
				if t.Type == "subtotal" {
					subtotal += t.Amount
				}
			}
			itemIDs = append(itemIDs, li.Item.ID)
		}
		for _, promo := range shopData.Promotions {
			if promo.Type == "free_shipping" {
				if promo.MinSubtotal > 0 && subtotal >= promo.MinSubtotal {
					freeShipping = true
					break
				}
				if len(promo.EligibleItemIDs) > 0 {
					for _, eligible := range promo.EligibleItemIDs {
						for _, itemID := range itemIDs {
							if eligible == itemID {
								freeShipping = true
								break
							}
						}
						if freeShipping {
							break
						}
					}
				}
			}
			if freeShipping {
				break
			}
		}
	}

	for _, rate := range rates {
		price := rate.Price
		title := rate.Title
		if freeShipping && rate.ServiceLevel == "standard" {
			price = 0
			title = "Free Standard Shipping"
		}
		options = append(options, FulfillmentOption{
			ID:    rate.ID,
			Title: title,
			Totals: []Total{
				{Type: "fulfillment", Amount: price},
				{Type: "total", Amount: price},
			},
		})
	}
	return options
}

func getCurrentShippingCost(co *Checkout) int {
	if co.Fulfillment == nil {
		return 0
	}
	for _, m := range co.Fulfillment.Methods {
		for _, g := range m.Groups {
			if g.SelectedOptionID != "" {
				for _, opt := range g.Options {
					if opt.ID == g.SelectedOptionID {
						for _, t := range opt.Totals {
							if t.Type == "total" {
								return t.Amount
							}
						}
					}
				}
			}
		}
	}
	return 0
}

func isFulfillmentComplete(co *Checkout) bool {
	if co.Fulfillment == nil {
		return false
	}
	for _, m := range co.Fulfillment.Methods {
		if m.SelectedDestinationID == "" {
			return false
		}
		hasOption := false
		for _, g := range m.Groups {
			if g.SelectedOptionID != "" {
				hasOption = true
				break
			}
		}
		if !hasOption {
			return false
		}
	}
	return len(co.Fulfillment.Methods) > 0
}

func applyDiscounts(discountsRaw interface{}, lineItems []LineItem) *Discounts {
	dMap, ok := discountsRaw.(map[string]interface{})
	if !ok {
		return nil
	}

	codesRaw, _ := dMap["codes"].([]interface{})
	if len(codesRaw) == 0 {
		return nil
	}

	// Calculate subtotal
	subtotal := 0
	for _, li := range lineItems {
		for _, t := range li.Totals {
			if t.Type == "subtotal" {
				subtotal += t.Amount
			}
		}
	}

	result := &Discounts{}
	for _, cRaw := range codesRaw {
		code, _ := cRaw.(string)
		if code == "" {
			continue
		}
		result.Codes = append(result.Codes, code)

		discount := findDiscountByCode(code)
		if discount == nil {
			continue // Unknown codes are silently ignored
		}

		var amount int
		switch discount.Type {
		case "percentage":
			amount = subtotal * discount.Value / 100
			subtotal -= amount // Apply sequentially for multiple discounts
		case "fixed_amount":
			amount = discount.Value
			subtotal -= amount
		}

		result.Applied = append(result.Applied, AppliedDiscount{
			Code:   discount.Code,
			Title:  discount.Description,
			Amount: amount,
		})
	}

	return result
}

func stringOr(m map[string]interface{}, key, def string) string {
	if v, ok := m[key].(string); ok && v != "" {
		return v
	}
	return def
}

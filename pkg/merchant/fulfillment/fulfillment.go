package fulfillment

import (
	"fmt"
	"strings"

	"github.com/owulveryck/ucp-merchant-test/pkg/model"
	"github.com/owulveryck/ucp-merchant-test/pkg/ucp"
)

// Address represents a known shipping destination from the merchant's address
// book, typically loaded from customer data (e.g., addresses.csv). Addresses
// are keyed by customer email and used during the UCP fulfillment flow to
// pre-populate destinations when a buyer's identity is linked.
type Address struct {
	ID            string
	CustomerID    string
	StreetAddress string
	City          string
	State         string
	PostalCode    string
	Country       ucp.Country
}

// ShippingRate represents a fulfillment cost for a specific country and service
// level (e.g., "standard", "express"). Rates are loaded from the merchant's
// shipping configuration (e.g., shipping_rates.csv) and used by
// [GenerateShippingOptions] to build the UCP fulfillment options presented to
// the buyer after destination selection.
type ShippingRate struct {
	ID           string
	CountryCode  string
	ServiceLevel string
	Price        int
	Title        string
}

// Promotion represents an automatic discount rule applied by the business
// during fulfillment option generation. The most common type is "free_shipping",
// which waives the standard shipping fee when the order meets certain criteria.
//
// Free shipping promotions are triggered by either:
//   - MinSubtotal: the order subtotal meets or exceeds this threshold (e.g., $100)
//   - EligibleItemIDs: the order contains at least one of the specified products
//     (e.g., "bouquet_roses" always qualifies for free shipping)
type Promotion struct {
	ID              string
	Type            string
	MinSubtotal     int
	EligibleItemIDs []string
	Description     string
}

// FulfillmentDataSource provides access to the merchant's address book, shipping
// rates, and promotion rules. This interface abstracts the data layer so the
// fulfillment logic can be tested independently of the data source (CSV files,
// JSON database, etc.).
//
// Implementations must support:
//   - Address lookup by buyer email (for destination pre-population)
//   - Dynamic address persistence (for new addresses submitted during checkout)
//   - Country-specific shipping rate retrieval
//   - Promotion rule retrieval for free shipping evaluation
type FulfillmentDataSource interface {
	// FindAddressesForEmail returns all known shipping addresses for the given
	// buyer email. Used to pre-populate destinations during checkout.
	FindAddressesForEmail(email string) []Address
	// SaveDynamicAddress persists a new address submitted during checkout for
	// the given buyer email. Returns the assigned address ID.
	SaveDynamicAddress(email string, addr Address) string
	// GetShippingRatesForCountry returns available shipping rates (service
	// levels and prices) for the given ISO country code.
	GetShippingRatesForCountry(country ucp.Country) []ShippingRate
	// GetPromotions returns all active promotion rules (e.g., free shipping
	// thresholds and eligible item lists).
	GetPromotions() []Promotion
}

// MatchExistingAddress performs case-insensitive matching of a submitted address
// against a list of known addresses. This supports the UCP fulfillment flow where
// a buyer submits a new destination without an ID — the server checks whether
// the address already exists in the buyer's address book to avoid creating
// duplicates.
//
// Returns the first matching address, or nil if no match is found. All string
// comparisons use case-insensitive folding per UCP conventions.
func MatchExistingAddress(addrs []Address, street, locality, region, postal string, country ucp.Country) *Address {
	for i := range addrs {
		a := &addrs[i]
		if strings.EqualFold(a.StreetAddress, street) &&
			strings.EqualFold(a.City, locality) &&
			strings.EqualFold(a.State, region) &&
			strings.EqualFold(a.PostalCode, postal) &&
			strings.EqualFold(string(a.Country), string(country)) {
			return a
		}
	}
	return nil
}

// ParseFulfillment processes the fulfillment section of a UCP checkout create or
// update request, implementing the progressive fulfillment flow defined by the
// Fulfillment Extension (dev.ucp.shopping.fulfillment).
//
// The function handles all stages of the fulfillment lifecycle:
//
//  1. Method initialization: creates fulfillment methods from the request,
//     linking all checkout line items to each method via LineItemIDs.
//
//  2. Destination population: if the request includes destinations, they are
//     parsed via [ParseDestination]. If no destinations are provided and the
//     method type is "shipping", the function auto-populates destinations from
//     the buyer's known addresses (looked up by email via the [FulfillmentDataSource]).
//
//  3. Destination selection: when SelectedDestinationID is set, the selected
//     destination is recorded in checkoutDestinations (for order creation) and
//     shipping options are generated for the destination's country via
//     [GenerateShippingOptions].
//
//  4. Option selection: when Groups with SelectedOptionID are provided, the
//     selected option title is recorded in checkoutOptionTitles (for order
//     fulfillment expectations). Existing options from the checkout's prior state
//     are preserved so the response includes the full option list.
//
// Returns nil when req is nil or contains no methods, indicating the checkout
// has no fulfillment requirements (appropriate for digital goods).
//
// The checkoutDestinations and checkoutOptionTitles maps accumulate state across
// multiple checkout updates, keyed by checkout ID. The addrSeqCounter and
// addrSeqMu provide thread-safe generation of dynamic address IDs.
func ParseFulfillment(
	req *model.FulfillmentRequest,
	buyer *model.Buyer,
	co *model.Checkout,
	ds FulfillmentDataSource,
	checkoutDestinations map[string]*model.FulfillmentDestination,
	checkoutOptionTitles map[string]string,
	addrSeqCounter *int,
	addrSeqMu interface {
		Lock()
		Unlock()
	},
) *model.Fulfillment {
	if req == nil || len(req.Methods) == 0 {
		return nil
	}

	f := &model.Fulfillment{}
	for _, mData := range req.Methods {
		method := model.FulfillmentMethod{}
		if mData.ID != "" {
			method.ID = mData.ID
		} else {
			method.ID = "method_shipping"
		}
		method.Type = mData.Type
		if co != nil {
			for _, li := range co.LineItems {
				method.LineItemIDs = append(method.LineItemIDs, li.ID)
			}
		}
		if method.LineItemIDs == nil {
			method.LineItemIDs = []string{}
		}

		if len(mData.Destinations) > 0 {
			for _, dReq := range mData.Destinations {
				dest := ParseDestination(dReq, buyer, ds, addrSeqCounter, addrSeqMu)
				method.Destinations = append(method.Destinations, dest)
			}
		} else if method.Type == "shipping" {
			email := ""
			if buyer != nil {
				email = buyer.Email
			} else if co != nil && co.Buyer != nil {
				email = co.Buyer.Email
			}
			if email != "" {
				addresses := ds.FindAddressesForEmail(email)
				for _, addr := range addresses {
					method.Destinations = append(method.Destinations, model.FulfillmentDestination{
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

		if mData.SelectedDestinationID != "" {
			method.SelectedDestinationID = mData.SelectedDestinationID

			if len(method.Destinations) == 0 && co != nil && co.Fulfillment != nil && len(co.Fulfillment.Methods) > 0 {
				method.Destinations = co.Fulfillment.Methods[0].Destinations
			}

			if co != nil {
				for _, d := range method.Destinations {
					if d.ID == mData.SelectedDestinationID {
						dest := d
						checkoutDestinations[co.ID] = &dest
						break
					}
				}
			}

			var destCountry ucp.Country
			for _, d := range method.Destinations {
				if d.ID == mData.SelectedDestinationID {
					destCountry = d.AddressCountry
					break
				}
			}
			if destCountry != "" {
				options := GenerateShippingOptions(destCountry, co, ds)
				groupLineItemIDs := method.LineItemIDs
				if groupLineItemIDs == nil {
					groupLineItemIDs = []string{}
				}
				method.Groups = []model.FulfillmentGroup{
					{ID: "group_1", LineItemIDs: groupLineItemIDs, Options: options},
				}
			}
		}

		if len(mData.Groups) > 0 {
			existingOptions := []model.FulfillmentOption{}
			if co != nil && co.Fulfillment != nil && len(co.Fulfillment.Methods) > 0 {
				existingMethod := co.Fulfillment.Methods[0]
				if len(existingMethod.Groups) > 0 {
					existingOptions = existingMethod.Groups[0].Options
				}
				if len(method.Destinations) == 0 {
					method.Destinations = existingMethod.Destinations
				}
				if method.SelectedDestinationID == "" {
					method.SelectedDestinationID = existingMethod.SelectedDestinationID
				}
			}

			for gi, gReq := range mData.Groups {
				groupLineItemIDs := method.LineItemIDs
				if groupLineItemIDs == nil {
					groupLineItemIDs = []string{}
				}
				group := model.FulfillmentGroup{
					ID:          fmt.Sprintf("group_%d", gi+1),
					LineItemIDs: groupLineItemIDs,
				}
				if gReq.SelectedOptionID != "" {
					group.SelectedOptionID = gReq.SelectedOptionID
					if co != nil {
						for _, opt := range existingOptions {
							if opt.ID == gReq.SelectedOptionID {
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

// ParseDestination converts a [model.FulfillmentDestinationRequest] into a
// [model.FulfillmentDestination] suitable for inclusion in a UCP checkout
// response.
//
// When the request includes an ID, the destination is used as-is (it references
// a known address). When no ID is provided, the function performs address
// deduplication:
//
//  1. If a buyer email is available, existing addresses for that email are
//     searched via [MatchExistingAddress] (case-insensitive field comparison).
//  2. If a match is found, the existing address ID is reused.
//  3. If no match is found, a new dynamic address ID is generated
//     ("addr_dyn_1", "addr_dyn_2", etc.) and the address is saved via the
//     [FulfillmentDataSource] for future lookups.
//  4. If no buyer email is available, a dynamic ID is assigned without saving.
//
// This supports the UCP flow where buyers submit new shipping addresses during
// checkout — the business ensures each unique address gets a stable ID that
// can be referenced in subsequent operations (destination selection, order
// fulfillment expectations).
func ParseDestination(
	dReq model.FulfillmentDestinationRequest,
	buyer *model.Buyer,
	ds FulfillmentDataSource,
	addrSeqCounter *int,
	addrSeqMu interface {
		Lock()
		Unlock()
	},
) model.FulfillmentDestination {
	dest := model.FulfillmentDestination{
		ID:              dReq.ID,
		FullName:        dReq.FullName,
		StreetAddress:   dReq.StreetAddress,
		AddressLocality: dReq.AddressLocality,
		AddressRegion:   dReq.AddressRegion,
		PostalCode:      dReq.PostalCode,
		AddressCountry:  dReq.AddressCountry,
	}

	if dest.ID == "" {
		email := ""
		if buyer != nil {
			email = buyer.Email
		}
		if email != "" {
			existingAddrs := ds.FindAddressesForEmail(email)
			matched := MatchExistingAddress(existingAddrs, dest.StreetAddress, dest.AddressLocality, dest.AddressRegion, dest.PostalCode, dest.AddressCountry)
			if matched != nil {
				dest.ID = matched.ID
			} else {
				addrSeqMu.Lock()
				*addrSeqCounter++
				dest.ID = fmt.Sprintf("addr_dyn_%d", *addrSeqCounter)
				addrSeqMu.Unlock()
				ds.SaveDynamicAddress(email, Address{
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
			*addrSeqCounter++
			dest.ID = fmt.Sprintf("addr_dyn_%d", *addrSeqCounter)
			addrSeqMu.Unlock()
		}
	}

	return dest
}

// GenerateShippingOptions creates the UCP fulfillment options array for a
// destination country. Options represent selectable shipping speeds with costs,
// displayed to the buyer after they select a destination address.
//
// The function queries the [FulfillmentDataSource] for country-specific shipping
// rates and evaluates free shipping promotions against the current checkout:
//
//   - Subtotal threshold: if the order subtotal meets or exceeds a promotion's
//     MinSubtotal, standard shipping is free (e.g., orders >= $100)
//   - Eligible items: if the order contains any product listed in the
//     promotion's EligibleItemIDs, standard shipping is free
//     (e.g., "bouquet_roses" always ships free)
//
// When free shipping applies, the standard rate's price is set to 0 and its
// title is changed to "Free Standard Shipping". Other service levels (express,
// next-day) are unaffected.
//
// Each returned [model.FulfillmentOption] includes:
//   - ID and Title for display
//   - Totals with "fulfillment" and "total" amounts (per UCP, never "shipping")
//
// Per UCP rendering guidelines, title + totals is sufficient for a platform to
// render any fulfillment option without understanding the specific method type.
func GenerateShippingOptions(country ucp.Country, co *model.Checkout, ds FulfillmentDataSource) []model.FulfillmentOption {
	rates := ds.GetShippingRatesForCountry(country)
	var options []model.FulfillmentOption

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
		for _, promo := range ds.GetPromotions() {
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
		options = append(options, model.FulfillmentOption{
			ID:    rate.ID,
			Title: title,
			Totals: []model.Total{
				{Type: "fulfillment", Amount: price},
				{Type: "total", Amount: price},
			},
		})
	}
	return options
}

// GetCurrentShippingCost extracts the total amount from the currently selected
// shipping option in a checkout's fulfillment structure. This value is used as
// the shippingCost parameter in [pricing.CalculateTotals] to include fulfillment
// cost in the checkout-level totals.
//
// The function traverses the fulfillment hierarchy: methods → groups →
// selected option → totals, looking for the "total" total type on the selected
// option. Returns 0 when no fulfillment is configured or no option is selected,
// which is appropriate for checkouts without physical delivery requirements.
func GetCurrentShippingCost(co *model.Checkout) int {
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

// IsFulfillmentComplete checks whether all required fulfillment selections have
// been made, which is a precondition for completing a UCP checkout session.
//
// A checkout's fulfillment is complete when every method has:
//   - A selected destination (SelectedDestinationID is non-empty)
//   - At least one group with a selected shipping option (SelectedOptionID is non-empty)
//
// The checkout can only transition to the "ready_for_complete" status (and
// subsequently be completed via the complete endpoint) when this function
// returns true. Returns false when fulfillment is nil (digital goods checkouts
// should not call this function — they do not require fulfillment).
func IsFulfillmentComplete(co *model.Checkout) bool {
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

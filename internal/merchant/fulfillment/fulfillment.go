package fulfillment

import (
	"fmt"
	"strings"

	"github.com/owulveryck/ucp-merchant-test/internal/model"
)

// Address represents a shipping destination.
type Address struct {
	ID            string
	CustomerID    string
	StreetAddress string
	City          string
	State         string
	PostalCode    string
	Country       string
}

// ShippingRate represents a fulfillment cost keyed by country code and service level.
type ShippingRate struct {
	ID           string
	CountryCode  string
	ServiceLevel string
	Price        int
	Title        string
}

// Promotion represents an automatic discount rule such as free shipping.
type Promotion struct {
	ID              string
	Type            string
	MinSubtotal     int
	EligibleItemIDs []string
	Description     string
}

// FulfillmentDataSource provides access to address, shipping, and promotion data.
type FulfillmentDataSource interface {
	FindAddressesForEmail(email string) []Address
	SaveDynamicAddress(email string, addr Address) string
	GetShippingRatesForCountry(country string) []ShippingRate
	GetPromotions() []Promotion
}

// MatchExistingAddress checks if a submitted address matches an existing one.
func MatchExistingAddress(addrs []Address, street, locality, region, postal, country string) *Address {
	for i := range addrs {
		a := &addrs[i]
		if strings.EqualFold(a.StreetAddress, street) &&
			strings.EqualFold(a.City, locality) &&
			strings.EqualFold(a.State, region) &&
			strings.EqualFold(a.PostalCode, postal) &&
			strings.EqualFold(a.Country, country) {
			return a
		}
	}
	return nil
}

// ParseFulfillment parses fulfillment data from a typed request.
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

			destCountry := ""
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

// ParseDestination parses a destination from a typed request.
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

// GenerateShippingOptions generates shipping options based on country and checkout.
func GenerateShippingOptions(country string, co *model.Checkout, ds FulfillmentDataSource) []model.FulfillmentOption {
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

// GetCurrentShippingCost extracts the selected shipping cost from a checkout.
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

// IsFulfillmentComplete checks if fulfillment has all required selections.
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

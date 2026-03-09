package main

import (
	"github.com/owulveryck/ucp-merchant-test/internal/merchant/discount"
	mfulfillment "github.com/owulveryck/ucp-merchant-test/internal/merchant/fulfillment"
	mpayment "github.com/owulveryck/ucp-merchant-test/internal/merchant/payment"
	"github.com/owulveryck/ucp-merchant-test/internal/merchant/pricing"
)

func buildLineItems(req map[string]interface{}) ([]LineItem, error) {
	return pricing.BuildLineItems(req, catalogInstance)
}

func calculateTotals(items []LineItem, shippingCost int, discounts *Discounts) []Total {
	return pricing.CalculateTotals(items, shippingCost, discounts)
}

func parsePayment(req map[string]interface{}) Payment {
	return mpayment.ParsePayment(req)
}

func defaultPayment() Payment {
	return mpayment.DefaultPayment()
}

func defaultPaymentHandlers() []map[string]interface{} {
	return mpayment.DefaultPaymentHandlers()
}

func parseBuyer(req map[string]interface{}) *Buyer {
	return mpayment.ParseBuyer(req)
}

func parseFulfillment(req map[string]interface{}, buyer *Buyer, co *Checkout) *Fulfillment {
	return mfulfillment.ParseFulfillment(req, buyer, co, shopData, checkoutDestinations, checkoutOptionTitles, &addrSeqCounter, &addrSeqMu)
}

func parseDestination(dMap map[string]interface{}, buyer *Buyer) FulfillmentDestination {
	return mfulfillment.ParseDestination(dMap, buyer, shopData, &addrSeqCounter, &addrSeqMu)
}

func generateShippingOptions(country string, co *Checkout) []FulfillmentOption {
	return mfulfillment.GenerateShippingOptions(country, co, shopData)
}

func getCurrentShippingCost(co *Checkout) int {
	return mfulfillment.GetCurrentShippingCost(co)
}

func isFulfillmentComplete(co *Checkout) bool {
	return mfulfillment.IsFulfillmentComplete(co)
}

func applyDiscounts(discountsRaw interface{}, lineItems []LineItem) *Discounts {
	return discount.ApplyDiscounts(discountsRaw, lineItems, shopData)
}

func stringOr(m map[string]interface{}, key, def string) string {
	if v, ok := m[key].(string); ok && v != "" {
		return v
	}
	return def
}

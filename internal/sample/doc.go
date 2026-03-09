// Package sample loads and manages the test dataset for the UCP conformance test
// suite from CSV files.
//
// The Universal Commerce Protocol defines a conformance test suite that validates
// merchant implementations against a standardized dataset. This package loads
// CSV and JSON files from the test data directory (e.g., flower_shop) into an
// in-memory DataSource that the merchant server queries at runtime.
//
// The DataSource implements the domain interfaces defined in the merchant packages
// (discount.DiscountLookup, fulfillment.FulfillmentDataSource) by converting
// internal CSV types to the domain types those packages define.
package sample

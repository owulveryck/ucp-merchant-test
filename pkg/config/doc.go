// Package config provides the runtime configuration for a UCP Shopping Service
// merchant server.
//
// In the Universal Commerce Protocol, a Business (merchant) exposes commerce
// capabilities through a discoverable profile at /.well-known/ucp. The Config
// struct holds the parameters needed to run the business server: network binding
// (port, TLS), authentication secrets for simulation testing, the merchant
// display name, and the path to test data files.
//
// The Scheme method returns "https" or "http" based on TLS configuration, which
// is used to construct absolute URLs required by UCP (e.g., continue_url for
// checkout handoff, OAuth2 authorization server metadata issuer).
package config

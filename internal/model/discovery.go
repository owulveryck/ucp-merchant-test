package model

import "github.com/owulveryck/ucp-merchant-test/internal/ucp"

// UCPDiscovery is the response body for /.well-known/ucp.
type UCPDiscovery struct {
	UCP     UCPDiscoveryProfile `json:"ucp"`
	Payment UCPPaymentProfile   `json:"payment"`
}

// UCPDiscoveryProfile describes the UCP services and capabilities.
type UCPDiscoveryProfile struct {
	Version      string                                 `json:"version"`
	Services     map[ucp.UCPService][]UCPServiceBinding `json:"services"`
	Capabilities []UCPCapabilityEntry                   `json:"capabilities"`
}

// UCPServiceBinding describes a single transport binding for a UCP service.
// Multiple bindings (REST, MCP, etc.) are grouped in an array per service.
type UCPServiceBinding struct {
	Version   string `json:"version"`
	Transport string `json:"transport"`
	Endpoint  string `json:"endpoint,omitempty"`
	Spec      string `json:"spec,omitempty"`
	Schema    string `json:"schema,omitempty"`
}

// UCPCapabilityEntry describes a UCP capability.
type UCPCapabilityEntry struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Spec    string `json:"spec"`
	Schema  string `json:"schema"`
}

// UCPPaymentProfile describes available payment handlers.
// Handlers remains []map[string]any because payment handler configs are opaque per UCP spec.
type UCPPaymentProfile struct {
	Handlers []map[string]any `json:"handlers"`
}

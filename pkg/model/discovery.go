package model

import "github.com/owulveryck/ucp-merchant-test/pkg/ucp"

// UCPDiscovery is the response body for /.well-known/ucp.
type UCPDiscovery struct {
	UCP     UCPDiscoveryProfile `json:"ucp"`
	Payment UCPPaymentProfile   `json:"payment"`
}

// UCPDiscoveryProfile describes the UCP services and capabilities.
type UCPDiscoveryProfile struct {
	Version      string                             `json:"version"`
	Services     map[ucp.UCPService]UCPServiceEntry `json:"services"`
	Capabilities []UCPCapabilityEntry               `json:"capabilities"`
}

// UCPServiceEntry describes a UCP service with transport-specific bindings.
type UCPServiceEntry struct {
	Version string               `json:"version"`
	Spec    string               `json:"spec,omitempty"`
	Rest    *UCPTransportBinding `json:"rest,omitempty"`
	MCP     *UCPTransportBinding `json:"mcp,omitempty"`
	A2A     *UCPTransportBinding `json:"a2a,omitempty"`
}

// UCPTransportBinding describes a transport endpoint for a UCP service.
type UCPTransportBinding struct {
	Schema   string `json:"schema,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
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

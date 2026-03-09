package model

// UCPDiscovery is the response body for /.well-known/ucp.
type UCPDiscovery struct {
	UCP     UCPDiscoveryProfile `json:"ucp"`
	Payment UCPPaymentProfile   `json:"payment"`
}

// UCPDiscoveryProfile describes the UCP services and capabilities.
type UCPDiscoveryProfile struct {
	Version      string                     `json:"version"`
	Services     map[string]UCPServiceEntry `json:"services"`
	Capabilities []UCPCapabilityEntry       `json:"capabilities"`
}

// UCPServiceEntry describes a single UCP service endpoint.
type UCPServiceEntry struct {
	Version string         `json:"version"`
	Spec    string         `json:"spec"`
	REST    *UCPRESTConfig `json:"rest,omitempty"`
}

// UCPRESTConfig holds REST endpoint details for a UCP service.
type UCPRESTConfig struct {
	Endpoint string `json:"endpoint"`
	Schema   string `json:"schema"`
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

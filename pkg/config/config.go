package config

// Config holds server configuration.
type Config struct {
	ListenPort       int
	TLSEnabled       bool
	SimulationSecret string
	MerchantName     string
	DataDir          string
}

// Scheme returns "https" if TLS is enabled, otherwise "http".
func (c *Config) Scheme() string {
	if c.TLSEnabled {
		return "https"
	}
	return "http"
}

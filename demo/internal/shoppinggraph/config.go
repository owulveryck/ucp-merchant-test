package shoppinggraph

// Config is the shopping graph configuration loaded from YAML.
type Config struct {
	Merchants []MerchantConfig `yaml:"merchants"`
}

// MerchantConfig is the configuration for a single merchant.
type MerchantConfig struct {
	ID            string   `yaml:"id"`
	Name          string   `yaml:"name"`
	Endpoint      string   `yaml:"endpoint"`
	Score         int      `yaml:"score"`
	DiscountHints []string `yaml:"discount_hints"`
}

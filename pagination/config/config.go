// Package config provides configuration for pagination functionality.
package config

// Config holds pagination configuration settings
type Config struct {
	// DefaultLimit is the default number of records per page when not specified
	DefaultLimit int `json:"default_limit" yaml:"default_limit"`

	// MaxLimit is the maximum allowed limit per page
	MaxLimit int `json:"max_limit" yaml:"max_limit"`

	// DefaultSortField is the default field to sort by when not specified
	DefaultSortField string `json:"default_sort_field" yaml:"default_sort_field"`

	// DefaultSortOrder is the default sort order when not specified (asc|desc)
	DefaultSortOrder string `json:"default_sort_order" yaml:"default_sort_order"`

	// AllowedSortOrders contains valid sort order values
	AllowedSortOrders []string `json:"allowed_sort_orders" yaml:"allowed_sort_orders"`

	// ValidationEnabled enables parameter validation
	ValidationEnabled bool `json:"validation_enabled" yaml:"validation_enabled"`

	// StrictMode enables strict validation (fails on unknown parameters)
	StrictMode bool `json:"strict_mode" yaml:"strict_mode"`
}

// NewDefaultConfig creates a configuration with sensible defaults
func NewDefaultConfig() *Config {
	return &Config{
		DefaultLimit:      50,
		MaxLimit:          150,
		DefaultSortField:  "id",
		DefaultSortOrder:  "asc",
		AllowedSortOrders: []string{"asc", "desc", "ASC", "DESC"},
		ValidationEnabled: true,
		StrictMode:        false,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.DefaultLimit <= 0 {
		c.DefaultLimit = 50
	}

	if c.MaxLimit <= 0 {
		c.MaxLimit = 150
	}

	if c.DefaultLimit > c.MaxLimit {
		c.DefaultLimit = c.MaxLimit
	}

	if c.DefaultSortField == "" {
		c.DefaultSortField = "id"
	}

	if c.DefaultSortOrder == "" {
		c.DefaultSortOrder = "asc"
	}

	if len(c.AllowedSortOrders) == 0 {
		c.AllowedSortOrders = []string{"asc", "desc", "ASC", "DESC"}
	}

	return nil
}

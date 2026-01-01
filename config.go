package weexgo

import "time"

// Config represents the SDK configuration
type Config struct {
	APIKey     string
	SecretKey  string
	Passphrase string
	BaseURL    string
	Timeout    time.Duration
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		BaseURL: "https://api-contract.weex.com",
		Timeout: 30 * time.Second,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.APIKey == "" {
		return &ConfigError{Field: "APIKey", Message: "API key is required"}
	}
	if c.SecretKey == "" {
		return &ConfigError{Field: "SecretKey", Message: "Secret key is required"}
	}
	if c.Passphrase == "" {
		return &ConfigError{Field: "Passphrase", Message: "Passphrase is required"}
	}
	return nil
}

// ConfigError represents a configuration error
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}


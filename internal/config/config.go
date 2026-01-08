package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config holds application configuration
type Config struct {
	Server   ServerConfig   `toml:"server"`
	Providers []Provider    `toml:"providers"`
	Mappings  ModelMappings `toml:"mappings"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Host         string `toml:"host"`
	Port         int    `toml:"port"`
	ReadTimeout  int    `toml:"read_timeout"`
	WriteTimeout int    `toml:"write_timeout"`
}

// Provider represents an LLM provider configuration
type Provider struct {
	Name         string   `toml:"name"`
	Type         string   `toml:"type"`
	BaseURL      string   `toml:"api_base_url"`
	APIKey       string   `toml:"api_key"`
	Models       []string `toml:"models"`
	UseVertexAuth bool     `toml:"use_vertex_auth,omitempty"`
	VertexProject string   `toml:"vertex_project,omitempty"`
	VertexLocation string  `toml:"vertex_location,omitempty"`

	// Runtime fields (not in TOML)
	ParsedAPIKey   string
	IsBypass      bool
}

// ModelMappings holds model alias mappings
type ModelMappings map[string]string

// Load loads configuration from TOML file
func Load() (*Config, error) {
	configPath := getConfigPath()
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{}
	if err := toml.Unmarshal(configFile, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	setDefaults(cfg)

	// Parse API keys
	if err := cfg.ParseAPIKeys(); err != nil {
		return nil, fmt.Errorf("failed to parse API keys: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// ParseAPIKeys parses API keys for all providers
func (c *Config) ParseAPIKeys() error {
	for i := range c.Providers {
		key, bypass := parseAPIKey(c.Providers[i].APIKey)
		c.Providers[i].ParsedAPIKey = key
		c.Providers[i].IsBypass = bypass
	}
	return nil
}

// parseAPIKey parses an API key configuration
func parseAPIKey(apiKey string) (string, bool) {
	// Check for bypass/forward
	if apiKey == "bypass" || apiKey == "forward" {
		return "", true
	}

	// Check for environment variable
	if strings.HasPrefix(apiKey, "env:") {
		envKey := strings.TrimPrefix(apiKey, "env:")
		value := os.Getenv(envKey)
		return value, false
	}

	// Direct value
	return apiKey, false
}

// getConfigPath returns the path to the configuration file
func getConfigPath() string {
	// 1. Check CONFIG_PATH environment variable
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		return path
	}

	// 2. Check config.toml in current directory
	if _, err := os.Stat("config.toml"); err == nil {
		return "config.toml"
	}

	// 3. Check .llm-to-anthropic.toml in home directory
	home, err := os.UserHomeDir()
	if err == nil {
		configPath := filepath.Join(home, ".llm-to-anthropic.toml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	// Default to config.toml
	return "config.toml"
}

// setDefaults sets default values for configuration
func setDefaults(cfg *Config) {
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8082
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 120
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 120
	}

	if cfg.Mappings == nil {
		cfg.Mappings = make(ModelMappings)
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate server configuration
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	// Validate providers
	providerNames := make(map[string]bool)
	for i, provider := range c.Providers {
		if provider.Name == "" {
			return fmt.Errorf("provider %d: name is required", i)
		}
		if providerNames[provider.Name] {
			return fmt.Errorf("duplicate provider name: %s", provider.Name)
		}
		providerNames[provider.Name] = true

		if provider.Type == "" {
			return fmt.Errorf("provider %s: type is required", provider.Name)
		}

		if provider.BaseURL == "" {
			return fmt.Errorf("provider %s: api_base_url is required", provider.Name)
		}

		// Validate vertex auth configuration
		if provider.UseVertexAuth {
			if provider.VertexProject == "" {
				return fmt.Errorf("provider %s: vertex_project is required when use_vertex_auth is true", provider.Name)
			}
			if provider.VertexLocation == "" {
				return fmt.Errorf("provider %s: vertex_location is required when use_vertex_auth is true", provider.Name)
			}
		}
	}

	return nil
}

// GetProviderByName returns a provider by name
func (c *Config) GetProviderByName(name string) (*Provider, bool) {
	for i := range c.Providers {
		if c.Providers[i].Name == name {
			return &c.Providers[i], true
		}
	}
	return nil, false
}

// ParseModelMapping parses a model mapping string
// Returns provider name and model name
// Example: "openai/gpt-4.1-mini" → ("openai", "gpt-4.1-mini")
// Example: "ollama/custom/model:free" → ("ollama", "custom/model:free")
func ParseModelMapping(mapping string) (string, string) {
	parts := strings.SplitN(mapping, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", mapping
}

// GetHost returns the server host
func (c *Config) GetHost() string {
	return c.Server.Host
}

// GetPort returns the server port
func (c *Config) GetPort() int {
	return c.Server.Port
}

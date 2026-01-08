package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/kelseyhightower/envconfig"
)

// Config holds application configuration
// API keys come from environment variables, other settings from TOML file
type Config struct {
	// TOML Configuration
	Server   ServerConfig   `toml:"server"`
	General  GeneralConfig  `toml:"general"`
	Models   ModelsConfig   `toml:"models"`
	Google   GoogleConfig   `toml:"google"`
	Mappings map[string]string `toml:"mappings"`

	// Environment Variables (API Keys only)
	AnthropicAPIKey string `envconfig:"ANTHROPIC_API_KEY" required:"false"`
	OpenAIKey      string `envconfig:"OPENAI_API_KEY" required:"false"`
	GeminiAPIKey   string `envconfig:"GEMINI_API_KEY" required:"false"`
}

// ServerConfig represents server configuration from TOML
type ServerConfig struct {
	Host         string `toml:"host"`
	Port         int    `toml:"port"`
	ReadTimeout  int    `toml:"read_timeout"`
	WriteTimeout int    `toml:"write_timeout"`
}

// GeneralConfig represents general configuration from TOML
type GeneralConfig struct {
	PreferredProvider Provider `toml:"preferred_provider"`
	Verbose          bool     `toml:"verbose"`
}

// ModelsConfig represents model configuration from TOML
type ModelsConfig struct {
	SmallModel string `toml:"small_model"`
	BigModel   string `toml:"big_model"`
}

// GoogleConfig represents Google-specific configuration from TOML
type GoogleConfig struct {
	UseVertexAuth  bool   `toml:"use_vertex_auth"`
	VertexProject  string `toml:"vertex_project"`
	VertexLocation string `toml:"vertex_location"`
}

// Provider represents LLM provider
type Provider string

const (
	ProviderOpenAI    Provider = "openai"
	ProviderGoogle    Provider = "google"
	ProviderAnthropic Provider = "anthropic"
)

// Load loads configuration from TOML file and environment variables
// It first tries to load config.toml, then loads env vars
func Load() (*Config, error) {
	// Load TOML configuration
	configPath := getConfigPath()
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{}
	if err := toml.Unmarshal(configFile, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults if not specified
	setDefaults(cfg)

	// Load environment variables (API keys only)
	if err := envconfig.Process("", cfg); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// getConfigPath returns the path to the configuration file
// Searches in this order:
// 1. CONFIG_PATH environment variable
// 2. config.toml in current directory
// 3. .llm-to-anthropic.toml in home directory
func getConfigPath() string {
	// Check if CONFIG_PATH is set
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		return path
	}

	// Try config.toml in current directory
	if _, err := os.Stat("config.toml"); err == nil {
		return "config.toml"
	}

	// Try .llm-to-anthropic.toml in home directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		configPath := filepath.Join(homeDir, ".llm-to-anthropic.toml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	// Fallback to config.toml
	return "config.toml"
}

// setDefaults sets default values for configuration
func setDefaults(cfg *Config) {
	// Server defaults
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

	// General defaults
	if cfg.General.PreferredProvider == "" {
		cfg.General.PreferredProvider = ProviderOpenAI
	}

	// Models defaults
	if cfg.Models.SmallModel == "" {
		cfg.Models.SmallModel = "gpt-4.1-mini"
	}
	if cfg.Models.BigModel == "" {
		cfg.Models.BigModel = "gpt-4.1"
	}

	// Initialize mappings if nil
	if cfg.Mappings == nil {
		cfg.Mappings = make(map[string]string)
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate server configuration
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535")
	}

	// Validate provider
	if !isValidProvider(c.General.PreferredProvider) {
		return fmt.Errorf("invalid preferred provider: %s", c.General.PreferredProvider)
	}

	// Validate Google configuration
	if c.Google.UseVertexAuth {
		if c.Google.VertexProject == "" {
			return fmt.Errorf("vertex_project is required when use_vertex_auth is true")
		}
		if c.Google.VertexLocation == "" {
			return fmt.Errorf("vertex_location is required when use_vertex_auth is true")
		}
	}

	return nil
}

// ServerHost returns the server host
func (c *Config) ServerHost() string {
	return c.Server.Host
}

// ServerPort returns the server port
func (c *Config) ServerPort() int {
	return c.Server.Port
}

// Verbose returns whether verbose logging is enabled
func (c *Config) Verbose() bool {
	return c.General.Verbose
}

// GetDefaultBigModel returns the default big model for the preferred provider
func (c *Config) GetDefaultBigModel() string {
	if c.Models.BigModel != "" {
		return c.Models.BigModel
	}
	return c.General.PreferredProvider.GetDefaultBigModel()
}

// GetDefaultSmallModel returns the default small model for the preferred provider
func (c *Config) GetDefaultSmallModel() string {
	if c.Models.SmallModel != "" {
		return c.Models.SmallModel
	}
	return c.General.PreferredProvider.GetDefaultSmallModel()
}

// IsConfigured returns true if at least one provider is configured
// (either server-side key or client can provide key)
func (c *Config) IsConfigured() bool {
	// If any server-side API key is configured, return true
	if c.OpenAIKey != "" || c.GeminiAPIKey != "" || c.AnthropicAPIKey != "" {
		return true
	}

	// If no server-side keys, we can still work if clients provide keys
	// This is a valid configuration for a pure proxy
	return true
}

// GetDefaultBigModel returns the default big model for a given provider
func (p Provider) GetDefaultBigModel() string {
	switch p {
	case ProviderOpenAI:
		return "gpt-4.1"
	case ProviderGoogle:
		return "gemini-2.5-pro"
	case ProviderAnthropic:
		return "claude-sonnet-4-20250514"
	default:
		return ""
	}
}

// GetDefaultSmallModel returns the default small model for a given provider
func (p Provider) GetDefaultSmallModel() string {
	switch p {
	case ProviderOpenAI:
		return "gpt-4.1-mini"
	case ProviderGoogle:
		return "gemini-2.5-flash"
	case ProviderAnthropic:
		return "claude-haiku-4-20250514"
	default:
		return ""
	}
}

// GetPrefix returns the prefix for a given provider
func (p Provider) GetPrefix() string {
	switch p {
	case ProviderOpenAI:
		return "openai/"
	case ProviderGoogle:
		return "gemini/"
	case ProviderAnthropic:
		return "anthropic/"
	default:
		return ""
	}
}

// ParseProvider parses a provider string to Provider type
func ParseProvider(s string) (Provider, error) {
	switch strings.ToLower(s) {
	case "openai":
		return ProviderOpenAI, nil
	case "google", "gemini":
		return ProviderGoogle, nil
	case "anthropic", "claude":
		return ProviderAnthropic, nil
	default:
		return "", fmt.Errorf("unknown provider: %s", s)
	}
}

// isValidProvider checks if a provider is valid
func isValidProvider(p Provider) bool {
	switch p {
	case ProviderOpenAI, ProviderGoogle, ProviderAnthropic:
		return true
	default:
		return false
	}
}

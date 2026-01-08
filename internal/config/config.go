package config

import (
	"fmt"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

// Config holds the application configuration loaded from environment variables
type Config struct {
	// API Keys
	AnthropicAPIKey string `envconfig:"ANTHROPIC_API_KEY" required:"false"`
	OpenAIKey      string `envconfig:"OPENAI_API_KEY" required:"false"`
	GeminiAPIKey   string `envconfig:"GEMINI_API_KEY" required:"false"`

	// Vertex AI Configuration
	VertexProject  string `envconfig:"VERTEX_PROJECT" required:"false"`
	VertexLocation string `envconfig:"VERTEX_LOCATION" required:"false"`
	UseVertexAuth  bool   `envconfig:"USE_VERTEX_AUTH" required:"false" default:"false"`

	// Model Configuration
	PreferredProvider Provider `envconfig:"PREFERRED_PROVIDER" required:"false" default:"openai"`
	BigModel         string   `envconfig:"BIG_MODEL" required:"false"`
	SmallModel       string   `envconfig:"SMALL_MODEL" required:"false"`

	// Server Configuration
	ServerHost string `envconfig:"SERVER_HOST" required:"false" default:"0.0.0.0"`
	ServerPort int    `envconfig:"SERVER_PORT" required:"false" default:"8082"`

	// Logging
	Verbose bool `envconfig:"VERBOSE" required:"false" default:"false"`
}

// Provider represents the LLM provider
type Provider string

const (
	ProviderOpenAI    Provider = "openai"
	ProviderGoogle    Provider = "google"
	ProviderAnthropic Provider = "anthropic"
)

// Load loads configuration from environment variables
func Load() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Set default models based on preferred provider
	if cfg.BigModel == "" {
		switch cfg.PreferredProvider {
		case ProviderOpenAI:
			cfg.BigModel = "gpt-4.1"
		case ProviderGoogle:
			cfg.BigModel = "gemini-2.5-pro"
		case ProviderAnthropic:
			cfg.BigModel = "claude-sonnet-4-20250514"
		}
	}

	if cfg.SmallModel == "" {
		switch cfg.PreferredProvider {
		case ProviderOpenAI:
			cfg.SmallModel = "gpt-4.1-mini"
		case ProviderGoogle:
			cfg.SmallModel = "gemini-2.5-flash"
		case ProviderAnthropic:
			cfg.SmallModel = "claude-haiku-4-20250514"
		}
	}

	// Validate Vertex AI configuration if using Vertex auth
	if cfg.UseVertexAuth && (cfg.VertexProject == "" || cfg.VertexLocation == "") {
		return nil, fmt.Errorf("VERTEX_PROJECT and VERTEX_LOCATION are required when USE_VERTEX_AUTH is true")
	}

	// Validate that at least one provider is configured
	if cfg.OpenAIKey == "" && cfg.GeminiAPIKey == "" && cfg.AnthropicAPIKey == "" {
		return nil, fmt.Errorf("at least one API key (OPENAI_API_KEY, GEMINI_API_KEY, or ANTHROPIC_API_KEY) must be configured")
	}

	// Validate provider configuration
	switch cfg.PreferredProvider {
	case ProviderGoogle:
		if cfg.UseVertexAuth {
			if cfg.VertexProject == "" || cfg.VertexLocation == "" {
				return nil, fmt.Errorf("VERTEX_PROJECT and VERTEX_LOCATION are required for Google provider with Vertex auth")
			}
		} else if cfg.GeminiAPIKey == "" {
			return nil, fmt.Errorf("GEMINI_API_KEY is required for Google provider when not using Vertex auth")
		}
	case ProviderAnthropic:
		if cfg.AnthropicAPIKey == "" {
			return nil, fmt.Errorf("ANTHROPIC_API_KEY is required for Anthropic provider")
		}
	case ProviderOpenAI:
		if cfg.OpenAIKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY is required for OpenAI provider")
		}
	}

	return &cfg, nil
}

// GetDefaultBigModel returns the default big model for the given provider
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

// GetDefaultSmallModel returns the default small model for the given provider
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

// GetPrefix returns the prefix for the provider
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

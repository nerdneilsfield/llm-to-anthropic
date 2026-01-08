package proxy

import (
	"fmt"
	"strings"

	"github.com/nerdneilsfield/llm-to-anthropic/internal/config"
	"github.com/nerdneilsfield/llm-to-anthropic/pkg/api/proxy/openai"
	"github.com/nerdneilsfield/llm-to-anthropic/pkg/api/proxy/gemini"
)

const (
	// Anthropic model names that need mapping
	AnthropicModelHaiku  = "haiku"
	AnthropicModelSonnet = "sonnet"
	AnthropicModelOpus   = "opus"
)

// Model represents a model with its provider information
type Model struct {
	ID        string
	Provider  config.Provider
	Name      string // The actual model name (without prefix)
}

// ModelManager handles model mapping and routing
type ModelManager struct {
	cfg *config.Config
}

// NewModelManager creates a new model manager
func NewModelManager(cfg *config.Config) *ModelManager {
	return &ModelManager{cfg: cfg}
}

// ParseModel parses a model string and returns provider and actual model name
func (m *ModelManager) ParseModel(modelStr string) (*Model, error) {
	// Check if model has an explicit provider prefix
	if strings.HasPrefix(modelStr, "openai/") {
		return &Model{
			ID:       modelStr,
			Provider: config.ProviderOpenAI,
			Name:     strings.TrimPrefix(modelStr, "openai/"),
		}, nil
	}

	if strings.HasPrefix(modelStr, "gemini/") {
		return &Model{
			ID:       modelStr,
			Provider: config.ProviderGoogle,
			Name:     strings.TrimPrefix(modelStr, "gemini/"),
		}, nil
	}

	if strings.HasPrefix(modelStr, "anthropic/") {
		return &Model{
			ID:       modelStr,
			Provider: config.ProviderAnthropic,
			Name:     strings.TrimPrefix(modelStr, "anthropic/"),
		}, nil
	}

	// Check if it's an Anthropic model name that needs mapping
	switch modelStr {
	case AnthropicModelHaiku:
		providerModel := m.getProviderSmallModel()
		return &Model{
			ID:       providerModel,
			Provider: m.cfg.General.PreferredProvider,
			Name:     strings.TrimPrefix(providerModel, m.cfg.General.PreferredProvider.GetPrefix()),
		}, nil

	case AnthropicModelSonnet:
		providerModel := m.getProviderMediumModel()
		return &Model{
			ID:       providerModel,
			Provider: m.cfg.General.PreferredProvider,
			Name:     strings.TrimPrefix(providerModel, m.cfg.General.PreferredProvider.GetPrefix()),
		}, nil

	case AnthropicModelOpus:
		providerModel := m.getProviderBigModel()
		return &Model{
			ID:       providerModel,
			Provider: m.cfg.General.PreferredProvider,
			Name:     strings.TrimPrefix(providerModel, m.cfg.General.PreferredProvider.GetPrefix()),
		}, nil
	}

	// Try to auto-detect provider based on known model names
	if isOpenAIModel(modelStr) {
		return &Model{
			ID:       "openai/" + modelStr,
			Provider: config.ProviderOpenAI,
			Name:     modelStr,
		}, nil
	}

	if isGeminiModel(modelStr) {
		return &Model{
			ID:       "gemini/" + modelStr,
			Provider: config.ProviderGoogle,
			Name:     modelStr,
		}, nil
	}

	// Default to preferred provider
	return &Model{
		ID:       m.cfg.General.PreferredProvider.GetPrefix() + modelStr,
		Provider: m.cfg.General.PreferredProvider,
		Name:     modelStr,
	}, nil
}

// getProviderBigModel returns the configured big model with provider prefix
func (m *ModelManager) getProviderBigModel() string {
	if m.cfg.Models.BigModel != "" {
		// If it already has a prefix, use as-is
		if strings.Contains(m.cfg.Models.BigModel, "/") {
			return m.cfg.Models.BigModel
		}
		// Otherwise, add the preferred provider's prefix
		return m.cfg.General.PreferredProvider.GetPrefix() + m.cfg.Models.BigModel
	}
	// Use default for preferred provider
	defaultModel := m.cfg.General.PreferredProvider.GetDefaultBigModel()
	return m.cfg.General.PreferredProvider.GetPrefix() + defaultModel
}

// getProviderMediumModel returns the configured medium model with provider prefix
func (m *ModelManager) getProviderMediumModel() string {
	if m.cfg.Models.MediumModel != "" {
		// If it already has a prefix, use as-is
		if strings.Contains(m.cfg.Models.MediumModel, "/") {
			return m.cfg.Models.MediumModel
		}
		// Otherwise, add the preferred provider's prefix
		return m.cfg.General.PreferredProvider.GetPrefix() + m.cfg.Models.MediumModel
	}
	// Use default for preferred provider
	defaultModel := m.cfg.General.PreferredProvider.GetDefaultMediumModel()
	return m.cfg.General.PreferredProvider.GetPrefix() + defaultModel
}

// getProviderSmallModel returns the configured small model with provider prefix
func (m *ModelManager) getProviderSmallModel() string {
	if m.cfg.Models.SmallModel != "" {
		// If it already has a prefix, use as-is
		if strings.Contains(m.cfg.Models.SmallModel, "/") {
			return m.cfg.Models.SmallModel
		}
		// Otherwise, add the preferred provider's prefix
		return m.cfg.General.PreferredProvider.GetPrefix() + m.cfg.Models.SmallModel
	}
	// Use default for preferred provider
	defaultModel := m.cfg.General.PreferredProvider.GetDefaultSmallModel()
	return m.cfg.General.PreferredProvider.GetPrefix() + defaultModel
}

// GetAvailableModels returns all available models
func (m *ModelManager) GetAvailableModels() []Model {
	models := []Model{}

	// Add Anthropic model mappings
	if m.cfg.AnthropicAPIKey != "" {
		models = append(models, Model{
			ID:       "anthropic/" + AnthropicModelHaiku,
			Provider: config.ProviderAnthropic,
			Name:     AnthropicModelHaiku,
		})
		models = append(models, Model{
			ID:       "anthropic/" + AnthropicModelSonnet,
			Provider: config.ProviderAnthropic,
			Name:     AnthropicModelSonnet,
		})
	}

	// Add OpenAI models
	if m.cfg.OpenAIKey != "" {
		for _, model := range openai.SupportedModels {
			models = append(models, Model{
				ID:       "openai/" + model,
				Provider: config.ProviderOpenAI,
				Name:     model,
			})
		}
	}

	// Add Gemini models
	if m.cfg.GeminiAPIKey != "" || m.cfg.Google.UseVertexAuth {
		for _, model := range gemini.SupportedModels {
			models = append(models, Model{
				ID:       "gemini/" + model,
				Provider: config.ProviderGoogle,
				Name:     model,
			})
		}
	}

	// Add mapped models
	if m.cfg.General.PreferredProvider == config.ProviderOpenAI && m.cfg.OpenAIKey != "" {
		models = append(models, Model{
			ID:       AnthropicModelHaiku,
			Provider: config.ProviderOpenAI,
			Name:     fmt.Sprintf("openai/%s (mapped)", m.getProviderSmallModel()),
		})
		models = append(models, Model{
			ID:       AnthropicModelSonnet,
			Provider: config.ProviderOpenAI,
			Name:     fmt.Sprintf("openai/%s (mapped)", m.getProviderBigModel()),
		})
	}

	if m.cfg.General.PreferredProvider == config.ProviderGoogle && (m.cfg.GeminiAPIKey != "" || m.cfg.Google.UseVertexAuth) {
		models = append(models, Model{
			ID:       AnthropicModelHaiku,
			Provider: config.ProviderGoogle,
			Name:     fmt.Sprintf("gemini/%s (mapped)", m.getProviderSmallModel()),
		})
		models = append(models, Model{
			ID:       AnthropicModelSonnet,
			Provider: config.ProviderGoogle,
			Name:     fmt.Sprintf("gemini/%s (mapped)", m.getProviderBigModel()),
		})
	}

	return models
}

// isOpenAIModel checks if a model name is a known OpenAI model
func isOpenAIModel(model string) bool {
	for _, supported := range openai.SupportedModels {
		if model == supported {
			return true
		}
	}
	return false
}

// isGeminiModel checks if a model name is a known Gemini model
func isGeminiModel(model string) bool {
	for _, supported := range gemini.SupportedModels {
		if model == supported {
			return true
		}
	}
	return false
}

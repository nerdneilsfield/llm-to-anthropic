package proxy

import (
	"fmt"
	"strings"

	"github.com/nerdneilsfield/llm-to-anthropic/internal/config"
)

const (
	// Anthropic model names that need mapping
	AnthropicModelHaiku  = "haiku"
	AnthropicModelSonnet = "sonnet"
	AnthropicModelOpus   = "opus"
)

// Model represents a model with its provider information
type Model struct {
	ID       string
	Provider *config.Provider
	Name     string // The actual model name (without prefix)
}

// ModelManager handles model mapping and routing
type ModelManager struct {
	cfg *config.Config
}

// NewModelManager creates a new model manager
func NewModelManager(cfg *config.Config) *ModelManager {
	return &ModelManager{
		cfg: cfg,
	}
}

// ParseModel parses a model string and returns to model information
// Supports formats:
// 1. "provider/model" - direct provider/model specification
// 2. "model_name" - looks up in mappings, then defaults
// 3. "haiku"/"sonnet"/"opus" - special mappings
func (m *ModelManager) ParseModel(modelStr string) (*Model, error) {
	// Check if it's a direct provider/model specification
	if strings.Contains(modelStr, "/") {
		return m.parseDirectModel(modelStr)
	}

	// Check for special model names
	switch modelStr {
	case AnthropicModelHaiku, AnthropicModelSonnet, AnthropicModelOpus:
		return m.parseSpecialModel(modelStr)
	}

	// Check if it's a mapping
	if mappedModel, ok := m.cfg.Mappings[modelStr]; ok {
		return m.parseDirectModel(mappedModel)
	}

	// Default to first provider's models
	return m.parseDefaultModel(modelStr)
}

// parseDirectModel parses a "provider/model" string
func (m *ModelManager) parseDirectModel(modelStr string) (*Model, error) {
	providerName, modelName := config.ParseModelMapping(modelStr)

	// Find provider
	provider, ok := m.cfg.GetProviderByName(providerName)
	if !ok {
		return nil, fmt.Errorf("provider '%s' not found", providerName)
	}

	// Validate model exists in provider's models
	if !m.modelExists(provider, modelName) {
		return nil, fmt.Errorf("model '%s' not found in provider '%s'", modelName, providerName)
	}

	return &Model{
		ID:       modelStr,
		Provider: provider,
		Name:     modelName,
	}, nil
}

// parseSpecialModel parses special model names (haiku, sonnet, opus)
func (m *ModelManager) parseSpecialModel(modelStr string) (*Model, error) {
	// Check if there's a mapping for this special model
	if mappedModel, ok := m.cfg.Mappings[modelStr]; ok {
		return m.parseDirectModel(mappedModel)
	}

	// No mapping, use default provider's default model
	return m.parseDefaultModel(modelStr)
}

// parseDefaultModel parses using default provider
func (m *ModelManager) parseDefaultModel(modelStr string) (*Model, error) {
	// Try to find a provider that has this model
	for i := range m.cfg.Providers {
		provider := &m.cfg.Providers[i]
		if m.modelExists(provider, modelStr) {
			return &Model{
				ID:       provider.Name + "/" + modelStr,
				Provider: provider,
				Name:     modelStr,
			}, nil
		}
	}

	return nil, fmt.Errorf("model '%s' not found in any provider", modelStr)
}

// modelExists checks if a model exists in a provider's model list
func (m *ModelManager) modelExists(provider *config.Provider, modelName string) bool {
	for _, model := range provider.Models {
		if model == modelName {
			return true
		}
	}
	return false
}

// GetAvailableModels returns all available models from all providers
func (m *ModelManager) GetAvailableModels() []Model {
	models := []Model{}

	for i := range m.cfg.Providers {
		provider := &m.cfg.Providers[i]
		for _, modelName := range provider.Models {
			models = append(models, Model{
				ID:       provider.Name + "/" + modelName,
				Provider: provider,
				Name:     modelName,
			})
		}
	}

	return models
}

// GetProvider returns to provider for a model
func (m *ModelManager) GetProvider(model *Model) *config.Provider {
	return model.Provider
}

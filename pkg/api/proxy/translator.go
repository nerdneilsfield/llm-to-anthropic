package proxy

import (
	"io"

	"github.com/nerdneilsfield/go-template/internal/config"
)

// ProviderClient interface defines the contract for backend provider clients
type ProviderClient interface {
	// SendRequest sends a non-streaming request to the provider
	// apiKey is optional - if provided, it overrides the default API key
	SendRequest(model string, req interface{}, apiKey ...string) ([]byte, error)

	// SendStream sends a streaming request to the provider
	// apiKey is optional - if provided, it overrides the default API key
	SendStream(model string, req interface{}, apiKey ...string) (io.ReadCloser, error)

	// GetProvider returns the provider type
	GetProvider() config.Provider

	// IsConfigured returns true if the provider is properly configured
	// (has either default API key or supports client-provided keys)
	IsConfigured() bool
}

package proxy

import (
	"io"

	"github.com/nerdneilsfield/go-template/internal/config"
)

// ProviderClient interface defines the contract for backend provider clients
type ProviderClient interface {
	// SendRequest sends a non-streaming request to the provider
	SendRequest(model string, req interface{}) ([]byte, error)

	// SendStream sends a streaming request to the provider
	SendStream(model string, req interface{}) (io.ReadCloser, error)

	// GetProvider returns the provider type
	GetProvider() config.Provider

	// IsConfigured returns true if the provider is properly configured
	IsConfigured() bool
}

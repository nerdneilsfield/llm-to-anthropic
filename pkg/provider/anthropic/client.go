package anthropic

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
	"bytes"

	"github.com/nerdneilsfield/llm-to-anthropic/internal/config"
	"github.com/valyala/fasthttp"
)

const (
	// MessagesEndpoint is the messages endpoint
	MessagesEndpoint = "/v1/messages"
	ChatCompletionEndpoint = "/v1/messages"
)

// Client implements ProviderClient for Anthropic
type Client struct {
	provider *config.Provider
	client    *fasthttp.Client
}

// NewClient creates a new Anthropic client
func NewClient(provider *config.Provider) *Client {
	return &Client{
		provider: provider,
		client: &fasthttp.Client{
			MaxConnsPerHost: 100,
			ReadTimeout:     120 * time.Second,
			WriteTimeout:    120 * time.Second,
		},
	}
}

// SendRequest sends a non-streaming request to Anthropic
// apiKey is optional - if provided, it overrides the provider's API key
func (c *Client) SendRequest(model string, req interface{}, apiKey ...string) ([]byte, error) {
	key := c.provider.ParsedAPIKey
	if c.provider.IsBypass && len(apiKey) > 0 && apiKey[0] != "" {
		key = apiKey[0]
	}

	if key == "" && !c.provider.IsBypass {
		return nil, fmt.Errorf("Anthropic API key not provided")
	}

	// Serialize request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	url := c.provider.BaseURL + MessagesEndpoint
	httpReq := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(httpReq)

	httpReq.SetRequestURI(url)
	httpReq.Header.SetMethod("POST")
	httpReq.Header.SetContentType("application/json")
	httpReq.Header.Set("x-api-key", key)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.SetBody(body)

	// Send request
	httpResp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(httpResp)

	if err := c.client.Do(httpReq, httpResp); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Check response status
	status := httpResp.StatusCode()
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("Anthropic API returned status %d: %s", status, httpResp.Body())
	}

	// Return response body
	result := make([]byte, len(httpResp.Body()))
	copy(result, httpResp.Body())
	return result, nil
}

// SendStreamRequest sends a streaming request to Anthropic
func (c *Client) SendStreamRequest(model string, req interface{}, apiKey ...string) (io.ReadCloser, error) {
	key := c.provider.ParsedAPIKey
	if c.provider.IsBypass && len(apiKey) > 0 && apiKey[0] != "" {
		key = apiKey[0]
	}

	if key == "" && !c.provider.IsBypass {
		return nil, fmt.Errorf("Anthropic API key not provided")
	}

	// Serialize request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	url := c.provider.BaseURL + MessagesEndpoint
	httpReq := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(httpReq)

	httpReq.SetRequestURI(url)
	httpReq.Header.SetMethod("POST")
	httpReq.Header.SetContentType("application/json")
	httpReq.Header.Set("x-api-key", key)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.SetBody(body)

	// Send streaming request
	// Note: fasthttp doesn't support streaming responses directly
	// We'll need to handle this differently
	return nil, fmt.Errorf("streaming not implemented for fasthttp")
}

// GetProvider returns the provider configuration
func (c *Client) GetProvider() config.Provider {
	return *c.provider
}

// IsConfigured returns true if the provider is properly configured
func (c *Client) IsConfigured() bool {
	return c.provider.ParsedAPIKey != "" || c.provider.IsBypass
}

// SendStream sends a streaming request to Anthropic


func (c *Client) SendStream(model string, req interface{}, apiKey ...string) (io.ReadCloser, error) {
	key := c.provider.ParsedAPIKey
	if c.provider.IsBypass && len(apiKey) > 0 && apiKey[0] != "" {
		key = apiKey[0]
	}

	if key == "" && !c.provider.IsBypass {
		return nil, fmt.Errorf("Anthropic API key not provided")
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.provider.BaseURL + ChatCompletionEndpoint
	httpReq := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(httpReq)

	httpReq.SetRequestURI(url)
	httpReq.Header.SetMethod("POST")
	httpReq.Header.SetContentType("application/json")
	httpReq.Header.Set("x-api-key", key)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.SetBody(body)

	httpResp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(httpResp)

	if err := c.client.Do(httpReq, httpResp); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	status := httpResp.StatusCode()
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("Anthropic API returned status %d: %s", status, httpResp.Body())
	}

	bodyCopy := make([]byte, len(httpResp.Body()))
	copy(bodyCopy, httpResp.Body())

	return io.NopCloser(bytes.NewReader(bodyCopy)), nil
}



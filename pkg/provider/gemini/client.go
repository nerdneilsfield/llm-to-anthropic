package gemini

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/nerdneilsfield/llm-to-anthropic/internal/config"
	"github.com/valyala/fasthttp"
)

const (
	// GenerateContentEndpoint is the generate content endpoint
	GenerateContentEndpoint = "/models/{model}:generateContent"
	// StreamGenerateContentEndpoint is the streaming generate content endpoint
	StreamGenerateContentEndpoint = "/models/{model}:streamGenerateContent"
)

// Client implements ProviderClient for Google Gemini
type Client struct {
	provider *config.Provider
	client    *fasthttp.Client
}

// NewClient creates a new Gemini client
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

// SendRequest sends a non-streaming request to Gemini
// apiKey is optional - if provided, it overrides the provider's API key
func (c *Client) SendRequest(model string, req interface{}, apiKey ...string) ([]byte, error) {
	key := c.provider.ParsedAPIKey
	if c.provider.IsBypass && len(apiKey) > 0 && apiKey[0] != "" {
		key = apiKey[0]
	}

	if key == "" && !c.provider.IsBypass && !c.provider.UseVertexAuth {
		return nil, fmt.Errorf("Gemini API key not provided")
	}

	// Serialize request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create URL
	// Replace {model} with actual model name
	url := c.provider.BaseURL + "/models/" + model + ":generateContent"
	if c.provider.UseVertexAuth {
		url = c.provider.BaseURL + "/models/" + model + ":generateContent"
	}

	// Create request
	httpReq := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(httpReq)

	httpReq.SetRequestURI(url)
	httpReq.Header.SetMethod("POST")
	httpReq.Header.SetContentType("application/json")

	// Set authentication
	if c.provider.UseVertexAuth {
		// Vertex AI uses OAuth token in Authorization header
		// For simplicity, we'll just use the key as bearer token
		httpReq.Header.Set("Authorization", "Bearer "+key)
	} else {
		// Public Gemini API uses query parameter
		// Add ?key= to URL
		httpReq.SetRequestURI(url + "?key=" + key)
	}

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
		return nil, fmt.Errorf("Gemini API returned status %d: %s", status, httpResp.Body())
	}

	// Return response body
	result := make([]byte, len(httpResp.Body()))
	copy(result, httpResp.Body())
	return result, nil
}

// SendStreamRequest sends a streaming request to Gemini
func (c *Client) SendStreamRequest(model string, req interface{}, apiKey ...string) (io.ReadCloser, error) {
	key := c.provider.ParsedAPIKey
	if c.provider.IsBypass && len(apiKey) > 0 && apiKey[0] != "" {
		key = apiKey[0]
	}

	if key == "" && !c.provider.IsBypass && !c.provider.UseVertexAuth {
		return nil, fmt.Errorf("Gemini API key not provided")
	}

	// Serialize request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create URL
	url := c.provider.BaseURL + "/models/" + model + ":streamGenerateContent"

	// Create request
	httpReq := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(httpReq)

	httpReq.SetRequestURI(url)
	httpReq.Header.SetMethod("POST")
	httpReq.Header.SetContentType("application/json")

	// Set authentication
	if c.provider.UseVertexAuth {
		httpReq.Header.Set("Authorization", "Bearer "+key)
	} else {
		httpReq.SetRequestURI(url + "?key=" + key)
	}

	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.SetBody(body)

	// Send streaming request
	// Note: fasthttp doesn't support streaming responses directly
	// We'll need to handle this differently
	return nil, fmt.Errorf("streaming not implemented for fasthttp")
}

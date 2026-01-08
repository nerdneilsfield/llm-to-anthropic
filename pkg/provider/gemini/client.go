package gemini

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/nerdneilsfield/llm-to-anthropic/internal/config"
	"github.com/nerdneilsfield/llm-to-anthropic/pkg/api/proxy/gemini"
	"github.com/valyala/fasthttp"
)

const (
	// BaseURL is the base URL for Gemini API
	BaseURL = "https://generativelanguage.googleapis.com/v1beta"
	// GenerateContentEndpoint is the generate content endpoint
	GenerateContentEndpoint = "/models/%s:generateContent"
)

// Client implements ProviderClient for Google Gemini
type Client struct {
	apiKey       string
	useVertexAuth bool
	vertexProject string
	vertexLocation string
	client       *fasthttp.Client
	cfg          *config.Config
}

// NewClient creates a new Gemini client
func NewClient(cfg *config.Config) *Client {
	return &Client{
		apiKey:        cfg.GeminiAPIKey,
		useVertexAuth:  cfg.Google.UseVertexAuth,
		vertexProject:  cfg.Google.VertexProject,
		vertexLocation: cfg.Google.VertexLocation,
		client: &fasthttp.Client{
			MaxConnsPerHost: 100,
			ReadTimeout:     120 * time.Second,
			WriteTimeout:    120 * time.Second,
		},
		cfg: cfg,
	}
}

// SendRequest sends a non-streaming request to Gemini
// apiKey is optional - if provided, it overrides default API key
func (c *Client) SendRequest(model string, req interface{}, apiKey ...string) ([]byte, error) {
	key := c.apiKey
	if len(apiKey) > 0 && apiKey[0] != "" {
		key = apiKey[0]
	}

	endpoint := c.getEndpoint(model, key)
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(httpReq)

	httpReq.SetRequestURI(endpoint)
	httpReq.Header.SetMethod("POST")
	httpReq.Header.SetContentType("application/json")

	httpReq.SetBody(reqBody)

	httpResp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(httpResp)

	if err := c.client.Do(httpReq, httpResp); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	body := httpResp.Body()
	statusCode := httpResp.StatusCode()

	if statusCode != fasthttp.StatusOK {
		var errorResp gemini.ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return nil, fmt.Errorf("request failed with status %d: %s", statusCode, string(body))
		}
		return nil, fmt.Errorf("Gemini API error: %s", errorResp.Error.Message)
	}

	// Return a copy of the body
	result := make([]byte, len(body))
	copy(result, body)
	return result, nil
}

// SendStream sends a streaming request to Gemini
// apiKey is optional - if provided, it overrides default API key
func (c *Client) SendStream(model string, req interface{}, apiKey ...string) (io.ReadCloser, error) {
	key := c.apiKey
	if len(apiKey) > 0 && apiKey[0] != "" {
		key = apiKey[0]
	}

	if !c.useVertexAuth && key == "" {
		return nil, fmt.Errorf("Gemini API key not provided")
	}

	endpoint := c.getEndpoint(model, key)
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(httpReq)

	httpReq.SetRequestURI(endpoint)
	httpReq.Header.SetMethod("POST")
	httpReq.Header.SetContentType("application/json")

	httpReq.SetBody(reqBody)

	httpResp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(httpResp)

	if err := c.client.Do(httpReq, httpResp); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if httpResp.StatusCode() != fasthttp.StatusOK {
		body := httpResp.Body()
		var errorResp gemini.ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return nil, fmt.Errorf("request failed with status %d: %s", httpResp.StatusCode(), string(body))
		}
		return nil, fmt.Errorf("Gemini API error: %s", errorResp.Error.Message)
	}

	// Return a stream reader wrapper
	return &streamReader{resp: httpResp}, nil
}

// getEndpoint returns the endpoint URL for a given model and API key
func (c *Client) getEndpoint(model string, apiKey string) string {
	if c.useVertexAuth {
		// Vertex AI endpoint format (uses ADC, not API key)
		return fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:streamGenerateContent",
			c.vertexLocation, c.vertexProject, c.vertexLocation, model)
	}
	// Standard Gemini API endpoint format - add key to URL
	return fmt.Sprintf("%s/%s?key=%s", BaseURL, fmt.Sprintf(GenerateContentEndpoint, model), apiKey)
}

// GetProvider returns the provider type
func (c *Client) GetProvider() config.Provider {
	return config.ProviderGoogle
}

// IsConfigured returns true if the client is properly configured
func (c *Client) IsConfigured() bool {
	if c.useVertexAuth {
		return c.vertexProject != "" && c.vertexLocation != ""
	}
	return c.apiKey != ""
}

// streamReader wraps fasthttp.Response for streaming
type streamReader struct {
	resp   *fasthttp.Response
	stream io.Reader
}

func (sr *streamReader) Read(p []byte) (n int, err error) {
	if sr.stream == nil {
		sr.stream = sr.resp.BodyStream()
	}
	return sr.stream.Read(p)
}

func (sr *streamReader) Close() error {
	return nil
}

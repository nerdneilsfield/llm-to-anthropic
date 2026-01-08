package openai

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/nerdneilsfield/llm-to-anthropic/internal/config"
	"github.com/nerdneilsfield/llm-to-anthropic/pkg/api/proxy/openai"
	"github.com/valyala/fasthttp"
)

const (
	// BaseURL is the OpenAI API base URL
	BaseURL = "https://api.openai.com/v1"
	// ChatCompletionEndpoint is the chat completion endpoint
	ChatCompletionEndpoint = "/chat/completions"
)

// Client implements ProviderClient for OpenAI
type Client struct {
	apiKey string
	client *fasthttp.Client
	cfg     *config.Config
}

// NewClient creates a new OpenAI client
func NewClient(cfg *config.Config) *Client {
	return &Client{
		apiKey: cfg.OpenAIKey,
		client: &fasthttp.Client{
			MaxConnsPerHost: 100,
			ReadTimeout:     120 * time.Second,
			WriteTimeout:    120 * time.Second,
		},
		cfg: cfg,
	}
}

// SendRequest sends a non-streaming request to OpenAI
// apiKey is optional - if provided, it overrides the default API key
func (c *Client) SendRequest(model string, req interface{}, apiKey ...string) ([]byte, error) {
	key := c.apiKey
	if len(apiKey) > 0 && apiKey[0] != "" {
		key = apiKey[0]
	}

	if key == "" {
		return nil, fmt.Errorf("OpenAI API key not provided")
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(httpReq)

	httpReq.SetRequestURI(BaseURL + ChatCompletionEndpoint)
	httpReq.Header.SetMethod("POST")
	httpReq.Header.SetContentType("application/json")
	httpReq.Header.Set("Authorization", "Bearer "+key)
	httpReq.SetBody(reqBody)

	httpResp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(httpResp)

	if err := c.client.Do(httpReq, httpResp); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	body := httpResp.Body()
	statusCode := httpResp.StatusCode()

	if statusCode != fasthttp.StatusOK {
		var errorResp openai.ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return nil, fmt.Errorf("request failed with status %d: %s", statusCode, string(body))
		}
		return nil, fmt.Errorf("OpenAI API error: %s", errorResp.Error.Message)
	}

	// Return a copy of the body
	result := make([]byte, len(body))
	copy(result, body)
	return result, nil
}

// SendStream sends a streaming request to OpenAI
// apiKey is optional - if provided, it overrides default API key
func (c *Client) SendStream(model string, req interface{}, apiKey ...string) (io.ReadCloser, error) {
	key := c.apiKey
	if len(apiKey) > 0 && apiKey[0] != "" {
		key = apiKey[0]
	}

	if key == "" {
		return nil, fmt.Errorf("OpenAI API key not provided")
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(httpReq)

	httpReq.SetRequestURI(BaseURL + ChatCompletionEndpoint)
	httpReq.Header.SetMethod("POST")
	httpReq.Header.SetContentType("application/json")
	httpReq.Header.Set("Authorization", "Bearer "+key)
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.SetBody(reqBody)

	httpResp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(httpResp)

	if err := c.client.Do(httpReq, httpResp); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if httpResp.StatusCode() != fasthttp.StatusOK {
		body := httpResp.Body()
		var errorResp openai.ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return nil, fmt.Errorf("request failed with status %d: %s", httpResp.StatusCode(), string(body))
		}
		return nil, fmt.Errorf("OpenAI API error: %s", errorResp.Error.Message)
	}

	// Return a stream reader wrapper
	return &streamReader{resp: httpResp}, nil
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

// GetProvider returns the provider type
func (c *Client) GetProvider() config.Provider {
	return config.ProviderOpenAI
}

// IsConfigured returns true if the client is properly configured
func (c *Client) IsConfigured() bool {
	return c.apiKey != ""
}

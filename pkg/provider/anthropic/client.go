package anthropic

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/nerdneilsfield/go-template/internal/config"
	"github.com/nerdneilsfield/go-template/pkg/api/proxy/anthropic"
	"github.com/valyala/fasthttp"
)

const (
	// BaseURL is the Anthropic API base URL
	BaseURL = "https://api.anthropic.com/v1"
	// MessagesEndpoint is the messages endpoint
	MessagesEndpoint = "/messages"
)

// Client implements ProviderClient for Anthropic
type Client struct {
	apiKey string
	client *fasthttp.Client
	cfg     *config.Config
}

// NewClient creates a new Anthropic client
func NewClient(cfg *config.Config) *Client {
	return &Client{
		apiKey: cfg.AnthropicAPIKey,
		client: &fasthttp.Client{
			MaxConnsPerHost: 100,
			ReadTimeout:     120 * time.Second,
			WriteTimeout:    120 * time.Second,
		},
		cfg: cfg,
	}
}

// SendRequest sends a non-streaming request to Anthropic
func (c *Client) SendRequest(model string, req interface{}) ([]byte, error) {
	// For direct Anthropic proxy, we don't translate the request
	// We just forward it as-is
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(httpReq)

	httpReq.SetRequestURI(BaseURL + MessagesEndpoint)
	httpReq.Header.SetMethod("POST")
	httpReq.Header.SetContentType("application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.SetBody(reqBody)

	httpResp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(httpResp)

	if err := c.client.Do(httpReq, httpResp); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	body := httpResp.Body()
	statusCode := httpResp.StatusCode()

	if statusCode != fasthttp.StatusOK {
		var errorResp anthropic.ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return nil, fmt.Errorf("request failed with status %d: %s", statusCode, string(body))
		}
		return nil, fmt.Errorf("Anthropic API error: %s", errorResp.Error.Message)
	}

	// Return a copy of the body
	result := make([]byte, len(body))
	copy(result, body)
	return result, nil
}

// SendStream sends a streaming request to Anthropic
func (c *Client) SendStream(model string, req interface{}) (io.ReadCloser, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(httpReq)

	httpReq.SetRequestURI(BaseURL + MessagesEndpoint)
	httpReq.Header.SetMethod("POST")
	httpReq.Header.SetContentType("application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.SetBody(reqBody)

	httpResp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(httpResp)

	if err := c.client.Do(httpReq, httpResp); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if httpResp.StatusCode() != fasthttp.StatusOK {
		body := httpResp.Body()
		var errorResp anthropic.ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return nil, fmt.Errorf("request failed with status %d: %s", httpResp.StatusCode(), string(body))
		}
		return nil, fmt.Errorf("Anthropic API error: %s", errorResp.Error.Message)
	}

	// Return a stream reader wrapper
	return &streamReader{resp: httpResp}, nil
}

// GetProvider returns the provider type
func (c *Client) GetProvider() config.Provider {
	return config.ProviderAnthropic
}

// IsConfigured returns true if the client is properly configured
func (c *Client) IsConfigured() bool {
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

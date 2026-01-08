package openai

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
	"strings"
	"bytes"
	"bufio"

	"github.com/nerdneilsfield/llm-to-anthropic/internal/config"
	"github.com/valyala/fasthttp"
)

const (
	// ChatCompletionEndpoint is the chat completion endpoint
	ChatCompletionEndpoint = "/chat/completions"
)

// Client implements ProviderClient for OpenAI
type Client struct {
	provider *config.Provider
	client    *fasthttp.Client
}

// NewClient creates a new OpenAI client
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

// SendRequest sends a non-streaming request to OpenAI
// apiKey is optional - if provided, it overrides the provider's API key
func (c *Client) SendRequest(model string, req interface{}, apiKey ...string) ([]byte, error) {
	key := c.provider.ParsedAPIKey
	if c.provider.IsBypass && len(apiKey) > 0 && apiKey[0] != "" {
		key = apiKey[0]
	}

	if key == "" && !c.provider.IsBypass {
		return nil, fmt.Errorf("OpenAI API key not provided")
	}

	// Serialize request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	url := c.provider.BaseURL + ChatCompletionEndpoint
	httpReq := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(httpReq)

	httpReq.SetRequestURI(url)
	httpReq.Header.SetMethod("POST")
	httpReq.Header.SetContentType("application/json")
	httpReq.Header.Set("Authorization", "Bearer "+key)
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
		return nil, fmt.Errorf("OpenAI API returned status %d: %s", status, httpResp.Body())
	}

	// Return response body
	result := make([]byte, len(httpResp.Body()))
	copy(result, httpResp.Body())
	return result, nil
}

// SendStreamRequest sends a streaming request to OpenAI
func (c *Client) SendStreamRequest(model string, req interface{}, apiKey ...string) (io.ReadCloser, error) {
	key := c.provider.ParsedAPIKey
	if c.provider.IsBypass && len(apiKey) > 0 && apiKey[0] != "" {
		key = apiKey[0]
	}

	if key == "" && !c.provider.IsBypass {
		return nil, fmt.Errorf("OpenAI API key not provided")
	}

	// Serialize request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	url := c.provider.BaseURL + ChatCompletionEndpoint
	httpReq := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(httpReq)

	httpReq.SetRequestURI(url)
	httpReq.Header.SetMethod("POST")
	httpReq.Header.SetContentType("application/json")
	httpReq.Header.Set("Authorization", "Bearer "+key)
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

// IsConfigured returns true if provider has API key or supports bypass
func (c *Client) IsConfigured() bool {
	return c.provider.ParsedAPIKey != "" || c.provider.IsBypass
}

// SendStream sends a streaming request to OpenAI

func (c *Client) SendStream(model string, req interface{}, apiKey ...string) (io.ReadCloser, error) {
	key := c.provider.ParsedAPIKey
	if c.provider.IsBypass && len(apiKey) > 0 && apiKey[0] != "" {
		key = apiKey[0]
	}

	if key == "" && !c.provider.IsBypass {
		return nil, fmt.Errorf("OpenAI API key not provided")
	}

	// Serialize request
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Parse to map and add stream=true
	var reqMap map[string]interface{}
	if err := json.Unmarshal(reqBytes, &reqMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}

	reqMap["stream"] = true

	if model != "" {
		reqMap["model"] = model
	}

	body, err := json.Marshal(reqMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.provider.BaseURL + ChatCompletionEndpoint
	httpReq := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(httpReq)

	httpReq.SetRequestURI(url)
	httpReq.Header.SetMethod("POST")
	httpReq.Header.SetContentType("application/json")
	httpReq.Header.Set("Authorization", "Bearer "+key)
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.SetBody(body)

	httpResp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(httpResp)

	if err := c.client.Do(httpReq, httpResp); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	status := httpResp.StatusCode()
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("OpenAI API returned status %d: %s", status, httpResp.Body())
	}

	bodyCopy := make([]byte, len(httpResp.Body()))
	copy(bodyCopy, httpResp.Body())

	return io.NopCloser(bytes.NewReader(bodyCopy)), nil
}



// ParseOpenAIStream parses OpenAI SSE stream
func ParseOpenAIStream(r io.Reader) (<-chan *StreamChunk, <-chan error) {
	chunks := make(chan *StreamChunk)
	errs := make(chan error, 1)

	go func() {
		defer close(chunks)
		defer close(errs)

		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			
			if line == "" || strings.HasPrefix(line, ":") {
				continue
			}

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			
			if data == "[DONE]" {
				break
			}

			var chunk StreamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				errs <- fmt.Errorf("failed to parse chunk: %w", err)
				return
			}

			chunks <- &chunk
		}

		if err := scanner.Err(); err != nil {
			errs <- fmt.Errorf("scanner error: %w", err)
		}
	}()

	return chunks, errs
}

// StreamChunk represents an OpenAI streaming chunk
type StreamChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64   `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int     `json:"index"`
		Delta        Delta  `json:"delta"`
		FinishReason *string `json:"finish_reason,omitempty"`
	} `json:"choices"`
}

// Delta represents a delta in a stream chunk
type Delta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

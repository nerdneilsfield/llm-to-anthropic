package openai

// ChatCompletionRequest represents OpenAI chat completion API request
type ChatCompletionRequest struct {
	Model            string                 `json:"model"`
	Messages         []Message             `json:"messages"`
	MaxTokens        int                    `json:"max_tokens,omitempty"`
	Temperature      *float64               `json:"temperature,omitempty"`
	TopP             *float64               `json:"top_p,omitempty"`
	N                *int                   `json:"n,omitempty"`
	Stream           bool                   `json:"stream,omitempty"`
	Stop             interface{}            `json:"stop,omitempty"` // Can be string or []string
	PresencePenalty  *float64               `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64               `json:"frequency_penalty,omitempty"`
	User             string                 `json:"user,omitempty"`
}

// Message represents a message in OpenAI format
type Message struct {
	Role    string `json:"role"` // system, user, assistant, tool
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

// ChatCompletionResponse represents OpenAI chat completion API response
type ChatCompletionResponse struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	Choices           []Choice `json:"choices"`
	Usage             *Usage   `json:"usage,omitempty"`
	SystemFingerprint string   `json:"system_fingerprint,omitempty"`
}

// Choice represents a completion choice
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason *string `json:"finish_reason"`
	Logprobs     *LogProbs `json:"logprobs,omitempty"`
}

// LogProbs represents log probabilities
type LogProbs struct {
	Content []TokenLogProb `json:"content"`
}

// TokenLogProb represents token log probability
type TokenLogProb struct {
	Token   string   `json:"token"`
	LogProb float64  `json:"logprob"`
	Bytes   []int    `json:"bytes,omitempty"`
	TopLogProbs []TopLogProb `json:"top_logprobs,omitempty"`
}

// TopLogProb represents top log probability
type TopLogProb struct {
	Token   string  `json:"token"`
	LogProb float64 `json:"logprob"`
	Bytes   []int   `json:"bytes,omitempty"`
}

// Usage represents token usage
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ErrorResponse represents OpenAI API error response
type ErrorResponse struct {
	Error *Error `json:"error"`
}

// Error represents an error detail
type Error struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param,omitempty"`
	Code    string `json:"code,omitempty"`
}

// StreamChunk represents a streaming response chunk
type StreamChunk struct {
	ID                string      `json:"id"`
	Object            string      `json:"object"`
	Created           int64       `json:"created"`
	Model             string      `json:"model"`
	Choices           []Choice    `json:"choices"`
	Usage             *Usage      `json:"usage,omitempty"`
	SystemFingerprint string      `json:"system_fingerprint,omitempty"`
}

// ModelsResponse represents response from models endpoint
type ModelsResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

// Model represents a model
type Model struct {
	ID      string `json:"id"`
	Created int64  `json:"created"`
	Object  string `json:"object"`
	OwnedBy string `json:"owned_by"`
}

// Supported OpenAI models
var SupportedModels = []string{
	"o3-mini",
	"o1",
	"o1-mini",
	"o1-pro",
	"gpt-4.5-preview",
	"gpt-4o",
	"gpt-4o-audio-preview",
	"chatgpt-4o-latest",
	"gpt-4o-mini",
	"gpt-4o-mini-audio-preview",
	"gpt-4.1",
	"gpt-4.1-mini",
}

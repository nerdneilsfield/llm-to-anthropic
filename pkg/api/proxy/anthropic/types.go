package anthropic

// MessageRequest represents Anthropic API v1 messages request
type MessageRequest struct {
	Model       string          `json:"model"`
	Messages    []Message       `json:"messages"`
	MaxTokens   int             `json:"max_tokens"`
	Stream      bool            `json:"stream,omitempty"`
	Temperature *float64        `json:"temperature,omitempty"`
	TopP        *float64        `json:"top_p,omitempty"`
	TopK        *int            `json:"top_k,omitempty"`
	StopSequences []string      `json:"stop_sequences,omitempty"`
	Metadata    *Metadata       `json:"metadata,omitempty"`
}

// Message represents a single message in the conversation
type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // Can be string or []ContentBlock
}

// ContentBlock represents a block of content
type ContentBlock struct {
	Type  string      `json:"type"` // "text" or "image"
	Text  string      `json:"text,omitempty"`
	Source *ImageSource `json:"source,omitempty"`
}

// ImageSource represents image source
type ImageSource struct {
	Type      string `json:"type"`       // "base64"
	MediaType string `json:"media_type"` // "image/jpeg", "image/png", "image/gif", "image/webp"
	Data      string `json:"data"`
}

// Metadata represents request metadata
type Metadata struct {
	UserID string `json:"user_id"`
}

// Usage represents token usage
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// MessageResponse represents Anthropic API v1 messages response
type MessageResponse struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   string         `json:"stop_reason"`
	StopSequence *string        `json:"stop_sequence,omitempty"`
	Usage        Usage          `json:"usage"`
}

// ErrorResponse represents Anthropic API error response
type ErrorResponse struct {
	Type    string `json:"type"`
	Error   *Error `json:"error"`
}

// Error represents an error detail
type Error struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// StreamEvent represents a server-sent event for streaming
type StreamEvent struct {
	Type string `json:"type"`
	// Depending on type, this will be different fields
	Index      *int            `json:"index,omitempty"`
	Delta      *ContentDelta   `json:"delta,omitempty"`
	Message    *MessageDelta   `json:"message,omitempty"`
	Usage      *Usage         `json:"usage,omitempty"`
}

// ContentDelta represents content delta in streaming
type ContentDelta struct {
	Type string `json:"type"` // "text_delta", "input_json_delta"
	Text string `json:"text,omitempty"`
}

// MessageDelta represents message delta in streaming
type MessageDelta struct {
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	StopReason   *string        `json:"stop_reason,omitempty"`
	StopSequence *string        `json:"stop_sequence,omitempty"`
}

// ModelsResponse represents the response from /v1/models endpoint
type ModelsResponse struct {
	Data []Model `json:"data"`
}

// Model represents a model
type Model struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	MaxTokens  int    `json:"max_tokens"`
	Type       string `json:"type"`
	Display    string `json:"display"`
	CreatedAt  string `json:"created_at"`
}

// Constants for streaming event types
const (
	EventTypeMessageStart   = "message_start"
	EventTypeMessageDelta   = "message_delta"
	EventTypeMessageStop    = "message_stop"
	EventTypeContentBlockStart = "content_block_start"
	EventTypeContentBlockDelta = "content_block_delta"
	EventTypeContentBlockStop = "content_block_stop"
	EventTypePing          = "ping"
	EventTypeError         = "error"
)

// Constants for stop reasons
const (
	StopReasonEndTurn       = "end_turn"
	StopReasonMaxTokens     = "max_tokens"
	StopReasonStopSequence  = "stop_sequence"
)

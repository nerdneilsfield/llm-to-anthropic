package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/nerdneilsfield/go-template/internal/config"
	"github.com/nerdneilsfield/go-template/pkg/api/proxy/anthropic"
)

// Translator implements Anthropic to OpenAI translation
type Translator struct{}

// NewTranslator creates a new OpenAI translator
func NewTranslator() *Translator {
	return &Translator{}
}

// RequestToProvider translates Anthropic request to OpenAI format
func (t *Translator) RequestToProvider(req *anthropic.MessageRequest) (interface{}, error) {
	openaiReq := &ChatCompletionRequest{
		Model:     req.Model,
		MaxTokens: req.MaxTokens,
		Stream:    req.Stream,
	}

	// Copy temperature, top_p if provided
	if req.Temperature != nil {
		openaiReq.Temperature = req.Temperature
	}
	if req.TopP != nil {
		openaiReq.TopP = req.TopP
	}

	// Translate messages
	openaiReq.Messages = make([]Message, 0, len(req.Messages))
	for _, msg := range req.Messages {
		openaiMsg, err := t.translateMessage(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to translate message: %w", err)
		}
		openaiReq.Messages = append(openaiReq.Messages, openaiMsg)
	}

	// Handle stop sequences
	if len(req.StopSequences) > 0 {
		if len(req.StopSequences) == 1 {
			openaiReq.Stop = req.StopSequences[0]
		} else {
			openaiReq.Stop = req.StopSequences
		}
	}

	// Handle metadata
	if req.Metadata != nil && req.Metadata.UserID != "" {
		openaiReq.User = req.Metadata.UserID
	}

	return openaiReq, nil
}

// translateMessage translates a single message from Anthropic to OpenAI format
func (t *Translator) translateMessage(msg anthropic.Message) (Message, error) {
	openaiMsg := Message{
		Role: t.translateRole(msg.Role),
	}

	// Handle content (can be string or []ContentBlock)
	switch content := msg.Content.(type) {
	case string:
		openaiMsg.Content = content
	case []interface{}:
		// Parse content blocks
		contentBlocks := make([]anthropic.ContentBlock, 0)
		for _, block := range content {
			// Unmarshal each block
			blockBytes, err := json.Marshal(block)
			if err != nil {
				return Message{}, fmt.Errorf("failed to marshal content block: %w", err)
			}

			var contentBlock anthropic.ContentBlock
			if err := json.Unmarshal(blockBytes, &contentBlock); err != nil {
				return Message{}, fmt.Errorf("failed to unmarshal content block: %w", err)
			}
			contentBlocks = append(contentBlocks, contentBlock)
		}

		// Convert content blocks to text
		text, err := t.convertContentBlocksToText(contentBlocks)
		if err != nil {
			return Message{}, fmt.Errorf("failed to convert content blocks: %w", err)
		}
		openaiMsg.Content = text
	default:
		return Message{}, fmt.Errorf("unsupported content type: %T", msg.Content)
	}

	return openaiMsg, nil
}

// convertContentBlocksToText converts Anthropic content blocks to a single text string
func (t *Translator) convertContentBlocksToText(blocks []anthropic.ContentBlock) (string, error) {
	var textParts []string

	for _, block := range blocks {
		switch block.Type {
		case "text":
			textParts = append(textParts, block.Text)
		case "image":
			// Images are not fully supported in chat completions
			// For now, we'll include a placeholder
			textParts = append(textParts, fmt.Sprintf("[Image: %s]", block.Source.MediaType))
		default:
			return "", fmt.Errorf("unsupported content block type: %s", block.Type)
		}
	}

	return strings.Join(textParts, "\n"), nil
}

// translateRole translates Anthropic role to OpenAI role
func (t *Translator) translateRole(role string) string {
	switch role {
	case "user":
		return "user"
	case "assistant":
		return "assistant"
	case "system":
		return "system"
	default:
		return "user" // Default to user
	}
}

// ResponseToAnthropic translates OpenAI response to Anthropic format
func (t *Translator) ResponseToAnthropic(resp []byte) (*anthropic.MessageResponse, error) {
	var openaiResp ChatCompletionResponse
	if err := json.Unmarshal(resp, &openaiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OpenAI response: %w", err)
	}

	// Create Anthropic response
	anthropicResp := &anthropic.MessageResponse{
		ID:   openaiResp.ID,
		Type: "message",
		Role: "assistant",
		Content: []anthropic.ContentBlock{
			{
				Type: "text",
				Text: openaiResp.Choices[0].Message.Content,
			},
		},
		Model:      openaiResp.Model,
		StopReason: t.translateFinishReason(openaiResp.Choices[0].FinishReason),
		Usage: anthropic.Usage{
			InputTokens:  openaiResp.Usage.PromptTokens,
			OutputTokens: openaiResp.Usage.CompletionTokens,
		},
	}

	return anthropicResp, nil
}

// translateFinishReason translates OpenAI finish reason to Anthropic format
func (t *Translator) translateFinishReason(reason *string) string {
	if reason == nil {
		return anthropic.StopReasonEndTurn
	}

	switch *reason {
	case "stop":
		return anthropic.StopReasonEndTurn
	case "length":
		return anthropic.StopReasonMaxTokens
	case "content_filter":
		return anthropic.StopReasonStopSequence
	default:
		return anthropic.StopReasonEndTurn
	}
}

// StreamToAnthropic translates OpenAI streaming response to Anthropic SSE format
func (t *Translator) StreamToAnthropic(providerStream io.Reader, anthropicStream io.Writer) error {
	decoder := json.NewDecoder(providerStream)

	// Send message_start event
	if err := t.writeSSEEvent(anthropicStream, "message_start", map[string]interface{}{
		"type": "message",
		"message": map[string]interface{}{
			"id":   "msg_" + generateRandomID(),
			"type": "message",
			"role": "assistant",
			"content": []anthropic.ContentBlock{
				{Type: "text", Text: ""},
			},
			"model":        "",
			"stop_reason":  nil,
			"stop_sequence": nil,
			"usage": anthropic.Usage{
				InputTokens:  0,
				OutputTokens: 0,
			},
		},
	}); err != nil {
		return fmt.Errorf("failed to send message_start event: %w", err)
	}

	// Send content_block_start event
	if err := t.writeSSEEvent(anthropicStream, "content_block_start", map[string]interface{}{
		"index": 0,
		"content_block": map[string]interface{}{
			"type": "text",
			"text": "",
		},
	}); err != nil {
		return fmt.Errorf("failed to send content_block_start event: %w", err)
	}

	// Process OpenAI stream chunks
	for decoder.More() {
		var openaiChunk StreamChunk
		if err := decoder.Decode(&openaiChunk); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode OpenAI stream chunk: %w", err)
		}

		// Check if we have content delta
		if len(openaiChunk.Choices) > 0 {
			delta := openaiChunk.Choices[0].Message.Content
			if delta != "" {
				// Send content_block_delta event
				if err := t.writeSSEEvent(anthropicStream, "content_block_delta", map[string]interface{}{
					"index": 0,
					"delta": map[string]interface{}{
						"type": "text_delta",
						"text": delta,
					},
				}); err != nil {
					return fmt.Errorf("failed to send content_block_delta event: %w", err)
				}
			}

			// Check for finish reason
			if openaiChunk.Choices[0].FinishReason != nil {
				stopReason := t.translateFinishReason(openaiChunk.Choices[0].FinishReason)

				// Send content_block_stop event
				if err := t.writeSSEEvent(anthropicStream, "content_block_stop", map[string]interface{}{
					"index": 0,
				}); err != nil {
					return fmt.Errorf("failed to send content_block_stop event: %w", err)
				}

				// Send message_delta event with stop reason
				if err := t.writeSSEEvent(anthropicStream, "message_delta", map[string]interface{}{
					"stop_reason":  stopReason,
					"stop_sequence": nil,
				}); err != nil {
					return fmt.Errorf("failed to send message_delta event: %w", err)
				}

				// Send message_stop event
				if err := t.writeSSEEvent(anthropicStream, "message_stop", nil); err != nil {
					return fmt.Errorf("failed to send message_stop event: %w", err)
				}

				break
			}
		}
	}

	return nil
}

// writeSSEEvent writes a server-sent event
func (t *Translator) writeSSEEvent(w io.Writer, eventType string, data interface{}) error {
	var buf bytes.Buffer

	// Write event type
	if _, err := fmt.Fprintf(&buf, "event: %s\n", eventType); err != nil {
		return err
	}

	// Write event data
	if data != nil {
		dataBytes, err := json.Marshal(data)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(&buf, "data: %s\n", dataBytes); err != nil {
			return err
		}
	}

	// Write empty line to end event
	if _, err := fmt.Fprintln(&buf); err != nil {
		return err
	}

	_, err := w.Write(buf.Bytes())
	return err
}

// GetProvider returns the provider type
func (t *Translator) GetProvider() config.Provider {
	return config.ProviderOpenAI
}

// generateRandomID generates a random ID for messages
func generateRandomID() string {
	return strconv.FormatInt(int64(1000000+int(randomInt())), 10)
}

func randomInt() int {
	// Simple random number generator
	// In production, use crypto/rand or math/rand properly seeded
	return 0 // Placeholder
}

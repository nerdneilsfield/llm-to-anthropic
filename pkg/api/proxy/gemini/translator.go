package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/nerdneilsfield/llm-to-anthropic/internal/config"
	"github.com/nerdneilsfield/llm-to-anthropic/pkg/api/proxy/anthropic"
)

// Translator implements Anthropic to Gemini translation
type Translator struct{}

// NewTranslator creates a new Gemini translator
func NewTranslator() *Translator {
	return &Translator{}
}

// RequestToProvider translates Anthropic request to Gemini format
func (t *Translator) RequestToProvider(req *anthropic.MessageRequest) (interface{}, error) {
	geminiReq := &GenerateContentRequest{
		Contents: make([]Content, 0, len(req.Messages)),
	}

	// Convert Anthropic messages to Gemini contents
	for i, msg := range req.Messages {
		content, err := t.translateMessage(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to translate message at index %d: %w", i, err)
		}
		geminiReq.Contents = append(geminiReq.Contents, content)
	}

	// Set generation config
	genConfig := GenerationConfig{
		MaxOutputTokens: req.MaxTokens,
	}

	if req.Temperature != nil {
		genConfig.Temperature = *req.Temperature
	}
	if req.TopP != nil {
		genConfig.TopP = *req.TopP
	}
	if req.TopK != nil {
		genConfig.TopK = *req.TopK
	}
	if len(req.StopSequences) > 0 {
		genConfig.StopSequences = req.StopSequences
	}

	geminiReq.GenerationConfig = &genConfig

	return geminiReq, nil
}

// translateMessage translates a single message from Anthropic to Gemini format
func (t *Translator) translateMessage(msg anthropic.Message) (Content, error) {
	content := Content{
		Role: t.translateRole(msg.Role),
		Parts: make([]Part, 0),
	}

	// Handle content (can be string or []ContentBlock)
	switch c := msg.Content.(type) {
	case string:
		content.Parts = append(content.Parts, Part{
			Text: c,
		})
	case []interface{}:
		// Parse content blocks
		contentBlocks := make([]anthropic.ContentBlock, 0)
		for _, block := range c {
			blockBytes, err := json.Marshal(block)
			if err != nil {
				return Content{}, fmt.Errorf("failed to marshal content block: %w", err)
			}

			var contentBlock anthropic.ContentBlock
			if err := json.Unmarshal(blockBytes, &contentBlock); err != nil {
				return Content{}, fmt.Errorf("failed to unmarshal content block: %w", err)
			}
			contentBlocks = append(contentBlocks, contentBlock)
		}

		// Convert content blocks to parts
		for _, block := range contentBlocks {
			part, err := t.convertContentBlockToPart(block)
			if err != nil {
				return Content{}, fmt.Errorf("failed to convert content block: %w", err)
			}
			content.Parts = append(content.Parts, part)
		}
	default:
		return Content{}, fmt.Errorf("unsupported content type: %T", msg.Content)
	}

	return content, nil
}

// convertContentBlockToPart converts an Anthropic content block to a Gemini part
func (t *Translator) convertContentBlockToPart(block anthropic.ContentBlock) (Part, error) {
	switch block.Type {
	case "text":
		return Part{
			Text: block.Text,
		}, nil
	case "image":
		return Part{
			InlineData: &InlineData{
				MimeType: block.Source.MediaType,
				Data:     block.Source.Data,
			},
		}, nil
	default:
		return Part{}, fmt.Errorf("unsupported content block type: %s", block.Type)
	}
}

// translateRole translates Anthropic role to Gemini role
func (t *Translator) translateRole(role string) string {
	switch role {
	case "user":
		return "user"
	case "assistant":
		return "model"
	case "system":
		// Gemini doesn't have a separate system role
		// System messages should be handled differently
		return "user"
	default:
		return "user"
	}
}

// ResponseToAnthropic translates Gemini response to Anthropic format
func (t *Translator) ResponseToAnthropic(resp []byte) (*anthropic.MessageResponse, error) {
	var geminiResp GenerateContentResponse
	if err := json.Unmarshal(resp, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Gemini response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in response")
	}

	candidate := geminiResp.Candidates[0]

	// Extract content from candidate
	contentBlocks, err := t.extractContentBlocks(candidate.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to extract content blocks: %w", err)
	}

	// Create Anthropic response
	anthropicResp := &anthropic.MessageResponse{
		ID:   generateMessageID(),
		Type:  "message",
		Role:  "assistant",
		Content: contentBlocks,
		Model: "",
		StopReason: t.translateFinishReason(candidate.FinishReason),
	}

	if geminiResp.UsageMetadata != nil {
		anthropicResp.Usage = anthropic.Usage{
			InputTokens:  geminiResp.UsageMetadata.PromptTokenCount,
			OutputTokens: geminiResp.UsageMetadata.CandidatesTokenCount,
		}
	}

	return anthropicResp, nil
}

// extractContentBlocks extracts Anthropic content blocks from Gemini content
func (t *Translator) extractContentBlocks(content *Content) ([]anthropic.ContentBlock, error) {
	if content == nil {
		return []anthropic.ContentBlock{}, nil
	}

	blocks := make([]anthropic.ContentBlock, 0, len(content.Parts))

	for _, part := range content.Parts {
		switch {
		case part.Text != "":
			blocks = append(blocks, anthropic.ContentBlock{
				Type: "text",
				Text: part.Text,
			})
		case part.InlineData != nil:
			blocks = append(blocks, anthropic.ContentBlock{
				Type: "image",
				Source: &anthropic.ImageSource{
					Type:      "base64",
					MediaType: part.InlineData.MimeType,
					Data:      part.InlineData.Data,
				},
			})
		}
	}

	return blocks, nil
}

// translateFinishReason translates Gemini finish reason to Anthropic format
func (t *Translator) translateFinishReason(reason string) string {
	switch reason {
	case FinishReasonStop:
		return anthropic.StopReasonEndTurn
	case FinishReasonMaxTokens:
		return anthropic.StopReasonMaxTokens
	case FinishReasonSafety:
		return anthropic.StopReasonStopSequence
	case FinishReasonRecitation:
		return anthropic.StopReasonStopSequence
	default:
		return anthropic.StopReasonEndTurn
	}
}

// StreamToAnthropic translates Gemini streaming response to Anthropic SSE format
func (t *Translator) StreamToAnthropic(providerStream io.Reader, anthropicStream io.Writer) error {
	decoder := json.NewDecoder(providerStream)

	// Send message_start event
	if err := t.writeSSEEvent(anthropicStream, "message_start", map[string]interface{}{
		"type": "message",
		"message": map[string]interface{}{
			"id":   generateMessageID(),
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

	// Process Gemini stream chunks
	for decoder.More() {
		var geminiChunk StreamChunk
		if err := decoder.Decode(&geminiChunk); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode Gemini stream chunk: %w", err)
		}

		// Process candidates
		if len(geminiChunk.Candidates) > 0 {
			candidate := geminiChunk.Candidates[0]

			// Extract text delta
			if candidate.Content != nil && len(candidate.Content.Parts) > 0 {
				for _, part := range candidate.Content.Parts {
					if part.Text != "" {
						// Send content_block_delta event
						if err := t.writeSSEEvent(anthropicStream, "content_block_delta", map[string]interface{}{
							"index": 0,
							"delta": map[string]interface{}{
								"type": "text_delta",
								"text": part.Text,
							},
						}); err != nil {
							return fmt.Errorf("failed to send content_block_delta event: %w", err)
						}
					}
				}
			}

			// Check for finish reason
			if candidate.FinishReason != "" && candidate.FinishReason != FinishReasonUnspecified {
				stopReason := t.translateFinishReason(candidate.FinishReason)

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
	return config.ProviderGoogle
}

// generateMessageID generates a message ID
func generateMessageID() string {
	return "msg_" + randomString(8)
}

// randomString generates a random string of the given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)] // Simple random, use crypto/rand in production
	}
	return string(b)
}

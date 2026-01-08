package translators

import (
	"encoding/json"
	"fmt"

	"github.com/nerdneilsfield/llm-to-anthropic/pkg/api/proxy/anthropic"
)

// OpenAI Request/Response structures
type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	ID      string        `json:"id"`
	Object  string        `json:"object"`
	Created int64         `json:"created"`
	Model   string        `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
	Usage   OpenAIUsage   `json:"usage"`
}

type OpenAIChoice struct {
	Index        int          `json:"index"`
	Message      OpenAIMessage `json:"message"`
	FinishReason string       `json:"finish_reason"`
}

type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// TranslateAnthropicToOpenAI converts Anthropic request to OpenAI format
func TranslateAnthropicToOpenAI(req *anthropic.MessageRequest, modelName string) (*OpenAIRequest, error) {
	messages := make([]OpenAIMessage, 0, len(req.Messages))
	
	for _, msg := range req.Messages {
		content := ""
		// Handle both string and []ContentBlock content
		switch v := msg.Content.(type) {
		case string:
			content = v
		case []anthropic.ContentBlock:
			if len(v) > 0 {
				content = v[0].Text
			}
		}
		
		messages = append(messages, OpenAIMessage{
			Role:    msg.Role,
			Content: content,
		})
	}
	
	return &OpenAIRequest{
		Model:       modelName,
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Temperature: 0.7, // Default temperature
		Stream:      false,
	}, nil
}

// TranslateOpenAIToAnthropic converts OpenAI response to Anthropic format
func TranslateOpenAIToAnthropic(resp []byte) (*anthropic.MessageResponse, error) {
	var openaiResp OpenAIResponse
	if err := json.Unmarshal(resp, &openaiResp); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}
	
	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in OpenAI response")
	}
	
	choice := openaiResp.Choices[0]
	
	return &anthropic.MessageResponse{
		ID:      openaiResp.ID,
		Type:    "message",
		Role:    "assistant",
		Content: []anthropic.ContentBlock{
			{
				Type: "text",
				Text: choice.Message.Content,
			},
		},
		Model:       openaiResp.Model,
		StopReason:  choice.FinishReason,
		Usage: anthropic.Usage{
			InputTokens:  openaiResp.Usage.PromptTokens,
			OutputTokens: openaiResp.Usage.CompletionTokens,
		},
	}, nil
}

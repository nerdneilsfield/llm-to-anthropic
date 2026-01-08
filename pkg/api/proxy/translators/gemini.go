package translators

import (
	"encoding/json"
	"fmt"

	"github.com/nerdneilsfield/llm-to-anthropic/pkg/api/proxy/anthropic"
)

// Gemini Request/Response structures
type GeminiRequest struct {
	Contents         []GeminiContent          `json:"contents,omitempty"`
	GenerationConfig *GeminiGenerationConfig `json:"generationConfig,omitempty"`
	Stream           bool                     `json:"stream,omitempty"`
}

type GeminiContent struct {
	Role  string          `json:"role,omitempty"`  // "user" or "model"
	Parts []GeminiPart    `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text,omitempty"`
}

type GeminiGenerationConfig struct {
	Temperature float64 `json:"temperature,omitempty"`
	MaxTokens   int     `json:"maxOutputTokens,omitempty"`
	TopP        float64 `json:"topP,omitempty"`
	TopK        int     `json:"topK,omitempty"`
}

type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
	Usage      *GeminiUsage      `json:"usageMetadata,omitempty"`
}

type GeminiCandidate struct {
	Content GeminiContent `json:"content"`
	Finish string        `json:"finishReason,omitempty"`
}

type GeminiUsage struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount     int `json:"totalTokenCount"`
}

// TranslateAnthropicToGemini converts Anthropic request to Gemini format
func TranslateAnthropicToGemini(req *anthropic.MessageRequest, modelName string) (*GeminiRequest, error) {
	contents := make([]GeminiContent, 0, len(req.Messages))
	
	for _, msg := range req.Messages {
		// Handle both string and []ContentBlock content
		text := ""
		switch v := msg.Content.(type) {
		case string:
			text = v
		case []anthropic.ContentBlock:
			if len(v) > 0 && v[0].Type == "text" {
				text = v[0].Text
			}
		}
		
		// Map Anthropic roles to Gemini roles
		role := "user"
		if msg.Role == "assistant" {
			role = "model"
		}
		
		if text != "" {
			contents = append(contents, GeminiContent{
				Role: role,
				Parts: []GeminiPart{
					{Text: text},
				},
			})
		}
	}
	
	// Build generation config
	config := &GeminiGenerationConfig{
		Temperature: 0.7, // Default temperature
		MaxTokens:   req.MaxTokens,
	}
	
	// Use request temperature if provided
	if req.Temperature != nil {
		config.Temperature = *req.Temperature
	}
	
	return &GeminiRequest{
		Contents:         contents,
		GenerationConfig: config,
		Stream:           false,
	}, nil
}

// TranslateGeminiToAnthropic converts Gemini response to Anthropic format
func TranslateGeminiToAnthropic(resp []byte) (*anthropic.MessageResponse, error) {
	var geminiResp GeminiResponse
	if err := json.Unmarshal(resp, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to parse Gemini response: %w", err)
	}
	
	if len(geminiResp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in Gemini response")
	}
	
	candidate := geminiResp.Candidates[0]
	
	// Extract text from response
	text := ""
	if len(candidate.Content.Parts) > 0 {
		text = candidate.Content.Parts[0].Text
	}
	
	// Map usage
	usage := anthropic.Usage{}
	if geminiResp.Usage != nil {
		usage.InputTokens = geminiResp.Usage.PromptTokenCount
		usage.OutputTokens = geminiResp.Usage.CandidatesTokenCount
	}
	
	// Map finish reason
	stopReason := "end_turn"
	if candidate.Finish != "" {
		stopReason = candidate.Finish
	}
	
	return &anthropic.MessageResponse{
		Type: "message",
		Role: "assistant",
		Content: []anthropic.ContentBlock{
			{
				Type: "text",
				Text: text,
			},
		},
		StopReason: stopReason,
		Usage:      usage,
	}, nil
}

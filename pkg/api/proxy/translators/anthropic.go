package translators

import (
	"encoding/json"

	"github.com/nerdneilsfield/llm-to-anthropic/pkg/api/proxy/anthropic"
)

// TranslateAnthropicToAnthropic passes through Anthropic format
func TranslateAnthropicToAnthropic(req *anthropic.MessageRequest) (*anthropic.MessageRequest, error) {
	// Anthropic format - pass through directly
	// But we need to make sure model name doesn't include provider prefix
	// The caller will handle model name separation
	return req, nil
}

// TranslateAnthropicToAnthropicResponse parses Anthropic response
func TranslateAnthropicToAnthropicResponse(resp []byte) (*anthropic.MessageResponse, error) {
	var anthropicResp anthropic.MessageResponse
	if err := json.Unmarshal(resp, &anthropicResp); err != nil {
		return nil, err
	}
	return &anthropicResp, nil
}

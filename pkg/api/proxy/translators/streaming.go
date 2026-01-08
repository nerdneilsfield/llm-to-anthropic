package translators

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"

	"github.com/nerdneilsfield/llm-to-anthropic/pkg/provider/openai"
)

// TranslateOpenAIStreamToAnthropicSSE converts OpenAI SSE stream to Anthropic format
func TranslateOpenAIStreamToAnthropicSSE(stream io.Reader, w io.Writer) error {
	chunks, errs := openai.ParseOpenAIStream(stream)
	
	for {
		select {
		case chunk, ok := <-chunks:
			if !ok {
				chunks = nil
				break
			}
			
			if len(chunk.Choices) > 0 {
				choice := chunk.Choices[0]
				
				if choice.FinishReason != nil {
					delta := map[string]interface{}{
						"type": "message_stop",
						"stop_reason": *choice.FinishReason,
					}
					if err := writeSSE(w, delta); err != nil {
						return err
					}
				} else if choice.Delta.Content != "" {
					delta := map[string]interface{}{
						"type": "content_block_delta",
						"index": 0,
						"delta": map[string]string{
							"type": "text_delta",
							"text": choice.Delta.Content,
						},
					}
					if err := writeSSE(w, delta); err != nil {
						return err
					}
				}
			}
			
		case err, ok := <-errs:
			if !ok {
				errs = nil
				break
			}
			return err
		}
		
		if chunks == nil && errs == nil {
			break
		}
	}
	
	return nil
}

// TranslateAnthropicStreamToAnthropicSSE passes through Anthropic stream
func TranslateAnthropicStreamToAnthropicSSE(stream io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(stream)
	
	for scanner.Scan() {
		line := scanner.Text()
		
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		if _, err := w.Write([]byte(line + "\n\n")); err != nil {
			return err
		}
	}

	return scanner.Err()
}

// TranslateGeminiStreamToAnthropicSSE converts Gemini SSE stream to Anthropic format
func TranslateGeminiStreamToAnthropicSSE(stream io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(stream)
	
	for scanner.Scan() {
		line := scanner.Text()
		
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		var chunk map[string]interface{}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if candidates, ok := chunk["candidates"].([]interface{}); ok && len(candidates) > 0 {
			if candidate, ok := candidates[0].(map[string]interface{}); ok {
				if content, ok := candidate["content"].(map[string]interface{}); ok {
					if parts, ok := content["parts"].([]interface{}); ok && len(parts) > 0 {
						if part, ok := parts[0].(map[string]interface{}); ok {
							if text, ok := part["text"].(string); ok {
								delta := map[string]interface{}{
									"type": "content_block_delta",
									"index": 0,
									"delta": map[string]string{
										"type": "text_delta",
										"text": text,
									},
								}
								if err := writeSSE(w, delta); err != nil {
									return err
								}
							}
						}
					}
				}
				
				if finishReason, ok := candidate["finishReason"].(string); ok {
					delta := map[string]interface{}{
						"type": "message_stop",
						"stop_reason": finishReason,
					}
					if err := writeSSE(w, delta); err != nil {
						return err
					}
				}
			}
		}
	}

	return scanner.Err()
}

// writeSSE writes an SSE event
func writeSSE(w io.Writer, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = w.Write([]byte("data: " + string(jsonData) + "\n\n"))
	return err
}

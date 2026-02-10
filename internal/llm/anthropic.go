package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	AnthropicBaseURL      = "https://api.anthropic.com/v1"
	AnthropicDefaultModel = "claude-sonnet-4-20250514"
	AnthropicAPIVersion   = "2023-06-01"
)

var anthropicValidModels = map[string]bool{
	"claude-sonnet-4-20250514":    true,
	"claude-3-5-haiku-20241022":   true,
	"claude-opus-4-20250514":      true,
}

var anthropicDeprecatedModels = map[string]string{
	"claude-3-opus-20240229":   "claude-opus-4-20250514",
	"claude-3-sonnet-20240229": "claude-sonnet-4-20250514",
	"claude-3-haiku-20240307":  "claude-3-5-haiku-20241022",
}

func validateAnthropicModel(model string) string {
	if model == "" {
		return AnthropicDefaultModel
	}
	if anthropicValidModels[model] {
		return model
	}
	if replacement, ok := anthropicDeprecatedModels[model]; ok {
		return replacement
	}
	return AnthropicDefaultModel
}

func (c *Client) generateAnthropic(ctx context.Context, prompt string) (string, error) {
	url := fmt.Sprintf("%s/messages", AnthropicBaseURL)

	reqBody := anthropicRequest{
		Model: c.model,
		Messages: []anthropicMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens: 8192,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", AnthropicAPIVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling Anthropic API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Anthropic API error %d: %s", resp.StatusCode, string(body))
	}

	var anthropicResp anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	if len(anthropicResp.Content) == 0 {
		return "", fmt.Errorf("empty response from Anthropic")
	}

	// Concatenate all text blocks
	var result string
	for _, block := range anthropicResp.Content {
		if block.Type == "text" {
			result += block.Text
		}
	}

	if result == "" {
		return "", fmt.Errorf("no text content in Anthropic response")
	}

	return result, nil
}

// Anthropic API types
type anthropicRequest struct {
	Model     string             `json:"model"`
	Messages  []anthropicMessage `json:"messages"`
	MaxTokens int                `json:"max_tokens"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []anthropicContentBlock `json:"content"`
}

type anthropicContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

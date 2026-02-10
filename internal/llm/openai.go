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
	OpenAIBaseURL      = "https://api.openai.com/v1"
	OpenAIDefaultModel = "gpt-4o"
)

var openAIValidModels = map[string]bool{
	"gpt-4o":        true,
	"gpt-4o-mini":   true,
	"gpt-4-turbo":   true,
	"gpt-4.1":       true,
	"gpt-4.1-mini":  true,
	"gpt-4.1-nano":  true,
	"o3-mini":       true,
}

var openAIDeprecatedModels = map[string]string{
	"gpt-4":              "gpt-4o",
	"gpt-3.5-turbo":      "gpt-4o-mini",
	"gpt-4-turbo-preview": "gpt-4-turbo",
}

func validateOpenAIModel(model string) string {
	if model == "" {
		return OpenAIDefaultModel
	}
	if openAIValidModels[model] {
		return model
	}
	if replacement, ok := openAIDeprecatedModels[model]; ok {
		return replacement
	}
	return OpenAIDefaultModel
}

func (c *Client) generateOpenAI(ctx context.Context, prompt string) (string, error) {
	url := fmt.Sprintf("%s/chat/completions", OpenAIBaseURL)

	reqBody := openAIRequest{
		Model: c.model,
		Messages: []openAIMessage{
			{Role: "user", Content: prompt},
		},
		Temperature: floatPtr(0.2),
		MaxTokens:   8192,
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
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API error %d: %s", resp.StatusCode, string(body))
	}

	var openAIResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("empty response from OpenAI")
	}

	return openAIResp.Choices[0].Message.Content, nil
}

func floatPtr(f float64) *float64 {
	return &f
}

// OpenAI API types
type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	Temperature *float64        `json:"temperature,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []openAIChoice `json:"choices"`
}

type openAIChoice struct {
	Message openAIMessage `json:"message"`
}

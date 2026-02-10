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
	GeminiBaseURL      = "https://generativelanguage.googleapis.com/v1beta"
	GeminiDefaultModel = "gemini-2.0-flash"
)

var geminiValidModels = map[string]bool{
	"gemini-2.0-flash":       true,
	"gemini-2.5-flash":       true,
	"gemini-2.5-pro":         true,
	"gemini-3-flash-preview": true,
	"gemini-3-pro-preview":   true,
}

var geminiDeprecatedModels = map[string]string{
	"gemini-1.5-pro":   "gemini-2.0-flash",
	"gemini-1.5-flash": "gemini-2.0-flash",
	"gemini-pro":       "gemini-2.0-flash",
	"gemini-ultra":     "gemini-2.5-pro",
}

func validateGeminiModel(model string) string {
	if model == "" {
		return GeminiDefaultModel
	}
	if geminiValidModels[model] {
		return model
	}
	if replacement, ok := geminiDeprecatedModels[model]; ok {
		return replacement
	}
	return GeminiDefaultModel
}

func (c *Client) generateGemini(ctx context.Context, prompt string) (string, error) {
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", GeminiBaseURL, c.model, c.apiKey)

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: prompt}}},
		},
		GenerationConfig: &geminiGenerationConfig{
			Temperature:     0.2,
			MaxOutputTokens: 8192,
		},
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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling Gemini API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Gemini API error %d: %s", resp.StatusCode, string(body))
	}

	var geminiResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from Gemini")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

// Gemini API types
type geminiRequest struct {
	Contents         []geminiContent          `json:"contents"`
	GenerationConfig *geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	Temperature     float64 `json:"temperature"`
	MaxOutputTokens int     `json:"maxOutputTokens"`
}

type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
}

type geminiCandidate struct {
	Content geminiContent `json:"content"`
}

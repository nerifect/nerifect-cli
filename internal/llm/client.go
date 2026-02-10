package llm

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const (
	ProviderGemini    = "gemini"
	ProviderOpenAI    = "openai"
	ProviderAnthropic = "anthropic"
)

var ValidProviders = []string{ProviderGemini, ProviderOpenAI, ProviderAnthropic}

type Client struct {
	provider   string
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewClient(provider, apiKey, model string) *Client {
	if provider == "" {
		provider = ProviderGemini
	}
	model = ValidateModel(provider, model)
	return &Client{
		provider:   provider,
		apiKey:     apiKey,
		model:      model,
		httpClient: &http.Client{Timeout: 180 * time.Second},
	}
}

// GenerateContent sends a prompt to the configured LLM provider and returns the text response.
func (c *Client) GenerateContent(ctx context.Context, prompt string) (string, error) {
	switch c.provider {
	case ProviderOpenAI:
		return c.generateOpenAI(ctx, prompt)
	case ProviderAnthropic:
		return c.generateAnthropic(ctx, prompt)
	default:
		return c.generateGemini(ctx, prompt)
	}
}

// ValidateModel checks that the model is valid for the given provider, falling back to the default.
func ValidateModel(provider, model string) string {
	switch provider {
	case ProviderOpenAI:
		return validateOpenAIModel(model)
	case ProviderAnthropic:
		return validateAnthropicModel(model)
	default:
		return validateGeminiModel(model)
	}
}

// IsValidProvider returns true if the given provider name is supported.
func IsValidProvider(provider string) bool {
	for _, p := range ValidProviders {
		if p == provider {
			return true
		}
	}
	return false
}

// DefaultModelForProvider returns the default model for the given provider.
func DefaultModelForProvider(provider string) string {
	switch provider {
	case ProviderOpenAI:
		return OpenAIDefaultModel
	case ProviderAnthropic:
		return AnthropicDefaultModel
	default:
		return GeminiDefaultModel
	}
}

// ProviderLabel returns a human-readable label for a provider.
func ProviderLabel(provider string) string {
	switch provider {
	case ProviderOpenAI:
		return "OpenAI"
	case ProviderAnthropic:
		return "Anthropic"
	default:
		return "Google Gemini"
	}
}

// ValidModelsForProvider returns the list of valid models for a given provider.
func ValidModelsForProvider(provider string) []string {
	switch provider {
	case ProviderOpenAI:
		var models []string
		for m := range openAIValidModels {
			models = append(models, m)
		}
		return models
	case ProviderAnthropic:
		var models []string
		for m := range anthropicValidModels {
			models = append(models, m)
		}
		return models
	default:
		var models []string
		for m := range geminiValidModels {
			models = append(models, m)
		}
		return models
	}
}

// APIKeyEnvVar returns the environment variable name for the provider's API key.
func APIKeyEnvVar(provider string) string {
	switch provider {
	case ProviderOpenAI:
		return "OPENAI_API_KEY"
	case ProviderAnthropic:
		return "ANTHROPIC_API_KEY"
	default:
		return "GEMINI_API_KEY"
	}
}

// HasAPIKey checks if the client has a non-empty API key.
func (c *Client) HasAPIKey() bool {
	return c.apiKey != ""
}

// ErrMissingAPIKey returns a formatted error for a missing API key.
func ErrMissingAPIKey(provider string) error {
	return fmt.Errorf("%s is required for %s provider (set via config or env var)", APIKeyEnvVar(provider), ProviderLabel(provider))
}

package fixer

import (
	"context"
	"fmt"

	"github.com/nerifect/nerifect-cli/internal/llm"
)

// FixResult holds the output of LLM fix generation.
type FixResult struct {
	FixDescription string  `json:"fix_description"`
	FixDiff        string  `json:"fix_diff"`
	Confidence     float64 `json:"confidence"`
}

// Fixer generates AI-powered fixes for compliance violations.
type Fixer struct {
	client *llm.Client
}

func NewFixer(client *llm.Client) *Fixer {
	return &Fixer{client: client}
}

// GenerateFix produces a fix for a given violation.
func (f *Fixer) GenerateFix(ctx context.Context, ruleDesc, filePath, severity, violationDesc, fileContent string) (*FixResult, error) {
	prompt := llm.BuildFixPrompt(ruleDesc, filePath, severity, violationDesc, fileContent)

	responseText, err := f.client.GenerateContent(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("fix generation failed: %w", err)
	}

	var result FixResult
	if err := llm.ParseJSONResponse(responseText, &result); err != nil {
		// If JSON parsing fails, use the raw text as description
		return &FixResult{
			FixDescription: truncate(responseText, 200),
			Confidence:     0.5,
		}, nil
	}

	return &result, nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

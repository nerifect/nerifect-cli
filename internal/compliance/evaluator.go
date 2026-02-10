package compliance

import (
	"context"
	"fmt"

	"github.com/nerifect/nerifect-cli/internal/llm"
)

// Evaluator performs LLM-powered compliance evaluation.
type Evaluator struct {
	client          *llm.Client
	maxFilesPerBatch int
	maxCharsPerFile  int
}

func NewEvaluator(client *llm.Client) *Evaluator {
	return &Evaluator{
		client:          client,
		maxFilesPerBatch: 15,
		maxCharsPerFile:  4000,
	}
}

// EvaluationResult holds the output of a compliance evaluation.
type EvaluationResult struct {
	Violations      []ViolationResult
	ComplianceScore int
	FilesScanned    int
	RulesEvaluated  int
}

// Evaluate runs LLM-based compliance evaluation of files against policies.
func (e *Evaluator) Evaluate(ctx context.Context, policies []map[string]interface{}, files map[string]string) (*EvaluationResult, error) {
	if len(policies) == 0 {
		return &EvaluationResult{
			ComplianceScore: 100,
			FilesScanned:    len(files),
		}, nil
	}

	prompt := llm.BuildCompliancePrompt(policies, files, e.maxFilesPerBatch, e.maxCharsPerFile)

	responseText, err := e.client.GenerateContent(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM evaluation failed: %w", err)
	}

	violations, score := parseEvaluationResponse(responseText)

	// Recalculate score for consistency
	if len(violations) > 0 {
		score = CalculateScore(violations)
	}

	// Count rules evaluated
	rulesCount := 0
	for _, p := range policies {
		if rj, ok := p["rules_json"].(string); ok && rj != "" {
			rulesCount += 10 // Approximate
		}
	}

	return &EvaluationResult{
		Violations:      violations,
		ComplianceScore: score,
		FilesScanned:    len(files),
		RulesEvaluated:  rulesCount,
	}, nil
}

func parseEvaluationResponse(text string) ([]ViolationResult, int) {
	var result struct {
		Violations      []ViolationResult `json:"violations"`
		ComplianceScore int               `json:"compliance_score"`
	}

	if err := llm.ParseJSONResponse(text, &result); err != nil {
		return nil, 100
	}

	// Truncate code snippets
	for i := range result.Violations {
		if len(result.Violations[i].CodeSnippet) > 200 {
			result.Violations[i].CodeSnippet = result.Violations[i].CodeSnippet[:200]
		}
	}

	return result.Violations, result.ComplianceScore
}

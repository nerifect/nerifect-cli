package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nerifect/nerifect-cli/internal/llm"
)

// ParsedPolicy is the LLM-extracted policy structure.
type ParsedPolicy struct {
	RegulationName string       `json:"regulation_name"`
	RegulationType string       `json:"regulation_type"`
	Version        string       `json:"version"`
	Summary        string       `json:"summary"`
	Rules          []ParsedRule `json:"rules"`
}

// ParsedRule is a single compliance rule extracted by the LLM.
type ParsedRule struct {
	RuleID          string   `json:"rule_id"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	Severity        string   `json:"severity"`
	Category        string   `json:"category"`
	CheckType       string   `json:"check_type"`
	Pattern         string   `json:"pattern,omitempty"`
	Recommendations []string `json:"recommendations"`
	ClauseReference string   `json:"clause_reference,omitempty"`
	Topic           string   `json:"topic,omitempty"`
	SourceExcerpt   string   `json:"source_excerpt,omitempty"`
}

// Parser extracts compliance rules from regulation documents using LLM.
type Parser struct {
	client    *llm.Client
	chunkSize int
	overlap   int
}

func NewParser(client *llm.Client) *Parser {
	return &Parser{
		client:    client,
		chunkSize: 40000,
		overlap:   2000,
	}
}

// Parse processes document text and extracts structured policy rules.
func (p *Parser) Parse(ctx context.Context, text string) (*ParsedPolicy, error) {
	chunks := p.splitText(text)

	var results []*ParsedPolicy
	for i, chunk := range chunks {
		parsed, err := p.extractChunk(ctx, chunk, i)
		if err != nil {
			fmt.Fprintf(nil, "")
			continue
		}
		results = append(results, parsed)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no rules could be extracted from the document")
	}

	merged := p.mergePolicies(results)

	// Set metadata from first successful chunk
	for _, r := range results {
		if len(r.Rules) > 0 {
			merged.RegulationName = r.RegulationName
			merged.RegulationType = r.RegulationType
			merged.Version = r.Version
			break
		}
	}

	return merged, nil
}

func (p *Parser) extractChunk(ctx context.Context, chunk string, index int) (*ParsedPolicy, error) {
	prompt := llm.BuildPolicyExtractionPrompt(chunk)

	responseText, err := p.client.GenerateContent(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("chunk %d LLM call failed: %w", index, err)
	}

	var parsed ParsedPolicy
	if err := llm.ParseJSONResponse(responseText, &parsed); err != nil {
		return nil, fmt.Errorf("chunk %d JSON parse failed: %w", index, err)
	}

	return &parsed, nil
}

func (p *Parser) mergePolicies(policies []*ParsedPolicy) *ParsedPolicy {
	uniqueRules := make(map[string]ParsedRule)
	var allRules []ParsedRule

	for _, pol := range policies {
		for _, rule := range pol.Rules {
			rid := strings.ToUpper(strings.TrimSpace(rule.RuleID))
			if _, exists := uniqueRules[rid]; !exists {
				uniqueRules[rid] = rule
				allRules = append(allRules, rule)
			}
		}
	}

	return &ParsedPolicy{
		RegulationName: "Merged Policy",
		RegulationType: "OTHER",
		Version:        "1.0",
		Summary:        fmt.Sprintf("Aggregated from %d chunks, %d unique rules.", len(policies), len(allRules)),
		Rules:          allRules,
	}
}

func (p *Parser) splitText(text string) []string {
	if len(text) <= p.chunkSize {
		return []string{text}
	}

	var chunks []string
	start := 0
	for start < len(text) {
		end := start + p.chunkSize
		if end > len(text) {
			end = len(text)
		}
		chunks = append(chunks, text[start:end])
		start = end - p.overlap
		if start < 0 {
			start = 0
		}
	}
	return chunks
}

// RulesJSON converts parsed rules to the JSON format stored in the database.
func RulesJSON(rules []ParsedRule) string {
	data, _ := json.MarshalIndent(map[string]interface{}{"rules": rules}, "", "  ")
	return string(data)
}

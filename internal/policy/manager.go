package policy

import (
	"context"
	"fmt"
	"strings"

	"github.com/nerifect/nerifect-cli/internal/llm"
	"github.com/nerifect/nerifect-cli/internal/store"
)

// Manager handles policy lifecycle: add, list, remove.
type Manager struct {
	fetcher *Fetcher
	parser  *Parser
}

func NewManager(llmClient *llm.Client) *Manager {
	return &Manager{
		fetcher: NewFetcher(),
		parser:  NewParser(llmClient),
	}
}

// AddFromURL fetches a URL, parses it with LLM, and stores the policy.
func (m *Manager) AddFromURL(ctx context.Context, url string) (*store.Policy, error) {
	text, err := m.fetcher.FetchURL(url)
	if err != nil {
		return nil, fmt.Errorf("fetching document: %w", err)
	}

	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("document at %s is empty", url)
	}

	parsed, err := m.parser.Parse(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("parsing document: %w", err)
	}

	rulesJSON := RulesJSON(parsed.Rules)

	policy, err := store.CreatePolicy(
		parsed.RegulationName,
		parsed.Summary,
		store.PolicyCategoryCompliance,
		store.SeverityMedium,
		url,
		rulesJSON,
		parsed.RegulationType,
		len(parsed.Rules),
	)
	if err != nil {
		return nil, fmt.Errorf("saving policy: %w", err)
	}

	return policy, nil
}

// AddFromFile reads a local file, parses it with LLM, and stores the policy.
func (m *Manager) AddFromFile(ctx context.Context, path string) (*store.Policy, error) {
	text, err := m.fetcher.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("file %s is empty", path)
	}

	parsed, err := m.parser.Parse(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("parsing file: %w", err)
	}

	rulesJSON := RulesJSON(parsed.Rules)

	policy, err := store.CreatePolicy(
		parsed.RegulationName,
		parsed.Summary,
		store.PolicyCategoryCompliance,
		store.SeverityMedium,
		path,
		rulesJSON,
		parsed.RegulationType,
		len(parsed.Rules),
	)
	if err != nil {
		return nil, fmt.Errorf("saving policy: %w", err)
	}

	return policy, nil
}

// List returns all stored policies.
func (m *Manager) List() ([]store.Policy, error) {
	return store.ListPolicies()
}

// Remove deletes a policy by ID.
func (m *Manager) Remove(id int64) error {
	return store.DeletePolicy(id)
}

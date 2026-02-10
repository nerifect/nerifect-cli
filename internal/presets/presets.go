package presets

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/nerifect/nerifect-cli/internal/store"
)

// Rule represents a single compliance rule within a preset.
type Rule struct {
	RuleID          string   `json:"rule_id"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	Severity        string   `json:"severity"`
	Category        string   `json:"category"`
	CheckType       string   `json:"check_type"`
	Pattern         string   `json:"pattern"`
	Recommendations []string `json:"recommendations"`
	ClauseReference string   `json:"clause_reference,omitempty"`
}

// Preset represents a built-in compliance framework rule pack.
type Preset struct {
	Name           string
	Slug           string
	Description    string
	Category       store.PolicyCategory
	Severity       store.Severity
	RegulationType string
	Rules          []Rule
}

// registry holds all built-in presets keyed by slug.
var registry = map[string]Preset{}

func register(p Preset) {
	registry[p.Slug] = p
}

// List returns all available presets sorted by slug.
func List() []Preset {
	var presets []Preset
	for _, p := range registry {
		presets = append(presets, p)
	}
	sort.Slice(presets, func(i, j int) bool {
		return presets[i].Slug < presets[j].Slug
	})
	return presets
}

// Get returns a preset by slug.
func Get(slug string) (Preset, bool) {
	p, ok := registry[slug]
	return p, ok
}

// Install loads a preset into the database as a policy.
func Install(slug string) (*store.Policy, error) {
	p, ok := registry[slug]
	if !ok {
		return nil, fmt.Errorf("unknown preset: %q (use 'nerifect policy list-presets' to see available presets)", slug)
	}

	rulesJSON, err := json.Marshal(map[string]interface{}{
		"rules": p.Rules,
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling rules: %w", err)
	}

	return store.CreatePolicy(
		p.Name,
		p.Description,
		p.Category,
		p.Severity,
		"builtin://"+p.Slug,
		string(rulesJSON),
		p.RegulationType,
		len(p.Rules),
	)
}

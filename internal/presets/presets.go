package presets

import (
	"embed"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/nerifect/nerifect-cli/internal/store"
	"gopkg.in/yaml.v3"
)

//go:embed data/*.yaml
var presetFiles embed.FS

// Rule represents a single compliance rule within a preset.
type Rule struct {
	RuleID          string   `json:"rule_id" yaml:"rule_id"`
	Title           string   `json:"title" yaml:"title"`
	Description     string   `json:"description" yaml:"description"`
	Severity        string   `json:"severity" yaml:"severity"`
	Category        string   `json:"category" yaml:"category"`
	CheckType       string   `json:"check_type" yaml:"check_type"`
	Pattern         string   `json:"pattern" yaml:"pattern"`
	Recommendations []string `json:"recommendations" yaml:"recommendations"`
	ClauseReference string   `json:"clause_reference,omitempty" yaml:"clause_reference,omitempty"`
}

// Preset represents a built-in compliance framework rule pack.
type Preset struct {
	Name           string               `yaml:"name"`
	Slug           string               `yaml:"slug"`
	Description    string               `yaml:"description"`
	Category       store.PolicyCategory `yaml:"category"`
	Severity       store.Severity       `yaml:"severity"`
	RegulationType string               `yaml:"regulation_type"`
	Rules          []Rule               `yaml:"rules"`
}

// registry holds all built-in presets keyed by slug.
var registry = map[string]Preset{}

func init() {
	entries, err := presetFiles.ReadDir("data")
	if err != nil {
		panic(fmt.Sprintf("reading embedded preset directory: %v", err))
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := presetFiles.ReadFile("data/" + entry.Name())
		if err != nil {
			panic(fmt.Sprintf("reading embedded preset %s: %v", entry.Name(), err))
		}
		var p Preset
		if err := yaml.Unmarshal(data, &p); err != nil {
			panic(fmt.Sprintf("parsing preset %s: %v", entry.Name(), err))
		}
		registry[p.Slug] = p
	}
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

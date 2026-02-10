package output

import (
	"fmt"
	"strings"

	"github.com/nerifect/nerifect-cli/internal/store"
)

// SARIF 2.1.0 types

type sarifReport struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name            string      `json:"name"`
	Version         string      `json:"version"`
	InformationURI  string      `json:"informationUri"`
	Rules           []sarifRule `json:"rules"`
}

type sarifRule struct {
	ID               string            `json:"id"`
	ShortDescription sarifMessage      `json:"shortDescription"`
	FullDescription  *sarifMessage     `json:"fullDescription,omitempty"`
	DefaultConfig    sarifRuleConfig   `json:"defaultConfiguration"`
	Properties       map[string]string `json:"properties,omitempty"`
}

type sarifRuleConfig struct {
	Level string `json:"level"`
}

type sarifMessage struct {
	Text string `json:"text"`
}

type sarifResult struct {
	RuleID    string           `json:"ruleId"`
	RuleIndex int              `json:"ruleIndex"`
	Level     string           `json:"level"`
	Message   sarifMessage     `json:"message"`
	Locations []sarifLocation  `json:"locations,omitempty"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
	Region           *sarifRegion          `json:"region,omitempty"`
}

type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	StartLine int            `json:"startLine,omitempty"`
	EndLine   int            `json:"endLine,omitempty"`
	Snippet   *sarifSnippet  `json:"snippet,omitempty"`
}

type sarifSnippet struct {
	Text string `json:"text"`
}

// severityToSARIF maps nerifect severity to SARIF level and security-severity score.
func severityToSARIF(severity string) (string, string) {
	switch strings.ToUpper(severity) {
	case "CRITICAL":
		return "error", "9.0"
	case "HIGH":
		return "error", "7.0"
	case "MEDIUM":
		return "warning", "5.0"
	case "LOW":
		return "note", "3.0"
	default:
		return "note", "1.0"
	}
}

// RenderSARIF outputs scan results in SARIF 2.1.0 JSON format.
func RenderSARIF(scan *store.Scan, violations []store.Violation, detections []store.AIDetection) {
	// Build unique rules from violations
	ruleIndex := map[string]int{}
	var rules []sarifRule

	for _, v := range violations {
		if _, exists := ruleIndex[v.RuleID]; exists {
			continue
		}
		level, secSeverity := severityToSARIF(string(v.Severity))
		rule := sarifRule{
			ID:               v.RuleID,
			ShortDescription: sarifMessage{Text: v.Title},
			DefaultConfig:    sarifRuleConfig{Level: level},
			Properties: map[string]string{
				"security-severity": secSeverity,
			},
		}
		if v.Description != "" {
			rule.FullDescription = &sarifMessage{Text: v.Description}
		}
		ruleIndex[v.RuleID] = len(rules)
		rules = append(rules, rule)
	}

	// Add rules for AI detections
	for _, d := range detections {
		ruleID := "ai-detection/" + strings.ToLower(strings.ReplaceAll(d.Name, " ", "-"))
		if _, exists := ruleIndex[ruleID]; exists {
			continue
		}
		level, secSeverity := severityToSARIF(d.RiskLevel)
		rule := sarifRule{
			ID:               ruleID,
			ShortDescription: sarifMessage{Text: fmt.Sprintf("AI/ML framework detected: %s", d.Name)},
			DefaultConfig:    sarifRuleConfig{Level: level},
			Properties: map[string]string{
				"security-severity": secSeverity,
				"eu-ai-act-risk":    d.EUAIActRisk,
			},
		}
		ruleIndex[ruleID] = len(rules)
		rules = append(rules, rule)
	}

	// Build results from violations
	var results []sarifResult
	for _, v := range violations {
		idx := ruleIndex[v.RuleID]
		level, _ := severityToSARIF(string(v.Severity))

		msg := v.Title
		if v.Recommendation != "" {
			msg += ". " + v.Recommendation
		}

		result := sarifResult{
			RuleID:    v.RuleID,
			RuleIndex: idx,
			Level:     level,
			Message:   sarifMessage{Text: msg},
		}

		if v.FilePath != "" {
			loc := sarifLocation{
				PhysicalLocation: sarifPhysicalLocation{
					ArtifactLocation: sarifArtifactLocation{URI: v.FilePath},
				},
			}
			if v.LineStart > 0 {
				region := &sarifRegion{StartLine: v.LineStart}
				if v.LineEnd > 0 {
					region.EndLine = v.LineEnd
				}
				if v.CodeSnippet != "" {
					region.Snippet = &sarifSnippet{Text: v.CodeSnippet}
				}
				loc.PhysicalLocation.Region = region
			}
			result.Locations = []sarifLocation{loc}
		}

		results = append(results, result)
	}

	// Build results from AI detections
	for _, d := range detections {
		ruleID := "ai-detection/" + strings.ToLower(strings.ReplaceAll(d.Name, " ", "-"))
		idx := ruleIndex[ruleID]
		level, _ := severityToSARIF(d.RiskLevel)

		msg := fmt.Sprintf("Detected %s (%s) - Risk: %s, EU AI Act: %s",
			d.Name, d.Type, d.RiskLevel, d.EUAIActRisk)

		result := sarifResult{
			RuleID:    ruleID,
			RuleIndex: idx,
			Level:     level,
			Message:   sarifMessage{Text: msg},
		}

		if d.FilePath != "" {
			result.Locations = []sarifLocation{
				{
					PhysicalLocation: sarifPhysicalLocation{
						ArtifactLocation: sarifArtifactLocation{URI: d.FilePath},
					},
				},
			}
		}

		results = append(results, result)
	}

	report := sarifReport{
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
		Version: "2.1.0",
		Runs: []sarifRun{
			{
				Tool: sarifTool{
					Driver: sarifDriver{
						Name:           "Nerifect CLI",
						Version:        "0.1.0",
						InformationURI: "https://github.com/nerifect/nerifect-cli",
						Rules:          rules,
					},
				},
				Results: results,
			},
		},
	}

	PrintJSON(report)
}

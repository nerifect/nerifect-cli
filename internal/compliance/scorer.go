package compliance

import "strings"

// SeverityWeights maps severity levels to score penalties.
var SeverityWeights = map[string]int{
	"CRITICAL": 25,
	"HIGH":     15,
	"MEDIUM":   8,
	"LOW":      3,
	"INFO":     1,
}

// CalculateScore computes compliance score: 100 minus severity-weighted penalties per unique rule.
func CalculateScore(violations []ViolationResult) int {
	seenRules := make(map[string]bool)
	penalty := 0
	for _, v := range violations {
		if seenRules[v.RuleID] {
			continue
		}
		seenRules[v.RuleID] = true
		w, ok := SeverityWeights[strings.ToUpper(v.Severity)]
		if !ok {
			w = 8
		}
		penalty += w
	}
	score := 100 - penalty
	if score < 0 {
		return 0
	}
	return score
}

// ViolationResult is used by the scorer and evaluator.
type ViolationResult struct {
	RuleID          string `json:"rule_id"`
	PolicyName      string `json:"policy_name"`
	Severity        string `json:"severity"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	FilePath        string `json:"file_path"`
	LineStart       int    `json:"line_start"`
	LineEnd         int    `json:"line_end"`
	CodeSnippet     string `json:"code_snippet"`
	ClauseReference string `json:"clause_reference"`
	Recommendation  string `json:"recommendation"`
}

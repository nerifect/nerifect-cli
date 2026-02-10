package compliance

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gobwas/glob"
	"github.com/nerifect/nerifect-cli/internal/store"
)

// PatternChecker evaluates policy rules using regex/glob pattern matching.
type PatternChecker struct{}

func NewPatternChecker() *PatternChecker {
	return &PatternChecker{}
}

// Check evaluates rules against file paths and contents. Returns violations found.
func (pc *PatternChecker) Check(rules []store.PolicyRule, files map[string]string, allPaths []string) []ViolationResult {
	var violations []ViolationResult
	regexCache := make(map[string]*regexp.Regexp)

	for _, rule := range rules {
		ct := strings.ToUpper(rule.CheckType)
		pattern := strings.TrimSpace(rule.Pattern)
		if pattern == "" {
			continue
		}

		switch ct {
		case "FILE_PATTERN":
			violations = append(violations, pc.checkFilePattern(rule, pattern, allPaths)...)
		case "CODE_PATTERN", "CONFIG_CHECK":
			violations = append(violations, pc.checkCodePattern(rule, pattern, files, regexCache)...)
		}
	}
	return violations
}

func (pc *PatternChecker) checkFilePattern(rule store.PolicyRule, pattern string, paths []string) []ViolationResult {
	var violations []ViolationResult

	if strings.HasPrefix(strings.ToLower(pattern), "missing:") {
		globPat := strings.TrimSpace(pattern[8:])
		if !anyPathMatches(paths, globPat) {
			violations = append(violations, makeViolation(rule, globPat, ""))
		}
	} else {
		for _, p := range paths {
			if pathMatches(p, pattern) {
				violations = append(violations, makeViolation(rule, p, ""))
			}
		}
	}
	return violations
}

func (pc *PatternChecker) checkCodePattern(rule store.PolicyRule, pattern string, files map[string]string, cache map[string]*regexp.Regexp) []ViolationResult {
	if strings.HasPrefix(strings.ToLower(pattern), "missing:") {
		return nil
	}

	var violations []ViolationResult
	re, ok := cache[pattern]
	if !ok {
		var err error
		re, err = regexp.Compile("(?im)" + pattern)
		if err != nil {
			return nil
		}
		cache[pattern] = re
	}

	for path, content := range files {
		loc := re.FindStringIndex(content)
		if loc != nil {
			end := loc[0] + 200
			if end > len(content) {
				end = len(content)
			}
			snippet := content[loc[0]:end]
			violations = append(violations, makeViolation(rule, path, snippet))
		}
	}
	return violations
}

func makeViolation(rule store.PolicyRule, filePath, snippet string) ViolationResult {
	rec := ""
	if len(rule.Recommendations) > 0 {
		rec = rule.Recommendations[0]
	}
	return ViolationResult{
		RuleID:          rule.RuleID,
		PolicyName:      rule.PolicyName,
		Severity:        strings.ToUpper(rule.Severity),
		Title:           rule.Title,
		Description:     rule.Description,
		FilePath:        filePath,
		CodeSnippet:     truncSnippet(snippet, 200),
		ClauseReference: rule.ClauseReference,
		Recommendation:  rec,
	}
}

func anyPathMatches(paths []string, pattern string) bool {
	g, err := glob.Compile(pattern)
	if err != nil {
		// Fallback to simple contains
		for _, p := range paths {
			if strings.Contains(p, pattern) {
				return true
			}
		}
		return false
	}
	for _, p := range paths {
		if g.Match(p) || g.Match(filepath.Base(p)) {
			return true
		}
	}
	return false
}

func pathMatches(path, pattern string) bool {
	g, err := glob.Compile(pattern)
	if err != nil {
		return strings.Contains(path, pattern)
	}
	return g.Match(path) || g.Match(filepath.Base(path))
}

func truncSnippet(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

package fixer

import (
	"regexp"
	"strings"
)

// ApplyDiff attempts to apply a diff to original content.
// Handles unified diff format, full replacement code, and markdown blocks.
func ApplyDiff(original, fixDiff string) string {
	cleaned := strings.TrimSpace(fixDiff)

	// Extract from markdown code blocks
	codeBlockRe := regexp.MustCompile(`(?s)` + "```(?:\\w+)?\\s*\n(.*?)\n```")
	if m := codeBlockRe.FindStringSubmatch(cleaned); len(m) > 1 {
		cleaned = strings.TrimSpace(m[1])
	}

	// Check if it's unified diff format
	if strings.Contains(cleaned, "@@") || strings.Count(cleaned, "\n+") > 2 || strings.Count(cleaned, "\n-") > 2 {
		result, err := applyUnifiedDiff(original, cleaned)
		if err == nil {
			return result
		}
	}

	// Check if it looks like replacement code (not description text)
	if !strings.HasPrefix(cleaned, "---") && !strings.HasPrefix(cleaned, "+++") && !strings.HasPrefix(cleaned, "diff ") && len(cleaned) > 50 {
		lines := strings.Split(cleaned, "\n")
		maxLines := 10
		if len(lines) < maxLines {
			maxLines = len(lines)
		}
		totalWords := 0
		for _, line := range lines[:maxLines] {
			totalWords += len(strings.Fields(line))
		}
		avgWords := float64(totalWords) / float64(maxLines)

		specialChars := 0
		end := 500
		if len(cleaned) < end {
			end = len(cleaned)
		}
		for _, c := range cleaned[:end] {
			switch c {
			case '{', '}', '[', ']', '(', ')', ';', '=', '<', '>', ':':
				specialChars++
			}
		}

		if avgWords < 8 || specialChars > 20 {
			return cleaned
		}
	}

	// Return original if unparseable
	return original
}

func applyUnifiedDiff(original, diffText string) (string, error) {
	resultLines := strings.Split(original, "\n")

	hunkPattern := regexp.MustCompile(`@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@`)
	lines := strings.Split(diffText, "\n")
	offset := 0

	i := 0
	for i < len(lines) {
		line := lines[i]

		match := hunkPattern.FindStringSubmatch(line)
		if match != nil {
			oldStart := 0
			fmt_sscanf(match[1], &oldStart)
			oldStart = oldStart - 1 + offset
			i++

			for i < len(lines) && !strings.HasPrefix(lines[i], "@@") {
				hunkLine := lines[i]
				if strings.HasPrefix(hunkLine, "-") {
					if oldStart >= 0 && oldStart < len(resultLines) {
						if strings.TrimSpace(resultLines[oldStart]) == strings.TrimSpace(hunkLine[1:]) {
							resultLines = append(resultLines[:oldStart], resultLines[oldStart+1:]...)
							offset--
						} else {
							oldStart++
						}
					} else {
						oldStart++
					}
				} else if strings.HasPrefix(hunkLine, "+") {
					newLine := hunkLine[1:]
					if oldStart > len(resultLines) {
						resultLines = append(resultLines, newLine)
					} else {
						resultLines = append(resultLines[:oldStart+1], resultLines[oldStart:]...)
						resultLines[oldStart] = newLine
					}
					oldStart++
					offset++
				} else if strings.HasPrefix(hunkLine, " ") || strings.TrimSpace(hunkLine) == "" {
					oldStart++
				}
				i++
			}
		} else {
			i++
		}
	}

	return strings.Join(resultLines, "\n"), nil
}

func fmt_sscanf(s string, v *int) {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	*v = n
}

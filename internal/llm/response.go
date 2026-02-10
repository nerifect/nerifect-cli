package llm

import (
	"encoding/json"
	"regexp"
	"strings"
)

var jsonFenceRe = regexp.MustCompile("(?s)```(?:json)?\\s*\n?(.*?)```")

// ExtractJSON strips markdown code fences and extracts JSON from LLM response text.
func ExtractJSON(text string) string {
	clean := strings.TrimSpace(text)

	// Try extracting from code fence
	if m := jsonFenceRe.FindStringSubmatch(clean); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}

	// Strip leading/trailing fences manually
	if strings.HasPrefix(clean, "```json") {
		clean = clean[7:]
	} else if strings.HasPrefix(clean, "```") {
		clean = clean[3:]
	}
	if strings.HasSuffix(clean, "```") {
		clean = clean[:len(clean)-3]
	}

	return strings.TrimSpace(clean)
}

// ParseJSONResponse extracts JSON from LLM text and unmarshals into target.
func ParseJSONResponse(text string, target interface{}) error {
	jsonStr := ExtractJSON(text)
	return json.Unmarshal([]byte(jsonStr), target)
}

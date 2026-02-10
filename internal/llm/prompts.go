package llm

import (
	"encoding/json"
	"fmt"
)

// ComplianceEvaluationPrompt is ported from backend/ai/compliance_evaluator.py
const ComplianceEvaluationPrompt = `You are a compliance expert evaluating source code against governance policies.

## Policy Rules
%s

## Source Files to Analyze
%s

## Instructions
1. Analyze each source file against the policy rules
2. Identify SPECIFIC violations in the code
3. Reference the exact policy clause being violated
4. Provide actionable recommendations to fix each violation

## Output Format
Return ONLY a valid JSON object with this exact structure:
{
  "violations": [
    {
      "rule_id": "unique rule ID from policy",
      "policy_name": "name of policy document",
      "severity": "CRITICAL|HIGH|MEDIUM|LOW",
      "title": "short violation title",
      "description": "detailed explanation of what code violates the policy",
      "file_path": "path/to/file.ext",
      "code_snippet": "the violating code (max 150 chars)",
      "clause_reference": "Section X.Y.Z of the policy",
      "recommendation": "specific fix recommendation"
    }
  ],
  "compliance_score": 0
}

IMPORTANT:
- Only report ACTUAL violations found in the code, not hypothetical issues
- If no violations found, return empty violations array and score of 100
- The compliance_score should be 100 minus penalties (CRITICAL=-25, HIGH=-15, MEDIUM=-8, LOW=-3 per unique rule)
- Return ONLY valid JSON, no markdown or explanations`

// FixGenerationPrompt is ported from backend/ai/main.py
const FixGenerationPrompt = `Analyze the following code violation and generate a fix.

Rule: %s
File: %s
Severity: %s

Violation description: %s

Code context:
%s

Return ONLY a valid JSON object with this exact structure:
{
  "fix_description": "clear explanation of what changes are needed and why",
  "fix_diff": "unified diff format showing the changes",
  "confidence": 0.0
}

IMPORTANT:
- The fix_diff should be in unified diff format with - and + lines
- confidence should be between 0.0 and 1.0
- Return ONLY valid JSON, no markdown or explanations`

// PolicyExtractionPrompt is ported from backend/ai/policy_parser.py
const PolicyExtractionPrompt = `You are a compliance expert. Analyze the following PARTIAL SEGMENT of a regulation document and extract compliance rules.

## INSTRUCTIONS
1. Analyze the text provided inside the <document_segment> tags below.
2. Extract every compliance rule fully contained or significantly present in this chunk.
3. Your goal is EXHAUSTIVE EXTRACTION.
4. For each rule, you must cite the exact clause/section/title.
5. Do NOT extract these instructions as rules. Only extract content from the <document_segment>.

## RULE STRUCTURE
For each rule, provide:
1. A unique rule_id (e.g., "GDPR-A5-1")
2. Title
3. Description
4. Severity: CRITICAL, HIGH, MEDIUM, LOW, INFO
5. Category: DATA_PROTECTION, CONSENT, RETENTION, SECURITY, ACCESS_CONTROL, LOGGING, ENCRYPTION, etc.
6. Check type: FILE_PATTERN, CODE_PATTERN, CONFIG_CHECK, MANUAL
7. Pattern (regex/glob) if applicable
8. Recommendations
9. clause_reference
10. topic
11. source_excerpt

## DOCUMENT TO ANALYZE
<document_segment>
%s
</document_segment>

## OUTPUT FORMAT
Respond in valid JSON format matching this schema:
{
  "regulation_name": "string (extract from context if possible, else 'Part')",
  "regulation_type": "string",
  "version": "string or null",
  "summary": "Brief summary of checks in this part",
  "rules": [ ... ]
}`

// AIGovernancePrompt is ported from backend/ai/llm.py
const AIGovernancePrompt = `Analyze the following AI/ML components detected in a repository for governance compliance:

Detected AI Components:
%s

Assess each component against:
1. **EU AI Act Risk Classification**:
   - HIGH-RISK: AI in critical infrastructure, education, employment, law enforcement, biometrics
   - LIMITED-RISK: Chatbots, emotion recognition, deepfakes (require transparency)
   - MINIMAL-RISK: AI-enabled games, spam filters

2. **GDPR Art. 22 Compliance** (Automated Decision Making):
   - Does the AI make decisions significantly affecting individuals?
   - Is there human oversight capability?
   - Can users request human intervention?

3. **Bias and Fairness**:
   - Training data diversity considerations
   - Potential for discriminatory outcomes
   - Fairness metrics implementation

4. **Transparency and Explainability**:
   - Model documentation present
   - Explainability capabilities (SHAP, LIME, etc.)
   - User notification of AI interaction

Return ONLY a valid JSON array with format:
[
  {
    "name": "component name",
    "status": "COMPLIANT|REVIEW_REQUIRED|NON_COMPLIANT",
    "risk_level": "HIGH|MEDIUM|LOW",
    "issues": 0,
    "eu_ai_act_risk": "HIGH-RISK|LIMITED-RISK|MINIMAL-RISK",
    "reasoning": "brief explanation"
  }
]`

// BuildCompliancePrompt formats the evaluation prompt with policies and files.
func BuildCompliancePrompt(policies []map[string]interface{}, files map[string]string, maxFiles, maxCharsPerFile int) string {
	// Prepare policy rules summary
	var policiesSummary []map[string]interface{}
	for _, p := range policies {
		rulesJSON, ok := p["rules_json"].(string)
		if !ok || rulesJSON == "" {
			continue
		}
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(rulesJSON), &parsed); err != nil {
			continue
		}
		rules, _ := parsed["rules"].([]interface{})
		for _, r := range rules {
			if rule, ok := r.(map[string]interface{}); ok {
				policiesSummary = append(policiesSummary, map[string]interface{}{
					"rule_id":          rule["rule_id"],
					"policy_name":      p["name"],
					"title":            rule["title"],
					"description":      rule["description"],
					"severity":         rule["severity"],
					"category":         rule["category"],
					"clause_reference": rule["clause_reference"],
				})
			}
		}
	}
	if len(policiesSummary) > 30 {
		policiesSummary = policiesSummary[:30]
	}

	// Prepare files summary
	var filesSummary []map[string]string
	count := 0
	for path, content := range files {
		if count >= maxFiles {
			break
		}
		if len(content) > maxCharsPerFile {
			content = content[:maxCharsPerFile]
		}
		filesSummary = append(filesSummary, map[string]string{
			"path":    path,
			"content": content,
		})
		count++
	}

	policiesJSON, _ := json.MarshalIndent(policiesSummary, "", "  ")
	filesJSON, _ := json.MarshalIndent(filesSummary, "", "  ")

	return fmt.Sprintf(ComplianceEvaluationPrompt, string(policiesJSON), string(filesJSON))
}

// BuildFixPrompt formats the fix generation prompt.
func BuildFixPrompt(ruleDesc, filePath, severity, violationDesc, fileContent string) string {
	if len(fileContent) > 4000 {
		fileContent = fileContent[:4000]
	}
	return fmt.Sprintf(FixGenerationPrompt, ruleDesc, filePath, severity, violationDesc, fileContent)
}

// BuildPolicyExtractionPrompt formats the policy extraction prompt for a document chunk.
func BuildPolicyExtractionPrompt(documentChunk string) string {
	return fmt.Sprintf(PolicyExtractionPrompt, documentChunk)
}

// BuildAIGovernancePrompt formats the AI governance assessment prompt.
func BuildAIGovernancePrompt(modelsSummary string) string {
	return fmt.Sprintf(AIGovernancePrompt, modelsSummary)
}

package store

import (
	"database/sql"
	"encoding/json"
	"time"
)

func CreatePolicy(name, description string, category PolicyCategory, severity Severity, sourceURL, rulesJSON, regulationType string, ruleCount int) (*Policy, error) {
	now := time.Now()
	result, err := db.Exec(
		`INSERT INTO policies (name, description, category, severity, source_url, rules_json, regulation_type, rule_count, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		name, description, string(category), string(severity), sourceURL, rulesJSON, regulationType, ruleCount, now, now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return &Policy{
		ID:             id,
		Name:           name,
		Description:    description,
		Category:       category,
		Severity:       severity,
		SourceURL:      sourceURL,
		RulesJSON:      rulesJSON,
		RegulationType: regulationType,
		RuleCount:      ruleCount,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

func ListPolicies() ([]Policy, error) {
	rows, err := db.Query(`SELECT id, name, description, category, severity, source_url, rules_json, regulation_type, rule_count, created_at, updated_at FROM policies ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []Policy
	for rows.Next() {
		p := Policy{}
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Category, &p.Severity,
			&p.SourceURL, &p.RulesJSON, &p.RegulationType, &p.RuleCount, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}
	return policies, nil
}

func GetPolicy(id int64) (*Policy, error) {
	row := db.QueryRow(`SELECT id, name, description, category, severity, source_url, rules_json, regulation_type, rule_count, created_at, updated_at FROM policies WHERE id = ?`, id)
	p := &Policy{}
	err := row.Scan(&p.ID, &p.Name, &p.Description, &p.Category, &p.Severity,
		&p.SourceURL, &p.RulesJSON, &p.RegulationType, &p.RuleCount, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func DeletePolicy(id int64) error {
	result, err := db.Exec(`DELETE FROM policies WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// GetAllPoliciesForScan returns policies in the format needed by the compliance evaluator.
func GetAllPoliciesForScan() ([]map[string]interface{}, error) {
	policies, err := ListPolicies()
	if err != nil {
		return nil, err
	}
	var result []map[string]interface{}
	for _, p := range policies {
		result = append(result, map[string]interface{}{
			"id":         p.ID,
			"name":       p.Name,
			"rules_json": p.RulesJSON,
		})
	}
	return result, nil
}

// ExtractRulesFromPolicies extracts flat rule list from policies for pattern scanning.
type PolicyRule struct {
	PolicyID        int64
	PolicyName      string
	RuleID          string
	Title           string
	Description     string
	Severity        string
	Category        string
	CheckType       string
	Pattern         string
	ClauseReference string
	Topic           string
	SourceExcerpt   string
	Recommendations []string
}

func ExtractRulesFromPolicies(policies []Policy) []PolicyRule {
	var rules []PolicyRule
	for _, p := range policies {
		var parsed struct {
			Rules []json.RawMessage `json:"rules"`
		}
		if err := json.Unmarshal([]byte(p.RulesJSON), &parsed); err != nil {
			continue
		}
		for _, raw := range parsed.Rules {
			var r map[string]interface{}
			if err := json.Unmarshal(raw, &r); err != nil {
				continue
			}
			ruleID := strVal(r, "rule_id", "ruleId")
			checkType := strVal(r, "check_type", "checkType")
			if ruleID == "" || checkType == "" {
				continue
			}

			var recs []string
			if v, ok := r["recommendations"]; ok {
				if arr, ok := v.([]interface{}); ok {
					for _, item := range arr {
						if s, ok := item.(string); ok {
							recs = append(recs, s)
						}
					}
				}
			}

			rules = append(rules, PolicyRule{
				PolicyID:        p.ID,
				PolicyName:      p.Name,
				RuleID:          ruleID,
				Title:           strVal(r, "title"),
				Description:     strVal(r, "description"),
				Severity:        strValOr(r, "MEDIUM", "severity"),
				Category:        strVal(r, "category"),
				CheckType:       checkType,
				Pattern:         strVal(r, "pattern"),
				ClauseReference: strVal(r, "clause_reference", "clauseReference"),
				Topic:           strVal(r, "topic"),
				SourceExcerpt:   strVal(r, "source_excerpt", "sourceExcerpt"),
				Recommendations: recs,
			})
		}
	}
	return rules
}

func strVal(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}

func strValOr(m map[string]interface{}, fallback string, keys ...string) string {
	v := strVal(m, keys...)
	if v == "" {
		return fallback
	}
	return v
}

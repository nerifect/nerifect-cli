package store

import "time"

func CreateViolation(scanID, policyID int64, policyName, ruleID string, severity Severity, title, description, filePath string, lineStart, lineEnd int, codeSnippet, clauseRef, recommendation, checkType string) (*Violation, error) {
	now := time.Now()
	result, err := db.Exec(
		`INSERT INTO violations (scan_id, policy_id, policy_name, rule_id, severity, title, description, file_path, line_start, line_end, code_snippet, clause_reference, recommendation, check_type, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		scanID, policyID, policyName, ruleID, string(severity), title, description, filePath, lineStart, lineEnd, codeSnippet, clauseRef, recommendation, checkType, now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return &Violation{
		ID:              id,
		ScanID:          scanID,
		PolicyID:        policyID,
		PolicyName:      policyName,
		RuleID:          ruleID,
		Severity:        severity,
		Title:           title,
		Description:     description,
		FilePath:        filePath,
		LineStart:       lineStart,
		LineEnd:         lineEnd,
		CodeSnippet:     codeSnippet,
		ClauseReference: clauseRef,
		Recommendation:  recommendation,
		CheckType:       checkType,
		CreatedAt:       now,
	}, nil
}

func GetViolationsByScan(scanID int64) ([]Violation, error) {
	rows, err := db.Query(
		`SELECT id, scan_id, policy_id, policy_name, rule_id, severity, title, description, file_path, line_start, line_end, code_snippet, clause_reference, recommendation, check_type, created_at FROM violations WHERE scan_id = ? ORDER BY CASE severity WHEN 'CRITICAL' THEN 0 WHEN 'HIGH' THEN 1 WHEN 'MEDIUM' THEN 2 WHEN 'LOW' THEN 3 ELSE 4 END, id`,
		scanID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var violations []Violation
	for rows.Next() {
		v := Violation{}
		if err := rows.Scan(&v.ID, &v.ScanID, &v.PolicyID, &v.PolicyName, &v.RuleID, &v.Severity,
			&v.Title, &v.Description, &v.FilePath, &v.LineStart, &v.LineEnd, &v.CodeSnippet,
			&v.ClauseReference, &v.Recommendation, &v.CheckType, &v.CreatedAt); err != nil {
			return nil, err
		}
		violations = append(violations, v)
	}
	return violations, nil
}

func GetViolation(id int64) (*Violation, error) {
	row := db.QueryRow(
		`SELECT id, scan_id, policy_id, policy_name, rule_id, severity, title, description, file_path, line_start, line_end, code_snippet, clause_reference, recommendation, check_type, created_at FROM violations WHERE id = ?`,
		id,
	)
	v := &Violation{}
	err := row.Scan(&v.ID, &v.ScanID, &v.PolicyID, &v.PolicyName, &v.RuleID, &v.Severity,
		&v.Title, &v.Description, &v.FilePath, &v.LineStart, &v.LineEnd, &v.CodeSnippet,
		&v.ClauseReference, &v.Recommendation, &v.CheckType, &v.CreatedAt)
	if err != nil {
		return nil, err
	}
	return v, nil
}

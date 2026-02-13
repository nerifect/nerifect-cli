package store

import (
	"database/sql"
	"time"
)

// CreateAgentSource inserts a new monitored source URL.
func CreateAgentSource(url, name string) (*AgentSource, error) {
	now := time.Now()
	result, err := db.Exec(
		`INSERT INTO agent_sources (url, name, enabled, created_at, updated_at)
		 VALUES (?, ?, 1, ?, ?)`,
		url, name, now, now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return &AgentSource{
		ID:        id,
		URL:       url,
		Name:      name,
		Enabled:   true,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// ListAgentSources returns all sources ordered by ID.
func ListAgentSources() ([]AgentSource, error) {
	rows, err := db.Query(
		`SELECT id, url, name, enabled, content_hash, last_check_at,
		        last_error, linked_policy_id, created_at, updated_at
		 FROM agent_sources ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []AgentSource
	for rows.Next() {
		var s AgentSource
		var enabled int
		if err := rows.Scan(&s.ID, &s.URL, &s.Name, &enabled, &s.ContentHash,
			&s.LastCheckAt, &s.LastError, &s.LinkedPolicyID,
			&s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		s.Enabled = enabled == 1
		sources = append(sources, s)
	}
	return sources, nil
}

// ListEnabledAgentSources returns only enabled sources for the daemon check cycle.
func ListEnabledAgentSources() ([]AgentSource, error) {
	rows, err := db.Query(
		`SELECT id, url, name, enabled, content_hash, last_check_at,
		        last_error, linked_policy_id, created_at, updated_at
		 FROM agent_sources WHERE enabled = 1 ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []AgentSource
	for rows.Next() {
		var s AgentSource
		var enabled int
		if err := rows.Scan(&s.ID, &s.URL, &s.Name, &enabled, &s.ContentHash,
			&s.LastCheckAt, &s.LastError, &s.LinkedPolicyID,
			&s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		s.Enabled = enabled == 1
		sources = append(sources, s)
	}
	return sources, nil
}

// DeleteAgentSource removes a source by ID.
func DeleteAgentSource(id int64) error {
	result, err := db.Exec(`DELETE FROM agent_sources WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// UpdateAgentSourceCheck records a successful check with new hash and policy link.
func UpdateAgentSourceCheck(id int64, contentHash string, linkedPolicyID int64) error {
	_, err := db.Exec(
		`UPDATE agent_sources
		 SET content_hash = ?, last_check_at = ?, last_error = '',
		     linked_policy_id = ?, updated_at = ?
		 WHERE id = ?`,
		contentHash, time.Now(), linkedPolicyID, time.Now(), id,
	)
	return err
}

// UpdateAgentSourceError records a check failure.
func UpdateAgentSourceError(id int64, errMsg string) error {
	_, err := db.Exec(
		`UPDATE agent_sources
		 SET last_check_at = ?, last_error = ?, updated_at = ?
		 WHERE id = ?`,
		time.Now(), errMsg, time.Now(), id,
	)
	return err
}

// AgentSourceCount returns the total number of sources.
func AgentSourceCount() (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM agent_sources`).Scan(&count)
	return count, err
}

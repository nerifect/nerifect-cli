package store

import (
	"database/sql"
	"fmt"
	"sync"

	_ "modernc.org/sqlite"
)

var (
	db   *sql.DB
	once sync.Once
)

func Open(dbPath string) (*sql.DB, error) {
	var err error
	once.Do(func() {
		db, err = sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
		if err != nil {
			return
		}
		db.SetMaxOpenConns(1)
		err = migrate(db)
	})
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	return db, nil
}

func DB() *sql.DB {
	return db
}

func migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS scans (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		target TEXT NOT NULL,
		target_type TEXT NOT NULL DEFAULT 'local',
		scan_type TEXT NOT NULL DEFAULT 'FULL',
		status TEXT NOT NULL DEFAULT 'PENDING',
		compliance_score INTEGER,
		files_scanned INTEGER DEFAULT 0,
		violation_count INTEGER DEFAULT 0,
		ai_detection_count INTEGER DEFAULT 0,
		commit_sha TEXT DEFAULT '',
		started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		completed_at DATETIME
	);

	CREATE TABLE IF NOT EXISTS policies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT DEFAULT '',
		category TEXT NOT NULL DEFAULT 'COMPLIANCE',
		severity TEXT NOT NULL DEFAULT 'MEDIUM',
		source_url TEXT DEFAULT '',
		rules_json TEXT DEFAULT '{"rules":[]}',
		regulation_type TEXT DEFAULT '',
		rule_count INTEGER DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS violations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		scan_id INTEGER NOT NULL REFERENCES scans(id),
		policy_id INTEGER DEFAULT 0,
		policy_name TEXT DEFAULT '',
		rule_id TEXT NOT NULL,
		severity TEXT NOT NULL DEFAULT 'MEDIUM',
		title TEXT NOT NULL,
		description TEXT DEFAULT '',
		file_path TEXT NOT NULL,
		line_start INTEGER DEFAULT 0,
		line_end INTEGER DEFAULT 0,
		code_snippet TEXT DEFAULT '',
		clause_reference TEXT DEFAULT '',
		recommendation TEXT DEFAULT '',
		check_type TEXT DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS ai_detections (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		scan_id INTEGER NOT NULL REFERENCES scans(id),
		name TEXT NOT NULL,
		version TEXT DEFAULT '',
		type TEXT NOT NULL,
		risk_level TEXT NOT NULL DEFAULT 'MEDIUM',
		eu_ai_act_risk TEXT DEFAULT '',
		status TEXT NOT NULL DEFAULT 'REVIEW_REQUIRED',
		file_path TEXT NOT NULL,
		confidence REAL DEFAULT 0.0,
		detection_method TEXT DEFAULT '',
		details TEXT DEFAULT '{}',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS fixes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		violation_id INTEGER NOT NULL REFERENCES violations(id),
		scan_id INTEGER NOT NULL REFERENCES scans(id),
		fix_description TEXT NOT NULL,
		fix_diff TEXT NOT NULL,
		confidence REAL DEFAULT 0.0,
		status TEXT NOT NULL DEFAULT 'PENDING',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_violations_scan_id ON violations(scan_id);
	CREATE INDEX IF NOT EXISTS idx_ai_detections_scan_id ON ai_detections(scan_id);
	CREATE INDEX IF NOT EXISTS idx_fixes_violation_id ON fixes(violation_id);
	CREATE INDEX IF NOT EXISTS idx_fixes_scan_id ON fixes(scan_id);

	CREATE TABLE IF NOT EXISTS agent_sources (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		url TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL DEFAULT '',
		enabled INTEGER NOT NULL DEFAULT 1,
		content_hash TEXT DEFAULT '',
		last_check_at DATETIME,
		last_error TEXT DEFAULT '',
		linked_policy_id INTEGER DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_agent_sources_url ON agent_sources(url);
	`

	_, err := db.Exec(schema)
	return err
}

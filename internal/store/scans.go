package store

import (
	"database/sql"
	"time"
)

func CreateScan(target, targetType string, scanType ScanType) (*Scan, error) {
	now := time.Now()
	result, err := db.Exec(
		`INSERT INTO scans (target, target_type, scan_type, status, started_at) VALUES (?, ?, ?, ?, ?)`,
		target, targetType, string(scanType), string(ScanStatusInProgress), now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return &Scan{
		ID:         id,
		Target:     target,
		TargetType: targetType,
		ScanType:   scanType,
		Status:     ScanStatusInProgress,
		StartedAt:  now,
	}, nil
}

func CompleteScan(id int64, score *int, filesScanned, violationCount, aiDetectionCount int, commitSHA string) error {
	now := time.Now()
	_, err := db.Exec(
		`UPDATE scans SET status = ?, compliance_score = ?, files_scanned = ?, violation_count = ?, ai_detection_count = ?, commit_sha = ?, completed_at = ? WHERE id = ?`,
		string(ScanStatusCompleted), score, filesScanned, violationCount, aiDetectionCount, commitSHA, now, id,
	)
	return err
}

func FailScan(id int64) error {
	now := time.Now()
	_, err := db.Exec(
		`UPDATE scans SET status = ?, completed_at = ? WHERE id = ?`,
		string(ScanStatusFailed), now, id,
	)
	return err
}

func GetScan(id int64) (*Scan, error) {
	row := db.QueryRow(`SELECT id, target, target_type, scan_type, status, compliance_score, files_scanned, violation_count, ai_detection_count, commit_sha, started_at, completed_at FROM scans WHERE id = ?`, id)
	s := &Scan{}
	var completedAt sql.NullTime
	var score sql.NullInt64
	err := row.Scan(&s.ID, &s.Target, &s.TargetType, &s.ScanType, &s.Status, &score,
		&s.FilesScanned, &s.ViolationCount, &s.AIDetectionCount, &s.CommitSHA, &s.StartedAt, &completedAt)
	if err != nil {
		return nil, err
	}
	if completedAt.Valid {
		s.CompletedAt = &completedAt.Time
	}
	if score.Valid {
		v := int(score.Int64)
		s.ComplianceScore = &v
	}
	return s, nil
}

func ListScans(limit int) ([]Scan, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := db.Query(`SELECT id, target, target_type, scan_type, status, compliance_score, files_scanned, violation_count, ai_detection_count, commit_sha, started_at, completed_at FROM scans ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scans []Scan
	for rows.Next() {
		s := Scan{}
		var completedAt sql.NullTime
		var score sql.NullInt64
		if err := rows.Scan(&s.ID, &s.Target, &s.TargetType, &s.ScanType, &s.Status, &score,
			&s.FilesScanned, &s.ViolationCount, &s.AIDetectionCount, &s.CommitSHA, &s.StartedAt, &completedAt); err != nil {
			return nil, err
		}
		if completedAt.Valid {
			s.CompletedAt = &completedAt.Time
		}
		if score.Valid {
			v := int(score.Int64)
			s.ComplianceScore = &v
		}
		scans = append(scans, s)
	}
	return scans, nil
}

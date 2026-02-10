package store

import "time"

func CreateFix(violationID, scanID int64, fixDesc, fixDiff string, confidence float64) (*Fix, error) {
	now := time.Now()
	result, err := db.Exec(
		`INSERT INTO fixes (violation_id, scan_id, fix_description, fix_diff, confidence, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		violationID, scanID, fixDesc, fixDiff, confidence, string(FixStatusPending), now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return &Fix{
		ID:             id,
		ViolationID:    violationID,
		ScanID:         scanID,
		FixDescription: fixDesc,
		FixDiff:        fixDiff,
		Confidence:     confidence,
		Status:         FixStatusPending,
		CreatedAt:      now,
	}, nil
}

func GetFixesByViolation(violationID int64) ([]Fix, error) {
	rows, err := db.Query(
		`SELECT id, violation_id, scan_id, fix_description, fix_diff, confidence, status, created_at FROM fixes WHERE violation_id = ? ORDER BY id DESC`,
		violationID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fixes []Fix
	for rows.Next() {
		f := Fix{}
		if err := rows.Scan(&f.ID, &f.ViolationID, &f.ScanID, &f.FixDescription, &f.FixDiff,
			&f.Confidence, &f.Status, &f.CreatedAt); err != nil {
			return nil, err
		}
		fixes = append(fixes, f)
	}
	return fixes, nil
}

func GetFixesByScan(scanID int64) ([]Fix, error) {
	rows, err := db.Query(
		`SELECT id, violation_id, scan_id, fix_description, fix_diff, confidence, status, created_at FROM fixes WHERE scan_id = ? ORDER BY id DESC`,
		scanID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fixes []Fix
	for rows.Next() {
		f := Fix{}
		if err := rows.Scan(&f.ID, &f.ViolationID, &f.ScanID, &f.FixDescription, &f.FixDiff,
			&f.Confidence, &f.Status, &f.CreatedAt); err != nil {
			return nil, err
		}
		fixes = append(fixes, f)
	}
	return fixes, nil
}

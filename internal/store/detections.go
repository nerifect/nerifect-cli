package store

func CreateAIDetection(scanID int64, name, version, detType, riskLevel, euAIActRisk, status, filePath string, confidence float64, detectionMethod, details string) (*AIDetection, error) {
	result, err := db.Exec(
		`INSERT INTO ai_detections (scan_id, name, version, type, risk_level, eu_ai_act_risk, status, file_path, confidence, detection_method, details) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		scanID, name, version, detType, riskLevel, euAIActRisk, status, filePath, confidence, detectionMethod, details,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return &AIDetection{
		ID:              id,
		ScanID:          scanID,
		Name:            name,
		Version:         version,
		Type:            detType,
		RiskLevel:       riskLevel,
		EUAIActRisk:     euAIActRisk,
		Status:          status,
		FilePath:        filePath,
		Confidence:      confidence,
		DetectionMethod: detectionMethod,
		Details:         details,
	}, nil
}

func GetAIDetectionsByScan(scanID int64) ([]AIDetection, error) {
	rows, err := db.Query(
		`SELECT id, scan_id, name, version, type, risk_level, eu_ai_act_risk, status, file_path, confidence, detection_method, details, created_at FROM ai_detections WHERE scan_id = ? ORDER BY id`,
		scanID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var detections []AIDetection
	for rows.Next() {
		d := AIDetection{}
		if err := rows.Scan(&d.ID, &d.ScanID, &d.Name, &d.Version, &d.Type, &d.RiskLevel,
			&d.EUAIActRisk, &d.Status, &d.FilePath, &d.Confidence, &d.DetectionMethod, &d.Details, &d.CreatedAt); err != nil {
			return nil, err
		}
		detections = append(detections, d)
	}
	return detections, nil
}

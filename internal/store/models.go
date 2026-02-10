package store

import "time"

type PolicyCategory string

const (
	PolicyCategorySecurity       PolicyCategory = "SECURITY"
	PolicyCategoryCost           PolicyCategory = "COST"
	PolicyCategorySustainability PolicyCategory = "SUSTAINABILITY"
	PolicyCategoryCompliance     PolicyCategory = "COMPLIANCE"
)

type Severity string

const (
	SeverityLow      Severity = "LOW"
	SeverityMedium   Severity = "MEDIUM"
	SeverityHigh     Severity = "HIGH"
	SeverityCritical Severity = "CRITICAL"
	SeverityInfo     Severity = "INFO"
)

type FixStatus string

const (
	FixStatusPending  FixStatus = "PENDING"
	FixStatusApproved FixStatus = "APPROVED"
	FixStatusApplied  FixStatus = "APPLIED"
	FixStatusRejected FixStatus = "REJECTED"
)

type ScanType string

const (
	ScanTypeFull       ScanType = "FULL"
	ScanTypeCompliance ScanType = "COMPLIANCE"
	ScanTypeAI         ScanType = "AI"
)

type ScanStatus string

const (
	ScanStatusPending    ScanStatus = "PENDING"
	ScanStatusInProgress ScanStatus = "IN_PROGRESS"
	ScanStatusCompleted  ScanStatus = "COMPLETED"
	ScanStatusFailed     ScanStatus = "FAILED"
)

type Scan struct {
	ID               int64      `json:"id"`
	Target           string     `json:"target"`
	TargetType       string     `json:"target_type"` // "local" or "github"
	ScanType         ScanType   `json:"scan_type"`
	Status           ScanStatus `json:"status"`
	ComplianceScore  *int       `json:"compliance_score"`
	FilesScanned     int        `json:"files_scanned"`
	ViolationCount   int        `json:"violation_count"`
	AIDetectionCount int        `json:"ai_detection_count"`
	CommitSHA        string     `json:"commit_sha"`
	StartedAt        time.Time  `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at"`
}

type Policy struct {
	ID             int64          `json:"id"`
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	Category       PolicyCategory `json:"category"`
	Severity       Severity       `json:"severity"`
	SourceURL      string         `json:"source_url"`
	RulesJSON      string         `json:"rules_json"`
	RegulationType string         `json:"regulation_type"`
	RuleCount      int            `json:"rule_count"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type Violation struct {
	ID              int64     `json:"id"`
	ScanID          int64     `json:"scan_id"`
	PolicyID        int64     `json:"policy_id"`
	PolicyName      string    `json:"policy_name"`
	RuleID          string    `json:"rule_id"`
	Severity        Severity  `json:"severity"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	FilePath        string    `json:"file_path"`
	LineStart       int       `json:"line_start"`
	LineEnd         int       `json:"line_end"`
	CodeSnippet     string    `json:"code_snippet"`
	ClauseReference string    `json:"clause_reference"`
	Recommendation  string    `json:"recommendation"`
	CheckType       string    `json:"check_type"`
	CreatedAt       time.Time `json:"created_at"`
}

type AIDetection struct {
	ID              int64     `json:"id"`
	ScanID          int64     `json:"scan_id"`
	Name            string    `json:"name"`
	Version         string    `json:"version"`
	Type            string    `json:"type"`
	RiskLevel       string    `json:"risk_level"`
	EUAIActRisk     string    `json:"eu_ai_act_risk"`
	Status          string    `json:"status"`
	FilePath        string    `json:"file_path"`
	Confidence      float64   `json:"confidence"`
	DetectionMethod string    `json:"detection_method"`
	Details         string    `json:"details"`
	CreatedAt       time.Time `json:"created_at"`
}

type Fix struct {
	ID             int64     `json:"id"`
	ViolationID    int64     `json:"violation_id"`
	ScanID         int64     `json:"scan_id"`
	FixDescription string    `json:"fix_description"`
	FixDiff        string    `json:"fix_diff"`
	Confidence     float64   `json:"confidence"`
	Status         FixStatus `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

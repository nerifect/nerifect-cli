package ai

import (
	"path/filepath"
	"regexp"
	"strings"
)

// Detection represents a detected AI/ML component.
type Detection struct {
	Name            string  `json:"name"`
	Version         string  `json:"version"`
	Type            string  `json:"type"`
	RiskLevel       string  `json:"risk_level"`
	EUAIActRisk     string  `json:"eu_ai_act_risk"`
	Status          string  `json:"status"`
	FilePath        string  `json:"file_path"`
	Confidence      float64 `json:"confidence"`
	DetectionMethod string  `json:"detection_method"`
}

// FileReader abstracts file access for both local and remote scanning.
type FileReader interface {
	ListFiles() ([]string, error)
	ReadFile(path string) (string, error)
}

// Detector scans a file tree for AI/ML framework usage.
type Detector struct{}

func NewDetector() *Detector {
	return &Detector{}
}

// Scan runs the 4-phase detection algorithm and returns all detections.
func (d *Detector) Scan(reader FileReader) ([]Detection, error) {
	files, err := reader.ListFiles()
	if err != nil {
		return nil, err
	}

	var detections []Detection
	seen := make(map[string]bool)

	// Phase 1: Model file extensions
	for _, path := range files {
		lower := strings.ToLower(path)
		for _, ext := range ModelFileExtensions {
			if strings.HasSuffix(lower, ext.Extension) {
				name := extractModelName(path)
				detections = append(detections, Detection{
					Name:            name,
					Type:            ext.ModelType,
					RiskLevel:       ext.Risk,
					EUAIActRisk:     classifyEUAIActRisk(ext.ModelType),
					Status:          "REVIEW_REQUIRED",
					FilePath:        path,
					Confidence:      0.9,
					DetectionMethod: "file_extension",
				})
				break
			}
		}
	}

	// Phase 2: AI config files
	for _, path := range files {
		filename := strings.ToLower(filepath.Base(path))
		for _, configFile := range AIConfigFiles {
			if filename == strings.ToLower(configFile) {
				detections = append(detections, Detection{
					Name:            "AI Config: " + filepath.Base(path),
					Type:            "CONFIG",
					RiskLevel:       "MEDIUM",
					EUAIActRisk:     "MINIMAL-RISK",
					Status:          "REVIEW_REQUIRED",
					FilePath:        path,
					Confidence:      0.8,
					DetectionMethod: "config_file",
				})
				break
			}
		}
	}

	// Phase 3: Dependency file scanning
	depFileNames := map[string]bool{
		"requirements.txt": true, "setup.py": true, "pyproject.toml": true,
		"package.json": true, "go.mod": true, ".env": true, ".env.example": true,
	}
	for _, path := range files {
		fname := filepath.Base(path)
		if !depFileNames[fname] {
			continue
		}
		content, err := reader.ReadFile(path)
		if err != nil {
			continue
		}
		contentLower := strings.ToLower(content)
		for key, fw := range AIFrameworks {
			if seen[key] {
				continue
			}
			for _, dep := range fw.DepNames {
				if strings.Contains(contentLower, strings.ToLower(dep)) {
					version := extractVersion(content, dep)
					detections = append(detections, Detection{
						Name:            fw.Name,
						Version:         version,
						Type:            fw.Type,
						RiskLevel:       fw.RiskBase,
						EUAIActRisk:     classifyEUAIActRisk(fw.Type),
						Status:          "REVIEW_REQUIRED",
						FilePath:        path,
						Confidence:      0.95,
						DetectionMethod: "dependency",
					})
					seen[key] = true
					break
				}
			}
		}
	}

	// Phase 4: Import pattern matching in code files (limited to 50 files)
	codeExts := map[string]bool{".py": true, ".js": true, ".ts": true, ".go": true, ".java": true, ".rs": true}
	var codeFiles []string
	for _, path := range files {
		ext := strings.ToLower(filepath.Ext(path))
		if codeExts[ext] {
			codeFiles = append(codeFiles, path)
		}
		if len(codeFiles) >= 50 {
			break
		}
	}

	for _, path := range codeFiles {
		content, err := reader.ReadFile(path)
		if err != nil {
			continue
		}
		for key, fw := range AIFrameworks {
			if seen[key] {
				continue
			}
			for _, pattern := range fw.Patterns {
				re, err := regexp.Compile("(?i)" + pattern)
				if err != nil {
					continue
				}
				if re.MatchString(content) {
					detections = append(detections, Detection{
						Name:            fw.Name,
						Type:            fw.Type,
						RiskLevel:       fw.RiskBase,
						EUAIActRisk:     classifyEUAIActRisk(fw.Type),
						Status:          "REVIEW_REQUIRED",
						FilePath:        path,
						Confidence:      0.85,
						DetectionMethod: "code_pattern",
					})
					seen[key] = true
					break
				}
			}
		}
	}

	return detections, nil
}

func extractModelName(path string) string {
	filename := filepath.Base(path)
	ext := filepath.Ext(filename)
	return strings.TrimSuffix(filename, ext)
}

func extractVersion(content, packageName string) string {
	escaped := regexp.QuoteMeta(packageName)
	patterns := []string{
		escaped + `[=<>~!]=*\s*([\d.]+)`,
		`"` + escaped + `":\s*"[^"]*?([\d.]+)`,
	}
	for _, p := range patterns {
		re, err := regexp.Compile("(?i)" + p)
		if err != nil {
			continue
		}
		if m := re.FindStringSubmatch(content); len(m) > 1 {
			return m[1]
		}
	}
	return ""
}

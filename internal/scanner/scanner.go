package scanner

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nerifect/nerifect-cli/internal/ai"
	"github.com/nerifect/nerifect-cli/internal/compliance"
	"github.com/nerifect/nerifect-cli/internal/config"
	"github.com/nerifect/nerifect-cli/internal/llm"
	"github.com/nerifect/nerifect-cli/internal/store"
)

// ScanResult holds the complete scan output.
type ScanResult struct {
	Scan       *store.Scan
	Violations []store.Violation
	Detections []store.AIDetection
}

// ScanOptions holds optional scan parameters from repo config.
type ScanOptions struct {
	Branch    string
	PolicyIDs []int64
}

// RunScan orchestrates a full scan of a target (local path or GitHub URL).
func RunScan(ctx context.Context, target string, scanType store.ScanType, cfg *config.Config, opts ScanOptions) (*ScanResult, error) {
	var (
		scanDir   string
		cleanup   func()
		targetType = "local"
		commitSHA  string
	)

	// Resolve target
	if IsGitHubURL(target) {
		targetType = "github"
		owner, repo, err := ParseGitHubURL(target)
		if err != nil {
			return nil, fmt.Errorf("parsing GitHub URL: %w", err)
		}

		cloneURL := BuildCloneURL(owner, repo, cfg.GithubToken)
		scanDir, cleanup, err = CloneRepo(ctx, cloneURL, opts.Branch)
		if err != nil {
			return nil, fmt.Errorf("cloning repo: %w", err)
		}
		defer cleanup()
		commitSHA = GetCloneCommitSHA(scanDir)
	} else {
		// Local path
		absPath := target
		if !strings.HasPrefix(target, "/") {
			cwd, _ := os.Getwd()
			absPath = cwd + "/" + target
		}
		if target == "." {
			absPath, _ = os.Getwd()
		}
		info, err := os.Stat(absPath)
		if err != nil {
			return nil, fmt.Errorf("target path %q: %w", target, err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("target %q is not a directory", target)
		}
		scanDir = absPath
	}

	// Create scan record
	scan, err := store.CreateScan(target, targetType, scanType)
	if err != nil {
		return nil, fmt.Errorf("creating scan record: %w", err)
	}

	// Build file reader
	reader := NewLocalFileReader(scanDir, cfg.MaxFilesPerScan, cfg.MaxFileSizeKB)
	files, err := reader.ListFiles()
	if err != nil {
		store.FailScan(scan.ID)
		return nil, fmt.Errorf("listing files: %w", err)
	}

	var allDetections []store.AIDetection
	var allViolations []store.Violation

	// Phase 1: AI Detection
	if scanType == store.ScanTypeFull || scanType == store.ScanTypeAI {
		detector := ai.NewDetector()
		detections, err := detector.Scan(reader)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: AI detection failed: %v\n", err)
		} else {
			for _, d := range detections {
				det, err := store.CreateAIDetection(
					scan.ID, d.Name, d.Version, d.Type, d.RiskLevel,
					d.EUAIActRisk, d.Status, d.FilePath, d.Confidence,
					d.DetectionMethod, "{}",
				)
				if err == nil {
					allDetections = append(allDetections, *det)
				}
			}

			// LLM-based risk assessment if detections found and API key available
			if len(detections) > 0 && cfg.ActiveAPIKey() != "" {
				assessDetections(ctx, cfg, detections, &allDetections)
			}
		}
	}

	// Phase 2: Compliance scanning
	if scanType == store.ScanTypeFull || scanType == store.ScanTypeCompliance {
		// Load policies
		policies, err := store.ListPolicies()
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: loading policies failed: %v\n", err)
		}

		// Filter policies by IDs if configured
		if len(opts.PolicyIDs) > 0 && len(policies) > 0 {
			idSet := make(map[int64]bool, len(opts.PolicyIDs))
			for _, id := range opts.PolicyIDs {
				idSet[id] = true
			}
			var filtered []store.Policy
			for _, p := range policies {
				if idSet[p.ID] {
					filtered = append(filtered, p)
				}
			}
			policies = filtered
		}

		if len(policies) > 0 {
			// Read file contents for scanning
			fileContents, err := reader.ReadFilesContents()
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: reading files failed: %v\n", err)
			}

			// Pattern-based checks
			rules := store.ExtractRulesFromPolicies(policies)
			checker := compliance.NewPatternChecker()
			patternViolations := checker.Check(rules, fileContents, files)

			for _, v := range patternViolations {
				viol, err := store.CreateViolation(
					scan.ID, 0, v.PolicyName, v.RuleID, store.Severity(v.Severity),
					v.Title, v.Description, v.FilePath, v.LineStart, v.LineEnd,
					v.CodeSnippet, v.ClauseReference, v.Recommendation, "PATTERN",
				)
				if err == nil {
					allViolations = append(allViolations, *viol)
				}
			}

			// LLM semantic evaluation
			if cfg.ActiveAPIKey() != "" {
				llmClient := llm.NewClient(cfg.LLMProvider, cfg.ActiveAPIKey(), cfg.DefaultModel)
				evaluator := compliance.NewEvaluator(llmClient)

				policiesForLLM, _ := store.GetAllPoliciesForScan()
				result, err := evaluator.Evaluate(ctx, policiesForLLM, fileContents)
				if err != nil {
					fmt.Fprintf(os.Stderr, "warning: LLM evaluation failed: %v\n", err)
				} else if result != nil {
					// Deduplicate LLM violations against pattern violations
					existingKeys := make(map[string]bool)
					for _, v := range allViolations {
						key := v.RuleID + "|" + v.FilePath
						existingKeys[key] = true
					}

					for _, v := range result.Violations {
						key := v.RuleID + "|" + v.FilePath
						if existingKeys[key] {
							continue
						}
						viol, err := store.CreateViolation(
							scan.ID, 0, v.PolicyName, v.RuleID, store.Severity(v.Severity),
							v.Title, v.Description, v.FilePath, v.LineStart, v.LineEnd,
							v.CodeSnippet, v.ClauseReference, v.Recommendation, "LLM",
						)
						if err == nil {
							allViolations = append(allViolations, *viol)
						}
					}
				}
			} else if scanType == store.ScanTypeCompliance {
				fmt.Fprintf(os.Stderr, "warning: %s not set, skipping LLM evaluation\n", llm.APIKeyEnvVar(cfg.LLMProvider))
			}
		} else if scanType == store.ScanTypeCompliance {
			fmt.Fprintf(os.Stderr, "warning: no policies loaded, use 'nerifect policy add' to add policies\n")
		}
	}

	// Calculate compliance score
	var violationResults []compliance.ViolationResult
	for _, v := range allViolations {
		violationResults = append(violationResults, compliance.ViolationResult{
			RuleID:   v.RuleID,
			Severity: string(v.Severity),
		})
	}
	score := compliance.CalculateScore(violationResults)

	// Complete scan
	store.CompleteScan(scan.ID, &score, len(files), len(allViolations), len(allDetections), commitSHA)

	// Reload scan to get updated fields
	scan, _ = store.GetScan(scan.ID)

	return &ScanResult{
		Scan:       scan,
		Violations: allViolations,
		Detections: allDetections,
	}, nil
}

func assessDetections(ctx context.Context, cfg *config.Config, detections []ai.Detection, storedDetections *[]store.AIDetection) {
	var lines []string
	for _, d := range detections {
		lines = append(lines, fmt.Sprintf("- %s (%s): %s", d.Name, d.Type, d.FilePath))
	}
	modelsSummary := strings.Join(lines, "\n")

	llmClient := llm.NewClient(cfg.LLMProvider, cfg.ActiveAPIKey(), cfg.DefaultModel)
	prompt := llm.BuildAIGovernancePrompt(modelsSummary)

	responseText, err := llmClient.GenerateContent(ctx, prompt)
	if err != nil {
		return
	}

	var assessments []struct {
		Name       string `json:"name"`
		Status     string `json:"status"`
		RiskLevel  string `json:"risk_level"`
		Issues     int    `json:"issues"`
		EUAIActRisk string `json:"eu_ai_act_risk"`
		Reasoning  string `json:"reasoning"`
	}

	if err := llm.ParseJSONResponse(responseText, &assessments); err != nil {
		return
	}

	// Update stored detections with LLM assessments
	for i, det := range *storedDetections {
		for _, a := range assessments {
			if strings.EqualFold(a.Name, det.Name) {
				if a.Status != "" {
					(*storedDetections)[i].Status = a.Status
				}
				if a.RiskLevel != "" {
					(*storedDetections)[i].RiskLevel = a.RiskLevel
				}
				if a.EUAIActRisk != "" {
					(*storedDetections)[i].EUAIActRisk = a.EUAIActRisk
				}
				break
			}
		}
	}
}

package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/nerifect/nerifect-cli/internal/config"
	"github.com/nerifect/nerifect-cli/internal/output"
	"github.com/nerifect/nerifect-cli/internal/scanner"
	"github.com/nerifect/nerifect-cli/internal/store"
	"github.com/spf13/cobra"
)

func newScanCmd() *cobra.Command {
	var scanTypeFlag string
	var diffBase string

	cmd := &cobra.Command{
		Use:   "scan <path-or-url>",
		Short: "Scan a repository or directory",
		Long: `Scan a local directory or GitHub repository for compliance violations
and AI governance issues.`,
		Example: `  nerifect scan .
  nerifect scan /path/to/project
  nerifect scan https://github.com/owner/repo
  nerifect scan --type ai .
  nerifect scan --type compliance . --output json
  nerifect scan --diff .
  nerifect scan --diff main .`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScan(cmd, args[0], scanTypeFlag, diffBase)
		},
	}

	cmd.Flags().StringVar(&scanTypeFlag, "type", "full", "scan type: full, compliance, ai")
	cmd.Flags().StringVar(&diffBase, "diff", "", "scan only changed files (vs git ref, default HEAD if flag set without value)")
	return cmd
}

func runScan(cmd *cobra.Command, target, scanTypeStr, diffBase string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Init database
	if _, err := store.Open(cfg.DatabasePath); err != nil {
		return fmt.Errorf("initializing database: %w", err)
	}

	// Check if target matches a configured repo
	var branch string
	var policyIDs []int64
	if repo := cfg.FindRepo(target); repo != nil {
		// Use repo URL/path as target if matched by name
		if repo.URL != "" && !scanner.IsGitHubURL(target) && target == repo.Name {
			target = repo.URL
		} else if repo.Path != "" && target == repo.Name {
			target = repo.Path
		}
		branch = repo.Branch
		policyIDs = repo.Policies
		// Use repo's scan type as default if not explicitly set via flag
		if !cmd.Flags().Changed("type") && repo.ScanType != "" {
			scanTypeStr = repo.ScanType
		}
	}

	// Parse scan type
	var scanType store.ScanType
	switch strings.ToLower(scanTypeStr) {
	case "compliance":
		scanType = store.ScanTypeCompliance
	case "ai":
		scanType = store.ScanTypeAI
	default:
		scanType = store.ScanTypeFull
	}

	outFmt := output.ParseFormat(outputFormat)

	// Start scanning with progress
	var progress *output.Progress
	if outFmt != output.FormatJSON && outFmt != output.FormatSARIF {
		progress = output.NewProgress("Scanning " + target + "...")
	}

	opts := scanner.ScanOptions{
		Branch:    branch,
		PolicyIDs: policyIDs,
	}

	// Handle --diff flag: if flag was changed but value is empty, default to HEAD
	if cmd.Flags().Changed("diff") {
		if diffBase == "" {
			diffBase = "HEAD"
		}
		opts.DiffBase = diffBase
	}
	result, err := scanner.RunScan(cmd.Context(), target, scanType, cfg, opts)
	if err != nil {
		if progress != nil {
			progress.Fail("Scan failed: " + err.Error())
		}
		return err
	}

	if progress != nil {
		progress.Done("Scan completed")
	}

	// Render results
	output.RenderScanReport(result.Scan, result.Violations, result.Detections, outFmt)

	// Exit code 2 for critical violations (CI/CD gate)
	for _, v := range result.Violations {
		if strings.ToUpper(string(v.Severity)) == "CRITICAL" {
			os.Exit(2)
		}
	}

	return nil
}

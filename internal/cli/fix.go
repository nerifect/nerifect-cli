package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/nerifect/nerifect-cli/internal/config"
	"github.com/nerifect/nerifect-cli/internal/fixer"
	"github.com/nerifect/nerifect-cli/internal/llm"
	"github.com/nerifect/nerifect-cli/internal/output"
	"github.com/nerifect/nerifect-cli/internal/store"
	"github.com/spf13/cobra"
)

func newFixCmd() *cobra.Command {
	var fixAll bool

	cmd := &cobra.Command{
		Use:   "fix <violation-id>",
		Short: "Generate an AI-powered fix for a violation",
		Long: `Generate a compliance fix using AI. Pass a violation ID to fix a single
violation, or use --all with a scan ID to fix all violations.`,
		Example: `  nerifect fix 42
  nerifect fix --all 1`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFix(cmd, args[0], fixAll)
		},
	}

	cmd.Flags().BoolVar(&fixAll, "all", false, "fix all violations in a scan (argument is scan-id)")
	return cmd
}

func runFix(cmd *cobra.Command, idStr string, fixAll bool) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration error: %w (run 'nerifect init' to set up)", err)
	}
	if _, err := store.Open(cfg.DatabasePath); err != nil {
		return err
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid ID: %s", idStr)
	}

	llmClient := llm.NewClient(cfg.LLMProvider, cfg.ActiveAPIKey(), cfg.DefaultModel)
	f := fixer.NewFixer(llmClient)
	outFmt := output.ParseFormat(outputFormat)

	if fixAll {
		return fixAllViolations(cmd, f, id, outFmt)
	}
	return fixSingleViolation(cmd, f, id, outFmt)
}

func fixSingleViolation(cmd *cobra.Command, f *fixer.Fixer, violationID int64, outFmt output.Format) error {
	v, err := store.GetViolation(violationID)
	if err != nil {
		return fmt.Errorf("violation #%d not found: %w", violationID, err)
	}

	progress := output.NewProgress(fmt.Sprintf("Generating fix for violation #%d...", violationID))

	result, err := f.GenerateFix(cmd.Context(), v.RuleID, v.FilePath, string(v.Severity), v.Description, v.CodeSnippet)
	if err != nil {
		progress.Fail("Fix generation failed: " + err.Error())
		return err
	}

	fix, err := store.CreateFix(v.ID, v.ScanID, result.FixDescription, result.FixDiff, result.Confidence)
	if err != nil {
		progress.Fail("Saving fix failed: " + err.Error())
		return err
	}

	progress.Done("Fix generated")
	output.RenderFix(fix, v, outFmt)
	return nil
}

func fixAllViolations(cmd *cobra.Command, f *fixer.Fixer, scanID int64, outFmt output.Format) error {
	violations, err := store.GetViolationsByScan(scanID)
	if err != nil {
		return fmt.Errorf("loading violations for scan #%d: %w", scanID, err)
	}

	if len(violations) == 0 {
		fmt.Println("No violations found for scan #", scanID)
		return nil
	}

	fmt.Printf("Generating fixes for %d violations in scan #%d...\n\n", len(violations), scanID)

	for i, v := range violations {
		progress := output.NewProgress(fmt.Sprintf("[%d/%d] Fixing %s in %s...", i+1, len(violations), v.RuleID, v.FilePath))

		result, err := f.GenerateFix(cmd.Context(), v.RuleID, v.FilePath, string(v.Severity), v.Description, v.CodeSnippet)
		if err != nil {
			progress.Fail(fmt.Sprintf("Failed: %v", err))
			fmt.Fprintf(os.Stderr, "  Skipping violation #%d: %v\n", v.ID, err)
			continue
		}

		fix, err := store.CreateFix(v.ID, v.ScanID, result.FixDescription, result.FixDiff, result.Confidence)
		if err != nil {
			progress.Fail("Save failed")
			continue
		}

		progress.Done(fmt.Sprintf("Fix #%d (confidence: %.0f%%)", fix.ID, fix.Confidence*100))
	}

	fmt.Println()
	output.PrintSuccess(fmt.Sprintf("Fix generation complete for scan #%d", scanID))
	return nil
}

package cli

import (
	"fmt"
	"strconv"

	"github.com/nerifect/nerifect-cli/internal/config"
	"github.com/nerifect/nerifect-cli/internal/output"
	"github.com/nerifect/nerifect-cli/internal/store"
	"github.com/spf13/cobra"
)

func newReportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "report <scan-id>",
		Short: "Display a scan report",
		Long:  `Show detailed results for a previous scan, including violations and AI detections.`,
		Example: `  nerifect report 1
  nerifect report 1 --output json`,
		Args: cobra.ExactArgs(1),
		RunE: runReport,
	}
}

func runReport(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if _, err := store.Open(cfg.DatabasePath); err != nil {
		return err
	}

	scanID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid scan ID: %s", args[0])
	}

	scan, err := store.GetScan(scanID)
	if err != nil {
		return fmt.Errorf("scan #%d not found: %w", scanID, err)
	}

	violations, err := store.GetViolationsByScan(scanID)
	if err != nil {
		return fmt.Errorf("loading violations: %w", err)
	}

	detections, err := store.GetAIDetectionsByScan(scanID)
	if err != nil {
		return fmt.Errorf("loading AI detections: %w", err)
	}

	output.RenderScanReport(scan, violations, detections, output.ParseFormat(outputFormat))
	return nil
}

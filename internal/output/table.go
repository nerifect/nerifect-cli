package output

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/nerifect/nerifect-cli/internal/store"
)

func RenderScanReport(scan *store.Scan, violations []store.Violation, detections []store.AIDetection, format Format) {
	switch format {
	case FormatJSON:
		PrintJSON(map[string]interface{}{
			"scan":          scan,
			"violations":    violations,
			"ai_detections": detections,
			"summary":       buildSummary(scan, violations, detections),
		})
	case FormatPlain:
		renderScanPlain(scan, violations, detections)
	default:
		renderScanTable(scan, violations, detections)
	}
}

func buildSummary(scan *store.Scan, violations []store.Violation, detections []store.AIDetection) map[string]interface{} {
	critCount := 0
	highCount := 0
	for _, v := range violations {
		switch strings.ToUpper(string(v.Severity)) {
		case "CRITICAL":
			critCount++
		case "HIGH":
			highCount++
		}
	}
	score := 0
	if scan.ComplianceScore != nil {
		score = *scan.ComplianceScore
	}
	return map[string]interface{}{
		"compliance_score": score,
		"files_scanned":   scan.FilesScanned,
		"violation_count":  len(violations),
		"critical_count":   critCount,
		"high_count":       highCount,
		"ai_detections":    len(detections),
		"has_ai_presence":  len(detections) > 0,
	}
}

func renderScanTable(scan *store.Scan, violations []store.Violation, detections []store.AIDetection) {
	// Summary box
	score := 0
	if scan.ComplianceScore != nil {
		score = *scan.ComplianceScore
	}
	scoreStr := lipgloss.NewStyle().Bold(true).Foreground(ScoreColor(score)).Render(fmt.Sprintf("%d/100", score))

	duration := ""
	if scan.CompletedAt != nil {
		d := scan.CompletedAt.Sub(scan.StartedAt)
		duration = fmt.Sprintf(" in %s", d.Round(1e8))
	}

	summary := fmt.Sprintf(
		"%s  Scan #%d\n%s  %s\n\n%s %s   %s %d   %s %d   %s %d",
		HeaderStyle.Render("Nerifect"), scan.ID,
		DimStyle.Render("Target:"), scan.Target,
		BoldStyle.Render("Score:"), scoreStr,
		BoldStyle.Render("Files:"), scan.FilesScanned,
		BoldStyle.Render("Violations:"), len(violations),
		BoldStyle.Render("AI Detections:"), len(detections),
	)
	if duration != "" {
		summary += DimStyle.Render(duration)
	}

	fmt.Println(SummaryBox.Render(summary))

	// Violations table
	if len(violations) > 0 {
		fmt.Println(HeaderStyle.Render("\nCompliance Violations"))
		fmt.Println(strings.Repeat("─", 90))
		fmt.Printf("  %-5s %-10s %-20s %-30s %s\n",
			DimStyle.Render("ID"), DimStyle.Render("SEVERITY"), DimStyle.Render("RULE"), DimStyle.Render("FILE"), DimStyle.Render("TITLE"))
		fmt.Println(strings.Repeat("─", 90))

		for _, v := range violations {
			sevStr := SeverityStyle(string(v.Severity)).Render(fmt.Sprintf("%-10s", v.Severity))
			fmt.Printf("  %-5d %s %-20s %-30s %s\n",
				v.ID,
				sevStr,
				Truncate(v.RuleID, 20),
				Truncate(v.FilePath, 30),
				Truncate(v.Title, 40),
			)
		}
		fmt.Println()
	}

	// AI detections table
	if len(detections) > 0 {
		fmt.Println(HeaderStyle.Render("AI/ML Detections"))
		fmt.Println(strings.Repeat("─", 90))
		fmt.Printf("  %-25s %-12s %-10s %-15s %s\n",
			DimStyle.Render("FRAMEWORK"), DimStyle.Render("TYPE"), DimStyle.Render("RISK"), DimStyle.Render("EU AI ACT"), DimStyle.Render("FILE"))
		fmt.Println(strings.Repeat("─", 90))

		for _, d := range detections {
			riskStr := SeverityStyle(d.RiskLevel).Render(fmt.Sprintf("%-10s", d.RiskLevel))
			fmt.Printf("  %-25s %-12s %s %-15s %s\n",
				Truncate(d.Name, 25),
				Truncate(d.Type, 12),
				riskStr,
				Truncate(d.EUAIActRisk, 15),
				Truncate(d.FilePath, 40),
			)
		}
		fmt.Println()
	}

	if len(violations) == 0 && len(detections) == 0 {
		fmt.Println(SuccessStyle.Render("\n  No violations or AI detections found. Looking good!"))
		fmt.Println()
	}
}

func renderScanPlain(scan *store.Scan, violations []store.Violation, detections []store.AIDetection) {
	score := 0
	if scan.ComplianceScore != nil {
		score = *scan.ComplianceScore
	}

	fmt.Printf("Scan #%d | Target: %s | Score: %d/100 | Files: %d | Violations: %d | AI: %d\n",
		scan.ID, scan.Target, score, scan.FilesScanned, len(violations), len(detections))

	for _, v := range violations {
		fmt.Printf("[%s] %s - %s (%s) in %s\n", v.Severity, v.RuleID, v.Title, v.CheckType, v.FilePath)
	}
	for _, d := range detections {
		fmt.Printf("[AI] %s (%s) risk=%s eu_ai_act=%s in %s\n", d.Name, d.Type, d.RiskLevel, d.EUAIActRisk, d.FilePath)
	}
}

func RenderPolicies(policies []store.Policy, format Format) {
	if format == FormatJSON {
		PrintJSON(policies)
		return
	}

	if len(policies) == 0 {
		fmt.Println(DimStyle.Render("  No policies loaded. Use 'nerifect policy add <file-or-url>' to add one."))
		return
	}

	fmt.Println(HeaderStyle.Render("\nPolicies"))
	fmt.Println(strings.Repeat("─", 80))
	fmt.Printf("  %-5s %-30s %-12s %-10s %-8s %s\n",
		DimStyle.Render("ID"), DimStyle.Render("NAME"), DimStyle.Render("CATEGORY"), DimStyle.Render("SEVERITY"), DimStyle.Render("RULES"), DimStyle.Render("TYPE"))
	fmt.Println(strings.Repeat("─", 80))

	for _, p := range policies {
		fmt.Printf("  %-5d %-30s %-12s %-10s %-8d %s\n",
			p.ID,
			Truncate(p.Name, 30),
			p.Category,
			p.Severity,
			p.RuleCount,
			p.RegulationType,
		)
	}
	fmt.Println()
}

func RenderFix(fix *store.Fix, violation *store.Violation, format Format) {
	if format == FormatJSON {
		PrintJSON(map[string]interface{}{
			"fix":       fix,
			"violation": violation,
		})
		return
	}

	fmt.Println(HeaderStyle.Render(fmt.Sprintf("\nFix for Violation #%d", violation.ID)))
	fmt.Println(strings.Repeat("─", 70))
	fmt.Printf("  %s %s\n", BoldStyle.Render("Rule:"), violation.RuleID)
	fmt.Printf("  %s %s\n", BoldStyle.Render("File:"), violation.FilePath)
	fmt.Printf("  %s %s\n", BoldStyle.Render("Confidence:"), fmt.Sprintf("%.0f%%", fix.Confidence*100))
	fmt.Printf("  %s %s\n", BoldStyle.Render("Description:"), fix.FixDescription)
	if fix.FixDiff != "" {
		fmt.Println()
		fmt.Println(BoldStyle.Render("  Diff:"))
		for _, line := range strings.Split(fix.FixDiff, "\n") {
			if strings.HasPrefix(line, "+") {
				fmt.Println("  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#00CC00")).Render(line))
			} else if strings.HasPrefix(line, "-") {
				fmt.Println("  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render(line))
			} else {
				fmt.Println("  " + line)
			}
		}
	}
	fmt.Println()
}

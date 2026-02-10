package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatPlain Format = "plain"
	FormatSARIF Format = "sarif"
)

func ParseFormat(s string) Format {
	switch strings.ToLower(s) {
	case "json":
		return FormatJSON
	case "plain":
		return FormatPlain
	case "sarif":
		return FormatSARIF
	default:
		return FormatTable
	}
}

// Severity styles
var (
	CriticalStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)
	HighStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6600")).Bold(true)
	MediumStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAA00"))
	LowStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#00CC00"))
	InfoStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#0099FF"))

	HeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	DimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	BoldStyle   = lipgloss.NewStyle().Bold(true)
	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00CC00")).Bold(true)
	ErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)

	SummaryBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)
)

func SeverityStyle(severity string) lipgloss.Style {
	switch strings.ToUpper(severity) {
	case "CRITICAL":
		return CriticalStyle
	case "HIGH":
		return HighStyle
	case "MEDIUM":
		return MediumStyle
	case "LOW":
		return LowStyle
	default:
		return InfoStyle
	}
}

func ScoreColor(score int) lipgloss.Color {
	if score >= 80 {
		return lipgloss.Color("#00CC00")
	}
	if score >= 60 {
		return lipgloss.Color("#FFAA00")
	}
	return lipgloss.Color("#FF0000")
}

func PrintJSON(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func Truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

func PrintError(msg string) {
	fmt.Fprintln(os.Stderr, ErrorStyle.Render("Error: ")+msg)
}

func PrintSuccess(msg string) {
	fmt.Println(SuccessStyle.Render("✓ ") + msg)
}

func PrintWarning(msg string) {
	fmt.Fprintln(os.Stderr, MediumStyle.Render("Warning: ")+msg)
}

package output

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func GetLogo() string {
	// A stylized unicode geometric representation
	// This attempts to look modern and technical
	
	primary := lipgloss.NewStyle().Foreground(lipgloss.Color("#0077BE")) // Ocean Blue
	secondary := lipgloss.NewStyle().Foreground(lipgloss.Color("#00A86B")) // Jade Green
	
	// 𝐍𝐞𝐫𝐢𝐟𝐞𝐜𝐭
	logoText := `
    █▄ █ █▀▀ █▀▄ █ █▀▀ █▀▀ █▀▀ ▀█▀
    █ ▀█ █▀▀ █▀▄ █ █▀▀ █▀▀ █    █ 
    ▀  ▀ ▀▀▀ ▀ ▀ ▀ ▀   ▀▀▀ ▀▀▀  ▀ 
`
	
	styledLogo := primary.Render(logoText)
	
	// Add a nice tagline
	tagline := secondary.Render("    Cloud Governance & Compliance CLI")
	
	return fmt.Sprintf("%s\n%s\n", styledLogo, tagline)
}

func PrintBanner() {
	fmt.Println(GetLogo())
}

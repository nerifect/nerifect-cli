package cli

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/nerifect/nerifect-cli/internal/config"
	"github.com/nerifect/nerifect-cli/internal/llm"
	"github.com/nerifect/nerifect-cli/internal/output"
	"github.com/nerifect/nerifect-cli/internal/store"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize Nerifect CLI configuration",
		Long:  `Interactively set up LLM provider, API keys, model selection, and output preferences.`,
		RunE:  runInit,
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	fmt.Println(output.HeaderStyle.Render("\n  Welcome to Nerifect CLI Setup\n"))

	var provider string

	providerForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("LLM Provider").
				Description("Select the AI provider for scanning and analysis").
				Options(
					huh.NewOption("Google Gemini (recommended)", llm.ProviderGemini),
					huh.NewOption("OpenAI", llm.ProviderOpenAI),
					huh.NewOption("Anthropic", llm.ProviderAnthropic),
				).
				Value(&provider),
		),
	)

	if err := providerForm.Run(); err != nil {
		return fmt.Errorf("setup cancelled: %w", err)
	}

	var apiKey, githubToken, model, outFmt string

	apiKeyInput := huh.NewInput().
		Title(llm.ProviderLabel(provider) + " API Key").
		Description("Required for AI-powered scanning and analysis").
		EchoMode(huh.EchoModePassword).
		Value(&apiKey)

	switch provider {
	case llm.ProviderOpenAI:
		apiKeyInput.Placeholder("sk-...")
	case llm.ProviderAnthropic:
		apiKeyInput.Placeholder("sk-ant-...")
	default:
		apiKeyInput.Placeholder("AIza...")
	}

	modelSelect := buildModelSelect(provider, &model)

	detailsForm := huh.NewForm(
		huh.NewGroup(
			apiKeyInput,

			huh.NewInput().
				Title("GitHub Token (optional)").
				Description("Required for scanning private repositories").
				EchoMode(huh.EchoModePassword).
				Value(&githubToken),

			modelSelect,

			huh.NewSelect[string]().
				Title("Default Output Format").
				Options(
					huh.NewOption("Table (human-friendly)", "table"),
					huh.NewOption("JSON (CI/CD pipelines)", "json"),
					huh.NewOption("Plain (minimal)", "plain"),
				).
				Value(&outFmt),
		),
	)

	if err := detailsForm.Run(); err != nil {
		return fmt.Errorf("setup cancelled: %w", err)
	}

	cfg := config.DefaultConfig()
	cfg.LLMProvider = provider
	cfg.GithubToken = githubToken
	cfg.DefaultModel = model
	cfg.OutputFormat = outFmt

	switch provider {
	case llm.ProviderOpenAI:
		cfg.OpenAIAPIKey = apiKey
	case llm.ProviderAnthropic:
		cfg.AnthropicAPIKey = apiKey
	default:
		cfg.GeminiAPIKey = apiKey
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	// Ensure data dir
	os.MkdirAll(cfg.DataDir, 0700)

	// Initialize database
	if _, err := store.Open(cfg.DatabasePath); err != nil {
		return fmt.Errorf("initializing database: %w", err)
	}

	fmt.Println()
	output.PrintSuccess(fmt.Sprintf("Config saved to ~/.nerifect.yaml"))
	output.PrintSuccess(fmt.Sprintf("Provider: %s", llm.ProviderLabel(provider)))
	output.PrintSuccess(fmt.Sprintf("Database initialized at %s", cfg.DatabasePath))
	fmt.Println()
	fmt.Println(output.DimStyle.Render("  Get started:"))
	fmt.Println(output.DimStyle.Render("    nerifect scan .                    # Scan current directory"))
	fmt.Println(output.DimStyle.Render("    nerifect policy add <url>          # Add a compliance policy"))
	fmt.Println(output.DimStyle.Render("    nerifect scan --type compliance .  # Run compliance scan"))
	fmt.Println()

	return nil
}

func buildModelSelect(provider string, model *string) *huh.Select[string] {
	switch provider {
	case llm.ProviderOpenAI:
		return huh.NewSelect[string]().
			Title("Default OpenAI Model").
			Options(
				huh.NewOption("gpt-4o (recommended)", "gpt-4o"),
				huh.NewOption("gpt-4o-mini (fast)", "gpt-4o-mini"),
				huh.NewOption("gpt-4-turbo (high quality)", "gpt-4-turbo"),
				huh.NewOption("gpt-4.1 (latest)", "gpt-4.1"),
				huh.NewOption("gpt-4.1-mini (fast, latest)", "gpt-4.1-mini"),
			).
			Value(model)
	case llm.ProviderAnthropic:
		return huh.NewSelect[string]().
			Title("Default Anthropic Model").
			Options(
				huh.NewOption("claude-sonnet-4 (recommended)", "claude-sonnet-4-20250514"),
				huh.NewOption("claude-3.5-haiku (fast)", "claude-3-5-haiku-20241022"),
				huh.NewOption("claude-opus-4 (highest quality)", "claude-opus-4-20250514"),
			).
			Value(model)
	default:
		return huh.NewSelect[string]().
			Title("Default Gemini Model").
			Options(
				huh.NewOption("gemini-2.0-flash (fast, recommended)", "gemini-2.0-flash"),
				huh.NewOption("gemini-2.5-flash (balanced)", "gemini-2.5-flash"),
				huh.NewOption("gemini-2.5-pro (highest quality)", "gemini-2.5-pro"),
			).
			Value(model)
	}
}

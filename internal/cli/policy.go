package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nerifect/nerifect-cli/internal/config"
	"github.com/nerifect/nerifect-cli/internal/llm"
	"github.com/nerifect/nerifect-cli/internal/output"
	"github.com/nerifect/nerifect-cli/internal/policy"
	"github.com/nerifect/nerifect-cli/internal/presets"
	"github.com/nerifect/nerifect-cli/internal/store"
	"github.com/spf13/cobra"
)

func newPolicyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Manage governance policies",
		Long:  `Add, list, or remove compliance policies used for scanning.`,
	}
	cmd.AddCommand(newPolicyListCmd())
	cmd.AddCommand(newPolicyAddCmd())
	cmd.AddCommand(newPolicyRemoveCmd())
	cmd.AddCommand(newPolicyListPresetsCmd())
	cmd.AddCommand(newPolicyAddPresetCmd())
	return cmd
}

func newPolicyListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all loaded policies",
		RunE:  runPolicyList,
	}
}

func newPolicyAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <file-or-url>",
		Short: "Add a policy from a file or URL",
		Long: `Add a compliance policy by providing a URL to a regulation document
or a local file path. The document will be parsed using AI to extract
structured compliance rules.`,
		Example: `  nerifect policy add https://example.com/gdpr-regulation.html
  nerifect policy add /path/to/policy.txt
  nerifect policy add regulation.pdf`,
		Args: cobra.ExactArgs(1),
		RunE: runPolicyAdd,
	}
}

func newPolicyRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "remove <id>",
		Aliases: []string{"rm", "delete"},
		Short:   "Remove a policy by ID",
		Args:    cobra.ExactArgs(1),
		RunE:    runPolicyRemove,
	}
}

func runPolicyList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if _, err := store.Open(cfg.DatabasePath); err != nil {
		return err
	}

	policies, err := store.ListPolicies()
	if err != nil {
		return fmt.Errorf("listing policies: %w", err)
	}

	output.RenderPolicies(policies, output.ParseFormat(outputFormat))
	return nil
}

func runPolicyAdd(cmd *cobra.Command, args []string) error {
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

	source := args[0]
	llmClient := llm.NewClient(cfg.LLMProvider, cfg.ActiveAPIKey(), cfg.DefaultModel)
	mgr := policy.NewManager(llmClient)

	outFmt := output.ParseFormat(outputFormat)
	var progress *output.Progress
	if outFmt != output.FormatJSON {
		progress = output.NewProgress("Parsing policy document...")
	}

	var p *store.Policy
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		p, err = mgr.AddFromURL(cmd.Context(), source)
	} else {
		p, err = mgr.AddFromFile(cmd.Context(), source)
	}

	if err != nil {
		if progress != nil {
			progress.Fail("Failed to add policy: " + err.Error())
		}
		return err
	}

	if progress != nil {
		progress.Done(fmt.Sprintf("Policy added: %s (%d rules)", p.Name, p.RuleCount))
	}

	if outFmt == output.FormatJSON {
		output.PrintJSON(p)
	}

	return nil
}

func runPolicyRemove(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if _, err := store.Open(cfg.DatabasePath); err != nil {
		return err
	}

	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid policy ID: %s", args[0])
	}

	if err := store.DeletePolicy(id); err != nil {
		return fmt.Errorf("removing policy: %w", err)
	}

	output.PrintSuccess(fmt.Sprintf("Policy #%d removed", id))
	return nil
}

func newPolicyListPresetsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-presets",
		Short: "List available built-in compliance presets",
		Long:  `Show all built-in compliance framework presets that can be installed without an LLM API key.`,
		RunE:  runPolicyListPresets,
	}
}

func newPolicyAddPresetCmd() *cobra.Command {
	var all bool
	cmd := &cobra.Command{
		Use:   "add-preset <name>",
		Short: "Install a built-in compliance preset",
		Long: `Install a built-in compliance framework preset as a policy.
These presets provide pattern-based rules that work without an LLM API key.`,
		Example: `  nerifect policy add-preset owasp-top-10
  nerifect policy add-preset cis-docker
  nerifect policy add-preset soc2-basic
  nerifect policy add-preset gdpr-basic
  nerifect policy add-preset hipaa-basic
  nerifect policy add-preset pci-dss
  nerifect policy add-preset cis-kubernetes
  nerifect policy add-preset nist-800-53
  nerifect policy add-preset eu-ai-act
  nerifect policy add-preset --all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPolicyAddPreset(cmd, args, all)
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "install all available presets")
	return cmd
}

func runPolicyListPresets(cmd *cobra.Command, args []string) error {
	available := presets.List()
	outFmt := output.ParseFormat(outputFormat)

	if outFmt == output.FormatJSON {
		output.PrintJSON(available)
		return nil
	}

	fmt.Println(output.HeaderStyle.Render("\nAvailable Presets"))
	fmt.Println(strings.Repeat("─", 80))
	fmt.Printf("  %-20s %-45s %s\n",
		output.DimStyle.Render("SLUG"),
		output.DimStyle.Render("NAME"),
		output.DimStyle.Render("RULES"))
	fmt.Println(strings.Repeat("─", 80))

	for _, p := range available {
		fmt.Printf("  %-20s %-45s %d\n",
			p.Slug,
			output.Truncate(p.Name, 45),
			len(p.Rules),
		)
	}
	fmt.Println()
	fmt.Println(output.DimStyle.Render("  Install with: nerifect policy add-preset <slug>"))
	fmt.Println()
	return nil
}

func runPolicyAddPreset(cmd *cobra.Command, args []string, all bool) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if _, err := store.Open(cfg.DatabasePath); err != nil {
		return err
	}

	if all {
		available := presets.List()
		for _, p := range available {
			pol, err := presets.Install(p.Slug)
			if err != nil {
				return fmt.Errorf("installing preset %q: %w", p.Slug, err)
			}
			output.PrintSuccess(fmt.Sprintf("Preset installed: %s (%d rules)", pol.Name, pol.RuleCount))
		}
		return nil
	}

	if len(args) == 0 {
		return fmt.Errorf("provide a preset name or use --all (see 'nerifect policy list-presets')")
	}

	pol, err := presets.Install(args[0])
	if err != nil {
		return err
	}

	outFmt := output.ParseFormat(outputFormat)
	if outFmt == output.FormatJSON {
		output.PrintJSON(pol)
	} else {
		output.PrintSuccess(fmt.Sprintf("Preset installed: %s (%d rules)", pol.Name, pol.RuleCount))
	}
	return nil
}

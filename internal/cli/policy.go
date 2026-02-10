package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nerifect/nerifect-cli/internal/config"
	"github.com/nerifect/nerifect-cli/internal/llm"
	"github.com/nerifect/nerifect-cli/internal/output"
	"github.com/nerifect/nerifect-cli/internal/policy"
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

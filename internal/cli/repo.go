package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nerifect/nerifect-cli/internal/config"
	"github.com/nerifect/nerifect-cli/internal/output"
	"github.com/nerifect/nerifect-cli/internal/scanner"
	"github.com/spf13/cobra"
)

func newRepoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Manage tracked repositories",
		Long:  `Add, list, update, or remove tracked repositories and their scan settings.`,
	}
	cmd.AddCommand(newRepoAddCmd())
	cmd.AddCommand(newRepoListCmd())
	cmd.AddCommand(newRepoRemoveCmd())
	cmd.AddCommand(newRepoUpdateCmd())
	return cmd
}

func newRepoAddCmd() *cobra.Command {
	var (
		name     string
		branch   string
		scanType string
		policies []int64
	)

	cmd := &cobra.Command{
		Use:   "add <path-or-url>",
		Short: "Add a repository to track",
		Long: `Add a local directory or GitHub repository URL to the global config.
The repo's branch, scan type, and policy filters will be used as defaults
when scanning this target.`,
		Example: `  nerifect repo add .
  nerifect repo add /path/to/project --name my-project --branch main
  nerifect repo add https://github.com/owner/repo --branch release/v2 --scan-type compliance
  nerifect repo add . --policy 1 --policy 2`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRepoAdd(args[0], name, branch, scanType, policies)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "repo name (default: derived from path or URL)")
	cmd.Flags().StringVar(&branch, "branch", "", "git branch or tag")
	cmd.Flags().StringVar(&scanType, "scan-type", "", "scan type: full, compliance, ai")
	cmd.Flags().Int64SliceVar(&policies, "policy", nil, "policy ID to apply (can repeat)")
	return cmd
}

func newRepoListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all tracked repositories",
		RunE:  runRepoList,
	}
}

func newRepoRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "remove <name>",
		Aliases: []string{"rm", "delete"},
		Short:   "Remove a tracked repository",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRepoRemove(args[0])
		},
	}
}

func newRepoUpdateCmd() *cobra.Command {
	var (
		branch   string
		scanType string
		policies []int64
	)

	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update a tracked repository's settings",
		Example: `  nerifect repo update my-project --branch develop
  nerifect repo update my-project --scan-type compliance --policy 1 --policy 2`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRepoUpdate(args[0], branch, scanType, policies, cmd)
		},
	}

	cmd.Flags().StringVar(&branch, "branch", "", "git branch or tag")
	cmd.Flags().StringVar(&scanType, "scan-type", "", "scan type: full, compliance, ai")
	cmd.Flags().Int64SliceVar(&policies, "policy", nil, "policy IDs to apply (replaces existing)")
	return cmd
}

func runRepoAdd(target, name, branch, scanType string, policies []int64) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	repo := config.RepoConfig{
		Branch:   branch,
		ScanType: scanType,
		Policies: policies,
	}

	// Determine if target is a URL or local path
	if scanner.IsGitHubURL(target) {
		repo.URL = target
		if name == "" {
			owner, repoName, err := scanner.ParseGitHubURL(target)
			if err != nil {
				return fmt.Errorf("cannot parse URL for name: %w", err)
			}
			name = owner + "/" + repoName
		}
	} else {
		// Local path - resolve to absolute
		absPath, err := filepath.Abs(target)
		if err != nil {
			return fmt.Errorf("resolving path: %w", err)
		}
		info, err := os.Stat(absPath)
		if err != nil {
			return fmt.Errorf("path %q: %w", target, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("path %q is not a directory", target)
		}
		repo.Path = absPath
		if name == "" {
			name = filepath.Base(absPath)
		}
	}

	repo.Name = name

	if err := cfg.AddRepo(repo); err != nil {
		return err
	}

	output.PrintSuccess(fmt.Sprintf("Repo %q added", name))
	return nil
}

func runRepoList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	repos := cfg.ListRepos()
	outFmt := output.ParseFormat(outputFormat)

	if outFmt == output.FormatJSON {
		output.PrintJSON(repos)
		return nil
	}

	if len(repos) == 0 {
		fmt.Println("No repos configured. Use 'nerifect repo add' to add one.")
		return nil
	}

	fmt.Println(output.HeaderStyle.Render("Tracked Repositories"))
	fmt.Println()

	for _, r := range repos {
		target := r.Path
		if target == "" {
			target = r.URL
		}

		fmt.Printf("  %s\n", output.BoldStyle.Render(r.Name))
		fmt.Printf("    Target:    %s\n", target)
		if r.Branch != "" {
			fmt.Printf("    Branch:    %s\n", r.Branch)
		}
		if r.ScanType != "" {
			fmt.Printf("    Scan Type: %s\n", r.ScanType)
		}
		if len(r.Policies) > 0 {
			policyStrs := make([]string, len(r.Policies))
			for i, p := range r.Policies {
				policyStrs[i] = fmt.Sprintf("#%d", p)
			}
			fmt.Printf("    Policies:  %s\n", strings.Join(policyStrs, ", "))
		}
		fmt.Println()
	}

	fmt.Printf("%s\n", output.DimStyle.Render(fmt.Sprintf("%d repo(s) configured", len(repos))))
	return nil
}

func runRepoRemove(name string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if err := cfg.RemoveRepo(name); err != nil {
		return err
	}

	output.PrintSuccess(fmt.Sprintf("Repo %q removed", name))
	return nil
}

func runRepoUpdate(name, branch, scanType string, policies []int64, cmd *cobra.Command) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	err = cfg.UpdateRepo(name, func(r *config.RepoConfig) {
		if cmd.Flags().Changed("branch") {
			r.Branch = branch
		}
		if cmd.Flags().Changed("scan-type") {
			r.ScanType = scanType
		}
		if cmd.Flags().Changed("policy") {
			r.Policies = policies
		}
	})
	if err != nil {
		return err
	}

	output.PrintSuccess(fmt.Sprintf("Repo %q updated", name))
	return nil
}

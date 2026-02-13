package cli

import (
	"fmt"
	"strconv"
	"time"

	"github.com/nerifect/nerifect-cli/internal/agent"
	"github.com/nerifect/nerifect-cli/internal/config"
	"github.com/nerifect/nerifect-cli/internal/output"
	"github.com/nerifect/nerifect-cli/internal/store"
	"github.com/spf13/cobra"
)

func newAgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Manage the policy ingestion agent",
		Long: `The agent periodically monitors compliance document URLs and
automatically ingests new or updated policies using the LLM pipeline.
Disabled by default — use 'nerifect agent start' to enable.`,
	}
	cmd.AddCommand(newAgentStartCmd())
	cmd.AddCommand(newAgentStopCmd())
	cmd.AddCommand(newAgentStatusCmd())
	cmd.AddCommand(newAgentSourcesCmd())
	return cmd
}

func newAgentStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the background policy ingestion daemon",
		RunE:  runAgentStart,
	}
}

func newAgentStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the running daemon",
		RunE:  runAgentStop,
	}
}

func newAgentStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show daemon status and source statistics",
		RunE:  runAgentStatus,
	}
}

func newAgentSourcesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sources",
		Short: "Manage monitored source URLs",
	}
	cmd.AddCommand(newAgentSourcesListCmd())
	cmd.AddCommand(newAgentSourcesAddCmd())
	cmd.AddCommand(newAgentSourcesRemoveCmd())
	return cmd
}

func newAgentSourcesListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all monitored source URLs",
		RunE:  runAgentSourcesList,
	}
}

func newAgentSourcesAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <url>",
		Short: "Add a source URL to monitor",
		Args:  cobra.ExactArgs(1),
		RunE:  runAgentSourcesAdd,
	}
	cmd.Flags().String("name", "", "label for the source")
	return cmd
}

func newAgentSourcesRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "remove <id>",
		Aliases: []string{"rm", "delete"},
		Short:   "Remove a source by ID",
		Args:    cobra.ExactArgs(1),
		RunE:    runAgentSourcesRemove,
	}
}

// newRunDaemonCmd is the hidden command used as the re-exec target for the daemon.
func newRunDaemonCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "_run-daemon",
		Hidden: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Suppress the banner for the daemon process.
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			return agent.RunDaemon(cfg)
		},
	}
}

// --- Run functions ---

func runAgentStart(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("agent requires a valid LLM API key: %w", err)
	}

	// Ensure database is initialized.
	if _, err := store.Open(cfg.DatabasePath); err != nil {
		return err
	}

	if err := agent.StartDaemon(cfg); err != nil {
		return err
	}

	output.PrintSuccess(fmt.Sprintf("Agent started (check interval: %dh)", cfg.AgentCheckInterval))
	fmt.Printf("  Log: %s\n", agent.LogPath(cfg))
	fmt.Printf("  PID: %s\n", agent.PidPath(cfg))
	return nil
}

func runAgentStop(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if err := agent.StopDaemon(cfg); err != nil {
		return err
	}

	output.PrintSuccess("Agent stopped")
	return nil
}

func runAgentStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	st, err := agent.GetStatus(cfg)
	if err != nil {
		return err
	}

	if outputFormat == "json" {
		output.PrintJSON(st)
		return nil
	}

	fmt.Println(output.HeaderStyle.Render("Agent Status"))
	fmt.Println()

	if st.Running {
		fmt.Printf("  Status:     %s\n", output.SuccessStyle.Render("Running"))
		fmt.Printf("  PID:        %d\n", st.PID)
	} else {
		fmt.Printf("  Status:     %s\n", output.DimStyle.Render("Stopped"))
	}

	fmt.Printf("  Interval:   %dh\n", st.Interval)
	fmt.Printf("  Sources:    %d\n", st.SourceCount)

	if st.ErrorCount > 0 {
		fmt.Printf("  Errors:     %s\n", output.ErrorStyle.Render(strconv.Itoa(st.ErrorCount)))
	}

	if st.LastCheckAt != nil {
		fmt.Printf("  Last check: %s\n", st.LastCheckAt.Format(time.RFC3339))
	}
	if st.NextCheckAt != nil && st.Running {
		fmt.Printf("  Next check: %s\n", st.NextCheckAt.Format(time.RFC3339))
	}

	return nil
}

func runAgentSourcesList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if _, err := store.Open(cfg.DatabasePath); err != nil {
		return err
	}

	sources, err := store.ListAgentSources()
	if err != nil {
		return err
	}

	if outputFormat == "json" {
		output.PrintJSON(sources)
		return nil
	}

	if len(sources) == 0 {
		fmt.Println("No sources configured. Use 'nerifect agent sources add <url>' to add one.")
		return nil
	}

	fmt.Println(output.HeaderStyle.Render("Agent Sources"))
	fmt.Println()

	for _, s := range sources {
		enabledStr := output.SuccessStyle.Render("enabled")
		if !s.Enabled {
			enabledStr = output.DimStyle.Render("disabled")
		}

		name := s.Name
		if name == "" {
			name = "(unnamed)"
		}

		fmt.Printf("  [%d] %s (%s)\n", s.ID, output.BoldStyle.Render(name), enabledStr)
		fmt.Printf("      URL:  %s\n", s.URL)

		if s.LastCheckAt != nil {
			fmt.Printf("      Last: %s\n", s.LastCheckAt.Format(time.RFC3339))
		}
		if s.LastError != "" {
			fmt.Printf("      Err:  %s\n", output.ErrorStyle.Render(output.Truncate(s.LastError, 80)))
		}
		if s.LinkedPolicyID > 0 {
			fmt.Printf("      Policy ID: %d\n", s.LinkedPolicyID)
		}
		fmt.Println()
	}

	return nil
}

func runAgentSourcesAdd(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if _, err := store.Open(cfg.DatabasePath); err != nil {
		return err
	}

	url := args[0]
	name, _ := cmd.Flags().GetString("name")

	src, err := store.CreateAgentSource(url, name)
	if err != nil {
		return fmt.Errorf("adding source: %w", err)
	}

	output.PrintSuccess(fmt.Sprintf("Added source #%d: %s", src.ID, url))
	return nil
}

func runAgentSourcesRemove(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if _, err := store.Open(cfg.DatabasePath); err != nil {
		return err
	}

	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid source ID: %s", args[0])
	}

	if err := store.DeleteAgentSource(id); err != nil {
		return fmt.Errorf("removing source: %w", err)
	}

	output.PrintSuccess(fmt.Sprintf("Removed source #%d", id))
	return nil
}

package cli

import (
	"github.com/nerifect/nerifect-cli/internal/output"
	"github.com/spf13/cobra"
)

var Version = "dev"

var (
	cfgFile      string
	outputFormat string
	verbose      bool
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "nerifect",
		Short: "Nerifect - Cloud Governance CLI",
		Long: `Nerifect CLI scans repositories for compliance violations,
detects AI/ML framework usage, evaluates governance policies,
and generates AI-powered fixes using Google Gemini.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if outputFormat == "table" || outputFormat == "" {
				// Only print banner if we are not running the root command's Run (which handles it manually)
				if cmd.Use != "nerifect" {
					output.PrintBanner()
				}
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			if outputFormat == "table" || outputFormat == "" {
				output.PrintBanner()
			}
			cmd.Help()
		},
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default ~/.nerifect.yaml)")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "output format: table, json, plain, sarif")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newScanCmd())
	rootCmd.AddCommand(newPolicyCmd())
	rootCmd.AddCommand(newFixCmd())
	rootCmd.AddCommand(newReportCmd())
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newRepoCmd())
	rootCmd.AddCommand(newVersionCmd())

	return rootCmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version of nerifect",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("nerifect version", Version)
		},
	}
}

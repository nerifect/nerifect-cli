package cli

import (
	"fmt"
	"strings"

	"github.com/nerifect/nerifect-cli/internal/config"
	"github.com/nerifect/nerifect-cli/internal/output"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
	}
	cmd.AddCommand(newConfigGetCmd())
	cmd.AddCommand(newConfigSetCmd())
	cmd.AddCommand(newConfigListCmd())
	return cmd
}

func newConfigGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			val, err := cfg.Get(args[0])
			if err != nil {
				return err
			}
			// Mask secrets
			key := strings.ToLower(args[0])
			if strings.Contains(key, "key") || strings.Contains(key, "token") {
				if len(val) > 8 {
					val = val[:4] + "..." + val[len(val)-4:]
				} else if val != "" {
					val = "****"
				}
			}
			fmt.Println(val)
			return nil
		},
	}
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long:  fmt.Sprintf("Set a configuration value. Valid keys: %s", config.ValidConfigKeysStr()),
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if err := cfg.Set(args[0], args[1]); err != nil {
				return err
			}
			output.PrintSuccess(fmt.Sprintf("Set %s", args[0]))
			return nil
		},
	}
}

func newConfigListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configuration keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			for _, key := range config.ValidConfigKeys() {
				val, _ := cfg.Get(key)
				// Mask secrets
				if strings.Contains(key, "key") || strings.Contains(key, "token") {
					if len(val) > 8 {
						val = val[:4] + "..." + val[len(val)-4:]
					} else if val != "" {
						val = "****"
					}
				}
				fmt.Printf("  %s = %s\n", output.BoldStyle.Render(key), val)
			}
			return nil
		},
	}
}

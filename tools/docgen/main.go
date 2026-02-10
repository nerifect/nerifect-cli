package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nerifect/nerifect-cli/internal/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func main() {
	rootCmd := cli.NewRootCmd()
	outDir := filepath.Join("docs", "cli")

	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	if err := genMarkdownTree(rootCmd, outDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating docs: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("CLI documentation generated in %s/\n", outDir)
}

func genMarkdownTree(cmd *cobra.Command, dir string) error {
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := genMarkdownTree(c, dir); err != nil {
			return err
		}
	}

	filename := cmdFilename(cmd)
	f, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		return err
	}
	defer f.Close()

	md := genMarkdown(cmd)
	_, err = f.WriteString(md)
	return err
}

func genMarkdown(cmd *cobra.Command) string {
	var b strings.Builder

	name := cmd.CommandPath()
	b.WriteString(fmt.Sprintf("# %s\n\n", name))

	// Description
	long := cmd.Long
	if long == "" {
		long = cmd.Short
	}
	if long != "" {
		b.WriteString(long + "\n\n")
	}

	// Usage / Synopsis
	if cmd.Runnable() {
		b.WriteString("## Usage\n\n")
		b.WriteString(fmt.Sprintf("```\n%s\n```\n\n", cmd.UseLine()))
	}

	// Aliases
	if len(cmd.Aliases) > 0 {
		b.WriteString("## Aliases\n\n")
		b.WriteString(fmt.Sprintf("`%s`\n\n", strings.Join(cmd.Aliases, "`, `")))
	}

	// Examples
	if cmd.HasExample() {
		b.WriteString("## Examples\n\n")
		b.WriteString(fmt.Sprintf("```bash\n%s\n```\n\n", cmd.Example))
	}

	// Subcommands
	children := cmd.Commands()
	if len(children) > 0 {
		avail := make([]*cobra.Command, 0)
		for _, c := range children {
			if c.IsAvailableCommand() && !c.IsAdditionalHelpTopicCommand() {
				avail = append(avail, c)
			}
		}
		if len(avail) > 0 {
			b.WriteString("## Available Commands\n\n")
			b.WriteString("| Command | Description |\n")
			b.WriteString("|---------|-------------|\n")
			for _, c := range avail {
				link := cmdFilename(c)
				b.WriteString(fmt.Sprintf("| [`%s`](%s) | %s |\n", c.Name(), link, c.Short))
			}
			b.WriteString("\n")
		}
	}

	// Flags (non-inherited)
	if hasNonInheritedFlags(cmd) {
		b.WriteString("## Flags\n\n")
		b.WriteString(flagsToTable(cmd.NonInheritedFlags()))
		b.WriteString("\n")
	}

	// Inherited / global flags
	if hasInheritedFlags(cmd) {
		b.WriteString("## Global Flags\n\n")
		b.WriteString(flagsToTable(cmd.InheritedFlags()))
		b.WriteString("\n")
	}

	// See also
	if hasSeeAlso(cmd) {
		b.WriteString("## See Also\n\n")
		if cmd.HasParent() {
			parent := cmd.Parent()
			link := cmdFilename(parent)
			b.WriteString(fmt.Sprintf("- [`%s`](%s) - %s\n", parent.CommandPath(), link, parent.Short))
		}

		sort.Sort(byName(children))
		for _, c := range children {
			if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
				continue
			}
			link := cmdFilename(c)
			b.WriteString(fmt.Sprintf("- [`%s`](%s) - %s\n", c.CommandPath(), link, c.Short))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func flagsToTable(fs *pflag.FlagSet) string {
	var b strings.Builder
	b.WriteString("| Flag | Shorthand | Description | Default |\n")
	b.WriteString("|------|-----------|-------------|---------|\n")

	fs.VisitAll(func(f *pflag.Flag) {
		if f.Hidden {
			return
		}
		shorthand := ""
		if f.Shorthand != "" {
			shorthand = fmt.Sprintf("`-%s`", f.Shorthand)
		}
		name := fmt.Sprintf("`--%s`", f.Name)
		desc := strings.ReplaceAll(f.Usage, "|", "\\|")
		desc = strings.ReplaceAll(desc, "\n", " ")
		defVal := f.DefValue
		if defVal == "" {
			defVal = ""
		} else {
			defVal = fmt.Sprintf("`%s`", defVal)
		}
		b.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", name, shorthand, desc, defVal))
	})

	return b.String()
}

func cmdFilename(cmd *cobra.Command) string {
	name := strings.ReplaceAll(cmd.CommandPath(), " ", "_") + ".md"
	return name
}

func hasNonInheritedFlags(cmd *cobra.Command) bool {
	found := false
	cmd.NonInheritedFlags().VisitAll(func(f *pflag.Flag) {
		if !f.Hidden {
			found = true
		}
	})
	return found
}

func hasInheritedFlags(cmd *cobra.Command) bool {
	found := false
	cmd.InheritedFlags().VisitAll(func(f *pflag.Flag) {
		if !f.Hidden {
			found = true
		}
	})
	return found
}

func hasSeeAlso(cmd *cobra.Command) bool {
	if cmd.HasParent() {
		return true
	}
	for _, c := range cmd.Commands() {
		if c.IsAvailableCommand() && !c.IsAdditionalHelpTopicCommand() {
			return true
		}
	}
	return false
}

type byName []*cobra.Command

func (s byName) Len() int           { return len(s) }
func (s byName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byName) Less(i, j int) bool { return s[i].Name() < s[j].Name() }

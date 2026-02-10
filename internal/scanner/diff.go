package scanner

import (
	"fmt"
	"os/exec"
	"strings"
)

// GitChangedFiles returns relative paths of files changed compared to a git ref.
// It includes added, copied, modified, and renamed files (ACMR filter).
// If base is empty, it defaults to HEAD.
func GitChangedFiles(dir string, base string) ([]string, error) {
	if base == "" {
		base = "HEAD"
	}

	// First try diff against the ref (for committed changes)
	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter=ACMR", base)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff failed: %w", err)
	}

	files := parseGitOutput(string(out))

	// Also include unstaged and untracked files
	cmd2 := exec.Command("git", "diff", "--name-only", "--diff-filter=ACMR")
	cmd2.Dir = dir
	out2, _ := cmd2.Output()
	for _, f := range parseGitOutput(string(out2)) {
		files = appendUnique(files, f)
	}

	// Include untracked files
	cmd3 := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	cmd3.Dir = dir
	out3, _ := cmd3.Output()
	for _, f := range parseGitOutput(string(out3)) {
		files = appendUnique(files, f)
	}

	return files, nil
}

func parseGitOutput(out string) []string {
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files
}

func appendUnique(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}

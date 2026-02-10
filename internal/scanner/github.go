package scanner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// ParseGitHubURL extracts owner and repo from a GitHub URL.
// Supports https://github.com/owner/repo and git@github.com:owner/repo formats.
func ParseGitHubURL(url string) (owner, repo string, err error) {
	url = strings.TrimSpace(url)

	// https://github.com/owner/repo(.git)
	if strings.Contains(url, "github.com/") {
		parts := strings.SplitN(url, "github.com/", 2)
		if len(parts) < 2 {
			return "", "", fmt.Errorf("invalid GitHub URL: %s", url)
		}
		rest := strings.Split(parts[1], "?")[0]
		rest = strings.Split(rest, "#")[0]
		rest = strings.Trim(rest, "/")
		segments := strings.Split(rest, "/")
		if len(segments) >= 2 {
			owner = segments[0]
			repo = strings.TrimSuffix(segments[1], ".git")
			return owner, repo, nil
		}
	}

	// git@github.com:owner/repo(.git)
	if strings.HasPrefix(url, "git@github.com:") {
		rest := strings.SplitN(url, ":", 2)[1]
		segments := strings.Split(strings.Trim(rest, "/"), "/")
		if len(segments) >= 2 {
			owner = segments[0]
			repo = strings.TrimSuffix(segments[1], ".git")
			return owner, repo, nil
		}
	}

	return "", "", fmt.Errorf("cannot parse GitHub URL: %s", url)
}

// IsGitHubURL returns true if the string looks like a GitHub repository URL.
func IsGitHubURL(s string) bool {
	return strings.Contains(s, "github.com/") || strings.HasPrefix(s, "git@github.com:")
}

// CloneRepo performs a shallow clone of a GitHub repo to a temp directory.
// Returns the temp dir path and a cleanup function.
func CloneRepo(ctx context.Context, url, branch string) (string, func(), error) {
	tmpDir, err := os.MkdirTemp("", "nerifect-scan-*")
	if err != nil {
		return "", nil, fmt.Errorf("creating temp dir: %w", err)
	}
	cleanup := func() { os.RemoveAll(tmpDir) }

	args := []string{"clone", "--depth=1"}
	if branch != "" {
		args = append(args, "--branch", branch)
	}
	args = append(args, url, tmpDir)

	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("git clone failed: %s: %w", string(output), err)
	}

	return tmpDir, cleanup, nil
}

// GetCloneCommitSHA returns the HEAD commit SHA of a cloned repo.
func GetCloneCommitSHA(dir string) string {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// BuildCloneURL constructs an HTTPS clone URL, adding token auth if provided.
func BuildCloneURL(owner, repo, token string) string {
	if token != "" {
		return fmt.Sprintf("https://%s@github.com/%s/%s.git", token, owner, repo)
	}
	return fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
}

var githubURLRe = regexp.MustCompile(`github\.com[/:]([^/]+)/([^/.\s]+)`)

// ExtractGitHubInfo is a lenient extraction that works with various URL formats.
func ExtractGitHubInfo(url string) (owner, repo string) {
	m := githubURLRe.FindStringSubmatch(url)
	if len(m) >= 3 {
		return m[1], strings.TrimSuffix(m[2], ".git")
	}
	return "", ""
}

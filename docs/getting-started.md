# Getting Started

## Prerequisites

- **Go 1.22+** --- Required to build from source
- **Git** --- Required for scanning remote repositories
- **An LLM API key** --- Required for LLM-powered features (compliance evaluation, fix generation, policy parsing). Supported providers:
    - [Google Gemini](https://aistudio.google.com/apikey)
    - [OpenAI](https://platform.openai.com/api-keys)
    - [Anthropic](https://console.anthropic.com/settings/keys)

## Installation

Nerifect CLI is available on Linux, macOS, and Windows.

### Package Managers

**Homebrew (macOS/Linux)**

```bash
brew install nerifect/tap/nerifect
```

**Manual Install (Debian/Ubuntu/RHEL)**

Download the latest `.deb` or `.rpm` release from the [GitHub Releases](https://github.com/nerifect/nerifect-cli/releases) page.

```bash
# Debian/Ubuntu
sudo dpkg -i nerifect_*.deb

# RHEL/CentOS
sudo rpm -i nerifect_*.rpm
```

**Go Install**

```bash
go install github.com/nerifect/nerifect-cli/cmd/nerifect@latest
```

### Build from source

```bash
git clone https://github.com/nerifect/nerifect-cli.git
cd nerifect-cli
make build
```

The binary will be at `bin/nerifect`.

### Install to PATH

```bash
make install
```

This copies the binary to `$GOPATH/bin/nerifect`.

## Initial Setup

Run the interactive setup wizard to configure your API key and preferences:

```bash
nerifect init
```

This will prompt you for:

1. Your LLM provider (Gemini, OpenAI, or Anthropic)
2. Your API key for the chosen provider
3. Preferred model (provider-specific defaults available)
4. Default output format

Configuration is saved to `~/.nerifect.yaml`.

!!! tip
    You can also set the API key via environment variable: `export GEMINI_API_KEY=your-key` (or `OPENAI_API_KEY` / `ANTHROPIC_API_KEY` depending on your provider)

## Your First Scan

### Scan a local directory

```bash
nerifect scan .
```

This runs a full scan (AI detection + compliance) on the current directory.

### Scan for AI/ML frameworks only

```bash
nerifect scan --type ai .
```

### Scan a GitHub repository

```bash
nerifect scan https://github.com/owner/repo
```

Nerifect clones the repo with `--depth=1` to a temp directory and scans it.

### Add a compliance policy

```bash
nerifect policy add https://example.com/gdpr-regulation.html
```

The document is parsed using AI to extract structured compliance rules.

### Run a compliance scan

```bash
nerifect scan --type compliance .
```

### View a previous scan report

```bash
nerifect report 1
```

### Generate fixes for violations

```bash
# Fix a single violation
nerifect fix 42

# Fix all violations in a scan
nerifect fix --all 1
```

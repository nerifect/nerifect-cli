# Nerifect CLI

A standalone command-line tool for cloud governance, compliance scanning, and AI/ML framework detection. Built in Go with support for multiple LLM providers (Google Gemini, OpenAI, and Anthropic).

Nerifect CLI scans repositories (local or GitHub) for compliance violations, detects AI/ML framework usage, evaluates governance policies, and generates AI-powered fixes.

## Key Features

- **AI/ML Framework Detection** --- Detects 25+ frameworks (TensorFlow, PyTorch, OpenAI, Anthropic, LangChain, HuggingFace, etc.) via import patterns, dependency files, model files, and config files
- **EU AI Act Risk Classification** --- Automatically classifies detected AI components as HIGH-RISK, LIMITED-RISK, or MINIMAL-RISK
- **Compliance Evaluation** --- Pattern-based and LLM-powered semantic analysis against ingested governance policies
- **Policy Ingestion** --- Parse regulation documents (HTML, text) into structured compliance rules using AI
- **Fix Generation** --- AI-generated fixes with unified diffs and confidence scores
- **GitHub Integration** --- Scan remote repositories via `git clone --depth=1`
- **CI/CD Support** --- JSON output mode and exit code 2 on critical violations for pipeline gating
- **Local Storage** --- SQLite database with no external dependencies

## How It Works

```
nerifect scan <target>
  -> Resolve target (local path or GitHub URL)
  -> Walk files, filter binaries
  -> Run 4-phase AI/ML framework detection
  -> Run pattern-based + LLM compliance evaluation
  -> Calculate compliance score
  -> Save results to local SQLite database
  -> Output report (table or JSON)
```

## Quick Example

```bash
# Set up your API key
nerifect init

# Scan the current directory
nerifect scan .

# Scan a GitHub repository
nerifect scan https://github.com/owner/repo

# Add a compliance policy and scan
nerifect policy add https://example.com/regulation.html
nerifect scan --type compliance .

# Generate fixes for violations
nerifect fix --all 1
```

## Next Steps

- [Getting Started](getting-started.md) --- Installation and first scan
- [CLI Reference](cli/nerifect.md) --- Full command reference
- [Configuration](configuration.md) --- Config file and environment variables
- [AI/ML Detection](ai-detection.md) --- Supported frameworks and risk classification
- [CI/CD Integration](cicd.md) --- Pipeline integration guide

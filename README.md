# Nerifect CLI

<p align="center">
  <img src="docs/assets/images/nerifect_logo.png" alt="Nerifect Logo" width="200"/>
</p>

A standalone command-line tool for cloud governance, compliance scanning, and AI/ML framework detection. Built in Go with support for multiple LLM providers (Google Gemini, OpenAI, and Anthropic).

Nerifect CLI scans repositories (local or GitHub) for compliance violations, detects AI/ML framework usage, evaluates governance policies, and generates AI-powered fixes.

## Features

- **AI/ML Framework Detection** — Detects 25+ frameworks (TensorFlow, PyTorch, OpenAI, Anthropic, LangChain, HuggingFace, etc.) via import patterns, dependency files, model files, and config files
- **EU AI Act Risk Classification** — Automatically classifies detected AI components as HIGH-RISK, LIMITED-RISK, or MINIMAL-RISK
- **Compliance Evaluation** — Pattern-based and LLM-powered semantic analysis against ingested governance policies
- **Policy Ingestion** — Parse regulation documents (HTML, text) into structured compliance rules using AI
- **Fix Generation** — AI-generated fixes with unified diffs and confidence scores
- **GitHub Integration** — Scan remote repositories via `git clone --depth=1`
- **CI/CD Support** — JSON output mode and exit code 2 on critical violations for pipeline gating
- **Local Storage** — SQLite database with no external dependencies

## Installation

### Build from source

```bash
git clone https://github.com/nerifect/nerifect-cli.git
cd nerifect-cli
make build
```

The binary will be at `bin/nerifect`.

### Prerequisites

- Go 1.22+
- Git (for scanning remote repositories)
- An API key from one of the supported LLM providers:
    - [Google Gemini](https://aistudio.google.com/apikey)
    - [OpenAI](https://platform.openai.com/api-keys)
    - [Anthropic](https://console.anthropic.com/settings/keys)

## Quick Start

```bash
# 1. Set up your API key
nerifect init

# 2. Scan the current directory
nerifect scan .

# 3. Scan a GitHub repository
nerifect scan https://github.com/owner/repo

# 4. Add a compliance policy
nerifect policy add https://example.com/regulation.html

# 5. Run a compliance scan
nerifect scan --type compliance .

# 6. Generate fixes for violations
nerifect fix --all 1
```

## Commands

### `nerifect init`

Interactive setup wizard for LLM provider selection, API keys, model selection, and preferences.

```bash
nerifect init
```

### `nerifect scan <path-or-url>`

Scan a local directory or GitHub repository.

```bash
# Full scan (AI detection + compliance)
nerifect scan .
nerifect scan /path/to/project

# AI/ML framework detection only
nerifect scan --type ai .

# Compliance-only scan
nerifect scan --type compliance .

# JSON output for CI/CD
nerifect scan . --output json

# Scan a remote repository
nerifect scan https://github.com/owner/repo
```

**Exit codes:**
- `0` — Scan completed, no critical violations
- `1` — Tool error
- `2` — Critical violations found (use for CI/CD gating)

### `nerifect policy`

Manage governance policies.

```bash
# List all policies
nerifect policy list

# Add a policy from a URL (parsed with AI)
nerifect policy add https://example.com/gdpr.html

# Add a policy from a local file
nerifect policy add /path/to/regulation.txt

# Remove a policy
nerifect policy remove 1
```

### `nerifect fix`

Generate AI-powered fixes for compliance violations.

```bash
# Fix a single violation
nerifect fix 42

# Fix all violations in a scan
nerifect fix --all 1
```

### `nerifect report <scan-id>`

Display results from a previous scan.

```bash
nerifect report 1
nerifect report 1 --output json
```

### `nerifect config`

Manage CLI configuration.

```bash
# List all config values
nerifect config list

# Get a specific value
nerifect config get default_model

# Set a value
nerifect config set llm_provider openai
nerifect config set default_model gpt-4o
```

## Configuration

Configuration is stored in `~/.nerifect.yaml` and can be overridden with environment variables.

| Config Key | Env Variable | Default | Description |
|---|---|---|---|
| `llm_provider` | `NERIFECT_PROVIDER` | `gemini` | LLM provider (`gemini`, `openai`, `anthropic`) |
| `gemini_api_key` | `GEMINI_API_KEY` | — | Google Gemini API key |
| `openai_api_key` | `OPENAI_API_KEY` | — | OpenAI API key |
| `anthropic_api_key` | `ANTHROPIC_API_KEY` | — | Anthropic API key |
| `github_token` | `GITHUB_TOKEN` | — | GitHub token for private repos |
| `default_model` | `NERIFECT_MODEL` | `gemini-2.0-flash` | Model to use (provider-specific) |
| `output_format` | `NERIFECT_OUTPUT` | `table` | Default output format |
| `data_dir` | `NERIFECT_DATA_DIR` | `~/.nerifect` | Data directory |
| `max_files_per_scan` | — | `800` | Max files to scan |
| `max_file_size_kb` | — | `80` | Max file size in KB |

### Supported models

**Google Gemini:**

- `gemini-2.0-flash` (default, fast)
- `gemini-2.5-flash` (balanced)
- `gemini-2.5-pro` (highest quality)

**OpenAI:**

- `gpt-4o` (default, recommended)
- `gpt-4o-mini` (fast)
- `gpt-4-turbo` (high quality)
- `gpt-4.1` (latest)
- `gpt-4.1-mini` (fast, latest)

**Anthropic:**

- `claude-sonnet-4-20250514` (default, recommended)
- `claude-3-5-haiku-20241022` (fast)
- `claude-opus-4-20250514` (highest quality)

## AI/ML Frameworks Detected

| Category | Frameworks |
|---|---|
| **ML Frameworks** | TensorFlow, PyTorch, scikit-learn, Keras, XGBoost, LightGBM |
| **LLM APIs** | OpenAI, Anthropic, Google Vertex AI, Azure OpenAI, AWS Bedrock, Cohere, AI21, Mistral, Replicate, Together AI, Groq |
| **LLM Local** | Ollama, llama.cpp |
| **Orchestration** | LangChain, LlamaIndex, Semantic Kernel, AutoGen, CrewAI |
| **Pre-trained Models** | HuggingFace Transformers |
| **MLOps** | MLflow |

Detection methods: import pattern matching, dependency file scanning, model file extensions (`.pt`, `.h5`, `.onnx`, `.safetensors`, etc.), and AI config files.

## CI/CD Integration

```yaml
# GitHub Actions example
- name: Nerifect Compliance Scan
  run: |
    nerifect scan . --type compliance --output json > scan-results.json
    # Exit code 2 means critical violations found
```

## Project Structure

```
nerifect-cli/
├── cmd/nerifect/main.go           # Entry point
├── internal/
│   ├── cli/                       # Cobra command definitions
│   ├── config/                    # YAML config + env var loading
│   ├── scanner/                   # Scan orchestration, file walking, GitHub clone
│   ├── ai/                        # AI/ML framework detection (25+ frameworks)
│   ├── compliance/                # Pattern checker, LLM evaluator, scorer
│   ├── fixer/                     # Fix generation and diff application
│   ├── policy/                    # Policy fetching, LLM parsing, management
│   ├── llm/                       # Multi-provider LLM client (Gemini, OpenAI, Anthropic)
│   ├── store/                     # SQLite database and CRUD operations
│   └── output/                    # Table, JSON, and plain text formatting
├── go.mod
└── Makefile
```

## License

See [LICENSE](LICENSE) for details.

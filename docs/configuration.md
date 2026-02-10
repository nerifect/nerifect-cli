# Configuration

Nerifect CLI stores configuration in `~/.nerifect.yaml` with `0600` permissions for API key security. All values can be overridden with environment variables.

## Configuration File

The config file is created automatically when you run `nerifect init`. You can also create it manually:

```yaml
llm_provider: "gemini"
gemini_api_key: "your-api-key-here"
openai_api_key: ""
anthropic_api_key: ""
github_token: ""
default_model: "gemini-2.0-flash"
output_format: "table"
data_dir: "~/.nerifect"
max_files_per_scan: 800
max_file_size_kb: 80
```

## Configuration Reference

| Config Key | Env Variable | Default | Description |
|---|---|---|---|
| `llm_provider` | `NERIFECT_PROVIDER` | `gemini` | LLM provider (`gemini`, `openai`, `anthropic`) |
| `gemini_api_key` | `GEMINI_API_KEY` | --- | Google Gemini API key |
| `openai_api_key` | `OPENAI_API_KEY` | --- | OpenAI API key |
| `anthropic_api_key` | `ANTHROPIC_API_KEY` | --- | Anthropic API key |
| `github_token` | `GITHUB_TOKEN` | --- | GitHub token for scanning private repos |
| `default_model` | `NERIFECT_MODEL` | `gemini-2.0-flash` | Model to use (provider-specific) |
| `output_format` | `NERIFECT_OUTPUT` | `table` | Default output format (`table`, `json`, `plain`) |
| `data_dir` | `NERIFECT_DATA_DIR` | `~/.nerifect` | Data directory for SQLite database |
| `max_files_per_scan` | --- | `800` | Maximum number of files to scan per run |
| `max_file_size_kb` | --- | `80` | Maximum individual file size in KB |

## Environment Variables

Environment variables take precedence over config file values:

```bash
export NERIFECT_PROVIDER="gemini"       # or "openai" or "anthropic"
export GEMINI_API_KEY="your-api-key"    # for Gemini provider
export OPENAI_API_KEY="sk-..."          # for OpenAI provider
export ANTHROPIC_API_KEY="sk-ant-..."   # for Anthropic provider
export GITHUB_TOKEN="ghp_xxxx"
export NERIFECT_MODEL="gemini-2.5-pro"
export NERIFECT_OUTPUT="json"
export NERIFECT_DATA_DIR="/custom/path"
```

## Supported Models

### Google Gemini

| Model | Description |
|---|---|
| `gemini-2.0-flash` | Default. Fast responses, good for most use cases |
| `gemini-2.5-flash` | Balanced between speed and quality |
| `gemini-2.5-pro` | Highest quality, best for complex compliance analysis |

### OpenAI

| Model | Description |
|---|---|
| `gpt-4o` | Default. Recommended for most use cases |
| `gpt-4o-mini` | Fast, cost-effective |
| `gpt-4-turbo` | High quality |
| `gpt-4.1` | Latest generation |
| `gpt-4.1-mini` | Latest generation, fast |

### Anthropic

| Model | Description |
|---|---|
| `claude-sonnet-4-20250514` | Default. Balanced quality and speed |
| `claude-3-5-haiku-20241022` | Fast, cost-effective |
| `claude-opus-4-20250514` | Highest quality |

## Managing Config via CLI

```bash
# List all values (secrets are masked)
nerifect config list

# Get a specific value
nerifect config get default_model

# Set provider and model
nerifect config set llm_provider openai
nerifect config set default_model gpt-4o
```

## Tracked Repositories

You can store repository scan settings in the config file. See [nerifect repo](cli/nerifect_repo.md) for details.

```yaml
repos:
  - name: my-project
    path: /home/user/projects/my-project
    branch: main
    scan_type: full
    policies:
      - 1
      - 2
```

## Data Storage

Nerifect stores scan results, policies, violations, and fixes in a local SQLite database at `~/.nerifect/nerifect.db`. This requires no external database setup.

To use a custom location:

```bash
nerifect config set data_dir /path/to/custom/dir
```

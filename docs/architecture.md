# Architecture

## Project Structure

```
nerifect-cli/
├── cmd/nerifect/main.go           # Entry point
├── internal/
│   ├── cli/                       # Cobra command definitions
│   │   ├── root.go                # Root command, global flags
│   │   ├── init.go                # Interactive setup wizard
│   │   ├── scan.go                # Scan command
│   │   ├── policy.go              # Policy management
│   │   ├── fix.go                 # Fix generation
│   │   ├── report.go              # Report display
│   │   ├── config_cmd.go          # Config get/set
│   │   └── repo.go                # Repo tracking
│   ├── config/                    # YAML config + env var loading
│   ├── scanner/                   # Scan orchestration
│   │   ├── scanner.go             # Main orchestrator
│   │   ├── files.go               # Local file walker + binary filter
│   │   └── github.go              # Git clone + URL parsing
│   ├── ai/                        # AI/ML framework detection
│   │   ├── detector.go            # 4-phase detection engine
│   │   ├── patterns.go            # 25+ framework registry
│   │   └── risk.go                # EU AI Act risk classification
│   ├── compliance/                # Compliance evaluation
│   │   ├── evaluator.go           # LLM-powered analysis
│   │   ├── pattern.go             # Regex/glob pattern checker
│   │   └── scorer.go              # Score calculation
│   ├── fixer/                     # Fix generation
│   │   ├── fixer.go               # LLM fix generation
│   │   └── diff.go                # Unified diff parser
│   ├── policy/                    # Policy management
│   │   ├── manager.go             # CRUD orchestration
│   │   ├── parser.go              # LLM chunked extraction
│   │   └── fetcher.go             # HTTP fetch + HTML strip
│   ├── llm/                       # Multi-provider LLM client
│   │   ├── client.go              # Provider dispatcher + shared logic
│   │   ├── gemini.go              # Google Gemini REST API
│   │   ├── openai.go              # OpenAI Chat Completions API
│   │   ├── anthropic.go           # Anthropic Messages API
│   │   ├── prompts.go             # Prompt templates
│   │   └── response.go            # JSON extraction
│   ├── store/                     # SQLite storage
│   │   ├── db.go                  # Schema + migrations
│   │   ├── models.go              # Data structures
│   │   ├── scans.go               # Scan CRUD
│   │   ├── policies.go            # Policy CRUD
│   │   ├── violations.go          # Violation CRUD
│   │   ├── detections.go          # AI detection CRUD
│   │   └── fixes.go               # Fix CRUD
│   └── output/                    # Output formatting
│       ├── format.go              # Styles + format dispatch
│       ├── table.go               # Table rendering
│       └── spinner.go             # Progress spinner
├── docs/                          # MkDocs documentation
├── mkdocs.yml                     # MkDocs configuration
├── go.mod
├── go.sum
└── Makefile
```

## Design Decisions

### Standalone Binary

Nerifect CLI is a self-contained binary with no external service dependencies. All data is stored in a local SQLite database, and LLM calls go directly to the configured provider's REST API (Gemini, OpenAI, or Anthropic).

### Direct REST API Calls

Instead of using provider SDKs (which add large dependency trees), Nerifect uses direct HTTP calls to each provider's REST API. This keeps the binary small (~18MB) and reduces build complexity. Supported providers:

- **Google Gemini** --- `generativelanguage.googleapis.com/v1beta`
- **OpenAI** --- `api.openai.com/v1`
- **Anthropic** --- `api.anthropic.com/v1`

### Pure Go SQLite

Uses `modernc.org/sqlite`, a pure Go SQLite implementation that requires no CGO. This enables straightforward cross-compilation to any platform without needing C toolchains.

### Clone-First for GitHub Repos

Rather than making hundreds of GitHub API calls to read individual files, Nerifect performs a single `git clone --depth=1` to a temp directory. This is faster, simpler, and works with private repos via token authentication.

### Exit Code Convention

- `0` --- Scan completed successfully, no critical violations
- `1` --- Tool error (missing config, database error, etc.)
- `2` --- Critical compliance violations found

Exit code `2` is specifically designed for CI/CD gating: pipelines can fail builds when critical violations exist.

## Data Flow

### Scan Flow

```
Target (path or URL)
  │
  ├─ GitHub URL? ──> git clone --depth=1 ──> temp directory
  │                                               │
  └─ Local path? ─────────────────────────────────┘
                                                   │
                                            File Walker
                                    (filter binaries, cap at 800 files)
                                                   │
                             ┌─────────────────────┴──────────────────────┐
                             │                                            │
                      AI Detection                               Compliance Scan
                    (4-phase engine)                          (pattern + LLM eval)
                             │                                            │
                             │                                    ┌───────┴───────┐
                             │                                    │               │
                             │                              Pattern Check    LLM Evaluate
                             │                              (regex/glob)     (LLM API)
                             │                                    │               │
                             └─────────────────────┬──────────────┴───────────────┘
                                                   │
                                            Score Calculation
                                     (100 - severity-weighted penalties)
                                                   │
                                            Save to SQLite
                                                   │
                                            Render Report
                                         (table / JSON / plain)
```

### Policy Ingestion Flow

```
Document (URL or file)
  │
  ├─ URL? ──> HTTP fetch ──> strip HTML tags
  │
  └─ File? ──> read from disk
                    │
             Chunk text (40K chars, 2K overlap)
                    │
              ┌─────┴─────┐
              │     ...    │     (parallel chunks)
              │            │
        LLM extract    LLM extract
         rules JSON      rules JSON
              │            │
              └─────┬──────┘
                    │
              Deduplicate rules
                    │
              Save to SQLite
```

## Dependencies

| Package | Purpose |
|---|---|
| `github.com/spf13/cobra` | CLI framework |
| `gopkg.in/yaml.v3` | Config file parsing |
| `modernc.org/sqlite` | Pure Go SQLite (no CGO) |
| `github.com/charmbracelet/lipgloss` | Terminal styling |
| `github.com/charmbracelet/huh` | Interactive prompts |
| `github.com/briandowns/spinner` | Progress indicators |
| `github.com/gobwas/glob` | Gitignore-style glob matching |

No provider SDKs --- all LLM APIs are called directly via `net/http`.

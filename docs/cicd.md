# CI/CD Integration

Nerifect CLI is designed for CI/CD pipeline integration with JSON output and meaningful exit codes.

## Exit Codes

| Code | Meaning | CI/CD Action |
|---|---|---|
| `0` | Scan completed, no critical violations | Pass the build |
| `1` | Tool error (config missing, scan failure) | Fail the build (infrastructure issue) |
| `2` | Critical violations found | Fail the build (compliance gate) |

## GitHub Actions

### Basic compliance scan

```yaml
name: Compliance Check
on: [push, pull_request]

jobs:
  compliance:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Build Nerifect
        run: |
          git clone https://github.com/nerifect/nerifect-cli.git /tmp/nerifect-cli
          cd /tmp/nerifect-cli && make build
          cp bin/nerifect /usr/local/bin/

      - name: Run compliance scan
        env:
          GEMINI_API_KEY: ${{ secrets.GEMINI_API_KEY }}  # or OPENAI_API_KEY / ANTHROPIC_API_KEY
        run: |
          nerifect scan . --type compliance --output json > scan-results.json
          # Exit code 2 means critical violations found
```

### AI framework detection

```yaml
      - name: Detect AI frameworks
        run: |
          nerifect scan . --type ai --output json > ai-detections.json
```

### Full scan with artifact upload

```yaml
name: Nerifect Governance Scan
on:
  push:
    branches: [main]
  pull_request:

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Install Nerifect
        run: |
          git clone https://github.com/nerifect/nerifect-cli.git /tmp/nerifect-cli
          cd /tmp/nerifect-cli && make build
          sudo cp bin/nerifect /usr/local/bin/

      - name: Run full scan
        env:
          GEMINI_API_KEY: ${{ secrets.GEMINI_API_KEY }}  # or OPENAI_API_KEY / ANTHROPIC_API_KEY
        run: nerifect scan . --output json > scan-results.json
        continue-on-error: true

      - name: Upload scan results
        uses: actions/upload-artifact@v4
        with:
          name: nerifect-scan-results
          path: scan-results.json
```

## GitLab CI

```yaml
nerifect-scan:
  stage: test
  image: golang:1.22
  script:
    - git clone https://github.com/nerifect/nerifect-cli.git /tmp/nerifect-cli
    - cd /tmp/nerifect-cli && make build
    - cp bin/nerifect /usr/local/bin/
    - nerifect scan . --type compliance --output json > scan-results.json
  artifacts:
    paths:
      - scan-results.json
    when: always
  variables:
    GEMINI_API_KEY: $GEMINI_API_KEY  # or OPENAI_API_KEY / ANTHROPIC_API_KEY
```

## JSON Output Format

When using `--output json`, the scan results are structured as:

```json
{
  "scan": {
    "id": 1,
    "target": ".",
    "scan_type": "full",
    "status": "completed",
    "compliance_score": 85,
    "files_scanned": 42,
    "violation_count": 3,
    "ai_detection_count": 5
  },
  "violations": [...],
  "detections": [...]
}
```

This makes it easy to parse results in subsequent pipeline steps for reporting, notifications, or gating decisions.

## Tips

!!! tip "Use exit code 2 for gating"
    Exit code `2` specifically means critical compliance violations were found. Use this to gate merges or deployments while allowing non-critical warnings to pass.

!!! tip "Cache the binary"
    In CI/CD, cache the built Nerifect binary to avoid rebuilding on every run.

!!! tip "Policy pre-loading"
    For consistent CI/CD scans, pre-configure policies in a shared Nerifect data directory or include policy setup as a pipeline step.

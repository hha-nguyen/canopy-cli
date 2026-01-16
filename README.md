# Canopy CLI

Command-line interface for the Canopy mobile app policy compliance scanner.

## Installation

### Quick Install (Recommended)

**macOS / Linux:**
```bash
curl -fsSL https://raw.githubusercontent.com/hha-nguyen/canopy-cli/main/scripts/install.sh | sh
```

### Homebrew (macOS / Linux)
```bash
brew tap hha-nguyen/canopy-tap
brew install canopy
```

### Go Install
```bash
go install github.com/hha-nguyen/canopy-cli@latest
```

### Docker
```bash
docker run --rm -v $(pwd):/project ghcr.io/hha-nguyen/canopy-cli scan /project
```

## Quick Start

### 1. Authenticate
```bash
# Login with your API key
canopy auth login --api-key cpk_your_api_key_here

# Or set via environment variable
export CANOPY_API_KEY=cpk_your_api_key_here
```

### 2. Scan Your Project
```bash
# Scan current directory
canopy scan .

# Scan for specific platform
canopy scan . --platform apple

# Output JSON for CI
canopy scan . --format json --output results.json

# Output SARIF for GitHub Code Scanning
canopy scan . --format sarif --output canopy.sarif
```

## Commands

### `canopy scan`

Scan a project for policy violations.

```bash
canopy scan [path] [flags]

Flags:
  -p, --platform string    Target platform: apple, google, both (default "both")
  -f, --format string      Output format: text, json, sarif (default "text")
  -o, --output string      Write output to file
  -t, --threshold string   Minimum severity to fail: blocker, high, medium, low (default "blocker")
      --timeout duration   Scan timeout (default 5m)
      --no-progress        Disable progress updates
```

### `canopy auth`

Manage authentication.

```bash
# Store API key
canopy auth login --api-key cpk_xxx

# Check authentication status
canopy auth status

# Remove stored credentials
canopy auth logout

# Create new API key
canopy auth token create --name "CI Pipeline"

# List API keys
canopy auth token list

# Revoke API key
canopy auth token revoke <id>
```

### `canopy config`

Manage CLI configuration.

```bash
# Create default config file
canopy config init

# Get config value
canopy config get api_url

# Set config value
canopy config set default_platform apple

# List all config values
canopy config list
```

## Configuration

### Config File

Location: `~/.canopy/config.yaml`

```yaml
api_url: https://api.canopy.app
api_key: cpk_xxxxxxxxxxxx

defaults:
  platform: both
  format: text
  threshold: blocker
  timeout: 5m

output:
  color: true
  progress: true
  quiet: false
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `CANOPY_API_KEY` | API key for authentication |
| `CANOPY_API_URL` | API base URL |

### Config Precedence

1. Command-line flags (highest)
2. Environment variables
3. Config file (`~/.canopy/config.yaml`)
4. Default values (lowest)

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success, no issues at or above threshold |
| 1 | Issues found at or above threshold |
| 2 | Scan failed (error during processing) |
| 3 | Authentication error |
| 4 | Network/API error |
| 5 | Invalid arguments |
| 6 | Timeout |

## CI/CD Integration

### GitHub Actions

```yaml
name: Canopy Scan
on: [push, pull_request]

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Install Canopy CLI
        run: curl -fsSL https://raw.githubusercontent.com/hha-nguyen/canopy-cli/main/scripts/install.sh | sh
        
      - name: Run Canopy Scan
        env:
          CANOPY_API_KEY: ${{ secrets.CANOPY_API_KEY }}
        run: canopy scan . --format sarif --output canopy.sarif
        
      - name: Upload SARIF
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: canopy.sarif
```

### GitLab CI

```yaml
canopy-scan:
  stage: test
  image: ghcr.io/hha-nguyen/canopy-cli:latest
  script:
    - canopy scan . --format json --output gl-canopy-report.json
  artifacts:
    reports:
      codequality: gl-canopy-report.json
  variables:
    CANOPY_API_KEY: $CANOPY_API_KEY
```

### Bitbucket Pipelines

```yaml
pipelines:
  default:
    - step:
        name: Canopy Scan
        script:
          - curl -fsSL https://raw.githubusercontent.com/hha-nguyen/canopy-cli/main/scripts/install.sh | sh
          - canopy scan . --threshold high
```

## Output Formats

### Text (Default)

Human-readable output with colored severity indicators.

### JSON

Full scan result in JSON format, suitable for machine parsing.

```bash
canopy scan . --format json
```

### SARIF

Static Analysis Results Interchange Format, compatible with:
- GitHub Code Scanning
- GitLab Code Quality
- Azure DevOps

```bash
canopy scan . --format sarif --output results.sarif
```

## Development

### Build from Source

```bash
git clone https://github.com/hha-nguyen/canopy-cli.git
cd canopy-cli
make build
```

### Run Tests

```bash
make test
```

### Cross-Compile

```bash
make build-all
```

### Create Release

```bash
make release
```

## License

MIT License - see [LICENSE](LICENSE) for details.

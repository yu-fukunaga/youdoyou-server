# youdoyou-server

Personal life management backend server powered by Go, Firebase, and AI.

## Project Overview

`youdoyou-server` is a Go-based backend providing a chat interface integrated with AI (Genkit) and productivity tools (Notion). It manages conversation history in Firestore and automates life-logging or task management.

### Key Features
- **AI-Driven Chat**: Leverages Firebase Genkit for intelligent interactions.
- **Firestore Integration**: Persistent conversation history and state management.
- **Notion Integration**: Seamlessly syncs with Notion for task and note management.
- **Modular Design**: Clean architecture with separate handlers, services, and repositories.

## Tech Stack
- **Language**: Go 1.25.x
- **Infrastructure**: Firebase (Firestore, Emulators)
- **AI**: Firebase Genkit (Google AI Plugin)
- **Integrations**: Notion API

## Prerequisites
- Go 1.25.x or later installed.
- Firebase CLI installed (for local emulators).
- `gitleaks` (optional, for local security checks).

## Setup

1. **Install Dependencies**:
   ```bash
   make setup
   ```

2. **Configuration**:
   Copy `.env.example` to `.env` and fill in the required environment variables:
   ```bash
   cp .env.example .env
   ```
   Required variables:
   - `PORT`: Server port (default: 8081).
   - `FIRESTORE_PROJECT_ID`: Your Google Cloud Project ID.
   - `NOTION_TOKEN`: Notion Integration Token.
   - `GOOGLE_GENAI_API_KEY`: Google AI API Key.

## Development

All standard tasks are managed via `Makefile`.

| Command | Description |
| :--- | :--- |
| `make build` | Compiles the server and tools binaries into `./bin`. |
| `make run` | Runs the server locally. |
| `make test` | Runs all Go tests. |
| `make lint` | Runs `golangci-lint` check. |
| `make emulators` | Starts Firebase emulators (Firestore). |
| `make seed` | Seeds Firestore emulator with sample data. |
| `make check` | Runs a diagnostic tool to verify Firestore state. |
| `make create-message` | Creates a message in Firestore (requires `THREAD_ID` and `MESSAGE`). |
| `make semgrep` | Runs local security scan using Semgrep. |
| `make secrets` | Runs local secret leak detection using Gitleaks. |
| `make secure` | Runs both Semgrep and Gitleaks checks. |
| `make prepare-release` | Merges develop into main for release preparation. |
| `make ghtag` | Creates a GitHub release (auto-increment patch version). |
| `make ghtag/minor` | Creates a GitHub release (auto-increment minor version). |
| `make ghtag/major` | Creates a GitHub release (auto-increment major version). |
| `make release` | Deploys the latest release tag to Cloud Run. |
| `make release/v1.0.0` | Deploys a specific release tag to Cloud Run. |

### Creating Messages Directly in Firestore

You can create messages in Firestore without using the client app, which will trigger the agent processing via Eventarc (in production).

**Using Make:**
```bash
make create-message THREAD_ID=test-thread-001 MESSAGE="今日のスケジュールを教えて"
```

**Using Script Directly:**
```bash
./scripts/create_message.sh test-thread-001 "今日のタスクを確認して"
```

**Using gcloud CLI:**
```bash
# Set project ID (first time only)
export PROJECT_ID=youdoyou-intelligence

# Create message
gcloud firestore documents create \
  "projects/${PROJECT_ID}/databases/(default)/documents/threads/test-thread-001/messages/" \
  --project="${PROJECT_ID}" \
  --fields="role=user,content=今日の予定は？,status=unread,createdAt=timestamp:{seconds:$(date +%s)}"
```

**Important Notes:**
- Eventarc triggers only work in production Firestore (not in emulator)
- For local testing, use the API endpoint: `POST /v1/agent/chat` with `{"threadId": "..."}`
- The thread must exist before creating messages in it

## Deployment

This project uses release-based deployment workflow. All deployments to production are tagged with semantic versioning.

### Prerequisites

1. **gcloud CLI**: Ensure you are authenticated and the correct project is set.
   ```bash
   gcloud auth login
   gcloud config set project youdoyou-intelligence
   ```

2. **gh CLI** (optional, for GitHub release creation):
   ```bash
   brew install gh
   gh auth login
   ```

3. **Setup Secrets**: Before first deployment, ensure GCP secrets are configured.
   ```bash
   ./scripts/setup_secrets.sh
   ```

### Deployment Workflow

This project uses a three-step release workflow:

1. **Prepare Release**: Merge `develop` into `main` via PR
2. **Create Release**: Create a GitHub release (tags are created on `main` branch automatically)
3. **Deploy**: Deploy the release to Cloud Run

#### Step 1: Prepare Release (Merge develop to main)

```bash
make prepare-release
```

This command will:
- Fetch latest changes from remote
- Show commits that will be merged from `develop` to `main`
- Create a PR for review
- After PR is merged, `main` branch will be updated

#### Step 2: Create a GitHub Release

**Note**: You can run this from any branch. The release tag will be created on the `main` branch on GitHub.

```bash
# Patch version (v1.0.0 -> v1.0.1) - for bug fixes and small improvements
make ghtag

# Minor version (v1.0.0 -> v1.1.0) - for new features
make ghtag/minor

# Major version (v1.0.0 -> v2.0.0) - for breaking changes
make ghtag/major
```

The command will:
- Automatically calculate the next version
- Check for duplicate releases (prevents creating release without changes)
- Warn if `develop` has unmerged commits
- Create release and tag on GitHub's `main` branch

**With custom release notes:**

```bash
./scripts/create_release.sh patch "Bug fixes and performance improvements"
./scripts/create_release.sh minor "New feature: User authentication"
./scripts/create_release.sh major "Breaking: API v2 migration"
```

#### Step 3: Deploy to Cloud Run

After creating a release, deploy it to Cloud Run:

```bash
# Deploy the latest release
make release

# Or deploy a specific release
make release/v1.0.1
```

The deploy command will:
- Checkout the release tag to ensure exact code match
- Deploy to Cloud Run with proper revision naming
- Automatically restore your original branch

#### Complete Workflow Example

```bash
# 1. Prepare release (create PR from develop to main)
make prepare-release
# → Review and merge PR on GitHub

# 2. Create a new release (can run from any branch)
make ghtag

# 3. Deploy the latest release
make release
```

### Semantic Versioning

This project follows [Semantic Versioning](https://semver.org/) (MAJOR.MINOR.PATCH):

- **PATCH** (v1.0.0 → v1.0.1): Bug fixes, small improvements, no API changes
- **MINOR** (v1.0.0 → v1.1.0): New features, backwards-compatible changes
- **MAJOR** (v1.0.0 → v2.0.0): Breaking changes, major refactoring

### Deployment Details

- **Region**: asia-northeast2 (Seoul)
- **Service Name**: youdoyou-server
- **Authentication**: No public access (requires authentication)
- **Secrets**: NOTION_TOKEN and GOOGLE_GENAI_API_KEY are loaded from GCP Secret Manager
- **Revision Naming**: Each release deployment creates a revision with the format `youdoyou-server-{version}` (e.g., `youdoyou-server-1-0-0`)

## Security (Local)

To prevent accidental leakage of API keys or secrets, and to perform security audits locally, it is highly recommended to set up the following tools.

### 1. Gitleaks (Secret Detection)
- **Install Gitleaks**:
  ```bash
  brew install gitleaks
  ```

- **Configure Pre-commit Hook**:
  Create or update `.git/hooks/pre-commit` with the following:
  ```bash
  #!/bin/bash
  gitleaks protect --staged --exit-code 1
  if [ $? -ne 0 ]; then
    echo "❌ gitleaks: シークレットが検出されました。コミットを中止します。"
    exit 1
  fi
  ```

- **Make it Executable**:
  ```bash
  chmod +x .git/hooks/pre-commit
  ```

### 2. Semgrep (Static Analysis)
- **Install Semgrep**:
  ```bash
  brew install semgrep
  ```

- **Run Scan**:
  You can run a local security scan anytime using:
  ```bash
  make secure
  ```

This project uses GitHub Actions for automated verification and security scanning.

### Continuous Integration
The CI workflow (`.github/workflows/ci.yml`) runs on every push and pull request to the `main` branch:
- **setup-deps**: Prepares the Go environment and dependency cache.
- **lint**: Runs static analysis (`golangci-lint`).
- **build**: Verifies the project compiles.
- **test**: Runs the test suite.

### Security and Maintenance
- **Semgrep**: Static analysis security testing (SAST) runs on every PR (`.github/workflows/semgrep.yml`) to detect potential security vulnerabilities.
- **Dependabot**: Automatically monitors and creates pull requests for Go module and GitHub Actions updates (`.github/dependabot.yml`).
- **Gitleaks**: Prevent secret leakage via local `pre-commit` hook (see Security section).

## License

Copyright (c) 2025 Yu Fukunaga. All rights reserved.

This project is proprietary. Unauthorized copying, modification, distribution, or any other use of this software is strictly prohibited. This repository is intended for viewing purposes only.

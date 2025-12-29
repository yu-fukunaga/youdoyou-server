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

## Security (Local)

To prevent accidental leakage of API keys or secrets, it is highly recommended to set up `gitleaks` as a local `pre-commit` hook.

1. **Install Gitleaks**:
   ```bash
   brew install gitleaks
   ```

2. **Configure Pre-commit Hook**:
   Create or update `.git/hooks/pre-commit` with the following:
   ```bash
   #!/bin/bash
   gitleaks protect --staged --exit-code 1
   if [ $? -ne 0 ]; then
     echo "❌ gitleaks: シークレットが検出されました。コミットを中止します。"
     exit 1
   fi
   ```

3. **Make it Executable**:
   ```bash
   chmod +x .git/hooks/pre-commit
   ```

## CI/CD

This project uses GitHub Actions for automated verification. The CI workflow (`.github/workflows/ci.yml`) runs on every push and pull request to the `main` branch, performing:
- **setup-deps**: Prepares the Go environment and dependency cache.
- **lint**: Runs static analysis.
- **build**: Verifies the project compiles.
- **test**: Runs the test suite.

## License

Copyright (c) 2025 Yu Fukunaga. All rights reserved.

This project is proprietary. Unauthorized copying, modification, distribution, or any other use of this software is strictly prohibited. This repository is intended for viewing purposes only.

#!/bin/bash
set -e

# Load required secrets configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/secrets_config.sh"

# Check if release tag is provided
if [ -z "$1" ]; then
  echo "Error: Release tag is required"
  echo "Usage: $0 <release-tag>"
  echo "Example: $0 v1.0.0"
  exit 1
fi

RELEASE_TAG="$1"

# Save current branch/commit for restoration later
CURRENT_BRANCH=$(git symbolic-ref --short HEAD 2>/dev/null || git rev-parse --short HEAD)

# Function to restore original branch
restore_branch() {
  echo ""
  echo "Restoring original branch/commit: $CURRENT_BRANCH"
  git checkout "$CURRENT_BRANCH" 2>/dev/null || true
}

# Set trap to restore branch on exit (success or failure)
trap restore_branch EXIT

# Check for uncommitted changes
if ! git diff-index --quiet HEAD --; then
  echo "Error: You have uncommitted changes in your working directory"
  echo "Please commit or stash your changes before deploying"
  git status --short
  exit 1
fi

# Validate tag format (should start with 'v' followed by semver)
if [[ ! "$RELEASE_TAG" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?$ ]]; then
  echo "Warning: Tag '$RELEASE_TAG' does not follow semantic versioning (v1.2.3)"
  read -p "Do you want to continue anyway? (y/N): " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 1
  fi
fi

# Check if tag exists locally
if ! git rev-parse "$RELEASE_TAG" >/dev/null 2>&1; then
  echo "Error: Tag '$RELEASE_TAG' does not exist in local repository"
  echo "Please create the tag first:"
  echo "  git tag $RELEASE_TAG"
  echo "  git push origin $RELEASE_TAG"
  exit 1
fi

# Check if tag exists on remote
if ! git ls-remote --tags origin | grep -q "refs/tags/$RELEASE_TAG$"; then
  echo "Warning: Tag '$RELEASE_TAG' has not been pushed to remote"
  read -p "Do you want to push the tag and continue? (y/N): " -n 1 -r
  echo
  if [[ $REPLY =~ ^[Yy]$ ]]; then
    git push origin "$RELEASE_TAG"
  else
    exit 1
  fi
fi

# Check if GitHub release exists (requires gh CLI)
if command -v gh &> /dev/null; then
  if ! gh release view "$RELEASE_TAG" &>/dev/null; then
    echo "Warning: GitHub release for '$RELEASE_TAG' does not exist"
    read -p "Do you want to create the release now? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
      gh release create "$RELEASE_TAG" --generate-notes
    else
      echo "Continuing deployment without GitHub release..."
    fi
  fi
else
  echo "Note: 'gh' CLI not found. Skipping GitHub release check."
  echo "Install with: brew install gh"
fi

# Get GCP project ID from gcloud config if not set
if [ -z "$GCP_PROJECT_ID" ]; then
  GCP_PROJECT_ID=$(gcloud config get-value project 2>/dev/null)
  if [ -z "$GCP_PROJECT_ID" ]; then
    echo "Error: Could not determine GCP project ID"
    echo "Please run: gcloud config set project YOUR_PROJECT_ID"
    exit 1
  fi
fi

# Check if required secrets exist
MISSING_SECRETS=()
for SECRET_NAME in "${REQUIRED_SECRETS[@]}"; do
  if ! gcloud secrets describe "$SECRET_NAME" --project="$GCP_PROJECT_ID" &>/dev/null; then
    MISSING_SECRETS+=("$SECRET_NAME")
  fi
done

if [ ${#MISSING_SECRETS[@]} -ne 0 ]; then
  echo "⚠ Warning: The following secrets are missing:"
  for SECRET_NAME in "${MISSING_SECRETS[@]}"; do
    echo "  - $SECRET_NAME"
  done
  echo ""
  echo "Please run: ./scripts/setup_secrets.sh"
  echo ""
  read -p "Do you want to continue anyway? (y/N): " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 1
  fi
fi

# Generate revision suffix from tag (remove 'v' prefix and replace dots with hyphens)
# Cloud Run revision suffix must be lowercase alphanumeric and hyphens only
REVISION_SUFFIX=$(echo "$RELEASE_TAG" | sed 's/^v//' | tr '.' '-' | tr '[:upper:]' '[:lower:]')

# Generate traffic tag from release tag (prepend 'r' for 'release' and use same format as revision)
# Cloud Run traffic tag must start with a letter, use lowercase alphanumeric and hyphens only
TRAFFIC_TAG="r${REVISION_SUFFIX}"

echo "========================================"
echo "Deploying youdoyou-server to Cloud Run"
echo "========================================"
echo "Release Tag:     $RELEASE_TAG"
echo "Project:         $GCP_PROJECT_ID"
echo "Region:          asia-northeast2"
echo "Revision Suffix: $REVISION_SUFFIX"
echo "Traffic Tag:     $TRAFFIC_TAG"
echo "Current Branch:  $CURRENT_BRANCH"
echo "========================================"
echo ""

# Checkout the release tag
echo "Checking out tag: $RELEASE_TAG"
git checkout "$RELEASE_TAG"
echo ""

# Deploy to Cloud Run with release tag as revision suffix
gcloud run deploy youdoyou-server \
  --source . \
  --region asia-northeast2 \
  --platform managed \
  --no-allow-unauthenticated \
  --set-build-env-vars GOOGLE_BUILDABLE=./cmd/server \
  --set-env-vars FIRESTORE_PROJECT_ID=$GCP_PROJECT_ID \
  --update-secrets NOTION_TOKEN=NOTION_TOKEN:latest,GOOGLE_GENAI_API_KEY=GOOGLE_GENAI_API_KEY:latest \
  --revision-suffix "$REVISION_SUFFIX" \
  --project $GCP_PROJECT_ID

echo ""
echo "✅ Deployment completed successfully!"
echo "Release: $RELEASE_TAG"
echo "Revision: youdoyou-server-$REVISION_SUFFIX"

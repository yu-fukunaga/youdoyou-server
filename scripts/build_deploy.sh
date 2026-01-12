#!/bin/bash
set -e

# Load required secrets configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/secrets_config.sh"

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

echo "Deploying youdoyou-server to Cloud Run..."
echo "Project: $GCP_PROJECT_ID"

gcloud run deploy youdoyou-server \
  --source . \
  --region asia-northeast2 \
  --platform managed \
  --no-allow-unauthenticated \
  --set-build-env-vars GOOGLE_BUILDABLE=./cmd/server \
  --set-env-vars FIRESTORE_PROJECT_ID=$GCP_PROJECT_ID \
  --update-secrets NOTION_TOKEN=NOTION_TOKEN:latest,GOOGLE_GENAI_API_KEY=GOOGLE_GENAI_API_KEY:latest \
  --project $GCP_PROJECT_ID

echo "Deployment completed successfully!"

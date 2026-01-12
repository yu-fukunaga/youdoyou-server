#!/bin/bash
set -e

# Load required secrets configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/secrets_config.sh"

# 1. プロジェクトIDの取得
if [ -z "$GCP_PROJECT_ID" ]; then
  GCP_PROJECT_ID=$(gcloud config get-value project 2>/dev/null)
  if [ -z "$GCP_PROJECT_ID" ]; then
    echo "Error: Could not determine GCP project ID"
    echo "Please run: gcloud config set project YOUR_PROJECT_ID"
    exit 1
  fi
fi

# 2. プロジェクト番号（Number）の取得 ※これが大事！
echo "Getting Project Number..."
PROJECT_NUMBER=$(gcloud projects describe "$GCP_PROJECT_ID" --format="value(projectNumber)")
SERVICE_ACCOUNT="${PROJECT_NUMBER}-compute@developer.gserviceaccount.com"

echo "Target Project: $GCP_PROJECT_ID"
echo "Service Account: $SERVICE_ACCOUNT"
echo ""

for SECRET_NAME in "${REQUIRED_SECRETS[@]}"; do
  echo "------------------------------------------------"
  echo "Setting up $SECRET_NAME..."

  # 既に存在するか確認
  if gcloud secrets describe "$SECRET_NAME" --project="$GCP_PROJECT_ID" &>/dev/null; then
    echo "  ✓ Secret exists"
    read -p "  Do you want to update value? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
      # 更新しない場合でも、権限チェックに進むためにcontinueしない
      echo "  Skipping value update."
    else
      # 値の更新
      read -sp "  Enter new value for $SECRET_NAME: " SECRET_VALUE
      echo
      if [ -n "$SECRET_VALUE" ]; then
        echo -n "$SECRET_VALUE" | gcloud secrets versions add "$SECRET_NAME" \
          --data-file=- \
          --project="$GCP_PROJECT_ID" >/dev/null
        echo "  ✓ Value updated"
      fi
    fi
  else
    # 新規作成
    read -sp "  Enter value for $SECRET_NAME: " SECRET_VALUE
    echo
    if [ -z "$SECRET_VALUE" ]; then
      echo "  ⚠ Skipped (empty value)"
      continue
    fi
    echo -n "$SECRET_VALUE" | gcloud secrets create "$SECRET_NAME" \
      --data-file=- \
      --replication-policy="automatic" \
      --project="$GCP_PROJECT_ID" >/dev/null
    echo "  ✓ Secret created"
  fi

  # 4. IAM権限の自動付与 (Secret Accessor)
  echo "  Checking/Adding IAM policy binding..."
  gcloud secrets add-iam-policy-binding "$SECRET_NAME" \
    --member="serviceAccount:${SERVICE_ACCOUNT}" \
    --role="roles/secretmanager.secretAccessor" \
    --project="$GCP_PROJECT_ID" >/dev/null
  echo "  ✓ IAM role 'Secret Accessor' granted to service account"

done

echo "------------------------------------------------"
echo "🎉 Setup completed successfully!"
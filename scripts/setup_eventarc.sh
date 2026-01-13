#!/bin/bash
# scripts/setup_eventarc.sh
# Setup Eventarc trigger for Firestore document creation events

set -e

PROJECT_ID="${PROJECT_ID:-youdoyou-intelligence}"
TRIGGER_REGION="asia-northeast1"  # Must match Firestore database region
SERVICE_REGION="asia-northeast2"  # Cloud Run service region
SERVICE_NAME="youdoyou-server"
TRIGGER_NAME="firestore-message-trigger"
ENDPOINT_PATH="/v1/hooks/firestore"

# Get project number for service account
PROJECT_NUMBER=$(gcloud projects describe "$PROJECT_ID" --format="value(projectNumber)")
SERVICE_ACCOUNT="${PROJECT_NUMBER}-compute@developer.gserviceaccount.com"

echo "========================================"
echo "Setting up Eventarc Trigger"
echo "========================================"
echo "Project:         $PROJECT_ID"
echo "Project Number:  $PROJECT_NUMBER"
echo "Service Account: $SERVICE_ACCOUNT"
echo "Trigger Region:  $TRIGGER_REGION (Firestore)"
echo "Service Region:  $SERVICE_REGION (Cloud Run)"
echo "Service:         $SERVICE_NAME"
echo "Trigger:         $TRIGGER_NAME"
echo "Endpoint:        $ENDPOINT_PATH"
echo "========================================"
echo ""

# Check if trigger already exists
if gcloud eventarc triggers describe "$TRIGGER_NAME" \
  --location="$TRIGGER_REGION" \
  --project="$PROJECT_ID" &>/dev/null; then
  echo "⚠️  Trigger '$TRIGGER_NAME' already exists."
  echo ""
  read -p "Do you want to delete and recreate it? (y/N): " -n 1 -r
  echo
  if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Deleting existing trigger..."
    gcloud eventarc triggers delete "$TRIGGER_NAME" \
      --location="$TRIGGER_REGION" \
      --project="$PROJECT_ID" \
      --quiet
    echo "✅ Trigger deleted."
    echo ""
  else
    echo "Skipping trigger creation."
    exit 0
  fi
fi

# Create Eventarc trigger
echo "Creating Eventarc trigger..."
gcloud eventarc triggers create "$TRIGGER_NAME" \
  --location="$TRIGGER_REGION" \
  --destination-run-service="$SERVICE_NAME" \
  --destination-run-region="$SERVICE_REGION" \
  --destination-run-path="$ENDPOINT_PATH" \
  --event-filters="type=google.cloud.firestore.document.v1.created" \
  --event-filters="database=(default)" \
  --event-filters-path-pattern="document=threads/*/messages/*" \
  --event-data-content-type="application/protobuf" \
  --service-account="$SERVICE_ACCOUNT" \
  --project="$PROJECT_ID"

echo ""
echo "✅ Eventarc trigger created successfully!"
echo ""
echo "The trigger will automatically invoke the Cloud Run service when a new message"
echo "is created in Firestore at: threads/{threadId}/messages/{messageId}"
echo ""
echo "You can view the trigger with:"
echo "  gcloud eventarc triggers describe $TRIGGER_NAME --location=$TRIGGER_REGION --project=$PROJECT_ID"

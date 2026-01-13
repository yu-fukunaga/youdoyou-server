#!/bin/bash
# scripts/create_message.sh
# Create a message in Firestore to trigger agent processing

set -e

THREAD_ID="${1}"
MESSAGE="${2}"
PROJECT_ID="${PROJECT_ID:-youdoyou-intelligence}"

if [ -z "$THREAD_ID" ] || [ -z "$MESSAGE" ]; then
  echo "Usage: $0 <thread-id> <message>"
  echo "Example: $0 test-thread-001 '今日のスケジュールを教えて'"
  exit 1
fi

CURRENT_TIME=$(date +%s)

echo "Creating message in Firestore..."
echo "  Project: $PROJECT_ID"
echo "  Thread ID: $THREAD_ID"
echo "  Message: $MESSAGE"
echo ""

gcloud firestore documents create \
  "projects/${PROJECT_ID}/databases/(default)/documents/threads/${THREAD_ID}/messages/" \
  --project="${PROJECT_ID}" \
  --fields="role=user,content=${MESSAGE},status=unread,createdAt=timestamp:{seconds:${CURRENT_TIME}}"

echo ""
echo "✅ Message created successfully in thread: ${THREAD_ID}"
echo ""
echo "The message will be processed by the agent via Eventarc trigger (production environment)."
echo "For local testing, call the API endpoint directly:"
echo "  curl -X POST http://localhost:8081/v1/agent/chat -d '{\"threadId\":\"${THREAD_ID}\"}'"

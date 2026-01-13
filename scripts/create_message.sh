#!/bin/bash
# scripts/create_message.sh
# Create a message in Firestore to trigger agent processing

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Parse arguments and build go run command
if [ $# -eq 1 ]; then
  # Single argument: MESSAGE only (create new thread)
  MESSAGE="${1}"
  go run "$PROJECT_ROOT/cmd/create-message" --message "$MESSAGE"
elif [ $# -eq 2 ]; then
  # Two arguments: THREAD_ID + MESSAGE (use existing thread)
  THREAD_ID="${1}"
  MESSAGE="${2}"
  go run "$PROJECT_ROOT/cmd/create-message" --thread-id "$THREAD_ID" --message "$MESSAGE"
else
  echo "Usage: $0 [thread-id] <message>"
  echo ""
  echo "Examples:"
  echo "  # Create new thread:"
  echo "  $0 '今日のスケジュールを教えて'"
  echo ""
  echo "  # Add to existing thread:"
  echo "  $0 test-thread-001 '今日のスケジュールを教えて'"
  exit 1
fi

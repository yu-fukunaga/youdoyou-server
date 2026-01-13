#!/bin/bash
set -e

echo "========================================"
echo "Preparing Release"
echo "========================================"
echo "This will create a PR to merge 'develop' into 'main'"
echo "========================================"
echo ""

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
  echo "Error: 'gh' CLI is not installed"
  echo "Please install it with: brew install gh"
  echo "Then authenticate with: gh auth login"
  exit 1
fi

# Fetch latest changes
echo "Fetching latest changes from remote..."
git fetch origin

# Check if there are commits to merge
COMMITS_TO_MERGE=$(git log origin/main..origin/develop --oneline)

if [ -z "$COMMITS_TO_MERGE" ]; then
  echo ""
  echo "✅ Nothing to merge. 'main' is already up to date with 'develop'."
  exit 0
fi

echo ""
echo "The following commits will be merged into 'main':"
echo ""
echo "$COMMITS_TO_MERGE" | head -20

COMMIT_COUNT=$(echo "$COMMITS_TO_MERGE" | wc -l | tr -d ' ')
if [ "$COMMIT_COUNT" -gt 20 ]; then
  echo "... and $((COMMIT_COUNT - 20)) more commits"
fi

# Create PR from develop to main
echo ""
echo "Creating pull request..."

# Generate PR title with current date (Release YYYYMMDD)
PR_TITLE="Release $(date +%Y%m%d)"

# Build PR body with commit list
PR_BODY="## Release Preparation

This PR merges the latest changes from \`develop\` into \`main\` to prepare for a new release.

## Changes

$COMMITS_TO_MERGE

🤖 Generated with [Claude Code](https://claude.com/claude-code)"

PR_URL=$(gh pr create \
  --base main \
  --head develop \
  --title "$PR_TITLE" \
  --body "$PR_BODY")

echo ""
echo "✅ Pull request created successfully!"
echo "$PR_URL"
echo ""
echo "Next steps:"
echo "  1. Review and merge the PR on GitHub"
echo "  2. After merge, checkout main: git checkout main && git pull"
echo "  3. Create release: make ghtag"
echo "  4. Deploy: make release"

#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Get increment type or specific version (default: patch)
INCREMENT_TYPE="${1:-patch}"
RELEASE_NOTES="${2:-}"

# Fetch latest changes to ensure we have up-to-date tags
echo "Fetching latest tags from remote..."
git fetch --tags origin

# Calculate next version
RELEASE_TAG=$("$SCRIPT_DIR/next_version.sh" "$INCREMENT_TYPE")

if [ -z "$RELEASE_TAG" ]; then
  echo "Error: Could not determine next version"
  exit 1
fi

# Get current latest tag from main branch for display
LATEST_TAG=$(git describe --tags --abbrev=0 origin/main 2>/dev/null || echo "none")

echo "========================================"
echo "Creating GitHub Release"
echo "========================================"
echo "Current version: $LATEST_TAG"
echo "New version:     $RELEASE_TAG"
echo "Increment type:  $INCREMENT_TYPE"
echo "========================================"
echo ""

# Check if there are any changes since the last tag on main branch
if [ "$LATEST_TAG" != "none" ]; then
  if git diff --quiet "$LATEST_TAG" origin/main; then
    echo "Error: No changes detected since last tag ($LATEST_TAG) on main branch"
    echo "Cannot create a new release without any code changes."
    exit 1
  fi
fi

# Fetch latest remote branches
echo "Fetching latest remote branches..."
git fetch origin main develop 2>/dev/null || git fetch origin 2>/dev/null

# Check if there are unmerged commits in develop that are not in main
if git rev-parse --verify origin/develop >/dev/null 2>&1 && git rev-parse --verify origin/main >/dev/null 2>&1; then
  UNMERGED_COMMITS=$(git log origin/main..origin/develop --oneline 2>/dev/null)

  if [ -n "$UNMERGED_COMMITS" ]; then
    echo ""
    echo "⚠️  Warning: There are commits in 'develop' that have not been merged to 'main':"
    echo ""
    echo "$UNMERGED_COMMITS" | head -10

    COMMIT_COUNT=$(echo "$UNMERGED_COMMITS" | wc -l | tr -d ' ')
    if [ "$COMMIT_COUNT" -gt 10 ]; then
      echo "... and $((COMMIT_COUNT - 10)) more commits"
    fi

    echo ""
    echo "It is recommended to merge 'develop' into 'main' before creating a release."
    echo ""
    read -p "Do you want to continue creating the release anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
      echo "Release creation cancelled."
      exit 1
    fi
  fi
fi

echo ""

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
  echo "Error: 'gh' CLI is not installed"
  echo "Please install it with: brew install gh"
  echo "Then authenticate with: gh auth login"
  exit 1
fi

# Check if tag already exists on remote
if git ls-remote --tags origin | grep -q "refs/tags/$RELEASE_TAG$"; then
  echo "Error: Tag '$RELEASE_TAG' already exists on remote"
  echo "Cannot create a duplicate release."
  exit 1
fi

# Create GitHub release (this will create the tag on main branch)
echo "Creating GitHub release on 'main' branch..."
if [ -n "$RELEASE_NOTES" ]; then
  # Use provided release notes
  gh release create "$RELEASE_TAG" \
    --target main \
    --title "$RELEASE_TAG" \
    --notes "$RELEASE_NOTES"
else
  # Auto-generate release notes from commits
  gh release create "$RELEASE_TAG" \
    --target main \
    --title "$RELEASE_TAG" \
    --generate-notes
fi

echo ""
echo "✅ Release created successfully!"
echo "Tag: $RELEASE_TAG"
echo ""
echo "Next steps:"
echo "  1. Deploy with: make release/$RELEASE_TAG"
echo "  2. Or run: ./scripts/deploy_release.sh $RELEASE_TAG"

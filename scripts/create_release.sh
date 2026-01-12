#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Get increment type or specific version (default: patch)
INCREMENT_TYPE="${1:-patch}"
RELEASE_NOTES="${2:-}"

# Calculate next version
RELEASE_TAG=$("$SCRIPT_DIR/next_version.sh" "$INCREMENT_TYPE")

if [ -z "$RELEASE_TAG" ]; then
  echo "Error: Could not determine next version"
  exit 1
fi

# Get current latest tag for display
LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "none")

echo "========================================"
echo "Creating GitHub Release"
echo "========================================"
echo "Current version: $LATEST_TAG"
echo "New version:     $RELEASE_TAG"
echo "Increment type:  $INCREMENT_TYPE"
echo "========================================"
echo ""

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
  echo "Error: 'gh' CLI is not installed"
  echo "Please install it with: brew install gh"
  echo "Then authenticate with: gh auth login"
  exit 1
fi

# Check if tag already exists locally
if git rev-parse "$RELEASE_TAG" >/dev/null 2>&1; then
  echo "Warning: Tag '$RELEASE_TAG' already exists locally"
  read -p "Do you want to continue and push it? (y/N): " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 1
  fi
else
  # Check for uncommitted changes
  if ! git diff-index --quiet HEAD --; then
    echo "Warning: You have uncommitted changes"
    git status --short
    echo ""
    read -p "Do you want to continue creating the release? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
      exit 1
    fi
  fi

  # Create tag
  echo "Creating tag '$RELEASE_TAG'..."
  git tag -a "$RELEASE_TAG" -m "Release $RELEASE_TAG"
fi

# Push tag to remote
echo "Pushing tag to remote..."
git push origin "$RELEASE_TAG"

# Create GitHub release
echo "Creating GitHub release..."
if [ -n "$RELEASE_NOTES" ]; then
  # Use provided release notes
  gh release create "$RELEASE_TAG" --title "$RELEASE_TAG" --notes "$RELEASE_NOTES"
else
  # Auto-generate release notes from commits
  gh release create "$RELEASE_TAG" --title "$RELEASE_TAG" --generate-notes
fi

echo ""
echo "✅ Release created successfully!"
echo "Tag: $RELEASE_TAG"
echo ""
echo "Next steps:"
echo "  1. Deploy with: make release/$RELEASE_TAG"
echo "  2. Or run: ./scripts/deploy_release.sh $RELEASE_TAG"

#!/bin/bash
set -e

# Get the increment type (patch, minor, major)
INCREMENT_TYPE="${1:-patch}"

# Validate increment type
if [[ ! "$INCREMENT_TYPE" =~ ^(patch|minor|major)$ ]]; then
  echo "Error: Invalid increment type '$INCREMENT_TYPE'" >&2
  echo "Valid options: patch, minor, major" >&2
  exit 1
fi

# Get the latest tag from git
LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")

# Remove 'v' prefix if present
VERSION="${LATEST_TAG#v}"

# Parse version components
IFS='.' read -r MAJOR MINOR PATCH <<< "$VERSION"

# Remove any pre-release suffix from PATCH (e.g., "3-beta" -> "3")
PATCH="${PATCH%%-*}"

# Validate that we have numeric values
if ! [[ "$MAJOR" =~ ^[0-9]+$ ]] || ! [[ "$MINOR" =~ ^[0-9]+$ ]] || ! [[ "$PATCH" =~ ^[0-9]+$ ]]; then
  echo "Error: Could not parse version from tag '$LATEST_TAG'" >&2
  echo "Starting from v0.0.0" >&2
  MAJOR=0
  MINOR=0
  PATCH=0
fi

# Calculate next version based on increment type
case "$INCREMENT_TYPE" in
  patch)
    PATCH=$((PATCH + 1))
    ;;
  minor)
    MINOR=$((MINOR + 1))
    PATCH=0
    ;;
  major)
    MAJOR=$((MAJOR + 1))
    MINOR=0
    PATCH=0
    ;;
esac

# Output the next version
echo "v${MAJOR}.${MINOR}.${PATCH}"

#!/bin/bash
# tools/build-release.sh - Build in release mode with version tags

set -euo pipefail

# Check if we're on a clean git tree
if [ -n "$(git status --porcelain 2>/dev/null)" ]; then
  echo "❌ ERROR: Cannot build release with uncommitted changes"
  echo "Please commit or stash your changes first"
  exit 1
fi

echo "🚀 Building in release mode..."

# Set release mode
export RELEASE_MODE=true

# Get the version that will be used
VERSION=$(./tools/bazel_stamp_vars.sh | grep STABLE_VERSION_TAG | cut -d' ' -f2)
echo "📦 Images will be tagged with version: $VERSION"

# Confirm with user
read -p "Continue with release build? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  echo "❌ Release build cancelled"
  exit 1
fi

# Build all targets
echo "🔨 Building all targets..."
bazel build //...

# Test all targets
echo "🧪 Running tests..."
bazel test //...

echo "✅ Release build complete!"
echo "📦 Version: $VERSION"
echo "💡 All images are now tagged with the release version"
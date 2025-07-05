#!/bin/bash
# tools/bazel_stamp_vars.sh

set -euo pipefail

# Check for release mode via environment variable
RELEASE_MODE="${RELEASE_MODE:-false}"

# Get git information
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_TAG=$(git describe --tags --exact-match 2>/dev/null || echo "")
GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")

# Check if working directory is dirty
if [ -n "$(git status --porcelain 2>/dev/null)" ]; then
  GIT_COMMIT_SUFFIX="-dirty"
else
  GIT_COMMIT_SUFFIX=""
fi

STABLE_GIT_COMMIT="${GIT_COMMIT}${GIT_COMMIT_SUFFIX}"
echo "STABLE_GIT_COMMIT ${STABLE_GIT_COMMIT}"
echo "STABLE_GIT_BRANCH ${GIT_BRANCH}"

# Determine version tag based on mode
if [ "$RELEASE_MODE" = "true" ]; then
  # Release mode: use git tag if available, otherwise use version from MODULE.bazel
  if [ -n "$GIT_TAG" ]; then
    VERSION_TAG="$GIT_TAG"
  elif [ -n "$CI_TAG" ]; then
    VERSION_TAG="$CI_TAG"
  elif [ -n "$GITHUB_REF_NAME" ] && [[ "$GITHUB_REF_NAME" =~ ^v[0-9]+\.[0-9]+\.[0-9]+ ]]; then
    VERSION_TAG="$GITHUB_REF_NAME"
  else
    # Fallback: extract version from MODULE.bazel
    MODULE_VERSION=$(grep -E '^\s*version\s*=' MODULE.bazel | sed -E 's/.*version\s*=\s*"([^"]+)".*/\1/' || echo "0.1.0")
    VERSION_TAG="v${MODULE_VERSION}"
  fi
  
  # In release mode, ensure we're on a clean tree
  if [ "$GIT_COMMIT_SUFFIX" = "-dirty" ]; then
    echo "ERROR: Cannot build release with dirty working tree" >&2
    exit 1
  fi
else
  # Development mode: always use dev tag with commit
  if [ "$GIT_COMMIT" != "unknown" ]; then
    VERSION_TAG="dev-${GIT_COMMIT}${GIT_COMMIT_SUFFIX}"
  else
    VERSION_TAG="dev-local"
  fi
fi

echo "STABLE_VERSION_TAG ${VERSION_TAG}"
echo "STABLE_IMAGE_VERSION ${VERSION_TAG}"
echo "STABLE_RELEASE_MODE ${RELEASE_MODE}"

# Additional stamp variables
BUILD_TIMESTAMP=$(date +%s)
echo "BUILD_TIMESTAMP ${BUILD_TIMESTAMP}"
BUILD_TIMESTAMP_RFC3339=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
echo "BUILD_TIMESTAMP_RFC3339 ${BUILD_TIMESTAMP_RFC3339}"
BUILD_USER=$(whoami)
echo "BUILD_USER ${BUILD_USER}"
BUILD_HOST=$(hostname)
echo "BUILD_HOST ${BUILD_HOST}"
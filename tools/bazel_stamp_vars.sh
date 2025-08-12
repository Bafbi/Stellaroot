#!/bin/bash
# tools/bazel_stamp_vars.sh

set -euo pipefail

# Check for release mode via environment variable
RELEASE_MODE="${RELEASE_MODE:-false}"

# Track which VCS provided information
VCS="git"

# Get git information
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_TAG=$(git describe --tags --exact-match 2>/dev/null || echo "")
GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")

# If Git info is unknown, try jujutsu (jj) as a backend
if [ "$GIT_COMMIT" = "unknown" ]; then
  if command -v jj >/dev/null 2>&1; then
    JJ_COMMIT=$(jj log -r @ -n 1 --no-graph -T 'commit_id.short()' 2>/dev/null || true)
    if [ -n "${JJ_COMMIT:-}" ]; then
      GIT_COMMIT="$JJ_COMMIT"
      VCS="jj"
    fi
    # jj doesn't have branches like Git; use current bookmarks as a reasonable stand-in
    JJ_BOOKMARKS=$(jj log -r @ -n 1 --no-graph -T 'separate(" ", current_bookmarks())' 2>/dev/null || true)
    if [ -n "${JJ_BOOKMARKS:-}" ]; then
      # Take the first bookmark as the main "branch" indicator
      GIT_BRANCH="${JJ_BOOKMARKS%% *}"
    else
      GIT_BRANCH="unknown"
    fi
    # Best-effort tag from jj: if a bookmark looks like a semver tag, treat it as tag
    case "$GIT_BRANCH" in
      v[0-9]*.[0-9]*.[0-9]*) GIT_TAG="$GIT_BRANCH" ;;
    esac
  fi
fi

# Check if working directory is dirty
GIT_DIRTY="false"
DIRTY_OUTPUT="$(git status --porcelain 2>/dev/null || true)"
if [ -n "$DIRTY_OUTPUT" ]; then
  GIT_COMMIT_SUFFIX="-dirty"
  GIT_DIRTY="true"
else
  GIT_COMMIT_SUFFIX=""
fi

# If Git didn't indicate dirtiness and jj exists, check jj working copy
if [ "$GIT_DIRTY" = "false" ] && command -v jj >/dev/null 2>&1; then
  if jj diff -r @- --summary >/dev/null 2>&1; then
    if [ -n "$(jj diff -r @- --summary 2>/dev/null | head -n1)" ]; then
      GIT_COMMIT_SUFFIX="-dirty"
      GIT_DIRTY="true"
    fi
  fi
fi

STABLE_GIT_COMMIT="${GIT_COMMIT}${GIT_COMMIT_SUFFIX}"
echo "STABLE_GIT_COMMIT ${STABLE_GIT_COMMIT}"
 
# Handle detached HEADs more helpfully for branch display
if [ "$GIT_BRANCH" = "HEAD" ] || [ "$GIT_BRANCH" = "unknown" ]; then
  if [ -n "$GIT_TAG" ]; then
    GIT_BRANCH="$GIT_TAG"
  elif [ "$GIT_COMMIT" != "unknown" ]; then
    GIT_BRANCH="detached-${GIT_COMMIT}"
  fi
fi

echo "STABLE_GIT_BRANCH ${GIT_BRANCH}"
echo "STABLE_GIT_TAG ${GIT_TAG}"
echo "STABLE_GIT_DIRTY ${GIT_DIRTY}"
echo "STABLE_VCS ${VCS}"
if [ -n "${JJ_BOOKMARKS:-}" ]; then
  echo "STABLE_JJ_BOOKMARKS ${JJ_BOOKMARKS}"
fi
if [ "${VCS}" = "jj" ] && [ -n "${JJ_COMMIT:-}" ]; then
  echo "STABLE_JJ_COMMIT ${JJ_COMMIT}"
fi

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
    # Fallback: extract version from MODULE.bazel (robust to single-line or multi-line module())
    MODULE_VERSION=$(grep -Eo 'version[[:space:]]*=[[:space:]]*"[^"]+"' MODULE.bazel 2>/dev/null | head -n1 | sed -E 's/.*"([^"]+)".*/\1/' || true)
    if [ -z "${MODULE_VERSION:-}" ]; then
      MODULE_VERSION="0.1.0"
    fi
    VERSION_TAG="v${MODULE_VERSION}"
  fi

  # In release mode, ensure we're on a clean tree
  if [ "$GIT_COMMIT_SUFFIX" = "-dirty" ]; then
    echo "ERROR: Cannot build release with dirty working tree" >&2
    exit 1
  fi
else
  # Development mode: informative dev tag including branch and commit (sanitized)
  if [ "$GIT_COMMIT" != "unknown" ]; then
    SAN_BRANCH=$(echo "${GIT_BRANCH}" | tr -c 'A-Za-z0-9._-' '-')
    # Trim any trailing hyphens introduced by sanitization
    SAN_BRANCH="${SAN_BRANCH%-}"
    ADD_BRANCH=""
    if [ -n "$SAN_BRANCH" ] && [ "$SAN_BRANCH" != "unknown" ]; then
      if [[ "$SAN_BRANCH" != detached-* ]] && [[ "$SAN_BRANCH" != *"-$GIT_COMMIT" ]]; then
        ADD_BRANCH="-$SAN_BRANCH"
      fi
    fi
    VERSION_TAG="dev${ADD_BRANCH}-${GIT_COMMIT}${GIT_COMMIT_SUFFIX}"
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

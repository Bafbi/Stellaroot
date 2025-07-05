#!/bin/bash
# tools/build-dev.sh - Build in development mode with dev tags

set -euo pipefail

echo "🔧 Building in development mode..."
echo "📦 Images will be tagged with 'dev-' prefix"

# Ensure we're in dev mode
export RELEASE_MODE=false

# Build all targets
bazel build //...

echo "✅ Development build complete!"
echo "💡 Run 'bazel run //services/dashboard:dashboard' to start services"
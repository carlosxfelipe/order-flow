#!/usr/bin/env bash
set -e

# 1. Format Go code
if command -v go >/dev/null 2>&1; then
  echo "💧 Formatting Go code..."
  go fmt ./...
else
  echo "⚠️  'go' not found, skipping Go formatting."
fi

# 2. Format JavaScript files with Prettier (.js only, ignoring .ts)
if command -v prettier >/dev/null 2>&1; then
  echo "✨ Formatting JavaScript files (.js only with prettier)..."
  find . -name "*.js" -not -path "*/.*" -not -path "*/node_modules/*" -exec prettier --write {} +
elif command -v npx >/dev/null 2>&1; then
  echo "✨ Formatting JavaScript files (.js only with npx prettier)..."
  find . -name "*.js" -not -path "*/.*" -not -path "*/node_modules/*" -exec npx --yes prettier --write {} +
else
  echo "⚠️  Neither 'prettier' nor 'npx' found. Skipping JavaScript formatting."
fi

# 3. Format Shell scripts (prefers native shfmt from brew, falls back to uvx)
if command -v shfmt >/dev/null 2>&1; then
  echo "📜 Formatting shell scripts (shfmt)..."
  shopt -s globstar 2>/dev/null || true
  if shopt -q globstar 2>/dev/null; then
    shfmt -i 2 -w **/*.sh
  else
    find . -name "*.sh" -not -path "*/.*" -exec shfmt -i 2 -w {} +
  fi
elif command -v uvx >/dev/null 2>&1; then
  echo "📜 Formatting shell scripts (uvx shfmt-py)..."
  shopt -s globstar 2>/dev/null || true
  if shopt -q globstar 2>/dev/null; then
    uvx --from shfmt-py shfmt -i 2 -w **/*.sh
  else
    find . -name "*.sh" -not -path "*/.*" -exec uvx --from shfmt-py shfmt -i 2 -w {} +
  fi
else
  echo "⚠️  Neither 'shfmt' nor 'uvx' found. Install with: brew install shfmt"
fi

echo "✅ Done!"

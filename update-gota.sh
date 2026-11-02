#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "💧 Updating github.com/carlosxfelipe/gota to latest commit on main..."
GOPROXY=direct go get github.com/carlosxfelipe/gota@main
go mod tidy

echo "🧪 Running tests to verify build..."
go test ./...

VERSION=$(grep 'github.com/carlosxfelipe/gota' go.mod | awk '{print $NF}')
echo "✅ Gota successfully updated to: ${VERSION}"

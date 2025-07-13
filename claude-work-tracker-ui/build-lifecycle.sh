#!/bin/bash

# Build the lifecycle management tool
echo "🔨 Building lifecycle management tool..."

go build -o lifecycle ./cmd/lifecycle/main.go

if [ $? -eq 0 ]; then
    echo "✅ Built: lifecycle"
    echo ""
    echo "Usage examples:"
    echo "  ./lifecycle analyze      - Analyze system decay and health"
    echo "  ./lifecycle health       - Show system health metrics"
    echo "  ./lifecycle cleanup      - Interactive cleanup mode"
    echo "  ./lifecycle auto-cleanup - Execute all auto-safe cleanup actions"
    echo "  ./lifecycle refresh      - Refresh all activity scores"
else
    echo "❌ Build failed"
    exit 1
fi
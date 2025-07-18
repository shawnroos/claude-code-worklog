#!/bin/bash

# Build the enhanced consolidation tool
echo "🔨 Building hierarchical consolidation tool..."

go build -o consolidate-hierarchy ./cmd/consolidate-hierarchy/main.go

if [ $? -eq 0 ]; then
    echo "✅ Built: consolidate-hierarchy"
    echo ""
    echo "Usage examples:"
    echo "  ./consolidate-hierarchy analyze       - Find artifact clusters"
    echo "  ./consolidate-hierarchy interactive   - Interactive consolidation"
    echo "  ./consolidate-hierarchy auto-group    - Auto-create groups"
    echo "  ./consolidate-hierarchy ready-groups  - Show ready groups"
    echo "  ./consolidate-hierarchy consolidate <group-id> - Create Work from group"
else
    echo "❌ Build failed"
    exit 1
fi
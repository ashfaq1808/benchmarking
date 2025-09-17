#!/bin/bash

# Cassandra Schema Setup Script
# This script sets up the Cassandra keyspace, table, and indexes for the benchmark tool

set -e  # Exit on any error

CONFIG_FILE="${1:-config.yaml}"
SETUP_BINARY="./setup-schema"

echo "🚀 Cassandra Benchmark Schema Setup"
echo "=================================="
echo

# Check if config file exists
if [ ! -f "$CONFIG_FILE" ]; then
    echo "❌ Config file '$CONFIG_FILE' not found!"
    echo "Usage: $0 [config-file]"
    echo "Default config file: config.yaml"
    exit 1
fi

echo "📋 Using config file: $CONFIG_FILE"
echo

# Build the setup tool if it doesn't exist or is older than source
if [ ! -f "$SETUP_BINARY" ] || [ cmd/setup-schema/main.go -nt "$SETUP_BINARY" ]; then
    echo "🔨 Building schema setup tool..."
    go build -o setup-schema cmd/setup-schema/main.go
    echo "✅ Build completed"
    echo
fi

# Run the schema setup
echo "🏗️ Running schema setup..."
echo
$SETUP_BINARY "$CONFIG_FILE"

echo
echo "✨ Schema setup completed! Your Cassandra database is ready for benchmarking."
echo
echo "🎯 Next steps:"
echo "   1. Run: go build -o benchmark-new cmd/benchmark/main.go"
echo "   2. Run: ./benchmark-new"
echo "   3. View results in the Streamlit dashboard"
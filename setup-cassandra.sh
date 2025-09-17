#!/bin/bash

# Cassandra Schema Setup Script
# This script sets up the Cassandra keyspace, table, and indexes for the benchmark tool

set -e  # Exit on any error

CONFIG_FILE="${1:-config.yaml}"

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

# Check if Go module is available
if [ ! -f "go.mod" ]; then
    echo "❌ go.mod not found - not in a Go module directory"
    exit 1
fi

# Check if setup-schema source exists
if [ ! -d "cmd/setup-schema" ]; then
    echo "❌ setup-schema source directory not found"
    exit 1
fi

# Run the schema setup
echo "🏗️ Running schema setup..."
echo
go run cmd/setup-schema/main.go "$CONFIG_FILE"

echo
echo "✨ Schema setup completed! Your Cassandra database is ready for benchmarking."
echo
echo "🎯 Next steps:"
echo "   1. Run: go run cmd/benchmark/main.go"
echo "   2. Or use: ./benchmark-monitoring.sh"
echo "   3. View results in the Streamlit dashboard"
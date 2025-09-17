#!/bin/bash

set -e  # Exit on error

CONFIG_FILE="config.yaml"
MONITORING_PID=""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}🔧 $1${NC}"
}

# Function to cleanup monitoring on exit
cleanup() {
    if [ ! -z "$MONITORING_PID" ]; then
        print_info "Stopping system monitoring..."
        kill $MONITORING_PID 2>/dev/null || true
        wait $MONITORING_PID 2>/dev/null || true
        print_status "System monitoring stopped"
    fi
}

# Set trap to cleanup on exit
trap cleanup EXIT INT TERM

# Check if config file exists
if [ ! -f "$CONFIG_FILE" ]; then
    print_error "Config file '$CONFIG_FILE' not found"
    exit 1
fi

print_info "Starting Cassandra benchmark with monitoring..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

# Check if the benchmark binary exists, if not build it
BENCHMARK_BINARY="./benchmark-monitoring"
if [ ! -f "$BENCHMARK_BINARY" ] || [ ! -x "$BENCHMARK_BINARY" ]; then
    print_info "Building benchmark binary..."
    go build -o benchmark-monitoring ./cmd/benchmark/
    if [ $? -ne 0 ]; then
        print_error "Failed to build benchmark binary"
        exit 1
    fi
    print_status "Benchmark binary built successfully"
fi

# Start system monitoring in background
start_monitoring() {
    print_info "Initializing system monitoring..."

    # Create a simple monitoring script
    cat > monitor.sh << 'EOF'
#!/bin/bash
METRICS_FILE="system_metrics.json"
echo '{"metrics": [' > "$METRICS_FILE"
FIRST=true

while true; do
    TIMESTAMP=$(date +%s)
    CPU_USAGE=$(top -l 1 -s 0 | grep "CPU usage" | awk '{print $3}' | sed 's/%//')
    MEMORY_INFO=$(vm_stat | grep -E "Pages (free|active|inactive|speculative|throttled|wired|purgeable)" | awk '{print $3}' | sed 's/\.//')

    if [ "$FIRST" = false ]; then
        echo ',' >> "$METRICS_FILE"
    fi
    FIRST=false

    echo "  {" >> "$METRICS_FILE"
    echo "    \"timestamp\": $TIMESTAMP," >> "$METRICS_FILE"
    echo "    \"cpu_usage\": \"$CPU_USAGE\"," >> "$METRICS_FILE"
    echo "    \"memory_info\": \"$(echo $MEMORY_INFO | tr '\n' ' ')\"" >> "$METRICS_FILE"
    echo "  }" >> "$METRICS_FILE"

    sleep 5
done
EOF

    chmod +x monitor.sh
    ./monitor.sh &
    MONITORING_PID=$!
    print_status "System monitoring started (PID: $MONITORING_PID)"
}

# Check if monitoring is enabled in config
MONITORING_ENABLED=$(grep -A 10 "monitoring:" "$CONFIG_FILE" | grep "enabled:" | awk '{print $2}')
if [ "$MONITORING_ENABLED" = "true" ]; then
    start_monitoring
fi

# Setup schema first
print_info "Setting up Cassandra schema..."
if [ -f "./setup-schema" ]; then
    ./setup-schema
    if [ $? -eq 0 ]; then
        print_status "Schema setup completed"
    else
        print_warning "Schema setup had issues but continuing..."
    fi
else
    print_warning "setup-schema binary not found, skipping schema setup"
fi

# Run the benchmark
print_info "Starting benchmark execution..."

# Try to run the native binary first
if [ -f "$BENCHMARK_BINARY" ] && [ -x "$BENCHMARK_BINARY" ]; then
    # Check the architecture
    if file "$BENCHMARK_BINARY" | grep -q "$(uname -m)"; then
        print_info "Running native benchmark binary..."
        "$BENCHMARK_BINARY"
    else
        print_warning "Binary architecture mismatch, rebuilding..."
        go build -o benchmark-monitoring ./cmd/benchmark/
        "$BENCHMARK_BINARY"
    fi
else
    print_info "Building and running benchmark from source..."
    go run ./cmd/benchmark/
fi

if [ $? -eq 0 ]; then
    print_status "Benchmarking completed successfully"

    # Finalize monitoring data
    if [ ! -z "$MONITORING_PID" ]; then
        kill $MONITORING_PID 2>/dev/null || true
        wait $MONITORING_PID 2>/dev/null || true
        echo ']' >> system_metrics.json
        MONITORING_PID=""
        print_status "Monitoring data saved to system_metrics.json"
    fi

    # Launch visualization if available
    if [ -f "visualize.py" ]; then
        print_info "Launching Streamlit dashboard at http://localhost:8501..."
        if command -v streamlit &> /dev/null; then
            streamlit run visualize.py
        elif command -v python3 &> /dev/null; then
            python3 -m streamlit run visualize.py
        else
            print_warning "Streamlit not found. Results saved to result.json"
            print_info "Install streamlit with: pip install streamlit"
        fi
    else
        print_info "Results saved to result.json"
    fi
else
    print_error "Benchmark execution failed"
    exit 1
fi
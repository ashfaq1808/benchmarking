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

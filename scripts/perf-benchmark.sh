#!/bin/bash

# Performance benchmark script for CI pipeline
# Runs aztui performance tests with and without caching to show improvements

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
AZTUI_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
SRC_DIR="${AZTUI_ROOT}/src"
PERFBENCH_BIN="${AZTUI_ROOT}/bin/perfbench"
CONFIG_PATH="${AZTUI_ROOT}/conf/default.yaml"

# Configuration
ITERATIONS=${PERF_ITERATIONS:-5}
OUTPUT_DIR="${AZTUI_ROOT}/perf-results"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

echo "=== Azure TUI Performance Benchmark ==="
echo "Timestamp: $(date)"
echo "Iterations per test: ${ITERATIONS}"
echo "Config: ${CONFIG_PATH}"
echo "Output directory: ${OUTPUT_DIR}"
echo

# Create output directory
mkdir -p "${OUTPUT_DIR}"

# Build the performance benchmark tool
echo "Building performance benchmark tool..."
cd "${SRC_DIR}"
go build -o "${PERFBENCH_BIN}" cmd/perfbench/main.go
echo "✓ Built perfbench tool at ${PERFBENCH_BIN}"
echo

# Function to run benchmark and save results
run_benchmark() {
    local mode="$1"
    local cache_flag="$2"
    local output_file="${OUTPUT_DIR}/${mode}_${TIMESTAMP}.json"
    
    echo "Running benchmark: ${mode}"
    echo "Command: ${PERFBENCH_BIN} ${cache_flag} -iterations ${ITERATIONS} -json -config ${CONFIG_PATH}"
    
    # Set environment for Azure CLI (in case it's needed)
    export AZTUI_CONFIG_PATH="${CONFIG_PATH}"
    
    # Run the benchmark
    if "${PERFBENCH_BIN}" ${cache_flag} -iterations "${ITERATIONS}" -json -config "${CONFIG_PATH}" > "${output_file}"; then
        echo "✓ ${mode} benchmark completed"
        echo "  Results saved to: ${output_file}"
        
        # Extract key metrics for quick summary
        local total_ops=$(jq -r '.total_operations' "${output_file}")
        local avg_duration=$(jq -r '.average_duration_ns' "${output_file}")
        local cache_hits=$(jq -r '.cache_hits' "${output_file}")
        local cache_misses=$(jq -r '.cache_misses' "${output_file}")
        local hit_ratio=$(jq -r '.cache_hit_ratio' "${output_file}")
        
        echo "  Total operations: ${total_ops}"
        echo "  Average duration: ${avg_duration} ns"
        echo "  Cache hits: ${cache_hits}"
        echo "  Cache misses: ${cache_misses}"
        echo "  Cache hit ratio: $(echo "${hit_ratio} * 100" | bc -l | cut -d. -f1)%"
    else
        echo "✗ ${mode} benchmark failed"
        return 1
    fi
    echo
}

# Function to compare results
compare_results() {
    local cached_file="${OUTPUT_DIR}/with-cache_${TIMESTAMP}.json"
    local uncached_file="${OUTPUT_DIR}/without-cache_${TIMESTAMP}.json"
    local comparison_file="${OUTPUT_DIR}/comparison_${TIMESTAMP}.json"
    
    if [[ -f "${cached_file}" && -f "${uncached_file}" ]]; then
        echo "Generating performance comparison..."
        
        # Extract metrics using jq
        local cached_avg=$(jq -r '.average_duration_ns' "${cached_file}")
        local uncached_avg=$(jq -r '.average_duration_ns' "${uncached_file}")
        local cached_total=$(jq -r '.total_duration_ns' "${cached_file}")
        local uncached_total=$(jq -r '.total_duration_ns' "${uncached_file}")
        local cache_hits=$(jq -r '.cache_hits' "${cached_file}")
        local cache_misses=$(jq -r '.cache_misses' "${cached_file}")
        local hit_ratio=$(jq -r '.cache_hit_ratio' "${cached_file}")
        
        # Calculate improvement ratios (handle division by zero)
        local avg_improvement="N/A"
        local total_improvement="N/A"
        if [[ "${cached_avg}" != "0" && "${cached_avg}" != "null" ]]; then
            avg_improvement=$(echo "scale=2; ${uncached_avg} / ${cached_avg}" | bc -l)
        fi
        if [[ "${cached_total}" != "0" && "${cached_total}" != "null" ]]; then
            total_improvement=$(echo "scale=2; ${uncached_total} / ${cached_total}" | bc -l)
        fi
        
        # Create comparison JSON
        jq -n \
            --arg timestamp "$(date -Iseconds)" \
            --argjson cached_avg "${cached_avg}" \
            --argjson uncached_avg "${uncached_avg}" \
            --argjson cached_total "${cached_total}" \
            --argjson uncached_total "${uncached_total}" \
            --argjson cache_hits "${cache_hits}" \
            --argjson cache_misses "${cache_misses}" \
            --argjson hit_ratio "${hit_ratio}" \
            --arg avg_improvement "${avg_improvement}" \
            --arg total_improvement "${total_improvement}" \
            '{
                timestamp: $timestamp,
                cached_performance: {
                    average_duration_ns: $cached_avg,
                    total_duration_ns: $cached_total
                },
                uncached_performance: {
                    average_duration_ns: $uncached_avg,
                    total_duration_ns: $uncached_total
                },
                cache_statistics: {
                    hits: $cache_hits,
                    misses: $cache_misses,
                    hit_ratio: $hit_ratio
                },
                improvements: {
                    average_speedup: $avg_improvement,
                    total_speedup: $total_improvement
                }
            }' > "${comparison_file}"
        
        echo "✓ Comparison saved to: ${comparison_file}"
        echo
        echo "=== Performance Comparison Summary ==="
        echo "Cache hit ratio: $(echo "${hit_ratio} * 100" | bc -l | cut -d. -f1)%"
        echo "Average operation speedup: ${avg_improvement}x"
        echo "Total execution speedup: ${total_improvement}x"
        echo "Cached average: ${cached_avg} ns"
        echo "Uncached average: ${uncached_avg} ns"
        echo
    else
        echo "⚠ Cannot generate comparison - missing result files"
    fi
}

# Check dependencies
echo "Checking dependencies..."
command -v jq >/dev/null 2>&1 || { echo "✗ jq is required but not installed. Please install jq."; exit 1; }
command -v bc >/dev/null 2>&1 || { echo "✗ bc is required but not installed. Please install bc."; exit 1; }
echo "✓ All dependencies satisfied"
echo

# Run benchmarks
echo "Starting performance benchmarks..."
echo "============================================"

# Run without cache first (to ensure no cached data affects the test)
run_benchmark "without-cache" "-with-cache=false"

# Run with cache
run_benchmark "with-cache" "-with-cache=true"

# Compare results
compare_results

echo "============================================"
echo "Performance benchmark completed!"
echo "Results available in: ${OUTPUT_DIR}"
echo

# Exit with success if we got this far
exit 0
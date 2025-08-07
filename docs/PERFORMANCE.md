# Performance Benchmarking

This directory contains tools for benchmarking the performance improvements achieved by the caching system in aztui.

## Files

- `cmd/perfbench/main.go` - Performance benchmark tool that can run with or without caching
- `scripts/perf-benchmark.sh` - Script that runs benchmarks with both configurations and compares results

## Usage

### Manual Benchmarking

Run the benchmark tool directly:

```bash
# With caching enabled (default)
cd src && go run cmd/perfbench/main.go -iterations=5 -verbose

# Without caching
cd src && go run cmd/perfbench/main.go -with-cache=false -iterations=5 -verbose

# JSON output for automated analysis
cd src && go run cmd/perfbench/main.go -json > results.json
```

### Automated Benchmarking

Run the complete benchmark script:

```bash
# Use default settings (5 iterations)
./scripts/perf-benchmark.sh

# Custom iteration count
PERF_ITERATIONS=10 ./scripts/perf-benchmark.sh
```

This will:
1. Build the benchmark tool
2. Run benchmarks without caching
3. Run benchmarks with caching
4. Generate a comparison report
5. Save all results to `perf-results/`

### CI Integration

The benchmarks are automatically run in the GitHub Actions CI pipeline as part of the `performance-benchmark` job. Results are uploaded as artifacts and a summary is displayed in the CI logs.

## Output

The benchmark generates:
- Individual result files for each run mode (`with-cache_*.json`, `without-cache_*.json`)
- A comparison file (`comparison_*.json`) showing speedup metrics
- Console output with summary statistics

### Example Results

```
=== Performance Comparison Summary ===
Cache hit ratio: 62%
Average operation speedup: 1.14x
Total execution speedup: 1.14x
Cached average: 105334265 ns
Uncached average: 120503898 ns
```

## Configuration

Benchmarks use the same configuration as the main application (`conf/default.yaml`). The cache TTL and other settings will affect the results.

Environment variables:
- `PERF_ITERATIONS` - Number of iterations per benchmark (default: 5)
- `AZTUI_CONFIG_PATH` - Path to configuration file

## Dependencies

The benchmark script requires:
- `jq` - JSON processing
- `bc` - Mathematical calculations
- Go 1.20+ - Building and running the benchmark tool
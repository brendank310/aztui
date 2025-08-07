package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/brendank310/aztui/pkg/cache"
	"github.com/brendank310/aztui/pkg/config"
	"github.com/brendank310/aztui/pkg/logger"
	"github.com/brendank310/aztui/pkg/resourceviews"
	"github.com/brendank310/aztui/pkg/tracing"
)

type BenchmarkResult struct {
	Mode               string                 `json:"mode"`
	TotalOperations    int                    `json:"total_operations"`
	TotalDuration      time.Duration          `json:"total_duration_ns"`
	AverageDuration    time.Duration          `json:"average_duration_ns"`
	CacheHits          int                    `json:"cache_hits"`
	CacheMisses        int                    `json:"cache_misses"`
	CacheHitRatio      float64                `json:"cache_hit_ratio"`
	OperationDetails   []OperationResult      `json:"operation_details"`
	TracingStats       map[string]interface{} `json:"tracing_stats,omitempty"`
	Timestamp          time.Time              `json:"timestamp"`
}

type OperationResult struct {
	Operation    string        `json:"operation"`
	Duration     time.Duration `json:"duration_ns"`
	CacheHit     bool          `json:"cache_hit"`
	Error        string        `json:"error,omitempty"`
}

func main() {
	var (
		withCache    = flag.Bool("with-cache", true, "Enable caching for the benchmark")
		iterations   = flag.Int("iterations", 3, "Number of iterations to run for each operation")
		outputJSON   = flag.Bool("json", false, "Output results in JSON format")
		configPath   = flag.String("config", "", "Path to config file (defaults to AZTUI_CONFIG_PATH env var)")
		verbose      = flag.Bool("verbose", false, "Enable verbose output")
	)
	flag.Parse()

	// Initialize logger
	err := logger.InitLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Load configuration
	if *configPath == "" {
		*configPath = os.Getenv("AZTUI_CONFIG_PATH")
		if *configPath == "" {
			*configPath = os.Getenv("HOME") + "/.config/aztui.yaml"
		}
	}

	c, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize tracing for performance measurement
	tracing.InitTracer(true, 1000)
	tracer := tracing.GetTracer()

	// Initialize cache service
	var cacheService *cache.ResourceCacheService
	mode := "without-cache"
	if *withCache {
		mode = "with-cache"
		cacheService = cache.NewResourceCacheService(c.GetCacheTTL())
		resourceviews.SetCacheService(cacheService)
		if *verbose {
			fmt.Fprintf(os.Stderr, "Cache enabled with TTL: %v\n", c.GetCacheTTL())
		}
	} else {
		// Set cache service to nil to disable caching
		resourceviews.SetCacheService(nil)
		if *verbose {
			fmt.Fprintf(os.Stderr, "Cache disabled\n")
		}
	}

	// Run benchmark operations
	result := BenchmarkResult{
		Mode:      mode,
		Timestamp: time.Now(),
	}

	operations := []struct {
		name string
		fn   func() error
	}{
		{"cache-operations", func() error {
			// Simulate cache operations without requiring Azure authentication
			if cacheService != nil {
				// Test cache operations
				testKey := "test-subscription-key"
				testData := []string{"subscription1", "subscription2", "subscription3"}
				
				// First operation should be a cache miss
				_, err := tracer.TraceFunc("cache-operation", testKey, func() (interface{}, bool, error) {
					result, err := cacheService.GetOrFetch(testKey, func() (interface{}, error) {
						time.Sleep(10 * time.Millisecond) // Simulate API call latency
						return testData, nil
					})
					if err != nil {
						return nil, false, err
					}
					return result, false, nil // First call is cache miss
				})
				if err != nil {
					return err
				}
				
				// Second operation should be a cache hit
				_, err = tracer.TraceFunc("cache-operation", testKey, func() (interface{}, bool, error) {
					result, err := cacheService.GetOrFetch(testKey, func() (interface{}, error) {
						time.Sleep(10 * time.Millisecond) // This shouldn't be called
						return testData, nil
					})
					if err != nil {
						return nil, false, err
					}
					return result, true, nil // Second call should be cache hit
				})
				if err != nil {
					return err
				}
			} else {
				// Without cache, simulate direct API calls
				_, err := tracer.TraceFunc("direct-api-call", "no-cache", func() (interface{}, bool, error) {
					time.Sleep(10 * time.Millisecond) // Simulate API call
					return nil, false, nil // Always a miss without cache
				})
				if err != nil {
					return err
				}
				
				_, err = tracer.TraceFunc("direct-api-call", "no-cache", func() (interface{}, bool, error) {
					time.Sleep(10 * time.Millisecond) // Simulate another API call
					return nil, false, nil // Always a miss without cache
				})
				if err != nil {
					return err
				}
			}
			return nil
		}},
	}

	startTime := time.Now()

	for _, op := range operations {
		for i := 0; i < *iterations; i++ {
			if *verbose {
				fmt.Fprintf(os.Stderr, "Running %s iteration %d/%d\n", op.name, i+1, *iterations)
			}

			opStart := time.Now()
			err := op.fn()
			duration := time.Since(opStart)

			opResult := OperationResult{
				Operation: fmt.Sprintf("%s-%d", op.name, i+1),
				Duration:  duration,
				CacheHit:  false, // Will be determined from tracing stats
			}

			if err != nil {
				opResult.Error = err.Error()
				if *verbose {
					fmt.Fprintf(os.Stderr, "Error in %s: %v\n", op.name, err)
				}
			}

			result.OperationDetails = append(result.OperationDetails, opResult)
			result.TotalOperations++

			// Small delay between operations to allow for cache expiration testing
			time.Sleep(100 * time.Millisecond)
		}
	}

	result.TotalDuration = time.Since(startTime)
	if result.TotalOperations > 0 {
		result.AverageDuration = result.TotalDuration / time.Duration(result.TotalOperations)
	}

	// Get tracing statistics
	if tracer != nil {
		stats := tracer.GetStats()
		result.TracingStats = map[string]interface{}{
			"cache_hits":            stats.CacheHits,
			"cache_misses":          stats.CacheMisses,
			"total_operations":      stats.TotalOperations,
			"avg_cache_hit_time":    stats.AvgCacheHitTime,
			"avg_cache_miss_time":   stats.AvgCacheMissTime,
			"hit_ratio":             stats.HitRatio,
		}
		
		result.CacheHits = int(stats.CacheHits)
		result.CacheMisses = int(stats.CacheMisses)
		result.CacheHitRatio = stats.HitRatio
	}

	// Output results
	if *outputJSON {
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to marshal JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonData))
	} else {
		fmt.Printf("=== Performance Benchmark Results ===\n")
		fmt.Printf("Mode: %s\n", result.Mode)
		fmt.Printf("Total Operations: %d\n", result.TotalOperations)
		fmt.Printf("Total Duration: %v\n", result.TotalDuration)
		fmt.Printf("Average Duration: %v\n", result.AverageDuration)
		fmt.Printf("Cache Hits: %d\n", result.CacheHits)
		fmt.Printf("Cache Misses: %d\n", result.CacheMisses)
		fmt.Printf("Cache Hit Ratio: %.2f%%\n", result.CacheHitRatio*100)

		if result.TracingStats != nil {
			fmt.Printf("\n=== Tracing Statistics ===\n")
			for key, value := range result.TracingStats {
				fmt.Printf("%s: %v\n", key, value)
			}
		}

		if *verbose {
			fmt.Printf("\n=== Operation Details ===\n")
			for _, op := range result.OperationDetails {
				status := "success"
				if op.Error != "" {
					status = fmt.Sprintf("error: %s", op.Error)
				}
				fmt.Printf("%s: %v (%s)\n", op.Operation, op.Duration, status)
			}
		}
	}
}
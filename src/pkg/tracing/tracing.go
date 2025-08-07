package tracing

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/pprof"
	"sync"
	"time"
)

// PerformanceStats holds statistics about cache performance
type PerformanceStats struct {
	CacheHits          int64         `json:"cache_hits"`
	CacheMisses        int64         `json:"cache_misses"`
	TotalOperations    int64         `json:"total_operations"`
	HitRatio           float64       `json:"hit_ratio"`
	AvgCacheHitTime    time.Duration `json:"avg_cache_hit_time_ns"`
	AvgCacheMissTime   time.Duration `json:"avg_cache_miss_time_ns"`
	TotalCacheHitTime  time.Duration `json:"total_cache_hit_time_ns"`
	TotalCacheMissTime time.Duration `json:"total_cache_miss_time_ns"`
}

// OperationTrace represents a single operation trace
type OperationTrace struct {
	Operation  string        `json:"operation"`
	CacheKey   string        `json:"cache_key"`
	Duration   time.Duration `json:"duration_ns"`
	WasCacheHit bool         `json:"was_cache_hit"`
	Timestamp  time.Time     `json:"timestamp"`
}

// PerformanceTracer provides tracing and statistics for performance analysis
type PerformanceTracer struct {
	enabled    bool
	mutex      sync.RWMutex
	stats      PerformanceStats
	traces     []OperationTrace
	maxTraces  int
	cpuProfile *os.File
}

var (
	globalTracer *PerformanceTracer
	once         sync.Once
)

// InitTracer initializes the global performance tracer
func InitTracer(enabled bool, maxTraces int) {
	once.Do(func() {
		globalTracer = &PerformanceTracer{
			enabled:   enabled,
			maxTraces: maxTraces,
			traces:    make([]OperationTrace, 0, maxTraces),
		}
	})
}

// GetTracer returns the global tracer instance
func GetTracer() *PerformanceTracer {
	if globalTracer == nil {
		InitTracer(false, 1000) // Default disabled tracer
	}
	return globalTracer
}

// IsEnabled returns whether tracing is enabled
func (t *PerformanceTracer) IsEnabled() bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.enabled
}

// StartCPUProfile starts CPU profiling for flamegraph generation
func (t *PerformanceTracer) StartCPUProfile(filename string) error {
	if !t.enabled {
		return nil
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create CPU profile: %v", err)
	}

	if err := pprof.StartCPUProfile(file); err != nil {
		file.Close()
		return fmt.Errorf("could not start CPU profile: %v", err)
	}

	t.mutex.Lock()
	t.cpuProfile = file
	t.mutex.Unlock()

	return nil
}

// StopCPUProfile stops CPU profiling
func (t *PerformanceTracer) StopCPUProfile() {
	if !t.enabled {
		return
	}

	pprof.StopCPUProfile()
	
	t.mutex.Lock()
	if t.cpuProfile != nil {
		t.cpuProfile.Close()
		t.cpuProfile = nil
	}
	t.mutex.Unlock()
}

// TraceOperation records the timing and result of a cache operation
func (t *PerformanceTracer) TraceOperation(operation, cacheKey string, duration time.Duration, wasCacheHit bool) {
	if !t.enabled {
		return
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	// Update statistics
	t.stats.TotalOperations++
	if wasCacheHit {
		t.stats.CacheHits++
		t.stats.TotalCacheHitTime += duration
		t.stats.AvgCacheHitTime = t.stats.TotalCacheHitTime / time.Duration(t.stats.CacheHits)
	} else {
		t.stats.CacheMisses++
		t.stats.TotalCacheMissTime += duration
		t.stats.AvgCacheMissTime = t.stats.TotalCacheMissTime / time.Duration(t.stats.CacheMisses)
	}
	
	if t.stats.TotalOperations > 0 {
		t.stats.HitRatio = float64(t.stats.CacheHits) / float64(t.stats.TotalOperations)
	}

	// Add trace if we have room
	if len(t.traces) < t.maxTraces {
		trace := OperationTrace{
			Operation:   operation,
			CacheKey:    cacheKey,
			Duration:    duration,
			WasCacheHit: wasCacheHit,
			Timestamp:   time.Now(),
		}
		t.traces = append(t.traces, trace)
	}
}

// GetStats returns a copy of current performance statistics
func (t *PerformanceTracer) GetStats() PerformanceStats {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.stats
}

// GetTraces returns a copy of current traces
func (t *PerformanceTracer) GetTraces() []OperationTrace {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	
	traces := make([]OperationTrace, len(t.traces))
	copy(traces, t.traces)
	return traces
}

// OutputStats outputs performance statistics to stderr as JSON
func (t *PerformanceTracer) OutputStats() {
	if !t.enabled {
		return
	}

	stats := t.GetStats()
	traces := t.GetTraces()

	output := struct {
		Stats  PerformanceStats `json:"performance_stats"`
		Traces []OperationTrace `json:"traces,omitempty"`
	}{
		Stats:  stats,
		Traces: traces,
	}

	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling performance data: %v\n", err)
		return
	}

	fmt.Fprintf(os.Stderr, "%s\n", jsonData)
}

// Reset clears all statistics and traces
func (t *PerformanceTracer) Reset() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	t.stats = PerformanceStats{}
	t.traces = t.traces[:0]
}

// TraceFunc is a helper that measures the execution time of a function
func (t *PerformanceTracer) TraceFunc(operation, cacheKey string, fn func() (interface{}, bool, error)) (interface{}, error) {
	if !t.enabled {
		result, _, err := fn()
		return result, err
	}

	start := time.Now()
	result, wasCacheHit, err := fn()
	duration := time.Since(start)

	t.TraceOperation(operation, cacheKey, duration, wasCacheHit)
	return result, err
}
package tracing

import (
	"sync"
	"testing"
	"time"
)

func TestPerformanceTracer_Disabled(t *testing.T) {
	// Test that disabled tracer has no overhead
	tracer := &PerformanceTracer{enabled: false}

	start := time.Now()
	tracer.TraceOperation("test", "key", time.Millisecond, true)
	elapsed := time.Since(start)

	// Should be very fast when disabled
	if elapsed > time.Microsecond*10 {
		t.Errorf("Disabled tracer took too long: %v", elapsed)
	}

	stats := tracer.GetStats()
	if stats.TotalOperations != 0 {
		t.Errorf("Expected 0 operations, got %d", stats.TotalOperations)
	}
}

func TestPerformanceTracer_CacheStats(t *testing.T) {
	tracer := &PerformanceTracer{
		enabled:   true,
		maxTraces: 100,
		traces:    make([]OperationTrace, 0, 100),
	}

	// Record some cache hits
	tracer.TraceOperation("get_subscriptions", "subscriptions", time.Millisecond*10, true)
	tracer.TraceOperation("get_subscriptions", "subscriptions", time.Millisecond*5, true)
	
	// Record some cache misses
	tracer.TraceOperation("get_resource_groups", "rg:sub1", time.Millisecond*100, false)
	tracer.TraceOperation("get_resources", "res:sub1:rg1", time.Millisecond*200, false)

	stats := tracer.GetStats()

	// Verify counts
	if stats.CacheHits != 2 {
		t.Errorf("Expected 2 cache hits, got %d", stats.CacheHits)
	}
	if stats.CacheMisses != 2 {
		t.Errorf("Expected 2 cache misses, got %d", stats.CacheMisses)
	}
	if stats.TotalOperations != 4 {
		t.Errorf("Expected 4 total operations, got %d", stats.TotalOperations)
	}

	// Verify hit ratio
	expectedHitRatio := 0.5
	if stats.HitRatio != expectedHitRatio {
		t.Errorf("Expected hit ratio %f, got %f", expectedHitRatio, stats.HitRatio)
	}

	// Verify average times
	expectedAvgHitTime := (time.Millisecond*10 + time.Millisecond*5) / 2 // (10+5)/2
	if stats.AvgCacheHitTime != expectedAvgHitTime {
		t.Errorf("Expected avg hit time %v, got %v", expectedAvgHitTime, stats.AvgCacheHitTime)
	}

	expectedAvgMissTime := (time.Millisecond*100 + time.Millisecond*200) / 2 // (100+200)/2
	if stats.AvgCacheMissTime != expectedAvgMissTime {
		t.Errorf("Expected avg miss time %v, got %v", expectedAvgMissTime, stats.AvgCacheMissTime)
	}
}

func TestPerformanceTracer_Traces(t *testing.T) {
	tracer := &PerformanceTracer{
		enabled:   true,
		maxTraces: 2, // Small limit to test overflow
		traces:    make([]OperationTrace, 0, 2),
	}

	// Add traces
	tracer.TraceOperation("op1", "key1", time.Millisecond, true)
	tracer.TraceOperation("op2", "key2", time.Millisecond*2, false)
	tracer.TraceOperation("op3", "key3", time.Millisecond*3, true) // Should not be stored due to limit

	traces := tracer.GetTraces()

	// Should only have 2 traces due to limit
	if len(traces) != 2 {
		t.Errorf("Expected 2 traces, got %d", len(traces))
	}

	// Verify first trace
	if traces[0].Operation != "op1" {
		t.Errorf("Expected operation 'op1', got '%s'", traces[0].Operation)
	}
	if traces[0].CacheKey != "key1" {
		t.Errorf("Expected cache key 'key1', got '%s'", traces[0].CacheKey)
	}
	if !traces[0].WasCacheHit {
		t.Errorf("Expected cache hit for first trace")
	}

	// Verify second trace
	if traces[1].Operation != "op2" {
		t.Errorf("Expected operation 'op2', got '%s'", traces[1].Operation)
	}
	if traces[1].WasCacheHit {
		t.Errorf("Expected cache miss for second trace")
	}
}

func TestPerformanceTracer_Reset(t *testing.T) {
	tracer := &PerformanceTracer{
		enabled:   true,
		maxTraces: 100,
		traces:    make([]OperationTrace, 0, 100),
	}

	// Add some data
	tracer.TraceOperation("test", "key", time.Millisecond, true)
	
	// Verify data exists
	stats := tracer.GetStats()
	if stats.TotalOperations == 0 {
		t.Error("Expected operations before reset")
	}

	traces := tracer.GetTraces()
	if len(traces) == 0 {
		t.Error("Expected traces before reset")
	}

	// Reset
	tracer.Reset()

	// Verify data is cleared
	stats = tracer.GetStats()
	if stats.TotalOperations != 0 {
		t.Errorf("Expected 0 operations after reset, got %d", stats.TotalOperations)
	}

	traces = tracer.GetTraces()
	if len(traces) != 0 {
		t.Errorf("Expected 0 traces after reset, got %d", len(traces))
	}
}

func TestPerformanceTracer_TraceFunc(t *testing.T) {
	tracer := &PerformanceTracer{
		enabled:   true,
		maxTraces: 100,
		traces:    make([]OperationTrace, 0, 100),
	}

	// Test cache hit scenario
	result, err := tracer.TraceFunc("test_op", "test_key", func() (interface{}, bool, error) {
		time.Sleep(time.Millisecond) // Simulate some work
		return "cached_result", true, nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != "cached_result" {
		t.Errorf("Expected 'cached_result', got %v", result)
	}

	stats := tracer.GetStats()
	if stats.CacheHits != 1 {
		t.Errorf("Expected 1 cache hit, got %d", stats.CacheHits)
	}

	traces := tracer.GetTraces()
	if len(traces) != 1 {
		t.Errorf("Expected 1 trace, got %d", len(traces))
	}
	if traces[0].Operation != "test_op" {
		t.Errorf("Expected operation 'test_op', got '%s'", traces[0].Operation)
	}
	if !traces[0].WasCacheHit {
		t.Error("Expected cache hit in trace")
	}
}

func TestGlobalTracer(t *testing.T) {
	// Reset global tracer
	globalTracer = nil
	once = sync.Once{}

	// Initialize global tracer
	InitTracer(true, 100)

	tracer := GetTracer()
	if tracer == nil {
		t.Error("Expected non-nil global tracer")
	}
	if !tracer.IsEnabled() {
		t.Error("Expected tracer to be enabled")
	}

	// Test that subsequent calls return the same instance
	tracer2 := GetTracer()
	if tracer != tracer2 {
		t.Error("Expected same tracer instance")
	}
}
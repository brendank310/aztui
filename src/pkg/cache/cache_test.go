package cache

import (
	"testing"
	"time"
)

func TestMemoryCache(t *testing.T) {
	cache := NewMemoryCache()
	
	// Test Set and Get
	cache.Set("test_key", "test_value", 1*time.Second)
	
	value, found := cache.Get("test_key")
	if !found {
		t.Error("Expected to find cached value")
	}
	
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got %v", value)
	}
	
	// Test TTL expiration
	cache.Set("expire_key", "expire_value", 100*time.Millisecond)
	
	// Should find it immediately
	_, found = cache.Get("expire_key")
	if !found {
		t.Error("Expected to find cached value before expiration")
	}
	
	// Wait for expiration
	time.Sleep(150 * time.Millisecond)
	
	// Should not find it after expiration
	_, found = cache.Get("expire_key")
	if found {
		t.Error("Expected cached value to be expired")
	}
	
	// Test Delete
	cache.Set("delete_key", "delete_value", 1*time.Minute)
	cache.Delete("delete_key")
	
	_, found = cache.Get("delete_key")
	if found {
		t.Error("Expected cached value to be deleted")
	}
	
	// Test Clear
	cache.Set("clear_key1", "value1", 1*time.Minute)
	cache.Set("clear_key2", "value2", 1*time.Minute)
	
	cache.Clear()
	
	_, found = cache.Get("clear_key1")
	if found {
		t.Error("Expected all cached values to be cleared")
	}
	
	_, found = cache.Get("clear_key2")
	if found {
		t.Error("Expected all cached values to be cleared")
	}
}

func TestResourceCacheService(t *testing.T) {
	service := NewResourceCacheService(1 * time.Second)
	
	fetchCount := 0
	fetchFunc := func() (interface{}, error) {
		fetchCount++
		return "fetched_data", nil
	}
	
	// First call should fetch data
	data, err := service.GetOrFetch("test_key", fetchFunc)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	if data != "fetched_data" {
		t.Errorf("Expected 'fetched_data', got %v", data)
	}
	
	if fetchCount != 1 {
		t.Errorf("Expected fetch count to be 1, got %d", fetchCount)
	}
	
	// Second call should use cache
	data, err = service.GetOrFetch("test_key", fetchFunc)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	if data != "fetched_data" {
		t.Errorf("Expected 'fetched_data', got %v", data)
	}
	
	if fetchCount != 1 {
		t.Errorf("Expected fetch count to remain 1, got %d", fetchCount)
	}
	
	// Wait for cache expiration
	time.Sleep(1100 * time.Millisecond)
	
	// Third call should fetch again
	data, err = service.GetOrFetch("test_key", fetchFunc)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	if data != "fetched_data" {
		t.Errorf("Expected 'fetched_data', got %v", data)
	}
	
	if fetchCount != 2 {
		t.Errorf("Expected fetch count to be 2, got %d", fetchCount)
	}
}

func TestCacheKeyGeneration(t *testing.T) {
	// Test subscription key
	subKey := GenerateSubscriptionKey()
	if subKey != "subscriptions" {
		t.Errorf("Expected 'subscriptions', got %s", subKey)
	}
	
	// Test resource group key
	rgKey := GenerateResourceGroupKey("sub123")
	expected := "resourcegroups:sub123"
	if rgKey != expected {
		t.Errorf("Expected '%s', got %s", expected, rgKey)
	}
	
	// Test resource key
	resourceKey := GenerateResourceKey("sub123", "rg123", "Microsoft.Compute/virtualMachines")
	expected = "resources:sub123:rg123:Microsoft.Compute/virtualMachines"
	if resourceKey != expected {
		t.Errorf("Expected '%s', got %s", expected, resourceKey)
	}
	
	// Test VM key
	vmKey := GenerateVMKey("sub123", "rg123")
	expected = "vms:sub123:rg123"
	if vmKey != expected {
		t.Errorf("Expected '%s', got %s", expected, vmKey)
	}
	
	// Test AKS key
	aksKey := GenerateAKSKey("sub123", "rg123")
	expected = "aks:sub123:rg123"
	if aksKey != expected {
		t.Errorf("Expected '%s', got %s", expected, aksKey)
	}
}
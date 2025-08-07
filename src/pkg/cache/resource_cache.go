package cache

import (
	"fmt"
	"time"
)

// ResourceCacheService provides caching for Azure resources
type ResourceCacheService struct {
	cache    Cache
	defaultTTL time.Duration
}

// NewResourceCacheService creates a new resource cache service
func NewResourceCacheService(defaultTTL time.Duration) *ResourceCacheService {
	return &ResourceCacheService{
		cache:      NewMemoryCache(),
		defaultTTL: defaultTTL,
	}
}

// GetOrFetch retrieves data from cache or fetches it using the provided function
func (s *ResourceCacheService) GetOrFetch(key string, fetchFunc func() (interface{}, error)) (interface{}, error) {
	return s.GetOrFetchWithTTL(key, s.defaultTTL, fetchFunc)
}

// GetOrFetchWithTTL retrieves data from cache or fetches it with custom TTL
func (s *ResourceCacheService) GetOrFetchWithTTL(key string, ttl time.Duration, fetchFunc func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	if cached, found := s.cache.Get(key); found {
		return cached, nil
	}
	
	// Not in cache, fetch the data
	data, err := fetchFunc()
	if err != nil {
		return nil, err
	}
	
	// Store in cache
	s.cache.Set(key, data, ttl)
	
	return data, nil
}

// InvalidateKey removes a specific key from the cache
func (s *ResourceCacheService) InvalidateKey(key string) {
	s.cache.Delete(key)
}

// InvalidatePattern removes all keys matching a pattern (simple prefix matching)
func (s *ResourceCacheService) InvalidatePattern(prefix string) {
	// For this simple implementation, we'll clear all cache
	// A more sophisticated implementation could iterate through keys
	s.cache.Clear()
}

// Clear removes all entries from the cache
func (s *ResourceCacheService) Clear() {
	s.cache.Clear()
}

// GenerateSubscriptionKey creates a cache key for subscription list
func GenerateSubscriptionKey() string {
	return "subscriptions"
}

// GenerateResourceGroupKey creates a cache key for resource groups
func GenerateResourceGroupKey(subscriptionID string) string {
	return fmt.Sprintf("resourcegroups:%s", subscriptionID)
}

// GenerateResourceKey creates a cache key for resources
func GenerateResourceKey(subscriptionID, resourceGroup, resourceType string) string {
	return fmt.Sprintf("resources:%s:%s:%s", subscriptionID, resourceGroup, resourceType)
}

// GenerateVMKey creates a cache key for virtual machines
func GenerateVMKey(subscriptionID, resourceGroup string) string {
	return fmt.Sprintf("vms:%s:%s", subscriptionID, resourceGroup)
}

// GenerateAKSKey creates a cache key for AKS clusters
func GenerateAKSKey(subscriptionID, resourceGroup string) string {
	return fmt.Sprintf("aks:%s:%s", subscriptionID, resourceGroup)
}
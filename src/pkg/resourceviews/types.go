package resourceviews

import "github.com/brendank310/aztui/pkg/cache"

var AvailableResourceTypes = []string{
	"Virtual Machines",
	"AKS Clusters",
}

// Global cache service instance
var globalCacheService *cache.ResourceCacheService

// SetCacheService sets the global cache service instance
func SetCacheService(cacheService *cache.ResourceCacheService) {
	globalCacheService = cacheService
}

// GetCacheService returns the global cache service instance
func GetCacheService() *cache.ResourceCacheService {
	return globalCacheService
}

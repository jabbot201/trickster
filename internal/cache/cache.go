/**
* Copyright 2018 Comcast Cable Communications Management, LLC
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
* http://www.apache.org/licenses/LICENSE-2.0
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

// Package cache defines the Trickster cache interfaces and provides
// general cache functionality
package cache

import (
	"errors"
	"fmt"
	"time"

	"github.com/Comcast/trickster/internal/cache/status"
	"github.com/Comcast/trickster/internal/config"
	"github.com/Comcast/trickster/internal/util/metrics"
)

// ErrKNF represents the error "key not found in cache"
var ErrKNF = errors.New("key not found in cache")

// Cache is the interface for the supported caching fabrics
// When making new cache types, Retrieve() must return an error on cache miss
type Cache interface {
	Connect() error
	Store(cacheKey string, data []byte, ttl time.Duration) error
	Retrieve(cacheKey string, allowExpired bool) ([]byte, status.LookupStatus, error)
	SetTTL(cacheKey string, ttl time.Duration)
	Remove(cacheKey string)
	BulkRemove(cacheKeys []string, noLock bool)
	Close() error
	Configuration() *config.CachingConfig
}

// MemoryCache is the interface for an in-memory cache
// This offers an additional method for storing references to bypass serialization
type MemoryCache interface {
	Connect() error
	Store(cacheKey string, data []byte, ttl time.Duration) error
	Retrieve(cacheKey string, allowExpired bool) ([]byte, status.LookupStatus, error)
	SetTTL(cacheKey string, ttl time.Duration)
	Remove(cacheKey string)
	BulkRemove(cacheKeys []string, noLock bool)
	Close() error
	Configuration() *config.CachingConfig
	StoreReference(cacheKey string, data ReferenceObject, ttl time.Duration) error
	RetrieveReference(cacheKey string, allowExpired bool) (interface{}, status.LookupStatus, error)
}

// ReferenceObject defines an interface for a cache object possessing the ability to report
// the approximate comprehensive byte size of its members, to assist with cache size management
type ReferenceObject interface {
	Size() int
}

// ObserveCacheMiss returns a standard Cache Miss response
func ObserveCacheMiss(cacheKey, cacheName, cacheType string) ([]byte, error) {
	ObserveCacheOperation(cacheName, cacheType, "get", "miss", 0)
	return nil, fmt.Errorf("value for key [%s] not in cache", cacheKey)
}

// ObserveCacheDel records a cache deletion event
func ObserveCacheDel(cache, cacheType string, count float64) {
	ObserveCacheOperation(cache, cacheType, "del", "none", count)
}

// CacheError returns an empty cache object and the formatted error
func CacheError(cacheKey, cacheName, cacheType string, msg string) ([]byte, error) {
	ObserveCacheEvent(cacheName, cacheType, "error", msg)
	return nil, fmt.Errorf(msg, cacheKey)
}

// ObserveCacheOperation increments counters as cache operations occur
func ObserveCacheOperation(cache, cacheType, operation, status string, bytes float64) {
	metrics.CacheObjectOperations.WithLabelValues(cache, cacheType, operation, status).Inc()
	if bytes > 0 {
		metrics.CacheByteOperations.WithLabelValues(cache, cacheType, operation, status).Add(float64(bytes))
	}
}

// ObserveCacheEvent increments counters as cache events occur
func ObserveCacheEvent(cache, cacheType, event, reason string) {
	metrics.CacheEvents.WithLabelValues(cache, cacheType, event, reason).Inc()
}

// ObserveCacheSizeChange adjust counters and gauges as the cache size changes due to object operations
func ObserveCacheSizeChange(cache, cacheType string, byteCount, objectCount int64) {
	metrics.CacheObjects.WithLabelValues(cache, cacheType).Set(float64(objectCount))
	metrics.CacheBytes.WithLabelValues(cache, cacheType).Set(float64(byteCount))
}

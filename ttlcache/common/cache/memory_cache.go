package cache

import (
	"context"
	"thanhldt060802/model"
	"time"

	"github.com/jellydator/ttlcache/v3"
	log "github.com/sirupsen/logrus"
)

var MemoryCacheInstance1 IMemoryCache[string, string]
var MemoryCacheInstance2 IMemoryCache[string, *model.DataStruct]

type MemoryCache[K comparable, V any] struct {
	cache *ttlcache.Cache[K, V]
}

type IMemoryCache[K comparable, V any] interface {
	Set(key K, value V)
	SetTTL(key K, value V, ttl time.Duration)
	Get(key K) (V, bool)
	Del(key K)
}

func NewMemoryCache[K comparable, V any](defaultTTL time.Duration) IMemoryCache[K, V] {
	cache := ttlcache.New(
		ttlcache.WithTTL[K, V](defaultTTL),
	)
	cache.OnInsertion(func(ctx context.Context, item *ttlcache.Item[K, V]) {
		log.Infof("MemoryCache: inserted [%v, %v], expires at %v", item.Key(), item.Value(), item.ExpiresAt())
	})
	cache.OnEviction(func(ctx context.Context, reason ttlcache.EvictionReason, item *ttlcache.Item[K, V]) {
		switch reason {
		case ttlcache.EvictionReasonDeleted:
			log.Infof("MemoryCache-deleted: Removed [%v, %v]", item.Key(), item.Value())
		case ttlcache.EvictionReasonCapacityReached:
			log.Infof("MemoryCache-capacity-reached: Removed [%v, %v]", item.Key(), item.Value())
		case ttlcache.EvictionReasonExpired:
			log.Infof("MemoryCache-expired: Removed [%v, %v]", item.Key(), item.Value())
		case ttlcache.EvictionReasonMaxCostExceeded:
			log.Infof("MemoryCache-max-cost-exceeded: Removed [%v, %v]", item.Key(), item.Value())
		}
	})
	go cache.Start()

	return &MemoryCache[K, V]{
		cache: cache,
	}
}

func (mCache *MemoryCache[K, V]) Set(key K, value V) {
	mCache.cache.Set(key, value, ttlcache.DefaultTTL)
}

func (mCache *MemoryCache[K, V]) SetTTL(key K, value V, ttl time.Duration) {
	mCache.cache.Set(key, value, ttl)
}

func (mCache *MemoryCache[K, V]) Get(key K) (V, bool) {
	item := mCache.cache.Get(key)
	if item == nil {
		var zero V
		return zero, false
	}
	return item.Value(), true
}

func (mCache *MemoryCache[K, V]) Del(key K) {
	mCache.cache.Delete(key)
}

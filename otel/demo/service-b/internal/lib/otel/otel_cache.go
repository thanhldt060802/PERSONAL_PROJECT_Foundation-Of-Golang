package otel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Error definitions for Cache.
var (
	// ErrCacheUnconfigured occurs when using Cache without including Cache option when initializing Otel Observer.
	ErrCacheUnconfigured = errors.New("cache is unconfigured")
)

// Cache provides storage for Trace Carriers across async boundaries.
type Cache interface {
	getTraceCarrierFromGroup(group string, key string) (TraceCarrier, error)
	setTraceCarrierFromGroup(group string, key string, traceCarrier TraceCarrier) error
	deleteTraceCarrierFromGroup(group string, key string) error
	deleteTraceCarrierGroup(group string) error
	clearTraceCarrier() error
}

// RedisConfig configures Redis connection for trace context storage.
type RedisConfig struct {
	Address         string // Redis connection address
	Database        int    // Redis database index
	Username        string // Redis username
	Password        string // Redis password
	PoolSize        int    // Redis connection pool size
	PoolTimeoutSec  int    // Redis connection pool timeout second
	IdleTimeoutSec  int    // Redis connection pool idle timeout second
	ReadTimeoutSec  int    // Redis connection pool read timeout second
	WriteTimeoutSec int    // Redis connection pool write timeout second
	Channel         string // Collection of keys managed
}

// redisCache implements Cache using Redis Cache
type redisCache struct {
	redisClient *redis.Client
	channel     string
}

// Default Redis settings
const (
	// defaultRedisPoolSize is pool size of connection
	defaultRedisPoolSize = 10
	// defaultRedisPoolTimeoutSec is pool connection timeout
	defaultRedisPoolTimeoutSec = 20
	// defaultRedisIdleTimeoutSec is pool connection idle timeout
	defaultRedisIdleTimeoutSec = 10
	// defaultRedisReadTimeoutSec is pool connection read timeout
	defaultRedisReadTimeoutSec = 20
	// defaultRedisWriteTimeoutSec is pool connection write timeout
	defaultRedisWriteTimeoutSec = 20
)

// Key prefix for Cache Trace Carriers
const traceCarrierRedisCacheKey = "OTEL:TRACECARRIER"

// initRedisCache initializes Redis connection and sets the global Cache
func initRedisCache(config *RedisConfig) *redisCache {
	// Create Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:            config.Address,
		Username:        config.Username,
		Password:        config.Password,
		DB:              config.Database,
		PoolSize:        config.PoolSize,
		PoolTimeout:     time.Duration(config.PoolTimeoutSec) * time.Second,
		ConnMaxIdleTime: time.Duration(config.IdleTimeoutSec) * time.Second,
		ReadTimeout:     time.Duration(config.ReadTimeoutSec) * time.Second,
		WriteTimeout:    time.Duration(config.WriteTimeoutSec) * time.Second,
	})

	// Ping to Redis for connection checking
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		stdLog.Fatalf("[error] Failed to ping to Redis: %v", err)
	}

	// Init redisCache
	cache := &redisCache{
		redisClient: redisClient,
		channel:     config.Channel,
	}

	// Return redisCache
	return cache
}

// getChannelKey constructs the full Redis key for all Trace Carrier in a channel
func (rCache *redisCache) getChannelKey() string {
	return traceCarrierRedisCacheKey + ":" + rCache.channel
}

// getGroupKey constructs the full Redis key for a Trace Carrier group
func (rCache *redisCache) getGroupKey(group string) string {
	return rCache.getChannelKey() + ":" + group
}

// getTraceCarrierFromGroup retrieves a Trace Carrier from Redis hash
func (rCache *redisCache) getTraceCarrierFromGroup(group string, key string) (TraceCarrier, error) {
	rawValue, err := rCache.redisClient.HGet(context.Background(), rCache.getGroupKey(group), key).Result()

	if err != nil {
		// Return empty carrier for non-existent keys
		if err == redis.Nil {
			return TraceCarrier{}, nil
		}
		return TraceCarrier{}, err
	}

	var carrier TraceCarrier
	if err := json.Unmarshal([]byte(rawValue), &carrier); err != nil {
		return TraceCarrier{}, err
	}

	return carrier, nil
}

// setTraceCarrierFromGroup stores a Trace Carrier in Redis.
func (rCache *redisCache) setTraceCarrierFromGroup(group string, key string, traceCarrier TraceCarrier) error {
	byteValue, err := json.Marshal(traceCarrier)
	if err != nil {
		return err
	}

	return rCache.redisClient.HSet(context.Background(), rCache.getGroupKey(group), key, string(byteValue)).Err()
}

// deleteTraceCarrierFromGroup removes a specific Trace Carrier from Redis.
func (rCache *redisCache) deleteTraceCarrierFromGroup(group string, key string) error {
	return rCache.redisClient.HDel(context.Background(), rCache.getGroupKey(group), key).Err()
}

// deleteTraceCarrierGroup removes an entire group of Trace Carriers.
func (rCache *redisCache) deleteTraceCarrierGroup(group string) error {
	return rCache.redisClient.Del(context.Background(), rCache.getGroupKey(group)).Err()
}

// clearTraceCarrier removes all groups of Trace Carriers.
func (rCache *redisCache) clearTraceCarrier() error {
	ctx := context.Background()

	var cursor uint64
	pattern := fmt.Sprintf("%s*", rCache.getChannelKey())
	keys := make([]string, 0)

	for {
		existingKeys, nextCursor, err := rCache.redisClient.Scan(context.Background(), cursor, pattern, 100).Result()
		if err != nil {
			stdLog.Printf("[error] Failed to scan partten '%s' with cursor '%d': %v", pattern, cursor, err)
		}
		keys = append(keys, existingKeys...)

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return rCache.redisClient.Del(ctx, keys...).Err()
}

// Public API functions with nil-safety checks.

// GetCacheTraceCarrierFromGroup retrieves a Trace Carrier from Cache.
// Returns ErrRedisUnconfigured if Redis was not initialized.
//
// Example:
//
//	carrier, err := observer.GetCacheTraceCarrierFromGroup("jobs", "job-123")
//	if err == nil && carrier != nil {
//	    ctx := carrier.ExtractContext()
//	}
func (o *Observer) GetCacheTraceCarrierFromGroup(group string, key string) (TraceCarrier, error) {
	if o.cache == nil {
		return TraceCarrier{}, ErrCacheUnconfigured
	}

	return o.cache.getTraceCarrierFromGroup(group, key)
}

// SetCacheTraceCarrierFromGroup stores a Trace Carrier in Cache.
// Returns ErrRedisUnconfigured if Redis was not initialized.
//
// Example:
//
//	carrier := otel.ExportTraceCarrier(ctx)
//	err := observer.SetCacheTraceCarrierFromGroup("jobs", "job-123", carrier)
func (o *Observer) SetCacheTraceCarrierFromGroup(group string, key string, traceCarrier TraceCarrier) error {
	if o.cache == nil {
		return ErrCacheUnconfigured
	}

	return o.cache.setTraceCarrierFromGroup(group, key, traceCarrier)
}

// DeleteCacheTraceCarrierFromGroup removes a Trace Carrier from Cache.
// Returns ErrRedisUnconfigured if Redis was not initialized.
//
// Example:
//
//	err := observer.DeleteCacheTraceCarrierFromGroup("jobs", "job-123")
func (o *Observer) DeleteCacheTraceCarrierFromGroup(group string, key string) error {
	if o.cache == nil {
		return ErrCacheUnconfigured
	}

	return o.cache.deleteTraceCarrierFromGroup(group, key)
}

// DeleteCacheTraceCarrierGroup removes all Trace Carriers in a group.
// Returns ErrRedisUnconfigured if Redis was not initialized.
//
// Example:
//
//	err := observer.DeleteCacheTraceCarrierGroup("jobs")
func (o *Observer) DeleteCacheTraceCarrierGroup(group string) error {
	if o.cache == nil {
		return ErrCacheUnconfigured
	}

	return o.cache.deleteTraceCarrierGroup(group)
}

// ClearCacheTraceCarrier removes all groups of Trace Carriers.
// Returns ErrRedisUnconfigured if Redis was not initialized.
//
// Example:
//
//	err := observer.ClearCacheTraceCarrier()
func (o *Observer) ClearCacheTraceCarrier() error {
	if o.cache == nil {
		return ErrCacheUnconfigured
	}

	return o.cache.clearTraceCarrier()
}

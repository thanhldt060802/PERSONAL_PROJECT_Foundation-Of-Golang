package otel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	// cache is global cache instance for internal storage
	cache Cache
)

var (
	ErrCacheUnconfigured = errors.New("cache is unconfigured")
)

// Cache provides storage for trace carriers across async boundaries
type Cache interface {
	getTraceCarrierFromGroup(group string, key string) (TraceCarrier, error)
	setTraceCarrierFromGroup(group string, key string, traceCarrier TraceCarrier) error
	deleteTraceCarrierFromGroup(group string, key string) error
	deleteTraceCarrierGroup(group string) error
	clearTraceCarrier() error
}

// RedisConfig configures Redis connection for trace context storage
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
}

// redisCache implements Cache using Redis hash maps
type redisCache struct {
	redisClient *redis.Client
}

// Default Redis settings
const (
	defaultRedisPoolSize        = 10
	defaultRedisPoolTimeoutSec  = 20
	defaultRedisIdleTimeoutSec  = 10
	defaultRedisReadTimeoutSec  = 20
	defaultRedisWriteTimeoutSec = 20
)

const (
	// Redis key prefix for trace carriers
	TraceCarrierCacheKey = "OTEL:TRACECARRIER"
)

// getGroupKey constructs the full Redis key for a trace carrier group
func getGroupKey(group string) string {
	return TraceCarrierCacheKey + ":" + group
}

// initRedisCache initializes Redis connection and sets the global cache
func initRedisCache(config *RedisConfig) {
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
		stdLog.Fatalf("Failed to ping to Redis: %v", err)
	}

	// Init Redis cache
	cache = &redisCache{
		redisClient: redisClient,
	}
}

// getTraceCarrierFromGroup retrieves a trace carrier from Redis hash
func (rCache *redisCache) getTraceCarrierFromGroup(group string, key string) (TraceCarrier, error) {
	rawValue, err := rCache.redisClient.HGet(context.Background(), getGroupKey(group), key).Result()

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

// setTraceCarrierFromGroup stores a trace carrier in Redis hash
func (rCache *redisCache) setTraceCarrierFromGroup(group string, key string, traceCarrier TraceCarrier) error {
	byteValue, err := json.Marshal(traceCarrier)
	if err != nil {
		return err
	}

	return rCache.redisClient.HSet(context.Background(), getGroupKey(group), key, string(byteValue)).Err()
}

// deleteTraceCarrierFromGroup removes a specific trace carrier from Redis has
func (rCache *redisCache) deleteTraceCarrierFromGroup(group string, key string) error {
	return rCache.redisClient.HDel(context.Background(), getGroupKey(group), key).Err()
}

// deleteTraceCarrierGroup removes an entire group of trace carriers
func (rCache *redisCache) deleteTraceCarrierGroup(group string) error {
	return rCache.redisClient.Del(context.Background(), getGroupKey(group)).Err()
}

// clearTraceCarrier removes all groups of trace carriers
func (rCache *redisCache) clearTraceCarrier() error {
	ctx := context.Background()

	var cursor uint64
	pattern := fmt.Sprintf("%s*", TraceCarrierCacheKey)
	keys := make([]string, 0)

	for {
		existingKeys, nextCursor, err := rCache.redisClient.Scan(context.Background(), cursor, pattern, 100).Result()
		if err != nil {
			return nil
		}
		keys = append(keys, existingKeys...)

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return rCache.redisClient.Del(ctx, keys...).Err()
}

// Public API functions with nil-safety checks

// GetCacheTraceCarrierFromGroup retrieves a trace carrier from cache.
// Returns ErrRedisUnconfigured if Redis was not initialized.
//
// Example:
//
//	carrier, err := otel.GetCacheTraceCarrierFromGroup("jobs", "job-123")
//	if err == nil && carrier != nil {
//	    ctx := carrier.ExtractContext()
//	}
func GetCacheTraceCarrierFromGroup(group string, key string) (TraceCarrier, error) {
	if cache == nil {
		return TraceCarrier{}, ErrCacheUnconfigured
	}

	return cache.getTraceCarrierFromGroup(group, key)
}

// SetCacheTraceCarrierFromGroup stores a trace carrier in cache.
// Returns ErrRedisUnconfigured if Redis was not initialized.
//
// Example:
//
//	carrier := otel.ExportTraceCarrier(ctx)
//	err := otel.SetCacheTraceCarrierFromGroup("jobs", "job-123", carrier)
func SetCacheTraceCarrierFromGroup(group string, key string, traceCarrier TraceCarrier) error {
	if cache == nil {
		return ErrCacheUnconfigured
	}

	return cache.setTraceCarrierFromGroup(group, key, traceCarrier)
}

// DeleteCacheTraceCarrierFromGroup removes a trace carrier from cache.
// Returns ErrRedisUnconfigured if Redis was not initialized.
func DeleteCacheTraceCarrierFromGroup(group string, key string) error {
	if cache == nil {
		return ErrCacheUnconfigured
	}

	return cache.deleteTraceCarrierFromGroup(group, key)
}

// DeleteCacheTraceCarrierGroup removes all trace carriers in a group.
// Returns ErrRedisUnconfigured if Redis was not initialized.
func DeleteCacheTraceCarrierGroup(group string) error {
	if cache == nil {
		return ErrCacheUnconfigured
	}

	return cache.deleteTraceCarrierGroup(group)
}

// ClearCacheTraceCarrier removes all groups of trace carriers.
// Returns ErrRedisUnconfigured if Redis was not initialized.
func ClearCacheTraceCarrier() error {
	if cache == nil {
		return ErrCacheUnconfigured
	}

	return cache.clearTraceCarrier()
}

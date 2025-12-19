package otel

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

// Cache defines the interface for managing otel operations.
type Cache interface {
	// Methods to store, retrieve, and delete trace carriers organized by groups.

	GetTraceCarrierFromGroup(group string, key string) (TraceCarrier, error)
	SetTraceCarrierFromGroup(group string, key string, traceCarrier TraceCarrier) error
	DeleteTraceCarrierFromGroup(group string, key string) error
	DeleteTraceCarrierGroup(group string) error
}

// RedisCache implements the Cache interface using Redis as the storage backend.
// It uses Redis hash structures to organize trace carriers by groups.
type RedisCache struct {
	rawClient redis.Cmdable
}

// TraceCarrierCacheKey is the Redis key prefix for all trace carrier data.
const (
	TraceCarrierCacheKey = "OTEL:TRACECARRIER"
)

// getGroupKey constructs the full Redis key for a trace carrier group.
// Format: "OTEL:TRACECARRIER:{group}"
func getGroupKey(group string) string {
	return TraceCarrierCacheKey + ":" + group
}

// NewOtelCacheWithRedisClient creates a new Cache implementation backed by Redis.
// The redisClient parameter can be either a redis.Client or redis.ClusterClient.
func NewOtelCacheWithRedisClient(redisClient redis.Cmdable) Cache {
	return &RedisCache{
		rawClient: redisClient,
	}
}

// GetTraceCarrierFromGroup retrieves a trace carrier from Redis hash storage.
// It returns an empty TraceCarrier (not an error) when the key doesn't exist.
// Other errors include Redis connection issues or JSON unmarshaling failures.
func (rCache *RedisCache) GetTraceCarrierFromGroup(group string, key string) (TraceCarrier, error) {
	rawValue, err := rCache.rawClient.HGet(context.Background(), getGroupKey(group), key).Result()

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

// SetTraceCarrierFromGroup stores a trace carrier in Redis hash storage.
// The trace carrier is serialized to JSON before storage.
func (rCache *RedisCache) SetTraceCarrierFromGroup(group string, key string, traceCarrier TraceCarrier) error {
	byteValue, err := json.Marshal(traceCarrier)
	if err != nil {
		return err
	}

	return rCache.rawClient.HSet(context.Background(), getGroupKey(group), key, string(byteValue)).Err()
}

// DeleteTraceCarrierFromGroup removes a specific field from the Redis hash.
// Returns nil if the field doesn't exist.
func (rCache *RedisCache) DeleteTraceCarrierFromGroup(group string, key string) error {
	return rCache.rawClient.HDel(context.Background(), getGroupKey(group), key).Err()
}

// DeleteTraceCarrierGroup removes the entire Redis hash for the specified group.
// This deletes all trace carriers within that group.
func (rCache *RedisCache) DeleteTraceCarrierGroup(group string) error {
	return rCache.rawClient.Del(context.Background(), getGroupKey(group)).Err()
}

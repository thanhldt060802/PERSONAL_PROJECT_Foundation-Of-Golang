package otel

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	cache Cache
)

type Cache interface {
	GetTraceCarrierFromGroup(group string, key string) (TraceCarrier, error)
	SetTraceCarrierFromGroup(group string, key string, traceCarrier TraceCarrier) error
	DeleteTraceCarrierFromGroup(group string, key string) error
	DeleteTraceCarrierGroup(group string) error
}

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

type redisCache struct {
	redisClient *redis.Client
}

var (
	ErrRedisUnconfigured = errors.New("redis is unconfigured")
)

const (
	defaultRedisPoolSize        = 10
	defaultRedisPoolTimeoutSec  = 20
	defaultRedisIdleTimeoutSec  = 10
	defaultRedisReadTimeoutSec  = 20
	defaultRedisWriteTimeoutSec = 20
)

const (
	TraceCarrierCacheKey = "OTEL:TRACECARRIER"
)

func getGroupKey(group string) string {
	return TraceCarrierCacheKey + ":" + group
}

func initRedisCache(config *RedisConfig) {
	// Create Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:            config.Address,
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

func (rCache *redisCache) GetTraceCarrierFromGroup(group string, key string) (TraceCarrier, error) {
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

func (rCache *redisCache) SetTraceCarrierFromGroup(group string, key string, traceCarrier TraceCarrier) error {
	byteValue, err := json.Marshal(traceCarrier)
	if err != nil {
		return err
	}

	return rCache.redisClient.HSet(context.Background(), getGroupKey(group), key, string(byteValue)).Err()
}

func (rCache *redisCache) DeleteTraceCarrierFromGroup(group string, key string) error {
	return rCache.redisClient.HDel(context.Background(), getGroupKey(group), key).Err()
}

func (rCache *redisCache) DeleteTraceCarrierGroup(group string) error {
	return rCache.redisClient.Del(context.Background(), getGroupKey(group)).Err()
}

func GetCacheTraceCarrierFromGroup(group string, key string) (TraceCarrier, error) {
	if cache == nil {
		return TraceCarrier{}, ErrRedisUnconfigured
	}

	return cache.GetTraceCarrierFromGroup(group, key)
}

func SetCacheTraceCarrierFromGroup(group string, key string, traceCarrier TraceCarrier) error {
	if cache == nil {
		return ErrRedisUnconfigured
	}

	return cache.SetTraceCarrierFromGroup(group, key, traceCarrier)
}

func DeleteCacheTraceCarrierFromGroup(group string, key string) error {
	if cache == nil {
		return ErrRedisUnconfigured
	}

	return cache.DeleteTraceCarrierFromGroup(group, key)
}

func DeleteCacheTraceCarrierGroup(group string) error {
	if cache == nil {
		return ErrRedisUnconfigured
	}

	return cache.DeleteTraceCarrierGroup(group)
}

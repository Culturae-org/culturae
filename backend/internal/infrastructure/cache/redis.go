// backend/internal/infrastructure/cache/redis.go

package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisClientInterface interface {
	Ping(ctx context.Context) error
	Close() error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	GetJSON(ctx context.Context, key string, dest interface{}) error
	SetCache(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	GetCache(ctx context.Context, key string, dest interface{}) error
	DeleteCache(ctx context.Context, key string) error
	CheckRateLimit(ctx context.Context, key string, limit int64, window time.Duration) (bool, int64, int64, error)
	Publish(ctx context.Context, channel string, message interface{}) error
	Subscribe(ctx context.Context, channels ...string) *redis.PubSub
	GetInfo(ctx context.Context) (map[string]interface{}, error)
	Scan(ctx context.Context, pattern string) ([]string, error)

	Expire(ctx context.Context, key string, expiration time.Duration) error
	RPush(ctx context.Context, key string, values ...interface{}) error
	LRem(ctx context.Context, key string, count int64, value interface{}) error
	LLen(ctx context.Context, key string) (int64, error)
	LPop(ctx context.Context, key string) (string, error)
	LPush(ctx context.Context, key string, values ...interface{}) error

	ZAdd(ctx context.Context, key string, members ...ZSetMember) error
	ZRem(ctx context.Context, key string, members ...string) error
	ZRangeByScore(ctx context.Context, key string, opts ZRangeByScoreOptions) ([]string, error)
	ZGetScore(ctx context.Context, key string, member string) (float64, error)

	MGet(ctx context.Context, keys ...string) ([]string, error)

	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error)

	NativeClient() redis.UniversalClient
}

type ZSetMember struct {
	Score  float64
	Member string
}

type ZRangeByScoreOptions struct {
	Min    string
	Max    string
	Offset int64
	Count  int64
}

type RedisClient struct {
	client *redis.Client
	logger *zap.Logger
}

func NewRedisClient(addr, password string, db int, logger *zap.Logger) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     20,
		MinIdleConns: 5,
		PoolTimeout:  4 * time.Second,
		ConnMaxIdleTime: 5 * time.Minute,
		ConnMaxLifetime: 30 * time.Minute,
		MaxRetries:   3,
	})

	return &RedisClient{
		client: rdb,
		logger: logger,
	}
}

func (r *RedisClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}

func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisClient) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	return count > 0, err
}

func (r *RedisClient) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return r.client.Set(ctx, key, jsonData, expiration).Err()
}

func (r *RedisClient) GetJSON(ctx context.Context, key string, dest interface{}) error {
	jsonData, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(jsonData), dest); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

func (r *RedisClient) SetCache(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	cacheKey := fmt.Sprintf("cache:%s", key)
	return r.SetJSON(ctx, cacheKey, value, expiration)
}

func (r *RedisClient) GetCache(ctx context.Context, key string, dest interface{}) error {
	cacheKey := fmt.Sprintf("cache:%s", key)
	return r.GetJSON(ctx, cacheKey, dest)
}

func (r *RedisClient) DeleteCache(ctx context.Context, key string) error {
	cacheKey := fmt.Sprintf("cache:%s", key)
	return r.client.Del(ctx, cacheKey).Err()
}

func (r *RedisClient) CheckRateLimit(ctx context.Context, key string, limit int64, window time.Duration) (bool, int64, int64, error) {
	rateLimitKey := fmt.Sprintf("rate_limit:%s", key)

	count, err := r.client.Incr(ctx, rateLimitKey).Result()
	if err != nil {
		return false, 0, 0, err
	}

	ttl, err := r.client.TTL(ctx, rateLimitKey).Result()
	if err != nil {
		_ = r.client.Expire(ctx, rateLimitKey, window).Err()
		resetAt := time.Now().Add(window).Unix()
		remaining := limit - count
		if remaining < 0 {
			remaining = 0
		}
		return count <= limit, remaining, resetAt, nil
	}

	if ttl <= 0 {
		_ = r.client.Expire(ctx, rateLimitKey, window).Err()
		ttl = window
	}

	resetAt := time.Now().Add(ttl).Unix()
	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	return count <= limit, remaining, resetAt, nil
}

func (r *RedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	return r.client.Publish(ctx, channel, jsonData).Err()
}

func (r *RedisClient) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.client.Subscribe(ctx, channels...)
}

func (r *RedisClient) GetInfo(ctx context.Context) (map[string]interface{}, error) {
	info, err := r.client.Info(ctx).Result()
	if err != nil {
		return nil, err
	}

	details := make(map[string]interface{})

	lines := strings.Split(info, "\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				switch key {
				case "redis_version":
					details["version"] = value
				case "uptime_in_seconds":
					details["uptime_seconds"] = value
				case "connected_clients":
					details["connected_clients"] = value
				case "used_memory":
					details["used_memory_bytes"] = value
				case "total_connections_received":
					details["total_connections"] = value
				case "total_commands_processed":
					details["total_commands"] = value
				case "keyspace_hits":
					details["keyspace_hits"] = value
				case "keyspace_misses":
					details["keyspace_misses"] = value
				case "evicted_keys":
					details["evicted_keys"] = value
				case "expired_keys":
					details["expired_keys"] = value
				}
			}
		}
	}

	if hitsStr, ok := details["keyspace_hits"]; ok {
		if missesStr, ok := details["keyspace_misses"]; ok {
			hits, ok1 := hitsStr.(string)
			misses, ok2 := missesStr.(string)
			if ok1 && ok2 {
				hitsVal := parseInt(hits)
				missesVal := parseInt(misses)
				total := hitsVal + missesVal
				if total > 0 {
					details["cache_hit_rate"] = float64(hitsVal) / float64(total) * 100
				}
			}
		}
	}

	return details, nil
}

func parseInt(s string) int64 {
	var result int64
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil {
		return 0
	}
	return result
}

func (r *RedisClient) Scan(ctx context.Context, pattern string) ([]string, error) {
	var keys []string
	var cursor uint64
	const scanCount = 100

	for {
		var scanKeys []string
		var err error

		scanKeys, cursor, err = r.client.Scan(ctx, cursor, pattern, scanCount).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan keys: %w", err)
		}

		keys = append(keys, scanKeys...)

		if cursor == 0 {
			break
		}
	}

	return keys, nil
}

func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

func (r *RedisClient) RPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.RPush(ctx, key, values...).Err()
}

func (r *RedisClient) LRem(ctx context.Context, key string, count int64, value interface{}) error {
	return r.client.LRem(ctx, key, count, value).Err()
}

func (r *RedisClient) LLen(ctx context.Context, key string) (int64, error) {
	return r.client.LLen(ctx, key).Result()
}

func (r *RedisClient) LPop(ctx context.Context, key string) (string, error) {
	val, err := r.client.LPop(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}
	return val, err
}

func (r *RedisClient) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.LPush(ctx, key, values...).Err()
}

func (r *RedisClient) ZAdd(ctx context.Context, key string, members ...ZSetMember) error {
	if len(members) == 0 {
		return nil
	}
	z := make([]redis.Z, len(members))
	for i, m := range members {
		z[i] = redis.Z{Score: m.Score, Member: m.Member}
	}
	return r.client.ZAdd(ctx, key, z...).Err()
}

func (r *RedisClient) ZRem(ctx context.Context, key string, members ...string) error {
	if len(members) == 0 {
		return nil
	}
	iface := make([]interface{}, len(members))
	for i, m := range members {
		iface[i] = m
	}
	return r.client.ZRem(ctx, key, iface...).Err()
}

func (r *RedisClient) ZRangeByScore(ctx context.Context, key string, opts ZRangeByScoreOptions) ([]string, error) {
	result, err := r.client.ZRangeArgs(ctx, redis.ZRangeArgs{
		Key:     key,
		Start:   opts.Min,
		Stop:    opts.Max,
		ByScore: true,
		Offset:  opts.Offset,
		Count:   opts.Count,
	}).Result()
	return result, err
}

func (r *RedisClient) ZGetScore(ctx context.Context, key string, member string) (float64, error) {
	result, err := r.client.ZScore(ctx, key, member).Result()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	return result, err
}

func (r *RedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, value, expiration).Result()
}

func (r *RedisClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return r.client.Eval(ctx, script, keys, args...).Result()
}

func (r *RedisClient) NativeClient() redis.UniversalClient {
	return r.client
}

func (r *RedisClient) MGet(ctx context.Context, keys ...string) ([]string, error) {
	results, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to mget keys: %w", err)
	}
	strs := make([]string, len(results))
	for i, v := range results {
		if s, ok := v.(string); ok {
			strs[i] = s
		}
	}
	return strs, nil
}

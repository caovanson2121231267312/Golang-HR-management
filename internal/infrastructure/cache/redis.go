package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"hr-management-system/internal/config"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client   *redis.Client
	prefix   string
	defaultTTL time.Duration
}

var cache *RedisCache

func NewRedisCache(cfg *config.RedisConfig) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	cache = &RedisCache{
		client:     client,
		prefix:     "hr:",
		defaultTTL: cfg.CacheTTL,
	}

	return cache, nil
}

func GetCache() *RedisCache {
	return cache
}

func (r *RedisCache) Client() *redis.Client {
	return r.client
}

func (r *RedisCache) Close() error {
	return r.client.Close()
}

// Key management
func (r *RedisCache) key(k string) string {
	return r.prefix + k
}

// Basic operations
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl ...time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	expiration := r.defaultTTL
	if len(ttl) > 0 {
		expiration = ttl[0]
	}

	return r.client.Set(ctx, r.key(key), data, expiration).Err()
}

func (r *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.Get(ctx, r.key(key)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return fmt.Errorf("failed to get key: %w", err)
	}

	return json.Unmarshal(data, dest)
}

func (r *RedisCache) Delete(ctx context.Context, keys ...string) error {
	fullKeys := make([]string, len(keys))
	for i, k := range keys {
		fullKeys[i] = r.key(k)
	}
	return r.client.Del(ctx, fullKeys...).Err()
}

func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, r.key(key)).Result()
	return result > 0, err
}

func (r *RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return r.client.Expire(ctx, r.key(key), ttl).Err()
}

func (r *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, r.key(key)).Result()
}

// Pattern-based deletion
func (r *RedisCache) DeleteByPattern(ctx context.Context, pattern string) error {
	iter := r.client.Scan(ctx, 0, r.key(pattern), 0).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) > 0 {
		return r.client.Del(ctx, keys...).Err()
	}
	return nil
}

// Hash operations
func (r *RedisCache) HSet(ctx context.Context, key string, values map[string]interface{}) error {
	args := make([]interface{}, 0, len(values)*2)
	for k, v := range values {
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		args = append(args, k, data)
	}
	return r.client.HSet(ctx, r.key(key), args...).Err()
}

func (r *RedisCache) HGet(ctx context.Context, key, field string, dest interface{}) error {
	data, err := r.client.HGet(ctx, r.key(key), field).Bytes()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return err
	}
	return json.Unmarshal(data, dest)
}

func (r *RedisCache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, r.key(key)).Result()
}

func (r *RedisCache) HDel(ctx context.Context, key string, fields ...string) error {
	return r.client.HDel(ctx, r.key(key), fields...).Err()
}

// List operations
func (r *RedisCache) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.LPush(ctx, r.key(key), values...).Err()
}

func (r *RedisCache) RPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.RPush(ctx, r.key(key), values...).Err()
}

func (r *RedisCache) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.LRange(ctx, r.key(key), start, stop).Result()
}

func (r *RedisCache) LLen(ctx context.Context, key string) (int64, error) {
	return r.client.LLen(ctx, r.key(key)).Result()
}

// Set operations
func (r *RedisCache) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SAdd(ctx, r.key(key), members...).Err()
}

func (r *RedisCache) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.client.SMembers(ctx, r.key(key)).Result()
}

func (r *RedisCache) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return r.client.SIsMember(ctx, r.key(key), member).Result()
}

func (r *RedisCache) SRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SRem(ctx, r.key(key), members...).Err()
}

// Sorted set operations
func (r *RedisCache) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	return r.client.ZAdd(ctx, r.key(key), members...).Err()
}

func (r *RedisCache) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.ZRange(ctx, r.key(key), start, stop).Result()
}

func (r *RedisCache) ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) ([]string, error) {
	return r.client.ZRangeByScore(ctx, r.key(key), opt).Result()
}

func (r *RedisCache) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.ZRem(ctx, r.key(key), members...).Err()
}

// Counter operations
func (r *RedisCache) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, r.key(key)).Result()
}

func (r *RedisCache) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.client.IncrBy(ctx, r.key(key), value).Result()
}

func (r *RedisCache) Decr(ctx context.Context, key string) (int64, error) {
	return r.client.Decr(ctx, r.key(key)).Result()
}

// Lock operations (distributed lock)
func (r *RedisCache) Lock(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
	return r.client.SetNX(ctx, r.key("lock:"+key), value, ttl).Result()
}

func (r *RedisCache) Unlock(ctx context.Context, key string, value string) error {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	return r.client.Eval(ctx, script, []string{r.key("lock:" + key)}, value).Err()
}

// Rate limiter
func (r *RedisCache) RateLimit(ctx context.Context, key string, limit int64, window time.Duration) (bool, int64, error) {
	now := time.Now().UnixNano()
	windowStart := now - int64(window)

	pipe := r.client.Pipeline()
	
	// Remove old entries
	pipe.ZRemRangeByScore(ctx, r.key("rl:"+key), "0", fmt.Sprintf("%d", windowStart))
	
	// Count current entries
	countCmd := pipe.ZCard(ctx, r.key("rl:"+key))
	
	// Add current request
	pipe.ZAdd(ctx, r.key("rl:"+key), redis.Z{Score: float64(now), Member: now})
	
	// Set expiration
	pipe.Expire(ctx, r.key("rl:"+key), window)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, err
	}
	
	count := countCmd.Val()
	allowed := count < limit
	remaining := limit - count - 1
	if remaining < 0 {
		remaining = 0
	}
	
	return allowed, remaining, nil
}

// OTP Cache
func (r *RedisCache) SetOTP(ctx context.Context, identifier string, otp string, ttl time.Duration) error {
	return r.client.Set(ctx, r.key("otp:"+identifier), otp, ttl).Err()
}

func (r *RedisCache) GetOTP(ctx context.Context, identifier string) (string, error) {
	return r.client.Get(ctx, r.key("otp:"+identifier)).Result()
}

func (r *RedisCache) DeleteOTP(ctx context.Context, identifier string) error {
	return r.client.Del(ctx, r.key("otp:"+identifier)).Err()
}

// Session cache
func (r *RedisCache) SetSession(ctx context.Context, sessionID string, data interface{}, ttl time.Duration) error {
	return r.Set(ctx, "session:"+sessionID, data, ttl)
}

func (r *RedisCache) GetSession(ctx context.Context, sessionID string, dest interface{}) error {
	return r.Get(ctx, "session:"+sessionID, dest)
}

func (r *RedisCache) DeleteSession(ctx context.Context, sessionID string) error {
	return r.Delete(ctx, "session:"+sessionID)
}

// User cache
func (r *RedisCache) SetUserCache(ctx context.Context, userID string, data interface{}) error {
	return r.Set(ctx, "user:"+userID, data)
}

func (r *RedisCache) GetUserCache(ctx context.Context, userID string, dest interface{}) error {
	return r.Get(ctx, "user:"+userID, dest)
}

func (r *RedisCache) InvalidateUserCache(ctx context.Context, userID string) error {
	return r.Delete(ctx, "user:"+userID)
}

// Permission cache
func (r *RedisCache) SetPermissions(ctx context.Context, userID string, permissions []string) error {
	return r.Set(ctx, "perms:"+userID, permissions, time.Hour)
}

func (r *RedisCache) GetPermissions(ctx context.Context, userID string) ([]string, error) {
	var perms []string
	err := r.Get(ctx, "perms:"+userID, &perms)
	return perms, err
}

func (r *RedisCache) InvalidatePermissions(ctx context.Context, userID string) error {
	return r.Delete(ctx, "perms:"+userID)
}

// Health check
func (r *RedisCache) HealthCheck(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Cache miss error
var ErrCacheMiss = fmt.Errorf("cache miss")

// Cache keys constants
const (
	KeyUserPrefix       = "user:"
	KeySessionPrefix    = "session:"
	KeyPermPrefix       = "perms:"
	KeyOTPPrefix        = "otp:"
	KeyRateLimitPrefix  = "rl:"
	KeyLockPrefix       = "lock:"
	KeyEmployeePrefix   = "emp:"
	KeyDepartmentPrefix = "dept:"
	KeyAttendancePrefix = "att:"
	KeyPayrollPrefix    = "payroll:"
)

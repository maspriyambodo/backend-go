package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

// Cache provides Redis-based caching functionality
type Cache struct {
	client *redis.Client
	ctx    context.Context
}

// NewCache creates a new cache instance
func NewCache(client *redis.Client) *Cache {
	return &Cache{
		client: client,
		ctx:    context.Background(),
	}
}

// Set serializes and stores data in Redis with expiration
func (c *Cache) Set(key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	return c.client.Set(c.ctx, key, data, expiration).Err()
}

// Get retrieves and deserializes data from Redis
func (c *Cache) Get(key string, dest interface{}) error {
	data, err := c.client.Get(c.ctx, key).Result()
	if err == redis.Nil {
		return fmt.Errorf("cache miss for key: %s", key)
	}
	if err != nil {
		return fmt.Errorf("failed to get from cache: %w", err)
	}

	err = json.Unmarshal([]byte(data), dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	return nil
}

// Delete removes a key from Redis
func (c *Cache) Delete(key string) error {
	return c.client.Del(c.ctx, key).Err()
}

// DeletePattern removes all keys matching a pattern
func (c *Cache) DeletePattern(pattern string) error {
	keys, err := c.client.Keys(c.ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to find keys for pattern %s: %w", pattern, err)
	}

	if len(keys) > 0 {
		return c.client.Del(c.ctx, keys...).Err()
	}

	return nil
}

// Exists checks if a key exists in Redis
func (c *Cache) Exists(key string) bool {
	count, err := c.client.Exists(c.ctx, key).Result()
	if err != nil {
		log.Printf("Error checking cache existence: %v", err)
		return false
	}
	return count > 0
}

// Increment increments a numeric value in Redis
func (c *Cache) Increment(key string) (int64, error) {
	return c.client.Incr(c.ctx, key).Result()
}

// SetNX sets a key only if it doesn't exist (useful for locking)
func (c *Cache) SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal cache data: %w", err)
	}

	return c.client.SetNX(c.ctx, key, data, expiration).Result()
}

// GetTTL returns the remaining time to live of a key
func (c *Cache) GetTTL(key string) (time.Duration, error) {
	return c.client.TTL(c.ctx, key).Result()
}

// Common cache key patterns
const (
	CacheKeyPrefix         = "cms:"
	CacheKeyMenuList       = CacheKeyPrefix + "menus:list"
	CacheKeyRolesList      = CacheKeyPrefix + "roles:list"
	CacheKeyUsersList      = CacheKeyPrefix + "users:list:%d:%d" // page:limit
	CacheKeyUsersCount     = CacheKeyPrefix + "users:count"
	CacheKeyMenuNavigation = CacheKeyPrefix + "menu:navigation"
	CacheKeyUser           = CacheKeyPrefix + "user:%s" // user_id
	CacheKeyRole           = CacheKeyPrefix + "role:%s" // role_id
	CacheKeyMenu           = CacheKeyPrefix + "menu:%s" // menu_id
)

// Default expirations
const (
	DefaultListExpiration       = 10 * time.Minute // For list endpoints
	DefaultDetailExpiration     = 5 * time.Minute  // For individual items
	DefaultCountExpiration      = 15 * time.Minute // For counts
	DefaultNavigationExpiration = 30 * time.Minute // For navigation (less frequent changes)
)

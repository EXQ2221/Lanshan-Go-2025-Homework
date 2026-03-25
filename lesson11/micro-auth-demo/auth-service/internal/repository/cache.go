package repository

import (
	"context"
	"encoding/json"
	"time"

	redisv9 "github.com/redis/go-redis/v9"
)

const (
	sessionCacheKeyPrefix    = "auth:session:"
	accessBlacklistKeyPrefix = "auth:access:blacklist:"
)

type SessionCacheEntry struct {
	UserID           int64  `json:"user_id"`
	Status           string `json:"status"`
	CurrentAccessJTI string `json:"current_access_jti"`
}

type AuthCache interface {
	SetSession(ctx context.Context, sessionID string, entry SessionCacheEntry, ttl time.Duration) error
	GetSession(ctx context.Context, sessionID string) (*SessionCacheEntry, error)
	BlacklistAccessToken(ctx context.Context, jti string, ttl time.Duration) error
	IsAccessTokenBlacklisted(ctx context.Context, jti string) (bool, error)
}

type RedisAuthCache struct {
	client *redisv9.Client
}

func NewAuthCache(client *redisv9.Client) *RedisAuthCache {
	return &RedisAuthCache{client: client}
}

func (r *RedisAuthCache) SetSession(ctx context.Context, sessionID string, entry SessionCacheEntry, ttl time.Duration) error {
	payload, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, sessionCacheKeyPrefix+sessionID, payload, ttl).Err()
}

func (r *RedisAuthCache) GetSession(ctx context.Context, sessionID string) (*SessionCacheEntry, error) {
	payload, err := r.client.Get(ctx, sessionCacheKeyPrefix+sessionID).Result()
	if err != nil {
		if err == redisv9.Nil {
			return nil, nil
		}
		return nil, err
	}

	var entry SessionCacheEntry
	if err := json.Unmarshal([]byte(payload), &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *RedisAuthCache) BlacklistAccessToken(ctx context.Context, jti string, ttl time.Duration) error {
	if jti == "" || ttl <= 0 {
		return nil
	}
	return r.client.Set(ctx, accessBlacklistKeyPrefix+jti, "1", ttl).Err()
}

func (r *RedisAuthCache) IsAccessTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	if jti == "" {
		return false, nil
	}
	count, err := r.client.Exists(ctx, accessBlacklistKeyPrefix+jti).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

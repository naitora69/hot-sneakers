package db

import "github.com/redis/go-redis/v9"

type RedisCache struct {
	RDB *redis.Client
}

func NewRedisCache(address string) (*RedisCache, error) {
	opt, err := redis.ParseURL(address)
	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient(opt)

	return &RedisCache{RDB: rdb}, nil
}

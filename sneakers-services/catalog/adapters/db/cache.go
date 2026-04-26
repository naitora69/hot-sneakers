package db

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"time"

	"hotsneakers/catalog/core"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	RDB  *redis.Client
	next core.Repository
	ttl  time.Duration
	log  *slog.Logger
}

func NewRedisCache(address string, next core.Repository, ttl time.Duration, log *slog.Logger) (*RedisCache, error) {
	opt, err := redis.ParseURL(address)
	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient(opt)

	return &RedisCache{
		RDB:  rdb,
		next: next,
		ttl:  ttl,
		log:  log,
	}, nil
}

func (rc RedisCache) GetAllSneakers(ctx context.Context) ([]core.Sneaker, error) {
	return rc.next.GetAllSneakers(ctx)
}

func (rc RedisCache) GetSneakerByID(ctx context.Context, id int) (core.Sneaker, error) {
	key := "sneaker:" + strconv.Itoa(id)

	val, err := rc.RDB.Get(ctx, key).Result()
	if err == redis.Nil {
		sneaker, err := rc.next.GetSneakerByID(ctx, id)
		if err != nil {
			rc.log.Error("error when get sneaker from pg", "error", err)
			return core.Sneaker{}, err
		}

		data, err := json.Marshal(sneaker)
		if err != nil {
			rc.log.Error("json marshal sneaker from pg error", "error", err)
			return core.Sneaker{}, err
		}

		err = rc.RDB.Set(ctx, key, data, rc.ttl).Err()
		if err != nil {
			rc.log.Error("redis save sneaker error", "error", err)
			return core.Sneaker{}, err
		}

		return sneaker, nil
	} else if err != nil {
		rc.log.Error("redis internal error", "error", err)
		return rc.next.GetSneakerByID(ctx, id)
	}

	res := core.Sneaker{}

	err = json.Unmarshal([]byte(val), &res)
	if err != nil {
		rc.log.Error("json unmarshal sneaker from redis error", "error", err)
		return core.Sneaker{}, err
	}

	return res, nil
}

func (rc RedisCache) CreateSneaker(ctx context.Context, sneaker core.CreateSneaker) (int64, error) {
	return rc.next.CreateSneaker(ctx, sneaker)
}

func (rc RedisCache) UpdateSneaker(ctx context.Context, sneaker core.UpdateSneaker) error {
	key := "sneaker:" + strconv.Itoa(int(sneaker.ID))

	err := rc.RDB.Del(ctx, key).Err()
	if err != nil {
		rc.log.Warn("cannot delete cache from redis", "key", key, "error", err)
	}

	return rc.next.UpdateSneaker(ctx, sneaker)
}

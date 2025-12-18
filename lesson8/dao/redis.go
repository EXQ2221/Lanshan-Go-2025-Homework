package dao

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var Redis *redis.Client

func InitRedis() {

	rd := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "123456",
		DB:       0,
	})

	if _, err := rd.Ping(context.Background()).Result(); err != nil {
		panic("redis init failed:" + err.Error())
	}

	Redis = rd
}

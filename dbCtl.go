package main

import (
	"github.com/go-redis/redis"
)

var redisDb *redis.Client

func initDbClient(err error) {
	redisDb = redis.NewClient(&redis.Options{
		Addr:     "localhost",
		Password: "",
		DB:       0,
	})
}

func getDataBackend() {
}

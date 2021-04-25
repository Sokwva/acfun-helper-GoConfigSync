package main

import (
	"log"

	"github.com/go-redis/redis"
)

var redisDb *redis.Client

func initDbClient() *redis.Client {
	redisDb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	return redisDb
}

func checkRedisServer(conn *redis.Client) bool {
	_, err := conn.Ping().Result()
	if err == nil {
		return true
	}
	log.Panic("Connect Redis Error.", err)
	return false
}

// func getDataBackend() {
// }

func dataSet(user string, data string) bool {
	handler := initDbClient()
	if !checkRedisServer(handler) {
		return false
	}
	err := handler.Set(user, data, 0).Err()
	if err != nil {
		log.Panic("Write data to Redis Failed.", err)
		return false
	}
	return true
}

func dataGet(user string) (data string, status bool) {
	handler := initDbClient()
	if !checkRedisServer(handler) {
		return "", false
	}
	result, err := handler.Get(user).Result()
	if err != nil {
		log.Panic("Can't get data from redis.")
		return "", false
	}
	return result, true
}

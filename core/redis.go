package core

import (
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

type rd struct {
	Client   *redis.Client
	Addr     string
	Password string
	Database int
}

var RedisManager *rd

func NewRedis() {
	db, err := strconv.Atoi(GetEnv("REDIS_DB", "1"))
	if err != nil {
		db = 0
	}

	client := redis.NewClient(&redis.Options{
		Addr:     GetEnv("REDIS_ADDR", "localhost:6379"),
		Password: GetEnv("REDIS_PW", ""), // no password set
		DB:       db,                     // use default DB
	})

	_, err = client.Ping().Result()

	if err != nil {
		os.Exit(0)
	}

	RedisManager = &rd{client, client.Options().Addr, client.Options().Password, client.Options().DB}
}

func (r *rd) Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return r.Client.Set(key, value, expiration)
}

func (r *rd) Get(key string) (string, error) {
	return r.Client.Get(key).Result()
}

func (r *rd) Keys(pattern string) ([]string, error) {
	return r.Client.Keys(pattern).Result()
}

func (r *rd) Delete(keys ...string) (int64, error) {
	return r.Client.Del(keys...).Result()
}

/*
	Your keys count must match the exist amount or not all keys exist
	Exists in redis doesn't return what keys actually existed so this function is pretty basic
*/
func (r *rd) Exist(keys ...string) bool {
	test, _ := r.Client.Exists(keys...).Result()
	keyCount := len(keys)

	return test == int64(keyCount)
}

func (r *rd) TTL(key string) (time.Duration, error) {
	return r.Client.TTL(key).Result()
}

package core

import (
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis"
)

type rd struct {
	Client   *redis.Client
	Addr     string
	Password string
}

var RedisManager *rd

func NewRedis() {
	client := redis.NewClient(&redis.Options{
		Addr:     GetEnv("REDIS_ADDR", "localhost:6379"),
		Password: GetEnv("REDIS_PW", ""), // no password set
		DB:       0,                      // use default DB
	})

	_, err := client.Ping().Result()

	if err != nil {
		os.Exit(0)
	}

	RedisManager = &rd{client, client.Options().Addr, client.Options().Password}
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

/* Custom Functions */
func (r *rd) SetUpdatedTime(key string, value interface{}) *redis.StatusCmd {
	return r.Set(fmt.Sprintf("%s_updated", key), value, 0)
}

func (r *rd) GetUpdatedTime(key string) (string, error) {
	return r.Get(fmt.Sprintf("%s_updated", key))
}

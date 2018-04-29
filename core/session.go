package core

import (
	"strconv"

	"github.com/gorilla/securecookie"
	sess "github.com/kataras/iris/sessions"
	"github.com/kataras/iris/sessions/sessiondb/redis"
	"github.com/kataras/iris/sessions/sessiondb/redis/service"
)

type Session struct {
	Session  *sess.Sessions
	Database *redis.Database
}

var SessionManager *Session

func NewSessions() {
	db := redis.New(service.Config{
		Network:     service.DefaultRedisNetwork,
		Addr:        RedisManager.Addr,
		Password:    RedisManager.Password,
		Database:    strconv.Itoa(RedisManager.Database),
		IdleTimeout: service.DefaultRedisIdleTimeout,
		Prefix:      GetEnv("REDIS_PREFIX", "gog_"),
	}) // optionally configure the bridge between your redis server

	// AES only supports key sizes of 16, 24 or 32 bytes.
	// You either need to provide exactly that amount or you derive the key from what you type in.
	hashKey := []byte(GetEnv("SESSION_HASH_KEY", "d2WEEsHDOZTy2qBOcwS8gPd0ZN7UEsxX"))
	blockKey := []byte(GetEnv("SESSION_BLOCK_KEY", "EkbCvPqBqBpEAumrE6LBEf7A0VV9IBFL"))
	secureCookie := securecookie.New(hashKey, blockKey)

	mySessions := sess.New(sess.Config{
		Cookie: GetEnv("SESSION_COOKIE_NAME", "go.glich"),
		Encode: secureCookie.Encode,
		Decode: secureCookie.Decode,
	})

	mySessions.UseDatabase(db)
	SessionManager = &Session{mySessions, db}
}

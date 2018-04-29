package core

import (
	"github.com/gorilla/securecookie"
	sess "github.com/kataras/iris/sessions"
)

type Session struct {
	Session *sess.Sessions
}

var SessionManager *Session

func NewSessions() {
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

	SessionManager = &Session{mySessions}
}

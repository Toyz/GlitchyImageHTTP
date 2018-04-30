package database

import (
	"log"
	"os"

	"github.com/Toyz/GlitchyImageHTTP/core"
	"github.com/globalsign/mgo"
)

type mongo struct {
	mgoSession *mgo.Session
	database   string
}

var MongoInstance *mongo

func NewMongo() {
	session, err := mgo.Dial(core.GetEnv("MONGO_HOST", "localhost"))
	if err != nil {
		// This should never happen but if it does we need to panic... this can cause some wonky effects if we don't...
		log.Panicln(err)
		os.Exit(9)
	}
	session.SetMode(mgo.Monotonic, true)

	userName := core.GetEnv("MONGO_USER", "")
	password := core.GetEnv("MONGO_PASS", "")
	if len(userName) > 0 && len(password) > 0 {
		err := session.Login(&mgo.Credential{
			Username: userName,
			Password: password,
		})

		if err != nil {
			// Panic when we failed to login because well... go build in logger has no warning...
			// Maybe i should replace the built in logger later... Iris has one built in that we could make public
			log.Panicln(err)
			os.Exit(9)
		}
	}

	MongoInstance = &mongo{
		mgoSession: session,
		database:   core.GetEnv("MONGO_DB", "glitch"),
	}
}

func (mg *mongo) session() *mgo.Session {
	return mg.mgoSession.Copy()
}

func (mg *mongo) collection(collection string) (*mgo.Session, *mgo.Collection) {
	session := mg.session()
	c := session.DB(mg.database).C(collection)

	return session, c
}

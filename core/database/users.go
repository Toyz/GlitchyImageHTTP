package database

import (
	"strings"
	"time"

	"github.com/globalsign/mgo/bson"
)

func (mg *mongo) GetUserByEmail(email string) User {
	session, c := mg.collection(USERS_COL)
	defer session.Close()

	email = strings.ToLower(email)
	var usr User
	c.Find(bson.M{"email": email}).One(&usr)

	if len(usr.Email) > 0 {
		return usr
	}

	return User{}
}

func (mg *mongo) GetUserByID(id bson.ObjectId) User {
	session, c := mg.collection(USERS_COL)
	defer session.Close()

	var usr User
	c.Find(bson.M{"_id": id}).One(&usr)

	if len(usr.Email) > 0 {
		return usr
	}

	return User{}
}

// This sets the user ID and the users Join & Update times
func (mg *mongo) InsertUser(user User) User {
	session, c := mg.collection(USERS_COL)
	defer session.Close()

	user.Email = strings.ToLower(user.Email)
	user.Username = strings.ToLower(user.Username)

	user.MGID = bson.NewObjectId()
	user.Joined = time.Now()
	user.Updated = user.Joined
	user.LastLogin = user.Joined

	c.Insert(user)

	return user
}

func (mg *mongo) ChangePasswordUser(user User, newPass string) User {
	session, c := mg.collection(USERS_COL)
	defer session.Close()

	user.Password = newPass
	user.Updated = time.Now()

	c.Update(
		bson.M{"_id": user.MGID},
		bson.M{"$set": bson.M{"password": user.Password, "updated": user.Updated}},
	)

	return user
}

func (mg *mongo) SetLastLogin(user User) User {
	session, c := mg.collection(USERS_COL)
	defer session.Close()

	user.LastLogin = time.Now()

	c.Update(
		bson.M{"_id": user.MGID},
		bson.M{"$set": bson.M{"ll": user.LastLogin}},
	)

	return user
}

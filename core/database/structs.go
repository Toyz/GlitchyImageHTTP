package database

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

const (
	ARTIDS_COL     = "artIds"
	ALERTS_COL     = "alerts" // unused... Will be used with @AlertItem
	EXPRESSION_COL = "exps"
	CATEGORY_COL   = "cats"
	UPLOADS_COL    = "uploads"
	USERS_COL      = "users"
)

type Upload struct {
	MGID        bson.ObjectId   `json:"-" bson:"_id,omitempty"`
	ImageID     bson.ObjectId   `json:"-" bson:"img_id`
	Expressions []bson.ObjectId `json:"expressions" bson:"exps"` // set if multiable are used (will always be 0 if empty)
	Tags        []bson.ObjectId `json:"tags"`
	User        bson.ObjectId   `json:"user" bson:"user"`
	Views       int             `json:"views"`
}

type ArtItem struct {
	MGID        bson.ObjectId `json:"-" bson:"_id,omitempty"`
	Folder      string        `json:"folder"`
	FileName    string        `json:"filename"`
	OrgFileName string        `json:"orgfilename"`
	Uploaded    time.Time     `json:"uploaded_on"`
	FileSize    int           `json:"file_size"`
	Width       int           `json:"width"`
	Height      int           `json:"height"`
}

type AlertItem struct {
	MGID  bson.ObjectId `json:"-" bson:"_id,omitempty"`
	Key   string        `json:"key"`
	Value string        `json:"message"`
	TTL   float64       `json:"ttl"` // Time To Live
}

type ExpressionItem struct {
	MGID          bson.ObjectId   `json:"-" bson:"_id,omitempty"`
	Expression    string          `json:"expression"`
	ExpressionCmp string          `json:"-" bson:"cid"`
	Usage         int             `json:"count"`
	CatIDs        []bson.ObjectId `json:"-" bson:"cats"` //enfoce using the bson object because this shit can fail...
}

type CategoryItem struct {
	MGID bson.ObjectId `json:"-" bson:"_id,omitempty"`
	Name string        `json:"name"`
}

type User struct {
	MGID     bson.ObjectId `json:"-" bson:"_id,omitempty"`
	Email    string        `json:"-"`
	Username string        `json:"-"`
	Password string        `json:"-"`
	Updated  time.Time     `json:"-"`
	Joined   time.Time     `json:"-"`
}

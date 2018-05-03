package database

import (
	"fmt"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

func (mg *mongo) GetUpload(id bson.ObjectId) Upload {
	session, c := mg.collection(UPLOADS_COL)
	defer session.Close()

	var up Upload
	c.Find(bson.M{"_id": id}).One(&up)

	return up
}

func (mg *mongo) AddUpload(item Upload) Upload {
	session, c := mg.collection(UPLOADS_COL)
	defer session.Close()

	item.MGID = bson.NewObjectId() // needed
	c.Insert(item)

	return item
}

func (mg *mongo) IncUploadViews(id bson.ObjectId) error {
	session, c := mg.collection(UPLOADS_COL)
	defer session.Close()

	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"views": 1}},
		ReturnNew: false,
	}
	_, err := c.Find(bson.M{"_id": id}).Apply(change, nil)
	if err != nil {
		return err
	}

	return nil
}

func (mg *mongo) OrderUploads(mode string, limit int) []Upload {
	items := make([]Upload, limit)

	session, c := mg.collection(UPLOADS_COL)
	defer session.Close()

	c.Find(bson.M{}).Sort(fmt.Sprintf("%sviews", mode)).Limit(limit).All(&items)

	return items
}

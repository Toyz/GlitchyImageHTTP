package database

import (
	"fmt"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

func (mg *mongo) SetImageInfo(upload ArtItem) ArtItem {
	session, c := mg.collection(ARTIDS_COL)
	defer session.Close()

	index := mgo.Index{
		Key:        []string{"id", "filename"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	c.EnsureIndex(index)

	upload.MGID = bson.NewObjectId()
	err := c.Insert(upload)
	if err != nil {
		return ArtItem{}
	}

	return upload
}

func (mg *mongo) GetImageInfo(id bson.ObjectId) ArtItem {
	session, c := mg.collection(ARTIDS_COL)
	defer session.Close()

	var image ArtItem
	c.Find(bson.M{"_id": id}).One(&image)

	if len(image.FileName) <= 0 {
		return ArtItem{}
	}

	return image
}

func (mg *mongo) GetArtByOrder(mode string, page, limit int) []ArtItem {
	items := make([]ArtItem, limit)

	session, c := mg.collection(ARTIDS_COL)
	defer session.Close()

	c.Find(bson.M{}).Sort(fmt.Sprintf("%sviews", mode)).Limit(limit).Skip(page * limit).All(&items)

	return items
}

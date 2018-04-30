package database

import (
	"errors"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

func (mg *mongo) WriteUploadInfo(upload *ArtItem) error {
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

	err := c.Insert(upload)
	if err != nil {
		return err
	}

	return nil
}

func (mg *mongo) GetUploadInfo(id string) (error, ArtItem) {
	session, c := mg.collection(ARTIDS_COL)
	defer session.Close()

	var image ArtItem
	c.Find(bson.M{"id": id}).One(&image)

	if len(image.FullPath) <= 0 {
		return errors.New("item doesn't exist"), ArtItem{}
	}

	return nil, image
}

func (mg *mongo) UploadInfoUpdateViews(art ArtItem) error {
	session, c := mg.collection(ARTIDS_COL)
	defer session.Close()

	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"views": 1}},
		ReturnNew: false,
	}
	_, err := c.Find(bson.M{"id": art.ID}).Apply(change, &art)
	if err != nil {
		return err
	}

	return nil
}

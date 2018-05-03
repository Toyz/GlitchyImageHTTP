package database

import "github.com/globalsign/mgo/bson"

func (mg *mongo) GetCategory(id bson.ObjectId) CategoryItem {
	session, c := mg.collection(CATEGORY_COL)
	defer session.Close()

	var item CategoryItem
	c.Find(bson.M{"_id": id}).One(&item)
	return item
}

func (mg *mongo) GetCategoryByName(name string) CategoryItem {
	session, c := mg.collection(CATEGORY_COL)
	defer session.Close()

	var item CategoryItem
	c.Find(bson.M{"name": name}).One(&item)
	return item
}

func (mg *mongo) AddCategory(item CategoryItem) CategoryItem {
	session, c := mg.collection(CATEGORY_COL)
	defer session.Close()

	item.MGID = bson.NewObjectId()
	c.Insert(item)
	return item
}

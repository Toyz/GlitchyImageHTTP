package database

import (
	"fmt"
	"strings"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

func (mg *mongo) UpdateExpressionUsage(id bson.ObjectId) ExpressionItem {
	session, c := mg.collection(EXPRESSION_COL)
	defer session.Close()

	exp := mg.GetExpression(id)

	if len(exp.ExpressionCmp) > 0 {
		change := mgo.Change{
			Update:    bson.M{"$inc": bson.M{"usage": 1}},
			ReturnNew: false,
		}
		c.Find(bson.M{"_id": id}).Apply(change, &exp)
		exp.Usage = exp.Usage + 1
		return exp
	}

	return ExpressionItem{}
}

func (mg *mongo) AddExpression(item ExpressionItem) ExpressionItem {
	expID := strings.Replace(item.Expression, " ", "", -1)
	item.ExpressionCmp = expID
	item.Usage = 1 // default to 1 because your the one to use it derp...

	session, c := mg.collection(EXPRESSION_COL)
	defer session.Close()

	index := mgo.Index{
		Key:        []string{"cid"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	c.EnsureIndex(index)

	item.MGID = bson.NewObjectId()
	c.Insert(item)

	return item
}

func (mg *mongo) GetExpression(id bson.ObjectId) ExpressionItem {
	session, c := mg.collection(EXPRESSION_COL)
	defer session.Close()

	var exp ExpressionItem
	c.Find(bson.M{"_id": id}).One(&exp)

	if len(exp.Expression) > 0 {
		return exp
	}

	return ExpressionItem{}
}

func (mg *mongo) GetExpressionByName(name string) ExpressionItem {
	name = strings.TrimSpace(name)
	name = strings.Replace(name, " ", "", -1)

	session, c := mg.collection(EXPRESSION_COL)
	defer session.Close()

	var exp ExpressionItem
	c.Find(bson.M{"cid": name}).One(&exp)

	if len(exp.Expression) > 0 {
		return exp
	}

	return ExpressionItem{}
}

func (mg *mongo) OrderExpression(mode string, limit int) []ExpressionItem {
	items := make([]ExpressionItem, limit)

	session, c := mg.collection(EXPRESSION_COL)
	defer session.Close()

	c.Find(bson.M{}).Sort(fmt.Sprintf("%susage", mode)).Limit(limit).All(&items)

	return items
}

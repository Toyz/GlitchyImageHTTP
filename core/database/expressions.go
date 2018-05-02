package database

import (
	"fmt"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

func (mg *mongo) UpdateExpression(expression string) ExpressionItem {
	session, c := mg.collection(EXPRESSION_COL)
	defer session.Close()

	exp := mg.GetExpression(expression)

	if len(exp.Expression) > 0 {
		change := mgo.Change{
			Update:    bson.M{"$inc": bson.M{"usage": 1}},
			ReturnNew: false,
		}
		c.Find(bson.M{"expression": exp.Expression}).Apply(change, &exp)
		exp.Usage = exp.Usage + 1
		return exp
	}

	exp = ExpressionItem{
		expression, 1,
	}

	mg.InsertExpression(exp)
	return exp
}

func (mg *mongo) InsertExpression(item ExpressionItem) {
	session, c := mg.collection(EXPRESSION_COL)
	defer session.Close()

	index := mgo.Index{
		Key:        []string{"expression"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	c.EnsureIndex(index)

	c.Insert(item)
}

func (mg *mongo) GetExpression(expression string) ExpressionItem {
	session, c := mg.collection(EXPRESSION_COL)
	defer session.Close()

	index := mgo.Index{
		Key:        []string{"expression"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	c.EnsureIndex(index)

	var exp ExpressionItem
	c.Find(bson.M{"expression": expression}).One(&exp)

	if len(exp.Expression) > 0 {
		change := mgo.Change{
			Update:    bson.M{"$inc": bson.M{"usage": 1}},
			ReturnNew: false,
		}
		c.Find(bson.M{"expression": exp.Usage}).Apply(change, &exp)
		exp.Usage = exp.Usage + 1
		return exp
	}

	return ExpressionItem{}
}

func (mg *mongo) GetMostUsedExpression(mode string, limit int) []ExpressionItem {
	items := make([]ExpressionItem, limit)

	session, c := mg.collection(EXPRESSION_COL)
	defer session.Close()

	index := mgo.Index{
		Key:        []string{"expression"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	c.EnsureIndex(index)

	c.Find(bson.M{}).Sort(fmt.Sprintf("%susage", mode)).Limit(limit).All(&items)

	return items
}

package database

import (
	"fmt"
	"strings"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

func (mg *mongo) UpdateExpression(expression string) ExpressionItem {
	expression = strings.TrimSpace(expression)
	expID := strings.Replace(expression, " ", "", -1)

	session, c := mg.collection(EXPRESSION_COL)
	defer session.Close()

	exp := mg.GetExpression(expression)

	if len(exp.ExpressionCmp) > 0 {
		change := mgo.Change{
			Update:    bson.M{"$inc": bson.M{"usage": 1}},
			ReturnNew: false,
		}
		c.Find(bson.M{"cid": expID}).Apply(change, &exp)
		exp.Usage = exp.Usage + 1
		return exp
	}

	exp = ExpressionItem{
		Expression:    expression,
		Usage:         1,
		ExpressionCmp: expID,
		MGID:          bson.NewObjectId(),
	}

	mg.InsertExpression(exp)
	return exp
}

func (mg *mongo) InsertExpression(item ExpressionItem) {
	expID := strings.Replace(item.Expression, " ", "", -1)
	item.ExpressionCmp = expID

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

	c.Insert(item)
}

func (mg *mongo) GetExpression(expression string) ExpressionItem {
	expression = strings.TrimSpace(expression)
	expID := strings.Replace(expression, " ", "", -1)

	session, c := mg.collection(EXPRESSION_COL)
	defer session.Close()

	var exp ExpressionItem
	c.Find(bson.M{"cid": expID}).One(&exp)

	if len(exp.Expression) > 0 {
		return exp
	}

	return ExpressionItem{}
}

func (mg *mongo) GetExpressionsByOrder(mode string, limit int) []ExpressionItem {
	items := make([]ExpressionItem, limit)

	session, c := mg.collection(EXPRESSION_COL)
	defer session.Close()

	c.Find(bson.M{}).Sort(fmt.Sprintf("%susage", mode)).Limit(limit).All(&items)

	return items
}

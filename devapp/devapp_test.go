/* В devapp мы разрабатываем наш квери билдер и тесты к нему/

Тесты и структуры(devapp.go,devapp_test.go) копируем в sampleapp из devapp,
кверибилдер(devapp_gen.go) для sampleapp генерируем из структур для проверки тому,
что наш сгенерированный код соответствует коду написанному в devapp_gen.go и проходит наши тесты
*/

package devapp

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/swelf19/tsqb/qtypes"
)

type QueryBuilderTestSuite struct {
	suite.Suite
}

func (suite *QueryBuilderTestSuite) SetupTest() {

}

func (suite *QueryBuilderTestSuite) TestBuildQuery() {
	bu := NewUserBuilder()
	expected := "select id,username,last_log from users"
	actual := bu.Select().SQL()
	suite.Equal(expected, actual)

	bs := NewStoreBuilder()
	expected = "select id,storename from stores"
	actual = bs.Select().SQL()
	suite.Equal(expected, actual)

}

func (suite *QueryBuilderTestSuite) TestConditionQuery() {
	b := NewUserBuilder()
	expected := "select id,username,last_log from users where users.id = $1 and users.id = $2"
	actual := b.Select().Where(b.CondEqID(1), b.CondEqID(2)).SQL()
	suite.Equal(expected, actual)
	expectedParams := []interface{}{
		1,
		2,
	}
	suite.Equal(expectedParams, b.GetStmtParams())

	// b = NewUserBuilder()
	expected = "select id,username,last_log from users where users.id = $1 and (users.username = $2 or users.last_log = $3)"
	actual = b.Select().Where(
		b.CondEqID(1),
		b.ComposeOr(
			b.CondEqUserName("swelf"),
			b.CondEqLastLog("admin"),
		),
	).SQL()
	suite.Equal(expected, actual)
	expectedParams = []interface{}{
		1,
		"swelf",
		"admin",
	}
	suite.Equal(expectedParams, b.GetStmtParams())

	expected = "select id,username,last_log from users where users.id = $1 and (users.username = $2 or users.username = $3 or (users.username = $4 and users.username = $5) or (users.username = $6 and users.username = $7))"
	actual = b.Select().Where(
		b.CondEqID(1),
		b.ComposeOr(
			b.CondEqUserName("swelf"),
			b.CondEqUserName("admin"),
			b.ComposeAnd(
				b.CondEqUserName("lalala"),
				b.CondEqUserName("lalala2"),
			),
			b.ComposeAnd(
				b.CondEqUserName("lalala3"),
				b.CondEqUserName("lalala4"),
			),
		),
	).SQL()
	suite.Equal(expected, actual)
	expectedParams = []interface{}{
		1,
		"swelf",
		"admin",
		"lalala",
		"lalala2",
		"lalala3",
		"lalala4",
	}
	suite.Equal(expectedParams, b.GetStmtParams())
}

func (suite *QueryBuilderTestSuite) TestLimitQuery() {
	b := NewUserBuilder()
	expected := "select id,username,last_log from users limit 10"
	actual := b.Select().Limit(10).SQL()
	suite.Equal(expected, actual)
}

func (suite *QueryBuilderTestSuite) TestOffsetQuery() {
	b := NewUserBuilder()
	expected := "select id,username,last_log from users offset 10"
	actual := b.Select().Offset(10).SQL()
	suite.Equal(expected, actual)
}

func (suite *QueryBuilderTestSuite) TestOrderingQuery() {
	b := NewUserBuilder()
	expected := "select id,username,last_log from users order by id, username desc"
	actual := b.Select().OrderBy(UserFieldID).OrderByDesc(UserFieldUserName).SQL()
	suite.Equal(expected, actual)
}

func (suite *QueryBuilderTestSuite) TestCompleteQuery() {
	b := NewUserBuilder()
	expected := "select id,username,last_log from users where users.id = $1 order by id, username desc offset 15 limit 10"
	actual := b.Select().Where(b.CondEqID(1)).Limit(10).Offset(15).OrderBy(UserFieldID).OrderByDesc(UserFieldUserName).SQL()
	suite.Equal(expected, actual)
}

func (suite *QueryBuilderTestSuite) TestRenderCondition() {
	name1 := UserCondition{
		Table: "test",
		Field: "name",
		Func:  "=",
		Value: "$1",
	}
	name2 := UserCondition{
		Table: "test",
		Field: "name",
		Func:  ">",
		Value: "$2",
	}
	nodename := UserCondNode{
		Conditions: []UserCondition{name1, name2},
		WhereLink:  qtypes.WhereAnd,
		Nodes:      []UserCondNode{},
	}
	expected := "(test.name = $1 and test.name > $2)"
	actual := nodename.String()
	suite.Equal(expected, actual)
}

func (suite *QueryBuilderTestSuite) TestInsertQuery() {
	b := NewUserBuilder()
	u := User{
		ID:       1,
		UserName: "lala",
		LastLog:  "today",
	}

	expected := "insert into users(id,username,last_log) values($1,$2,$3) returning id"
	actual := b.Insert(u).ReturningID().SQL()
	suite.Equal(expected, actual)
	insertParams := []interface{}{}
	for _, i := range b.InsertParams {
		insertParams = append(insertParams, i.Value)
	}
	expectedParams := []interface{}{
		1,
		"lala",
		"today",
	}
	suite.Equal(expectedParams, insertParams)

	u1 := User{
		ID:       0,
		UserName: "lala",
		LastLog:  "today",
	}
	expected = "insert into users(username,last_log) values($1,$2) returning id"
	actual = b.Insert(u1).ReturningID().SQL()
	suite.Equal(expected, actual)
	insertParams = []interface{}{}
	expectedParams = []interface{}{
		"lala",
		"today",
	}
	for _, i := range b.InsertParams {
		insertParams = append(insertParams, i.Value)
	}
	suite.Equal(expectedParams, insertParams)
	// c := b.Insert(u).ReturningID()
}

func (suite *QueryBuilderTestSuite) TestUpdateQuery() {
	b := NewUserBuilder()
	u := User{
		ID:       1,
		UserName: "lala",
		LastLog:  "today",
	}

	expected := "update users set username = $1, last_log = $2"
	actual := b.Update(u).SQL()
	suite.Equal(expected, actual)
	expectedParams := []interface{}{
		"lala",
		"today",
	}
	suite.Equal(expectedParams, b.GetStmtParams())

	expected = "update users set username = $2, last_log = $3 where users.id = $1"
	actual = b.Update(u).Where(b.CondEqID(u.ID)).SQL()
	suite.Equal(expected, actual)
	expectedParams = []interface{}{
		1,
		"lala",
		"today",
	}
	suite.Equal(expectedParams, b.GetStmtParams())

	// u1 := User{
	// 	ID:       0,
	// 	UserName: "lala",
	// 	LastLog:  "today",
	// }
	// expected = "insert into users(username,last_log) values($1,$2) returning id"
	// actual = b.Insert(u1).ReturningID().SQL()
	// suite.Equal(expected, actual)
	// c := b.Insert(u).ReturningID()
}

func TestTemplateProcessor(t *testing.T) {
	suite.Run(t, new(QueryBuilderTestSuite))
}

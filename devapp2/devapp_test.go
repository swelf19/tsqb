/* В devapp мы разрабатываем наш квери билдер и тесты к нему/

Тесты и структуры(devapp.go,devapp_test.go) копируем в sampleapp из devapp,
кверибилдер(devapp_gen.go) для sampleapp генерируем из структур для проверки тому,
что наш сгенерированный код соответствует коду написанному в devapp_gen.go и проходит наши тесты
*/

package devapp2

import (
	"fmt"
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
	bu := NewUserSelectBuilder()
	expected := "select id,username,last_log from users"
	actual := bu.Build().SQL()
	suite.Equal(expected, actual)
}

func (suite *QueryBuilderTestSuite) TestConditionQuery() {
	b := NewUserSelectBuilder()
	expected := "select id,username,last_log from users where users.id = $1 and users.id = $2"
	actualQuery := b.Where(b.CondEqID(1), b.CondEqID(2)).Build()
	fmt.Println(actualQuery.SQL())
	suite.Equal(expected, actualQuery.SQL())
	expectedParams := []interface{}{
		1,
		2,
	}
	suite.Equal(expectedParams, actualQuery.stmtParams)

	expected = "select id,username,last_log from users where users.id = $1 and (users.username = $2 or users.last_log = $3)"
	actualQuery = b.Where(
		b.CondEqID(1),
		b.ComposeOr(
			b.CondEqUserName("swelf"),
			b.CondEqLastLog("admin"),
		),
	).Build()
	suite.Equal(expected, actualQuery.SQL())
	expectedParams = []interface{}{
		1,
		"swelf",
		"admin",
	}
	suite.Equal(expectedParams, actualQuery.stmtParams)

	expected = "select id,username,last_log from users where users.id = $1 and (users.username = $2 or users.username = $3 or (users.username = $4 and users.username = $5) or (users.username = $6 and users.username = $7))"
	actualQuery = b.Where(
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
	).Build()
	suite.Equal(expected, actualQuery.SQL())
	expectedParams = []interface{}{
		1,
		"swelf",
		"admin",
		"lalala",
		"lalala2",
		"lalala3",
		"lalala4",
	}
	suite.Equal(expectedParams, actualQuery.stmtParams)
}

func (suite *QueryBuilderTestSuite) TestLimitQuery() {
	b := NewUserSelectBuilder()
	expected := "select id,username,last_log from users limit 10"
	actual := b.Limit(10).Build().SQL()
	suite.Equal(expected, actual)
}

func (suite *QueryBuilderTestSuite) TestOffsetQuery() {
	b := NewUserSelectBuilder()
	expected := "select id,username,last_log from users offset 10"
	actual := b.Offset(10).Build().SQL()
	suite.Equal(expected, actual)
}

func (suite *QueryBuilderTestSuite) TestOrderingQuery() {
	b := NewUserSelectBuilder()
	expected := "select id,username,last_log from users order by id, username desc"
	actual := b.OrderBy(UserFieldID).OrderByDesc(UserFieldUserName).Build().SQL()
	suite.Equal(expected, actual)
}

func (suite *QueryBuilderTestSuite) TestCompleteQuery() {
	b := NewUserSelectBuilder()
	expected := "select id,username,last_log from users where users.id = $1 order by id, username desc offset 15 limit 10"
	actual := b.Where(b.CondEqID(1)).Limit(10).Offset(15).OrderBy(UserFieldID).OrderByDesc(UserFieldUserName).Build().SQL()
	suite.Equal(expected, actual)
}

func (suite *QueryBuilderTestSuite) TestRenderCondition() {
	name1 := UserCondition{
		Field: "name",
		Func:  "=",
	}
	name2 := UserCondition{
		Field: "name",
		Func:  ">",
	}
	nodename := UserCondNode{
		Conditions:    []UserCondition{name1, name2},
		ComposeMethod: qtypes.WhereAnd,
		Nodes:         []UserCondNode{},
	}
	expected := "(test.name = $1 and test.name > $2)"
	actual := nodename.BuildCond(1, "test").sqlString
	suite.Equal(expected, actual)
}

func (suite *QueryBuilderTestSuite) TestInsertQuery() {
	b := NewUserInsertBuilder()
	u := User{
		ID:       1,
		UserName: "lala",
		LastLog:  "today",
	}

	expected := "insert into users(id,username,last_log) values($1,$2,$3) returning id"
	actual := b.Insert(u).Build().SQL()
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
	actual = b.Insert(u1).Build().SQL()
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
}

func (suite *QueryBuilderTestSuite) TestUpdateQuery() {
	b := NewUserUpdateBuilder()
	u := User{
		ID:       1,
		UserName: "lala",
		LastLog:  "today",
	}

	expected := "update users set username = $2, last_log = $3 where users.id = $1"
	actualQuery := b.UpdateUserName("lala").UpdateLastLog("tomorow").Where(b.CondEqID(1)).Build()
	suite.Equal(expected, actualQuery.SQL())
	expectedParams := []interface{}{
		1,
		"lala",
		"tomorow",
	}
	suite.Equal(expectedParams, actualQuery.getUpdateStmtParams())
	b = NewUserUpdateBuilder()
	expected = "update users set username = $2, last_log = $3 where users.id = $1"
	actualQuery = b.UpdateAllFields(u).Build()
	suite.Equal(expected, actualQuery.SQL())
}

func TestTemplateProcessor(t *testing.T) {
	suite.Run(t, new(QueryBuilderTestSuite))
}

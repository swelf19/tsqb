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
	suite.Equal(2, len(b.SelectParams))

	// b = NewUserBuilder()
	expected = "select id,username,last_log from users where users.id = $1 and (users.username = $2 or users.username = $3)"
	actual = b.Select().Where(
		b.CondEqID(1),
		b.ComposeOr(
			b.CondEqUserName("swelf"),
			b.CondEqUserName("admin"),
		),
	).SQL()
	suite.Equal(expected, actual)
	suite.Equal(3, len(b.SelectParams))

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
				b.CondEqUserName("lalala"),
				b.CondEqUserName("lalala2"),
			),
		),
	).SQL()
	suite.Equal(expected, actual)
	suite.Equal(7, len(b.SelectParams))
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
	expected := "insert into users(id,username,last_log) values($1,$2,$3)"
	actual := b.Insert().Values(u).SQL()
	suite.Equal(expected, actual)
}

func TestTemplateProcessor(t *testing.T) {
	suite.Run(t, new(QueryBuilderTestSuite))
}

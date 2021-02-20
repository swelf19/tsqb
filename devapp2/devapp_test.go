/* В devapp мы разрабатываем наш квери билдер и тесты к нему/

 */

package devapp2

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/swelf19/tsqb/qfuncs"
)

type QueryBuilderTestSuite struct {
	suite.Suite
}

func (suite *QueryBuilderTestSuite) SetupTest() {

}

func (suite *QueryBuilderTestSuite) equalStmtParams(expected []interface{}, actual []interface{}) bool {
	exp := []interface{}{}
	for _, e := range expected {
		exp = append(exp, e)
	}
	return suite.Equal(expected, actual)
}

func (suite *QueryBuilderTestSuite) TestBuildQuery() {
	bu := Select().User()
	expected := "select users.id, users.username, users.last_log from users"
	actual := bu.Build().SQL()
	suite.Equal(expected, actual)
}

func (suite *QueryBuilderTestSuite) TestLimitQuery() {
	bu := Select().User()
	expected := "select users.id, users.username, users.last_log from users limit 10"
	actual := bu.Limit(10).Build().SQL()
	suite.Equal(expected, actual)
}

func (suite *QueryBuilderTestSuite) TestOffsetQuery() {
	bu := Select().User()
	expected := "select users.id, users.username, users.last_log from users offset 10"
	actual := bu.Offset(10).Build().SQL()
	suite.Equal(expected, actual)

	expected = "select users.id, users.username, users.last_log from users offset 10 limit 20"
	actual = bu.Limit(20).Offset(10).Build().SQL()
	suite.Equal(expected, actual)
}

func (suite *QueryBuilderTestSuite) TestConditionQuery() {
	b := Select().User()
	expected := "select users.id, users.username, users.last_log from users where (users.id = $1 and users.id = $2)"
	actualQuery := b.Where(
		b.UserSchema.Fields.ID.Eq(1),
		b.UserSchema.Fields.ID.Eq(2),
	).Build()
	// fmt.Println(actualQuery.SQL())
	suite.Equal(expected, actualQuery.SQL())
	expectedParams := []interface{}{
		1,
		2,
	}
	suite.equalStmtParams(expectedParams, actualQuery.whereClause.StmtParams)

	expected = "select users.id, users.username, users.last_log from users where (users.id = $1 and (users.username = $2 or users.last_log = $3))"
	actualQuery = b.Where(
		b.UserSchema.Fields.ID.Eq(1),
		qfuncs.ComposeOr(
			b.UserSchema.Fields.UserName.Eq("swelf"),
			b.UserSchema.Fields.LastLog.Eq("admin"),
		),
	).Build()
	suite.Equal(expected, actualQuery.SQL())
	expectedParams = []interface{}{
		1,
		"swelf",
		"admin",
	}
	suite.equalStmtParams(expectedParams, actualQuery.whereClause.StmtParams)

	expected = "select users.id, users.username, users.last_log from users where (users.id = $1 and (users.username = $2 or users.username = $3 or (users.username = $4 and users.username = $5) or (users.username = $6 and users.username = $7)))"
	actualQuery = b.Where(
		b.UserSchema.Fields.ID.Eq(1),
		qfuncs.ComposeOr(
			b.UserSchema.Fields.UserName.Eq("swelf"),
			b.UserSchema.Fields.UserName.Eq("admin"),
			qfuncs.ComposeAnd(
				b.UserSchema.Fields.UserName.Eq("lalala"),
				b.UserSchema.Fields.UserName.Eq("lalala2"),
			),
			qfuncs.ComposeAnd(
				b.UserSchema.Fields.UserName.Eq("lalala3"),
				b.UserSchema.Fields.UserName.Eq("lalala4"),
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
	suite.equalStmtParams(expectedParams, actualQuery.whereClause.StmtParams)

	expected = "select users.id, users.username, users.last_log from users where users.username IN ($1,$2,$3)"
	actualQuery = b.Where(
		b.UserSchema.Fields.UserName.In("swelf", "admin", "user"),
	).Build()
	suite.Equal(expected, actualQuery.SQL())
	expectedParams = []interface{}{
		"swelf",
		"admin",
		"user",
	}
	suite.equalStmtParams(expectedParams, actualQuery.whereClause.StmtParams)

}

// func (suite *QueryBuilderTestSuite) TestLimitQuery() {
// 	b := Select().User()
// 	expected := "select users.id,users.username,users.last_log from users limit 10"
// 	actual := b.Limit(10).Build().SQL()
// 	suite.Equal(expected, actual)
// }

// func (suite *QueryBuilderTestSuite) TestOffsetQuery() {
// 	b := NewUserSelectBuilder()
// 	expected := "select users.id,users.username,users.last_log from users offset 10"
// 	actual := b.Offset(10).Build().SQL()
// 	suite.Equal(expected, actual)
// }

// func (suite *QueryBuilderTestSuite) TestOrderingQuery() {
// 	b := NewUserSelectBuilder()
// 	expected := "select users.id,users.username,users.last_log from users order by users.id, users.username desc"
// 	actual := b.OrderBy(UserFieldID).OrderByDesc(UserFieldUserName).Build().SQL()
// 	suite.Equal(expected, actual)
// }

// func (suite *QueryBuilderTestSuite) TestCompleteQuery() {
// 	b := NewUserSelectBuilder()
// 	expected := "select users.id,users.username,users.last_log from users where users.id = $1 order by users.id, users.username desc offset 15 limit 10"
// 	actual := b.Where(b.Cond.User.ID.Eq(1)).Limit(10).Offset(15).OrderBy(UserFieldID).OrderByDesc(UserFieldUserName).Build().SQL()
// 	suite.Equal(expected, actual)
// }

func (suite *QueryBuilderTestSuite) TestInsertQuery() {
	u := User{
		ID:       1,
		UserName: "lala",
		LastLog:  "today",
	}
	b := Insert().User(u)

	expected := "insert into users(id, username, last_log) values($1, $2, $3) returning id"
	actualQuery := b.Build()
	suite.Equal(expected, actualQuery.SQL())

	expectedParams := []interface{}{
		1,
		"lala",
		"today",
	}
	suite.Equal(expectedParams, actualQuery.params)

	u1 := User{
		ID:       0,
		UserName: "lala",
		LastLog:  "today",
	}
	expected = "insert into users(username, last_log) values($1, $2) returning id"
	actualQuery = Insert().User(u1).Build()
	suite.Equal(expected, actualQuery.SQL())
	expectedParams = []interface{}{
		"lala",
		"today",
	}
	suite.Equal(expectedParams, actualQuery.params)
}

func (suite *QueryBuilderTestSuite) TestUpdateQuery() {
	b := Update().User()
	expected := "update users set username = $2, last_log = $3 where users.id = $1"
	actualQuery := b.SetUserName("lala").SetLastLog("tomorow").Where(b.UserSchema.Fields.ID.Eq(1)).Build()
	suite.Equal(expected, actualQuery.SQL())
	expectedParams := []interface{}{
		1,
		"lala",
		"tomorow",
	}
	suite.Equal(expectedParams, actualQuery.params)

	u := User{
		ID:       1,
		UserName: "lala",
		LastLog:  "today",
	}
	b = Update().User()
	expected = "update users set username = $2, last_log = $3 where users.id = $1"
	actualQuery = b.SetAllFields(u).Build()
	suite.Equal(expected, actualQuery.SQL())
	expectedParams = []interface{}{
		1,
		"lala",
		"today",
	}
	suite.Equal(expectedParams, actualQuery.params)

	b = Update().User()
	expected = "update users set last_log = $1"
	actualQuery = b.SetLastLog("tomorow").Build()
	suite.Equal(expected, actualQuery.SQL())
	suite.Equal([]interface{}{"tomorow"}, actualQuery.params)
}

func (suite *QueryBuilderTestSuite) TestDeleteQuery() {
	b := Delete().User()
	expected := "delete from users"
	actualQuery := b.Build()
	suite.Equal(expected, actualQuery.SQL())
	suite.equalStmtParams(nil, actualQuery.whereClause.StmtParams)

	expected = "delete from users where users.id = $1"
	actualQuery = b.Where(b.UserSchema.Fields.ID.Eq(1)).Build()
	suite.Equal(expected, actualQuery.SQL())
	suite.equalStmtParams([]interface{}{1}, actualQuery.whereClause.StmtParams)

}

func TestTemplateProcessor(t *testing.T) {
	suite.Run(t, new(QueryBuilderTestSuite))
}

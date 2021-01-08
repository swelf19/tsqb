package gen

import (
	"fmt"
	"go/format"
	"testing"

	"github.com/kylelemons/godebug/diff"
	"github.com/stretchr/testify/suite"
)

type GenCodeTestSuite struct {
	suite.Suite
}

func (suite *GenCodeTestSuite) SetupTest() {

}

var CONDITION_EXPECTED = `type UserCondition struct {
	Table UserTableNameType
	Field UserFields
	Func  qtypes.EqualConditions
	Value string
}
func (c UserCondition) String() string {
	subQueryTemplate, _ := template.New("subquery").Parse("{{.Table}}.{{.Field}} {{.Func}} {{.Value}}")
	queryBuf := new(bytes.Buffer)
	_ = subQueryTemplate.Execute(queryBuf, map[string]string{
		"Table": string(c.Table),
		"Field": string(c.Field),
		"Func":  string(c.Func),
		"Value": c.Value,
	})
	return queryBuf.String()
}
type UserCondNode struct {
	Conditions []UserCondition
	WhereLink  qtypes.WhereLinks
	Nodes      []UserCondNode
	Not        bool
}
func (cn UserCondNode) String() string {
	conditions := []string{}
	for _, c := range cn.Conditions {
		conditions = append(conditions, c.String())
	}
	for _, n := range cn.Nodes {
		conditions = append(conditions, n.String())
	}
	condTemplate := "%s"
	if len(conditions) > 1 {
		condTemplate = "(%s)"
	}
	if cn.Not {
		condTemplate = "not " + condTemplate
	}
	return fmt.Sprintf(condTemplate, strings.Join(conditions, fmt.Sprintf(" %s ", cn.WhereLink)))
}`

var ORDERING_EXPECTED = `
	type UserOrderCond struct {
		Field     UserFields
		Direction qtypes.OrderDirection
	}

	func (o UserOrderCond) String() string {
		if o.Direction == qtypes.OrderAsc {
			return string(o.Field)
		} else {
			return fmt.Sprintf("%s %s", string(o.Field), string(o.Direction))
		}
	}`

var BUILDER_STRUCT_EXPECTED = `type UserBuilder struct {
	queryType       qtypes.QueryType
	Fields          []UserFields
	TableName       UserTableNameType
	Conditions      []UserCondNode
	LimitValue      int
	OffsetValue     int
	SelectParams    []interface{}
	InsertParams    map[string]interface{}
	lastPlaceHolder int
	order           []UserOrderCond
	connection      qtypes.DBConnection
}`

var BUILDER_RESET_METHOD = `func (b *UserBuilder) ResetBuilder() {
	b.Fields = []UserFields{UserFieldID,UserFieldUserName}
	b.TableName = UserTableName
	b.Conditions = []UserCondNode{}
	b.SelectParams = []interface{}{}
	b.InsertParams = map[string]interface{}{}
	b.lastPlaceHolder = 0
	b.LimitValue = 0
	b.OffsetValue = 0
	b.order = []UserOrderCond{}
}`

var ADDITIONAL_STRUCT_METHODS = `func (u *User) TSQBScanner() []interface{} {
	return []interface{}{
		&u.ID, &u.UserName,
	}
}

func (u User) TSQBSaver() map[string]interface{} {
	return map[string]interface{}{
		"id":       u.ID,
		"username": u.UserName,
	}
}`

var NEW_BUILDER_FUNCTION = `func NewUserBuilder() *UserBuilder {
	b := UserBuilder{}
	b.ResetBuilder()
	return &b
}`

var BASE_QUERY_METHODS = `func (b *UserBuilder) Where(conditions ...UserCondNode) *UserBuilder {
	b.Conditions = conditions
	return b
}

func (b *UserBuilder) ComposeAnd(conditions ...UserCondNode) UserCondNode {
	return b.compose(qtypes.WhereAnd, conditions...)
}

func (b *UserBuilder) ComposeOr(conditions ...UserCondNode) UserCondNode {
	return b.compose(qtypes.WhereOr, conditions...)
}

func (b *UserBuilder) compose(w qtypes.WhereLinks, conditions ...UserCondNode) UserCondNode {
	cn := UserCondNode{
		Conditions: []UserCondition{},
		Nodes:      []UserCondNode{},
		WhereLink:  w,
	}
	for _, node := range conditions {
		if len(node.Conditions) == 1 {
			cn.Conditions = append(cn.Conditions, node.Conditions...)
		} else {
			cn.Nodes = append(cn.Nodes, node)
		}
	}
	return cn
}



func (b *UserBuilder) OrderByDesc(field UserFields) *UserBuilder {
	b.order = append(b.order, UserOrderCond{Field: field, Direction: qtypes.OrderDesc})
	return b
}

func (b *UserBuilder) OrderBy(field UserFields) *UserBuilder {
	b.order = append(b.order, UserOrderCond{Field: field, Direction: qtypes.OrderAsc})
	return b
}`

var COMPRATION_HELPER = `func (b *UserBuilder) CondEqID(compareTo int) UserCondNode {
	b.lastPlaceHolder = b.lastPlaceHolder + 1
	placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
	c := UserCondition{
		Table: b.TableName,
		Field: UserFieldID,
		Func:  qtypes.Equal,
		Value: placeholder,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	b.SelectParams = append(b.SelectParams, compareTo)
	return cn
}`

var BUILDER_FETCH_METHODS = `func (b *UserBuilder) SetDBConnection(connection qtypes.DBConnection) *UserBuilder{
	b.connection = connection
	return b
}

func (b *UserBuilder) Fetch() ([]User, error) {
	if b.connection == nil {
		return nil, errors.New("Required to setup (SetDBConnection) connection before fetching")
	}
	values := []User{}
	rows, err := b.connection.Query(context.Background(), b.SQL(), b.SelectParams...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		v := User{}
		err = rows.Scan(v.TSQBScanner()...)
		if err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return values, nil
}`

//FormatAndCompare - Приводит код к форматированныому виду при помощи go/format,
// после чего сравнивает методом suite.Equal
func (suite *GenCodeTestSuite) FormatAndCompare(expected, actual string) bool {
	a1, err := format.Source([]byte(expected))

	suite.Assert().NoError(err)
	a2, err := format.Source([]byte(actual))
	suite.Assert().NoError(err)
	// fmt.Println(actual)
	// fmt.Println(string(a2))
	fmt.Println(diff.Diff(string(a1), string(a2)))
	return suite.Equal(a1, a2)
}

func (suite *GenCodeTestSuite) TestGenFields() {

	st := StructMeta{
		StructName: "User",
		TableName:  "users",
		Fields: []StructFieldMeta{
			{
				FieldName:    "ID",
				Type:         "int",
				SqlFieldName: "id",
			},
			{
				FieldName:    "UserName",
				Type:         "string",
				SqlFieldName: "username",
			},
		},
	}
	expected := `type UserFields string
		var (
			UserFieldID UserFields = "id"
			UserFieldUserName UserFields = "username"
		)`
	actual := st.genFieldConstantsBlock()
	suite.FormatAndCompare(expected, actual)
	expected = `type UserTableNameType string
		var (
			UserTableName UserTableNameType = "users"
		)`
	actual = st.genTableBlock()
	suite.FormatAndCompare(expected, actual)

	suite.FormatAndCompare(CONDITION_EXPECTED, st.genConditions())

	suite.FormatAndCompare(ORDERING_EXPECTED, st.genOrdering())

	suite.FormatAndCompare(BUILDER_STRUCT_EXPECTED, st.genBuilderStruct())

	suite.FormatAndCompare(BUILDER_RESET_METHOD, st.genBuilderResetMethod())

	suite.FormatAndCompare(ADDITIONAL_STRUCT_METHODS, st.genAdditionalStructMethods())

	suite.FormatAndCompare(NEW_BUILDER_FUNCTION, st.genCreateBuilderFunction())

	suite.FormatAndCompare(BASE_QUERY_METHODS, st.genBuilderBaseQueryMethods())

	suite.FormatAndCompare(COMPRATION_HELPER, st.genComprationForField("Eq", "qtypes.Equal", st.Fields[0]))

	suite.FormatAndCompare(BUILDER_FETCH_METHODS, st.genBuilderFetch())
}

func TestCodeGeneration(t *testing.T) {
	suite.Run(t, new(GenCodeTestSuite))
}

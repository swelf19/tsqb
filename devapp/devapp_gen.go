package devapp

import (
	"bytes"
	"errors"

	"fmt"

	"strings"

	"text/template"

	"github.com/swelf19/tsqb/qtypes"
	"golang.org/x/net/context"

	"github.com/swelf19/tsqb/qfuncs"
)

type UserFields string

var (
	UserFieldID       UserFields = "id"
	UserFieldUserName UserFields = "username"
	UserFieldLastLog  UserFields = "last_log"
)

type UserTableNameType string

var (
	UserTableName UserTableNameType = "users"
)

type UserCondition struct {
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
}

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
}

type UserBuilder struct {
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
}

func (b *UserBuilder) ResetBuilder() {
	b.Fields = []UserFields{UserFieldID, UserFieldUserName, UserFieldLastLog}
	b.TableName = UserTableName
	b.Conditions = []UserCondNode{}
	b.SelectParams = []interface{}{}
	b.InsertParams = map[string]interface{}{}
	b.lastPlaceHolder = 0
	b.LimitValue = 0
	b.OffsetValue = 0
	b.order = []UserOrderCond{}
}
func NewUserBuilder() *UserBuilder {
	b := UserBuilder{}
	b.ResetBuilder()
	return &b
}
func (b *UserBuilder) SQL() string {
	switch b.queryType {
	case qtypes.SelectQuery:
		return b.selectQuery()
	case qtypes.InsertQuery:
		return b.insertQuery()
	}
	return "You must init UserBuilder with one of initial method (Select,Insert,Update,Delete)"
}

func (b UserBuilder) getWhereString() string {
	WhereClause := ""
	if len(b.Conditions) > 0 {
		condditions := []string{}
		for _, c := range b.Conditions {
			condditions = append(condditions, c.String())
		}
		WhereClause = "where " + strings.Join(condditions, " and ")
	}
	return WhereClause
}

func (b UserBuilder) getFieldsString() string {
	fields := []string{}
	for _, f := range b.Fields {
		fields = append(fields, string(f))
	}
	return strings.Join(fields, ",")
}

func (b UserBuilder) getLimitString() string {
	limitString := ""
	if b.LimitValue > 0 {
		limitString = fmt.Sprintf("limit %d", b.LimitValue)
	}
	return limitString
}

func (b UserBuilder) getOffsetString() string {
	offsetString := ""
	if b.OffsetValue > 0 {
		offsetString = fmt.Sprintf("offset %d", b.OffsetValue)
	}
	return offsetString
}

func (b UserBuilder) getOrderString() string {
	orderString := ""
	if len(b.order) > 0 {
		orderStrings := []string{}
		for _, o := range b.order {
			orderStrings = append(orderStrings, o.String())
		}
		orderString = fmt.Sprintf("order by %s", strings.Join(orderStrings, ", "))
	}
	return orderString
}

func (b UserBuilder) selectQuery() string {
	fields := b.getFieldsString()
	WhereClause := b.getWhereString()
	limitString := b.getLimitString()
	offsetString := b.getOffsetString()
	orderString := b.getOrderString()
	baseQuery := fmt.Sprintf("select %s from %s", fields, string(b.TableName))
	directives := []string{
		baseQuery,
		WhereClause,
		orderString,
		offsetString,
		limitString,
	}
	directives = qfuncs.RemoveEmpty(directives)
	query := strings.Join(directives, " ")
	return query
}

func (b UserBuilder) getInsertPlaceholders() string {
	plc := 1
	placeholders := []string{}
	for range b.InsertParams {
		placeholders = append(placeholders, fmt.Sprintf("$%d", plc))
		plc = plc + 1
	}
	return strings.Join(placeholders, ",")
}

func (b UserBuilder) insertQuery() string {
	fields := b.getFieldsString()
	subQueryTemplate, _ := template.New("subquery").Parse("insert into {{.TableName}}({{.Fields}}) values({{.Placeholders}})")
	placeholders := b.getInsertPlaceholders()
	queryBuf := new(bytes.Buffer)
	_ = subQueryTemplate.Execute(queryBuf, map[string]string{
		"Fields":       fields,
		"TableName":    string(b.TableName),
		"Placeholders": placeholders,
	})
	query := strings.TrimSpace(queryBuf.String())
	return query
}

func (b *UserBuilder) setQueryType(qt qtypes.QueryType) {
	b.ResetBuilder()
	b.queryType = qt
}

func (b *UserBuilder) Select() *UserBuilder {
	b.setQueryType(qtypes.SelectQuery)
	return b
}

func (b *UserBuilder) Insert() *UserBuilder {
	b.setQueryType(qtypes.InsertQuery)
	return b
}

func (b *UserBuilder) Values(u User) *UserBuilder {
	b.InsertParams = u.TSQBSaver()
	return b
}

func (b *UserBuilder) Limit(limit int) *UserBuilder {
	b.LimitValue = limit
	return b
}

func (b *UserBuilder) Offset(offset int) *UserBuilder {
	b.OffsetValue = offset
	return b
}
func (b *UserBuilder) Where(conditions ...UserCondNode) *UserBuilder {
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
}
func (b *UserBuilder) CondEqID(compareTo int) UserCondNode {
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
}
func (b *UserBuilder) CondGtID(compareTo int) UserCondNode {
	b.lastPlaceHolder = b.lastPlaceHolder + 1
	placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
	c := UserCondition{
		Table: b.TableName,
		Field: UserFieldID,
		Func:  qtypes.GreaterThan,
		Value: placeholder,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	b.SelectParams = append(b.SelectParams, compareTo)
	return cn
}
func (b *UserBuilder) CondGteID(compareTo int) UserCondNode {
	b.lastPlaceHolder = b.lastPlaceHolder + 1
	placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
	c := UserCondition{
		Table: b.TableName,
		Field: UserFieldID,
		Func:  qtypes.GreaterOrEqualThan,
		Value: placeholder,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	b.SelectParams = append(b.SelectParams, compareTo)
	return cn
}
func (b *UserBuilder) CondLtID(compareTo int) UserCondNode {
	b.lastPlaceHolder = b.lastPlaceHolder + 1
	placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
	c := UserCondition{
		Table: b.TableName,
		Field: UserFieldID,
		Func:  qtypes.LessThan,
		Value: placeholder,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	b.SelectParams = append(b.SelectParams, compareTo)
	return cn
}
func (b *UserBuilder) CondLteID(compareTo int) UserCondNode {
	b.lastPlaceHolder = b.lastPlaceHolder + 1
	placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
	c := UserCondition{
		Table: b.TableName,
		Field: UserFieldID,
		Func:  qtypes.LessOrEqualThan,
		Value: placeholder,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	b.SelectParams = append(b.SelectParams, compareTo)
	return cn
}
func (b *UserBuilder) CondNeID(compareTo int) UserCondNode {
	b.lastPlaceHolder = b.lastPlaceHolder + 1
	placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
	c := UserCondition{
		Table: b.TableName,
		Field: UserFieldID,
		Func:  qtypes.NotEqual,
		Value: placeholder,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	b.SelectParams = append(b.SelectParams, compareTo)
	return cn
}
func (b *UserBuilder) CondEqUserName(compareTo string) UserCondNode {
	b.lastPlaceHolder = b.lastPlaceHolder + 1
	placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
	c := UserCondition{
		Table: b.TableName,
		Field: UserFieldUserName,
		Func:  qtypes.Equal,
		Value: placeholder,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	b.SelectParams = append(b.SelectParams, compareTo)
	return cn
}
func (b *UserBuilder) CondGtUserName(compareTo string) UserCondNode {
	b.lastPlaceHolder = b.lastPlaceHolder + 1
	placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
	c := UserCondition{
		Table: b.TableName,
		Field: UserFieldUserName,
		Func:  qtypes.GreaterThan,
		Value: placeholder,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	b.SelectParams = append(b.SelectParams, compareTo)
	return cn
}
func (b *UserBuilder) CondGteUserName(compareTo string) UserCondNode {
	b.lastPlaceHolder = b.lastPlaceHolder + 1
	placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
	c := UserCondition{
		Table: b.TableName,
		Field: UserFieldUserName,
		Func:  qtypes.GreaterOrEqualThan,
		Value: placeholder,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	b.SelectParams = append(b.SelectParams, compareTo)
	return cn
}
func (b *UserBuilder) CondLtUserName(compareTo string) UserCondNode {
	b.lastPlaceHolder = b.lastPlaceHolder + 1
	placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
	c := UserCondition{
		Table: b.TableName,
		Field: UserFieldUserName,
		Func:  qtypes.LessThan,
		Value: placeholder,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	b.SelectParams = append(b.SelectParams, compareTo)
	return cn
}
func (b *UserBuilder) CondLteUserName(compareTo string) UserCondNode {
	b.lastPlaceHolder = b.lastPlaceHolder + 1
	placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
	c := UserCondition{
		Table: b.TableName,
		Field: UserFieldUserName,
		Func:  qtypes.LessOrEqualThan,
		Value: placeholder,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	b.SelectParams = append(b.SelectParams, compareTo)
	return cn
}
func (b *UserBuilder) CondNeUserName(compareTo string) UserCondNode {
	b.lastPlaceHolder = b.lastPlaceHolder + 1
	placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
	c := UserCondition{
		Table: b.TableName,
		Field: UserFieldUserName,
		Func:  qtypes.NotEqual,
		Value: placeholder,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	b.SelectParams = append(b.SelectParams, compareTo)
	return cn
}
func (b *UserBuilder) CondGteLastLog(compareTo string) UserCondNode {
	b.lastPlaceHolder = b.lastPlaceHolder + 1
	placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
	c := UserCondition{
		Table: b.TableName,
		Field: UserFieldLastLog,
		Func:  qtypes.GreaterOrEqualThan,
		Value: placeholder,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	b.SelectParams = append(b.SelectParams, compareTo)
	return cn
}
func (b *UserBuilder) CondLtLastLog(compareTo string) UserCondNode {
	b.lastPlaceHolder = b.lastPlaceHolder + 1
	placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
	c := UserCondition{
		Table: b.TableName,
		Field: UserFieldLastLog,
		Func:  qtypes.LessThan,
		Value: placeholder,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	b.SelectParams = append(b.SelectParams, compareTo)
	return cn
}
func (b *UserBuilder) CondLteLastLog(compareTo string) UserCondNode {
	b.lastPlaceHolder = b.lastPlaceHolder + 1
	placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
	c := UserCondition{
		Table: b.TableName,
		Field: UserFieldLastLog,
		Func:  qtypes.LessOrEqualThan,
		Value: placeholder,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	b.SelectParams = append(b.SelectParams, compareTo)
	return cn
}
func (b *UserBuilder) CondNeLastLog(compareTo string) UserCondNode {
	b.lastPlaceHolder = b.lastPlaceHolder + 1
	placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
	c := UserCondition{
		Table: b.TableName,
		Field: UserFieldLastLog,
		Func:  qtypes.NotEqual,
		Value: placeholder,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	b.SelectParams = append(b.SelectParams, compareTo)
	return cn
}
func (b *UserBuilder) CondEqLastLog(compareTo string) UserCondNode {
	b.lastPlaceHolder = b.lastPlaceHolder + 1
	placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
	c := UserCondition{
		Table: b.TableName,
		Field: UserFieldLastLog,
		Func:  qtypes.Equal,
		Value: placeholder,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	b.SelectParams = append(b.SelectParams, compareTo)
	return cn
}
func (b *UserBuilder) CondGtLastLog(compareTo string) UserCondNode {
	b.lastPlaceHolder = b.lastPlaceHolder + 1
	placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
	c := UserCondition{
		Table: b.TableName,
		Field: UserFieldLastLog,
		Func:  qtypes.GreaterThan,
		Value: placeholder,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	b.SelectParams = append(b.SelectParams, compareTo)
	return cn
}
func (u *User) TSQBScanner() []interface{} {
	return []interface{}{
		&u.ID, &u.UserName, &u.LastLog,
	}
}

func (u User) TSQBSaver() map[string]interface{} {
	return map[string]interface{}{
		"id":       u.ID,
		"username": u.UserName,
		"last_log": u.LastLog,
	}
}

func (b *UserBuilder) SetDBConnection(connection qtypes.DBConnection) *UserBuilder {
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
}

type StoreFields string

var (
	StoreFieldID        StoreFields = "id"
	StoreFieldStoreName StoreFields = "storename"
)

type StoreTableNameType string

var (
	StoreTableName StoreTableNameType = "stores"
)

type StoreCondition struct {
	Table StoreTableNameType
	Field StoreFields
	Func  qtypes.EqualConditions
	Value string
}

func (c StoreCondition) String() string {
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

type StoreCondNode struct {
	Conditions []StoreCondition
	WhereLink  qtypes.WhereLinks
	Nodes      []StoreCondNode
	Not        bool
}

func (cn StoreCondNode) String() string {
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
}

type StoreOrderCond struct {
	Field     StoreFields
	Direction qtypes.OrderDirection
}

func (o StoreOrderCond) String() string {
	if o.Direction == qtypes.OrderAsc {
		return string(o.Field)
	} else {
		return fmt.Sprintf("%s %s", string(o.Field), string(o.Direction))
	}
}

type StoreBuilder struct {
	queryType       qtypes.QueryType
	Fields          []StoreFields
	TableName       StoreTableNameType
	Conditions      []StoreCondNode
	LimitValue      int
	OffsetValue     int
	SelectParams    []interface{}
	InsertParams    map[string]interface{}
	lastPlaceHolder int
	order           []StoreOrderCond
}

func (b *StoreBuilder) ResetBuilder() {
	b.Fields = []StoreFields{StoreFieldID, StoreFieldStoreName}
	b.TableName = StoreTableName
	b.Conditions = []StoreCondNode{}
	b.SelectParams = []interface{}{}
	b.InsertParams = map[string]interface{}{}
	b.lastPlaceHolder = 0
	b.LimitValue = 0
	b.OffsetValue = 0
	b.order = []StoreOrderCond{}
}
func NewStoreBuilder() *StoreBuilder {
	b := StoreBuilder{}
	b.ResetBuilder()
	return &b
}
func (b *StoreBuilder) SQL() string {
	switch b.queryType {
	case qtypes.SelectQuery:
		return b.selectQuery()
	case qtypes.InsertQuery:
		return b.insertQuery()
	}
	return "You must init StoreBuilder with one of initial method (Select,Insert,Update,Delete)"
}

func (b StoreBuilder) getWhereString() string {
	WhereClause := ""
	if len(b.Conditions) > 0 {
		condditions := []string{}
		for _, c := range b.Conditions {
			condditions = append(condditions, c.String())
		}
		WhereClause = "where " + strings.Join(condditions, " and ")
	}
	return WhereClause
}

func (b StoreBuilder) getFieldsString() string {
	fields := []string{}
	for _, f := range b.Fields {
		fields = append(fields, string(f))
	}
	return strings.Join(fields, ",")
}

func (b StoreBuilder) getLimitString() string {
	limitString := ""
	if b.LimitValue > 0 {
		limitString = fmt.Sprintf("limit %d", b.LimitValue)
	}
	return limitString
}

func (b StoreBuilder) getOffsetString() string {
	offsetString := ""
	if b.OffsetValue > 0 {
		offsetString = fmt.Sprintf("offset %d", b.OffsetValue)
	}
	return offsetString
}

func (b StoreBuilder) getOrderString() string {
	orderString := ""
	if len(b.order) > 0 {
		orderStrings := []string{}
		for _, o := range b.order {
			orderStrings = append(orderStrings, o.String())
		}
		orderString = fmt.Sprintf("order by %s", strings.Join(orderStrings, ", "))
	}
	return orderString
}

func (b StoreBuilder) selectQuery() string {
	fields := b.getFieldsString()
	WhereClause := b.getWhereString()
	limitString := b.getLimitString()
	offsetString := b.getOffsetString()
	orderString := b.getOrderString()
	baseQuery := fmt.Sprintf("select %s from %s", fields, string(b.TableName))
	directives := []string{
		baseQuery,
		WhereClause,
		orderString,
		offsetString,
		limitString,
	}
	directives = qfuncs.RemoveEmpty(directives)
	query := strings.Join(directives, " ")
	return query
}

func (b StoreBuilder) getInsertPlaceholders() string {
	plc := 1
	placeholders := []string{}
	for range b.InsertParams {
		placeholders = append(placeholders, fmt.Sprintf("$%d", plc))
		plc = plc + 1
	}
	return strings.Join(placeholders, ",")
}

func (b StoreBuilder) insertQuery() string {
	fields := b.getFieldsString()
	subQueryTemplate, _ := template.New("subquery").Parse("insert into {{.TableName}}({{.Fields}}) values({{.Placeholders}})")
	placeholders := b.getInsertPlaceholders()
	queryBuf := new(bytes.Buffer)
	_ = subQueryTemplate.Execute(queryBuf, map[string]string{
		"Fields":       fields,
		"TableName":    string(b.TableName),
		"Placeholders": placeholders,
	})
	query := strings.TrimSpace(queryBuf.String())
	return query
}

func (b *StoreBuilder) setQueryType(qt qtypes.QueryType) {
	b.ResetBuilder()
	b.queryType = qt
}

func (b *StoreBuilder) Select() *StoreBuilder {
	b.setQueryType(qtypes.SelectQuery)
	return b
}

func (b *StoreBuilder) Insert() *StoreBuilder {
	b.setQueryType(qtypes.InsertQuery)
	return b
}

func (b *StoreBuilder) Values(u User) *StoreBuilder {
	b.InsertParams = u.TSQBSaver()
	return b
}

func (b *StoreBuilder) Limit(limit int) *StoreBuilder {
	b.LimitValue = limit
	return b
}

func (b *StoreBuilder) Offset(offset int) *StoreBuilder {
	b.OffsetValue = offset
	return b
}
func (b *StoreBuilder) Where(conditions ...StoreCondNode) *StoreBuilder {
	b.Conditions = conditions
	return b
}

func (b *StoreBuilder) ComposeAnd(conditions ...StoreCondNode) StoreCondNode {
	return b.compose(qtypes.WhereAnd, conditions...)
}

func (b *StoreBuilder) ComposeOr(conditions ...StoreCondNode) StoreCondNode {
	return b.compose(qtypes.WhereOr, conditions...)
}

func (b *StoreBuilder) compose(w qtypes.WhereLinks, conditions ...StoreCondNode) StoreCondNode {
	cn := StoreCondNode{
		Conditions: []StoreCondition{},
		Nodes:      []StoreCondNode{},
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

func (b *StoreBuilder) OrderByDesc(field StoreFields) *StoreBuilder {
	b.order = append(b.order, StoreOrderCond{Field: field, Direction: qtypes.OrderDesc})
	return b
}

func (b *StoreBuilder) OrderBy(field StoreFields) *StoreBuilder {
	b.order = append(b.order, StoreOrderCond{Field: field, Direction: qtypes.OrderAsc})
	return b
}
func (b *StoreBuilder) CondEqID(compareTo int) StoreCondNode {
	b.lastPlaceHolder = b.lastPlaceHolder + 1
	placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
	c := StoreCondition{
		Table: b.TableName,
		Field: StoreFieldID,
		Func:  qtypes.Equal,
		Value: placeholder,
	}
	cn := StoreCondNode{
		Conditions: []StoreCondition{c},
	}
	b.SelectParams = append(b.SelectParams, compareTo)
	return cn
}
func (b *StoreBuilder) CondGtID(compareTo int) StoreCondNode {
	b.lastPlaceHolder = b.lastPlaceHolder + 1
	placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
	c := StoreCondition{
		Table: b.TableName,
		Field: StoreFieldID,
		Func:  qtypes.GreaterThan,
		Value: placeholder,
	}
	cn := StoreCondNode{
		Conditions: []StoreCondition{c},
	}
	b.SelectParams = append(b.SelectParams, compareTo)
	return cn
}
func (b *StoreBuilder) CondGteID(compareTo int) StoreCondNode {
	b.lastPlaceHolder = b.lastPlaceHolder + 1
	placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
	c := StoreCondition{
		Table: b.TableName,
		Field: StoreFieldID,
		Func:  qtypes.GreaterOrEqualThan,
		Value: placeholder,
	}
	cn := StoreCondNode{
		Conditions: []StoreCondition{c},
	}
	b.SelectParams = append(b.SelectParams, compareTo)
	return cn
}
func (b *StoreBuilder) CondLtID(compareTo int) StoreCondNode {
	b.lastPlaceHolder = b.lastPlaceHolder + 1
	placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
	c := StoreCondition{
		Table: b.TableName,
		Field: StoreFieldID,
		Func:  qtypes.LessThan,
		Value: placeholder,
	}
	cn := StoreCondNode{
		Conditions: []StoreCondition{c},
	}
	b.SelectParams = append(b.SelectParams, compareTo)
	return cn
}
func (b *StoreBuilder) CondLteID(compareTo int) StoreCondNode {
	b.lastPlaceHolder = b.lastPlaceHolder + 1
	placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
	c := StoreCondition{
		Table: b.TableName,
		Field: StoreFieldID,
		Func:  qtypes.LessOrEqualThan,
		Value: placeholder,
	}
	cn := StoreCondNode{
		Conditions: []StoreCondition{c},
	}
	b.SelectParams = append(b.SelectParams, compareTo)
	return cn
}
func (b *StoreBuilder) CondNeID(compareTo int) StoreCondNode {
	b.lastPlaceHolder = b.lastPlaceHolder + 1
	placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
	c := StoreCondition{
		Table: b.TableName,
		Field: StoreFieldID,
		Func:  qtypes.NotEqual,
		Value: placeholder,
	}
	cn := StoreCondNode{
		Conditions: []StoreCondition{c},
	}
	b.SelectParams = append(b.SelectParams, compareTo)
	return cn
}
func (b *StoreBuilder) CondGteStoreName(compareTo string) StoreCondNode {
	b.lastPlaceHolder = b.lastPlaceHolder + 1
	placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
	c := StoreCondition{
		Table: b.TableName,
		Field: StoreFieldStoreName,
		Func:  qtypes.GreaterOrEqualThan,
		Value: placeholder,
	}
	cn := StoreCondNode{
		Conditions: []StoreCondition{c},
	}
	b.SelectParams = append(b.SelectParams, compareTo)
	return cn
}
func (b *StoreBuilder) CondLtStoreName(compareTo string) StoreCondNode {
	b.lastPlaceHolder = b.lastPlaceHolder + 1
	placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
	c := StoreCondition{
		Table: b.TableName,
		Field: StoreFieldStoreName,
		Func:  qtypes.LessThan,
		Value: placeholder,
	}
	cn := StoreCondNode{
		Conditions: []StoreCondition{c},
	}
	b.SelectParams = append(b.SelectParams, compareTo)
	return cn
}
func (b *StoreBuilder) CondLteStoreName(compareTo string) StoreCondNode {
	b.lastPlaceHolder = b.lastPlaceHolder + 1
	placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
	c := StoreCondition{
		Table: b.TableName,
		Field: StoreFieldStoreName,
		Func:  qtypes.LessOrEqualThan,
		Value: placeholder,
	}
	cn := StoreCondNode{
		Conditions: []StoreCondition{c},
	}
	b.SelectParams = append(b.SelectParams, compareTo)
	return cn
}
func (b *StoreBuilder) CondNeStoreName(compareTo string) StoreCondNode {
	b.lastPlaceHolder = b.lastPlaceHolder + 1
	placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
	c := StoreCondition{
		Table: b.TableName,
		Field: StoreFieldStoreName,
		Func:  qtypes.NotEqual,
		Value: placeholder,
	}
	cn := StoreCondNode{
		Conditions: []StoreCondition{c},
	}
	b.SelectParams = append(b.SelectParams, compareTo)
	return cn
}
func (b *StoreBuilder) CondEqStoreName(compareTo string) StoreCondNode {
	b.lastPlaceHolder = b.lastPlaceHolder + 1
	placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
	c := StoreCondition{
		Table: b.TableName,
		Field: StoreFieldStoreName,
		Func:  qtypes.Equal,
		Value: placeholder,
	}
	cn := StoreCondNode{
		Conditions: []StoreCondition{c},
	}
	b.SelectParams = append(b.SelectParams, compareTo)
	return cn
}
func (b *StoreBuilder) CondGtStoreName(compareTo string) StoreCondNode {
	b.lastPlaceHolder = b.lastPlaceHolder + 1
	placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
	c := StoreCondition{
		Table: b.TableName,
		Field: StoreFieldStoreName,
		Func:  qtypes.GreaterThan,
		Value: placeholder,
	}
	cn := StoreCondNode{
		Conditions: []StoreCondition{c},
	}
	b.SelectParams = append(b.SelectParams, compareTo)
	return cn
}
func (u *Store) TSQBScanner() []interface{} {
	return []interface{}{
		&u.ID, &u.StoreName,
	}
}

func (u Store) TSQBSaver() map[string]interface{} {
	return map[string]interface{}{
		"id":        u.ID,
		"storename": u.StoreName,
	}
}

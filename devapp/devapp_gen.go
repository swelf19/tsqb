package devapp

import (
	"bytes"

	"fmt"

	"strings"

	"text/template"

	"github.com/swelf19/tsqb/qtypes"

	"github.com/swelf19/tsqb/qfuncs"

	"errors"

	"context"
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
	WhereLink  qtypes.ComposeMethod
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

type whereUserBuilder struct {
	Conditions []UserCondNode
}

type UserBuilder struct {
	queryType qtypes.QueryType
	Fields    []UserFields
	TableName UserTableNameType
	whereUserBuilder
	LimitValue         int
	OffsetValue        int
	PreparedStmtParams []interface{}
	InsertParams       []qtypes.InsertParam
	lastPlaceHolder    int
	order              []UserOrderCond
	connection         qtypes.DBConnection
	returningID        bool
}

func (b *UserBuilder) ResetBuilder() {
	b.Fields = []UserFields{UserFieldID, UserFieldUserName, UserFieldLastLog}
	b.TableName = UserTableName
	b.Conditions = []UserCondNode{}
	b.PreparedStmtParams = []interface{}{}
	b.InsertParams = []qtypes.InsertParam{}
	b.lastPlaceHolder = 0
	b.LimitValue = 0
	b.OffsetValue = 0
	b.order = []UserOrderCond{}
	b.returningID = false
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
	case qtypes.UpdateQuery:
		return b.updateQuery()
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

func (b UserBuilder) GetStmtParams() []interface{} {
	switch b.queryType {
	case qtypes.SelectQuery:
		return b.PreparedStmtParams
	case qtypes.InsertQuery:
		return append(b.PreparedStmtParams, b.getUpsertStmtParams()...)
	case qtypes.UpdateQuery:
		return append(b.PreparedStmtParams, b.getUpsertStmtParams()...)
	}
	return nil
}

func (b UserBuilder) getUpsertStmtParams() []interface{} {
	updates := []interface{}{}
	for _, i := range b.InsertParams {
		if b.queryType == qtypes.UpdateQuery && i.Name == "id" {
			continue
		}
		updates = append(updates, i.Value)
	}
	return updates
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

func (b UserBuilder) getInsertFieldsString() string {
	fields := []string{}
	for _, f := range b.InsertParams {

		fields = append(fields, string(f.Name))
	}
	return strings.Join(fields, ",")
}

func (b UserBuilder) insertQuery() string {
	fields := b.getInsertFieldsString()
	subQueryTemplate, _ := template.New("subquery").Parse("insert into {{.TableName}}({{.Fields}}) values({{.Placeholders}}) returning {{.ReturningfField}}")
	placeholders := b.getInsertPlaceholders()
	queryBuf := new(bytes.Buffer)
	_ = subQueryTemplate.Execute(queryBuf, map[string]string{
		"Fields":          fields,
		"TableName":       string(b.TableName),
		"Placeholders":    placeholders,
		"ReturningfField": "id",
	})
	query := strings.TrimSpace(queryBuf.String())
	return query
}

func (b *UserBuilder) getUpdateExpressions() string {
	updates := []string{}
	for _, v := range b.InsertParams {
		if v.Name == "id" {
			continue
		}
		b.lastPlaceHolder = b.lastPlaceHolder + 1
		updates = append(updates, fmt.Sprintf("%s = $%d", v.Name, b.lastPlaceHolder))
	}
	return strings.Join(updates, ", ")
}

func (b *UserBuilder) updateQuery() string {
	expression := b.getUpdateExpressions()
	WhereClause := b.getWhereString()
	limitString := b.getLimitString()
	baseQuery := fmt.Sprintf("update %s set", string(b.TableName))
	directives := []string{
		baseQuery,
		expression,
		WhereClause,
		limitString,
	}
	directives = qfuncs.RemoveEmpty(directives)
	query := strings.Join(directives, " ")
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

func (b *UserBuilder) Insert(u User) *UserBuilder {
	b.setQueryType(qtypes.InsertQuery)
	b.InsertParams = u.TSQBSaver()
	return b
}

func (b *UserBuilder) Update(u User) *UserBuilder {
	b.setQueryType(qtypes.UpdateQuery)
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
func (b *UserBuilder) ReturningID() *UserBuilder {
	b.returningID = true
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

func (b *UserBuilder) compose(w qtypes.ComposeMethod, conditions ...UserCondNode) UserCondNode {
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
	b.PreparedStmtParams = append(b.PreparedStmtParams, compareTo)
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
	b.PreparedStmtParams = append(b.PreparedStmtParams, compareTo)
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
	b.PreparedStmtParams = append(b.PreparedStmtParams, compareTo)
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
	b.PreparedStmtParams = append(b.PreparedStmtParams, compareTo)
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
	b.PreparedStmtParams = append(b.PreparedStmtParams, compareTo)
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
	b.PreparedStmtParams = append(b.PreparedStmtParams, compareTo)
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
	b.PreparedStmtParams = append(b.PreparedStmtParams, compareTo)
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
	b.PreparedStmtParams = append(b.PreparedStmtParams, compareTo)
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
	b.PreparedStmtParams = append(b.PreparedStmtParams, compareTo)
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
	b.PreparedStmtParams = append(b.PreparedStmtParams, compareTo)
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
	b.PreparedStmtParams = append(b.PreparedStmtParams, compareTo)
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
	b.PreparedStmtParams = append(b.PreparedStmtParams, compareTo)
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
	b.PreparedStmtParams = append(b.PreparedStmtParams, compareTo)
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
	b.PreparedStmtParams = append(b.PreparedStmtParams, compareTo)
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
	b.PreparedStmtParams = append(b.PreparedStmtParams, compareTo)
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
	b.PreparedStmtParams = append(b.PreparedStmtParams, compareTo)
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
	b.PreparedStmtParams = append(b.PreparedStmtParams, compareTo)
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
	b.PreparedStmtParams = append(b.PreparedStmtParams, compareTo)
	return cn
}
func (u *User) TSQBScanner() []interface{} {
	return []interface{}{
		&u.ID, &u.UserName, &u.LastLog,
	}
}

func (u User) TSQBSaver() []qtypes.InsertParam {
	params := []qtypes.InsertParam{}
	if u.ID > 0 {
		params = append(params, qtypes.InsertParam{Name: "id", Value: u.ID})
	}
	params = append(params, []qtypes.InsertParam{
		{Name: "username", Value: u.UserName},
		{Name: "last_log", Value: u.LastLog},
	}...)
	return params
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
	rows, err := b.connection.Query(context.Background(), b.SQL(), b.GetStmtParams()...)
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

func (b *UserBuilder) Exec() (int, error) {
	if b.connection == nil {
		return 0, errors.New("Required to setup (SetDBConnection) connection before fetching")
	}
	var ret int
	err := b.connection.QueryRow(context.Background(), b.SQL(), b.GetStmtParams()...).Scan(&ret)
	if err != nil {
		return 0, fmt.Errorf("insert User error: %w", err)
	}
	return ret, nil
}

func (b *UserBuilder) UpdateExec() error {
	if b.connection == nil {
		return errors.New("Required to setup (SetDBConnection) connection before fetching")
	}
	pgTag, err := b.connection.Exec(context.Background(), b.SQL(), b.GetStmtParams()...)
	if err != nil {
		return fmt.Errorf("Update User err: %w. %v", err, pgTag)
	}
	return nil
}

type whereBuilder struct {
	Nodes []UserCondNode
}

func (w *whereBuilder) Where(conditions ...UserCondNode) {

}

func (w whereBuilder) buildConditions(startPlaceholder int) (lastPlaceHolder int, conditionExp string) {
	return 0, ""
}

type BaseBuilder struct {
	placeholder int
}

type SelectBuilder struct {
	BaseBuilder
	whereBuilder
}

func (b *SelectBuilder) Where() *SelectBuilder {
	b.whereBuilder.Where()
	return b
}

func (b SelectBuilder) Build() QuerySelect {
	return QuerySelect{}
}

type UpdateBuilder struct {
	whereBuilder
}

func (b UpdateBuilder) Build() QueryUpdate {
	return QueryUpdate{}
}

type InsertBuilder struct {
}

func (b InsertBuilder) Build() QueryInsert {
	return QueryInsert{}
}

type QuerySelect struct {
	stmtParams []interface{}
}

type QueryUpdate struct {
}

type QueryInsert struct {
}

func (q QuerySelect) Fetch() ([]int, error) {
	return nil, nil
}

func (q QueryInsert) Exec() (int, error) {
	return 0, nil
}

func (q QueryUpdate) Exec() error {
	return nil
}

// type Builderer interface {
// 	Build() Query
// }

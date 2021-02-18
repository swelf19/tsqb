package devapp2

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
	Field     UserFields
	Func      qtypes.EqualConditions
	CompareTo interface{}
}

type UserCondNode struct {
	Conditions    []UserCondition
	ComposeMethod qtypes.ComposeMethod
	Nodes         []UserCondNode
	Not           bool
}

func (c UserCondition) BuildCond(nextPlcNumber int, tableName string) UserCompleteCondition {
	subQueryTemplate, _ := template.New("subquery").Parse("{{.Table}}.{{.Field}} {{.Func}} ${{.Value}}")
	queryBuf := new(bytes.Buffer)
	_ = subQueryTemplate.Execute(queryBuf, map[string]string{
		"Table": tableName,
		"Field": string(c.Field),
		"Func":  string(c.Func),
		"Value": fmt.Sprintf("%d", nextPlcNumber),
	})
	return UserCompleteCondition{
		sqlString:     queryBuf.String(),
		lastPlcNumber: nextPlcNumber + 1,
		stmtParams:    []interface{}{c.CompareTo},
	}
}

type UserCompleteCondition struct {
	sqlString     string
	lastPlcNumber int
	stmtParams    []interface{}
}

func (cn UserCondNode) BuildCond(nextPlcNumber int, tableName string) UserCompleteCondition {
	conditions := []string{}
	cc := UserCompleteCondition{}
	plcNumber := nextPlcNumber
	for _, c := range cn.Conditions {
		bc := c.BuildCond(plcNumber, tableName)
		cc.stmtParams = append(cc.stmtParams, bc.stmtParams...)
		conditions = append(conditions, bc.sqlString)
		plcNumber = plcNumber + len(bc.stmtParams)
	}
	for _, n := range cn.Nodes {
		bc := n.BuildCond(plcNumber, tableName)
		cc.stmtParams = append(cc.stmtParams, bc.stmtParams...)
		conditions = append(conditions, bc.sqlString)
		plcNumber = plcNumber + len(bc.stmtParams)
	}
	condTemplate := "%s"
	if len(conditions) > 1 {
		condTemplate = "(%s)"
	}
	if cn.Not {
		condTemplate = "not " + condTemplate
	}
	cc.sqlString = fmt.Sprintf(condTemplate, strings.Join(conditions, fmt.Sprintf(" %s ", cn.ComposeMethod)))
	cc.lastPlcNumber = plcNumber
	return cc
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

type UserSelectBuilder struct {
	Fields    []UserFields
	TableName UserTableNameType
	whereUserBuilder
	LimitValue         int
	OffsetValue        int
	PreparedStmtParams []interface{}
	order              []UserOrderCond
}

func NewUserSelectBuilder() *UserSelectBuilder {
	b := UserSelectBuilder{}
	b.Fields = []UserFields{UserFieldID, UserFieldUserName, UserFieldLastLog}
	b.TableName = UserTableName
	b.Conditions = []UserCondNode{}
	b.PreparedStmtParams = []interface{}{}
	b.LimitValue = 0
	b.OffsetValue = 0
	b.order = []UserOrderCond{}
	return &b
}

func (b *UserSelectBuilder) Build() UserSelectQuery {
	cc := b.getWhereString(b.TableName)
	return UserSelectQuery{
		Fields:      b.Fields,
		stmtParams:  cc.stmtParams,
		WhereString: cc.sqlString,
		TableName:   b.TableName,
		LimitValue:  b.LimitValue,
		OffsetValue: b.OffsetValue,
		order:       b.order,
	}
}
func (b *UserSelectBuilder) Limit(limit int) *UserSelectBuilder {
	b.LimitValue = limit
	return b
}

func (b *UserSelectBuilder) Offset(offset int) *UserSelectBuilder {
	b.OffsetValue = offset
	return b
}

func (b *UserSelectBuilder) Where(conditions ...UserCondNode) *UserSelectBuilder {
	b.Conditions = conditions
	return b
}
func (b *UserSelectBuilder) OrderByDesc(field UserFields) *UserSelectBuilder {
	b.order = append(b.order, UserOrderCond{Field: field, Direction: qtypes.OrderDesc})
	return b
}

func (b *UserSelectBuilder) OrderBy(field UserFields) *UserSelectBuilder {
	b.order = append(b.order, UserOrderCond{Field: field, Direction: qtypes.OrderAsc})
	return b
}

type UserInsertBuilder struct {
	Fields       []UserFields
	TableName    UserTableNameType
	InsertParams []qtypes.InsertParam
}

func NewUserInsertBuilder() *UserInsertBuilder {
	b := UserInsertBuilder{}
	b.Fields = []UserFields{UserFieldID, UserFieldUserName, UserFieldLastLog}
	b.TableName = UserTableName
	b.InsertParams = []qtypes.InsertParam{}
	return &b
}

func (b *UserInsertBuilder) Insert(u User) *UserInsertBuilder {
	b.InsertParams = u.TSQBSaver()
	return b
}

func (b *UserInsertBuilder) Build() UserInsertQuery {
	return UserInsertQuery{
		TableName:    b.TableName,
		InsertParams: b.InsertParams,
	}
}

type UserUpdateBuilder struct {
	TableName    UserTableNameType
	UpdateParams []qtypes.InsertParam
	whereUserBuilder
}

func NewUserUpdateBuilder() *UserUpdateBuilder {
	b := UserUpdateBuilder{}
	b.TableName = UserTableName
	return &b
}

func (b *UserUpdateBuilder) Where(conditions ...UserCondNode) *UserUpdateBuilder {
	b.Conditions = conditions
	return b
}

func (b *UserUpdateBuilder) UpdateAllFields(u User) *UserUpdateBuilder {
	b.UpdateParams = append(
		b.UpdateParams,
		[]qtypes.InsertParam{
			{Name: string(UserFieldUserName), Value: u.UserName},
			{Name: string(UserFieldLastLog), Value: u.LastLog},
		}...,
	)
	b.Where(b.CondEqID(u.ID))
	return b
}

func (b *UserUpdateBuilder) Build() UserUpdateQuery {
	cc := b.getWhereString(b.TableName)
	return UserUpdateQuery{
		WhereString:            cc.sqlString,
		TableName:              b.TableName,
		UpdateParams:           b.UpdateParams,
		whereStmtParams:        cc.stmtParams,
		startPlaceholderNumber: cc.lastPlcNumber,
	}
}

func (b *UserUpdateBuilder) UpdateUserName(value string) *UserUpdateBuilder {
	b.UpdateParams = append(
		b.UpdateParams,
		qtypes.InsertParam{
			Name:  string(UserFieldUserName),
			Value: value,
		},
	)
	return b
}
func (b *UserUpdateBuilder) UpdateLastLog(value string) *UserUpdateBuilder {
	b.UpdateParams = append(
		b.UpdateParams,
		qtypes.InsertParam{
			Name:  string(UserFieldLastLog),
			Value: value,
		},
	)
	return b
}

type whereUserBuilder struct {
	Conditions []UserCondNode
}

func (b whereUserBuilder) getWhereString(tableName UserTableNameType) UserCompleteCondition {
	cc := UserCompleteCondition{
		sqlString: "",
	}
	if len(b.Conditions) > 0 {
		condditions := []string{}
		nextPlc := 1
		for _, c := range b.Conditions {
			bc := c.BuildCond(nextPlc, string(tableName))
			condditions = append(condditions, bc.sqlString)
			nextPlc = bc.lastPlcNumber
			cc.stmtParams = append(cc.stmtParams, bc.stmtParams...)
		}
		cc.lastPlcNumber = nextPlc
		cc.sqlString = "where " + strings.Join(condditions, " and ")
	}
	return cc
}

func (b *whereUserBuilder) ComposeAnd(conditions ...UserCondNode) UserCondNode {
	return b.compose(qtypes.WhereAnd, conditions...)
}

func (b *whereUserBuilder) ComposeOr(conditions ...UserCondNode) UserCondNode {
	return b.compose(qtypes.WhereOr, conditions...)
}

func (b *whereUserBuilder) compose(w qtypes.ComposeMethod, conditions ...UserCondNode) UserCondNode {
	cn := UserCondNode{
		Conditions:    []UserCondition{},
		Nodes:         []UserCondNode{},
		ComposeMethod: w,
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
func (b *whereUserBuilder) CondEqID(compareTo int) UserCondNode {
	c := UserCondition{
		Field:     UserFieldID,
		Func:      qtypes.Equal,
		CompareTo: compareTo,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	return cn
}
func (b *whereUserBuilder) CondGtID(compareTo int) UserCondNode {
	c := UserCondition{
		Field:     UserFieldID,
		Func:      qtypes.GreaterThan,
		CompareTo: compareTo,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	return cn
}
func (b *whereUserBuilder) CondGteID(compareTo int) UserCondNode {
	c := UserCondition{
		Field:     UserFieldID,
		Func:      qtypes.GreaterOrEqualThan,
		CompareTo: compareTo,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	return cn
}
func (b *whereUserBuilder) CondLtID(compareTo int) UserCondNode {
	c := UserCondition{
		Field:     UserFieldID,
		Func:      qtypes.LessThan,
		CompareTo: compareTo,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	return cn
}
func (b *whereUserBuilder) CondLteID(compareTo int) UserCondNode {
	c := UserCondition{
		Field:     UserFieldID,
		Func:      qtypes.LessOrEqualThan,
		CompareTo: compareTo,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	return cn
}
func (b *whereUserBuilder) CondNeID(compareTo int) UserCondNode {
	c := UserCondition{
		Field:     UserFieldID,
		Func:      qtypes.NotEqual,
		CompareTo: compareTo,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	return cn
}
func (b *whereUserBuilder) CondEqUserName(compareTo string) UserCondNode {
	c := UserCondition{
		Field:     UserFieldUserName,
		Func:      qtypes.Equal,
		CompareTo: compareTo,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	return cn
}
func (b *whereUserBuilder) CondGtUserName(compareTo string) UserCondNode {
	c := UserCondition{
		Field:     UserFieldUserName,
		Func:      qtypes.GreaterThan,
		CompareTo: compareTo,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	return cn
}
func (b *whereUserBuilder) CondGteUserName(compareTo string) UserCondNode {
	c := UserCondition{
		Field:     UserFieldUserName,
		Func:      qtypes.GreaterOrEqualThan,
		CompareTo: compareTo,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	return cn
}
func (b *whereUserBuilder) CondLtUserName(compareTo string) UserCondNode {
	c := UserCondition{
		Field:     UserFieldUserName,
		Func:      qtypes.LessThan,
		CompareTo: compareTo,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	return cn
}
func (b *whereUserBuilder) CondLteUserName(compareTo string) UserCondNode {
	c := UserCondition{
		Field:     UserFieldUserName,
		Func:      qtypes.LessOrEqualThan,
		CompareTo: compareTo,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	return cn
}
func (b *whereUserBuilder) CondNeUserName(compareTo string) UserCondNode {
	c := UserCondition{
		Field:     UserFieldUserName,
		Func:      qtypes.NotEqual,
		CompareTo: compareTo,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	return cn
}
func (b *whereUserBuilder) CondEqLastLog(compareTo string) UserCondNode {
	c := UserCondition{
		Field:     UserFieldLastLog,
		Func:      qtypes.Equal,
		CompareTo: compareTo,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	return cn
}
func (b *whereUserBuilder) CondGtLastLog(compareTo string) UserCondNode {
	c := UserCondition{
		Field:     UserFieldLastLog,
		Func:      qtypes.GreaterThan,
		CompareTo: compareTo,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	return cn
}
func (b *whereUserBuilder) CondGteLastLog(compareTo string) UserCondNode {
	c := UserCondition{
		Field:     UserFieldLastLog,
		Func:      qtypes.GreaterOrEqualThan,
		CompareTo: compareTo,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	return cn
}
func (b *whereUserBuilder) CondLtLastLog(compareTo string) UserCondNode {
	c := UserCondition{
		Field:     UserFieldLastLog,
		Func:      qtypes.LessThan,
		CompareTo: compareTo,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	return cn
}
func (b *whereUserBuilder) CondLteLastLog(compareTo string) UserCondNode {
	c := UserCondition{
		Field:     UserFieldLastLog,
		Func:      qtypes.LessOrEqualThan,
		CompareTo: compareTo,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	return cn
}
func (b *whereUserBuilder) CondNeLastLog(compareTo string) UserCondNode {
	c := UserCondition{
		Field:     UserFieldLastLog,
		Func:      qtypes.NotEqual,
		CompareTo: compareTo,
	}
	cn := UserCondNode{
		Conditions: []UserCondition{c},
	}
	return cn
}

type UserSelectQuery struct {
	stmtParams  []interface{}
	Fields      []UserFields
	TableName   UserTableNameType
	WhereString string
	LimitValue  int
	OffsetValue int
	order       []UserOrderCond
}

func (q UserSelectQuery) getLimitString() string {
	limitString := ""
	if q.LimitValue > 0 {
		limitString = fmt.Sprintf("limit %d", q.LimitValue)
	}
	return limitString
}

func (q UserSelectQuery) getOffsetString() string {
	offsetString := ""
	if q.OffsetValue > 0 {
		offsetString = fmt.Sprintf("offset %d", q.OffsetValue)
	}
	return offsetString
}

func (q UserSelectQuery) getOrderString() string {
	orderString := ""
	if len(q.order) > 0 {
		orderStrings := []string{}
		for _, o := range q.order {
			orderStrings = append(orderStrings, o.String())
		}
		orderString = fmt.Sprintf("order by %s", strings.Join(orderStrings, ", "))
	}
	return orderString
}

func (q UserSelectQuery) getFieldsString() string {
	fields := []string{}
	for _, f := range q.Fields {
		fields = append(fields, string(f))
	}
	return strings.Join(fields, ",")
}

func (b UserSelectQuery) Fetch(ctx context.Context, connection qtypes.DBConnection) ([]User, error) {
	if connection == nil {
		return nil, errors.New("Required to setup (SetDBConnection) connection before fetching")
	}
	values := []User{}
	rows, err := connection.Query(ctx, b.SQL(), b.stmtParams...)
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

func (q UserSelectQuery) SQL() string {
	fields := q.getFieldsString()
	WhereClause := q.WhereString
	limitString := q.getLimitString()
	offsetString := q.getOffsetString()
	orderString := q.getOrderString()
	baseQuery := fmt.Sprintf("select %s from %s", fields, string(q.TableName))
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

type UserInsertQuery struct {
	TableName    UserTableNameType
	InsertParams []qtypes.InsertParam
}

func (q UserInsertQuery) Exec(ctx context.Context, connection qtypes.DBConnection) (int, error) {
	if connection == nil {
		return 0, errors.New("Required to setup (SetDBConnection) connection before fetching")
	}
	var ret int
	err := connection.QueryRow(ctx, q.SQL(), q.getInsertStmtParams()...).Scan(&ret)
	if err != nil {
		return 0, fmt.Errorf("insert User error: %w", err)
	}
	return ret, nil
}
func (b UserInsertQuery) getInsertPlaceholders() string {
	plc := 1
	placeholders := []string{}
	for range b.InsertParams {
		placeholders = append(placeholders, fmt.Sprintf("$%d", plc))
		plc = plc + 1
	}
	return strings.Join(placeholders, ",")
}

func (b UserInsertQuery) getInsertFieldsString() string {
	fields := []string{}
	for _, f := range b.InsertParams {
		fields = append(fields, string(f.Name))
	}
	return strings.Join(fields, ",")
}

func (b UserInsertQuery) getInsertStmtParams() []interface{} {
	params := []interface{}{}
	for _, f := range b.InsertParams {
		params = append(params, f.Value)
	}
	return params
}

func (b UserInsertQuery) SQL() string {
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

type UserUpdateQuery struct {
	TableName              UserTableNameType
	WhereString            string
	UpdateParams           []qtypes.InsertParam
	whereStmtParams        []interface{}
	startPlaceholderNumber int
}

func (q UserUpdateQuery) Exec(ctx context.Context, connection qtypes.DBConnection) error {
	if connection == nil {
		return errors.New("Required to setup (SetDBConnection) connection before fetching")
	}
	_, err := connection.Exec(ctx, q.SQL(), q.getUpdateStmtParams()...)
	if err != nil {
		return fmt.Errorf("update User error: %w", err)
	}
	return nil
}

func (b UserUpdateQuery) getUpdateFields() string {
	fields := []string{}
	plc := b.startPlaceholderNumber
	for _, f := range b.UpdateParams {
		fields = append(fields, fmt.Sprintf("%s = $%d", string(f.Name), plc))
		plc = plc + 1
	}
	return strings.Join(fields, ", ")
}

func (b UserUpdateQuery) getUpdateStmtParams() []interface{} {
	params := b.whereStmtParams
	for _, f := range b.UpdateParams {
		params = append(params, f.Value)
	}
	return params
}

func (b UserUpdateQuery) SQL() string {
	subQueryTemplate, _ := template.New("subquery").Parse("update {{.TableName}} set {{.Updates}} {{.Where}}")
	queryBuf := new(bytes.Buffer)
	_ = subQueryTemplate.Execute(queryBuf, map[string]string{
		"TableName": string(b.TableName),
		"Updates":   b.getUpdateFields(),
		"Where":     b.WhereString,
	})
	query := strings.TrimSpace(queryBuf.String())
	return query
}

func (u *User) TSQBScanner() []interface{} {
	return []interface{}{
		&u.ID, &u.UserName, &u.LastLog,
	}
}

func (u User) TSQBSaver() []qtypes.InsertParam {
	params := []qtypes.InsertParam{}
	if u.ID > 0 {
		params = append(params, qtypes.InsertParam{Name: string(UserFieldID), Value: u.ID})
	}
	params = append(params, []qtypes.InsertParam{
		{Name: string(UserFieldUserName), Value: u.UserName},
		{Name: string(UserFieldLastLog), Value: u.LastLog},
	}...)
	return params
}

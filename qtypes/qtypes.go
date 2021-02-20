package qtypes

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgconn"
	pgx "github.com/jackc/pgx/v4"
)

type EqualConditions string

var (
	Equal              EqualConditions = "="
	NotEqual           EqualConditions = "<>"
	GreaterThan        EqualConditions = ">"
	GreaterOrEqualThan EqualConditions = ">="
	LessThan           EqualConditions = "<"
	LessOrEqualThan    EqualConditions = "<="
	IN                 EqualConditions = "IN"
)

type ComposeMethod string

var (
	WhereAnd ComposeMethod = "and"
	WhereOr  ComposeMethod = "or"
)

type QueryType string

var (
	SelectQuery QueryType = "select"
	InsertQuery QueryType = "insert"
	UpdateQuery QueryType = "update"
	DeleteQuery QueryType = "delete"
)

type OrderDirection string

var (
	OrderAsc  OrderDirection = "asc"
	OrderDesc OrderDirection = "desc"
)

type DBConnection interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

type InsertParam struct {
	Name  string
	Value interface{}
}

type FieldI interface {
	GetFieldName() string
	GetTableName() string
	GetName() string
}

type SimpleCondition struct {
	Field     FieldI
	Func      EqualConditions
	CompareTo []interface{}
}

func (s SimpleCondition) getSQL(plcNumber *int) (string, PreparedStmtParams) {
	// fmt.Println(*plcNumber)
	tmp := ""
	if s.Func != IN {
		*plcNumber++
		tmp = fmt.Sprintf("$%d", *plcNumber)
	} else {
		t := []string{}
		for range s.CompareTo {
			*plcNumber++
			t = append(t, fmt.Sprintf("$%d", *plcNumber))
		}
		tmp = "(" + strings.Join(t, ",") + ")"
	}
	sql := s.Field.GetName() + " " + string(s.Func) + " " + tmp
	params := PreparedStmtParams(s.CompareTo)
	return sql, params
}

func (s SimpleCondition) Build() WhereClause {
	plc := 0
	sql, params := s.getSQL(&plc)
	return WhereClause{
		SqlString:       sql,
		LastPlaceholder: plc,
		StmtParams:      params,
	}
}

type ComposedCondition struct {
	ComposeMethod ComposeMethod
	Conditions    []Condition
}

func (s ComposedCondition) getSQL(plcNumber *int) (string, PreparedStmtParams) {
	conds := []string{}
	params := PreparedStmtParams{}
	for _, c := range s.Conditions {
		// fmt.Println(*plcNumber)
		sql, p := c.getSQL(plcNumber)
		conds = append(conds, sql)
		params = append(params, p...)
		// *plcNumber++
	}
	result := strings.Join(conds, fmt.Sprintf(" %s ", s.ComposeMethod))
	if len(s.Conditions) > 1 {
		result = "(" + result + ")"
	}
	return result, params
}

func (s ComposedCondition) Build() WhereClause {
	plc := 0
	sql, params := s.getSQL(&plc)
	return WhereClause{
		SqlString:       sql,
		LastPlaceholder: plc,
		StmtParams:      params,
	}
}

type PreparedStmtParams []interface{}

type WhereClause struct {
	SqlString       string
	StmtParams      PreparedStmtParams
	LastPlaceholder int
}

type Condition interface {
	getSQL(plcNumber *int) (string, PreparedStmtParams)
	Build() WhereClause
}

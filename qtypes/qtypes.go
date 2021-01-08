package qtypes

import (
	"context"

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
)

type WhereLinks string

var (
	WhereAnd WhereLinks = "and"
	WhereOr  WhereLinks = "or"
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

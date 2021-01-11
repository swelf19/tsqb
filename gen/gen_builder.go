package gen

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// func (s StructMeta) genOrdering() string {
// 	templ := `
// 	type {{.OrderingTypeName}} struct {
// 		Field     {{.FieldsTypeName}}
// 		Direction qtypes.OrderDirection
// 	}

// 	func (o {{.OrderingTypeName}}) String() string {
// 		if o.Direction == qtypes.OrderAsc {
// 			return string(o.Field)
// 		} else {
// 			return fmt.Sprintf("%s %s", string(o.Field), string(o.Direction))
// 		}
// 	}`
// 	subQueryTemplate, err := template.New("subquery").Parse(templ)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	queryBuf := new(bytes.Buffer)
// 	err = subQueryTemplate.Execute(queryBuf, map[string]string{
// 		"OrderingTypeName": s.getOrderTypeName(),
// 		"FieldsTypeName":   s.getFieldsTypeName(),
// 	})
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	return queryBuf.String()
// }

// func (s StructMeta) genBuilderStruct() string {
// 	templ := `type {{.BuilderName}} struct {
// 		queryType       qtypes.QueryType
// 		Fields          []{{.FieldsTypeName}}
// 		TableName       {{.TableTypeName}}
// 		Conditions      []{{.CondNodeNameType}}
// 		LimitValue      int
// 		OffsetValue     int
// 		PreparedStmtParams    []interface{}
// 		InsertParams    []qtypes.InsertParam
// 		lastPlaceHolder int
// 		order           []{{.OrderingTypeName}}
// 		connection      qtypes.DBConnection
// 		returningID     bool
// 	}`
// 	subQueryTemplate, err := template.New("subquery").Parse(templ)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	queryBuf := new(bytes.Buffer)
// 	err = subQueryTemplate.Execute(queryBuf, map[string]string{
// 		"OrderingTypeName": s.getOrderTypeName(),
// 		"FieldsTypeName":   s.getFieldsTypeName(),
// 		"BuilderName":      s.getSelectBuilderName(),
// 		"CondNodeNameType": s.getCondNodeTypeName(),
// 		"TableTypeName":    s.getTableTypeName(),
// 	})
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	return queryBuf.String()
// }

// func (s *StructMeta) genBuilderResetMethod() string {
// 	fields := []string{}
// 	for _, f := range s.Fields {
// 		fields = append(fields, s.getFieldName(f))
// 	}
// 	templ := `func (b *{{.BuilderName}}) ResetBuilder() {
// 		b.Fields = []{{.FieldsTypeName}}{{.Fields}}
// 		b.TableName = {{.TableName}}
// 		b.Conditions = []{{.CondNodeNameType}}{}
// 		b.PreparedStmtParams = []interface{}{}
// 		b.InsertParams = []qtypes.InsertParam{}
// 		b.lastPlaceHolder = 0
// 		b.LimitValue = 0
// 		b.OffsetValue = 0
// 		b.order = []{{.OrderingTypeName}}{}
// 		b.returningID = false
// 	}`
// 	subQueryTemplate, err := template.New("subquery").Parse(templ)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	queryBuf := new(bytes.Buffer)
// 	err = subQueryTemplate.Execute(queryBuf, map[string]string{
// 		"OrderingTypeName": s.getOrderTypeName(),
// 		"FieldsTypeName":   s.getFieldsTypeName(),
// 		"BuilderName":      s.getSelectBuilderName(),
// 		"CondNodeNameType": s.getCondNodeTypeName(),
// 		"TableName":        s.getTableVariableName(),
// 		"Fields":           "{" + strings.Join(fields, ",") + "}",
// 	})
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	return queryBuf.String()
// }

// func (s StructMeta) genCreateBuilderFunction() string {
// 	templ := `func New{{.BuilderName}}() *{{.BuilderName}} {
// 		b := {{.BuilderName}}{}
// 		b.ResetBuilder()
// 		return &b
// 	}`
// 	subQueryTemplate, err := template.New("subquery").Parse(templ)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	queryBuf := new(bytes.Buffer)
// 	err = subQueryTemplate.Execute(queryBuf, map[string]string{
// 		"BuilderName": s.getSelectBuilderName(),
// 	})
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	return queryBuf.String()
// }

// func (s StructMeta) genBuilderCommonMethods() string {
// 	insertQuery := "insert into {{.TableName}}({{.Fields}}) values({{.Placeholders}}) returning {{.ReturningfField}}"
// 	templ := `func (b *{{.BuilderName}}) SQL() string {
// 		switch b.queryType {
// 		case qtypes.SelectQuery:
// 			return b.selectQuery()
// 		case qtypes.InsertQuery:
// 			return b.insertQuery()
// 		case qtypes.UpdateQuery:
// 			return b.updateQuery()
// 		}
// 		return "You must init {{.BuilderName}} with one of initial method (Select,Insert,Update,Delete)"
// 	}

// 	func (b {{.BuilderName}}) getWhereString() string {
// 		WhereClause := ""
// 		if len(b.Conditions) > 0 {
// 			condditions := []string{}
// 			for _, c := range b.Conditions {
// 				condditions = append(condditions, c.String())
// 			}
// 			WhereClause = "where " + strings.Join(condditions, " and ")
// 		}
// 		return WhereClause
// 	}

// 	func (b {{.BuilderName}}) GetStmtParams() []interface{} {
// 		switch b.queryType {
// 		case qtypes.SelectQuery:
// 			return b.PreparedStmtParams
// 		case qtypes.InsertQuery:
// 			return append(b.PreparedStmtParams, b.getUpsertStmtParams()...)
// 		case qtypes.UpdateQuery:
// 			return append(b.PreparedStmtParams, b.getUpsertStmtParams()...)
// 		}
// 		return nil
// 	}

// 	func (b {{.BuilderName}}) getUpsertStmtParams() []interface{} {
// 		updates := []interface{}{}
// 		for _, i := range b.InsertParams {
// 			if b.queryType == qtypes.UpdateQuery && i.Name == "id" {
// 				continue
// 			}
// 			updates = append(updates, i.Value)
// 		}
// 		return updates
// 	}

// 	func (b {{.BuilderName}}) getFieldsString() string {
// 		fields := []string{}
// 		for _, f := range b.Fields {
// 			fields = append(fields, string(f))
// 		}
// 		return strings.Join(fields, ",")
// 	}

// 	func (b {{.BuilderName}}) getLimitString() string {
// 		limitString := ""
// 		if b.LimitValue > 0 {
// 			limitString = fmt.Sprintf("limit %d", b.LimitValue)
// 		}
// 		return limitString
// 	}

// 	func (b {{.BuilderName}}) getOffsetString() string {
// 		offsetString := ""
// 		if b.OffsetValue > 0 {
// 			offsetString = fmt.Sprintf("offset %d", b.OffsetValue)
// 		}
// 		return offsetString
// 	}

// 	func (b {{.BuilderName}}) getOrderString() string {
// 		orderString := ""
// 		if len(b.order) > 0 {
// 			orderStrings := []string{}
// 			for _, o := range b.order {
// 				orderStrings = append(orderStrings, o.String())
// 			}
// 			orderString = fmt.Sprintf("order by %s", strings.Join(orderStrings, ", "))
// 		}
// 		return orderString
// 	}

// 	func (b {{.BuilderName}}) selectQuery() string {
// 		fields := b.getFieldsString()
// 		WhereClause := b.getWhereString()
// 		limitString := b.getLimitString()
// 		offsetString := b.getOffsetString()
// 		orderString := b.getOrderString()
// 		baseQuery := fmt.Sprintf("select %s from %s", fields, string(b.TableName))
// 		directives := []string{
// 			baseQuery,
// 			WhereClause,
// 			orderString,
// 			offsetString,
// 			limitString,
// 		}
// 		directives = qfuncs.RemoveEmpty(directives)
// 		query := strings.Join(directives, " ")
// 		return query
// 	}

// 	func (b {{.BuilderName}}) getInsertPlaceholders() string {
// 		plc := 1
// 		placeholders := []string{}
// 		for range b.InsertParams {
// 			placeholders = append(placeholders, fmt.Sprintf("$%d", plc))
// 			plc = plc + 1
// 		}
// 		return strings.Join(placeholders, ",")
// 	}

// 	func (b {{.BuilderName}}) getInsertFieldsString() string {
// 		fields := []string{}
// 		for _, f := range b.InsertParams {

// 			fields = append(fields, string(f.Name))
// 		}
// 		return strings.Join(fields, ",")
// 	}

// 	func (b {{.BuilderName}}) insertQuery() string {
// 		fields := b.getInsertFieldsString()
// 		subQueryTemplate, _ := template.New("subquery").Parse("{{.InsertQuery}}")
// 		placeholders := b.getInsertPlaceholders()
// 		queryBuf := new(bytes.Buffer)
// 		_ = subQueryTemplate.Execute(queryBuf, map[string]string{
// 			"Fields":       fields,
// 			"TableName":    string(b.TableName),
// 			"Placeholders": placeholders,
// 			"ReturningfField": "id",
// 		})
// 		query := strings.TrimSpace(queryBuf.String())
// 		return query
// 	}

// 	func (b *{{.BuilderName}}) getUpdateExpressions() string {
// 		updates := []string{}
// 		for _, v := range b.InsertParams {
// 			if v.Name == "id" {
// 				continue
// 			}
// 			b.lastPlaceHolder = b.lastPlaceHolder + 1
// 			updates = append(updates, fmt.Sprintf("%s = $%d", v.Name, b.lastPlaceHolder))
// 		}
// 		return strings.Join(updates, ", ")
// 	}

// 	func (b *{{.BuilderName}}) updateQuery() string {
// 		expression := b.getUpdateExpressions()
// 		WhereClause := b.getWhereString()
// 		limitString := b.getLimitString()
// 		baseQuery := fmt.Sprintf("update %s set", string(b.TableName))
// 		directives := []string{
// 			baseQuery,
// 			expression,
// 			WhereClause,
// 			limitString,
// 		}
// 		directives = qfuncs.RemoveEmpty(directives)
// 		query := strings.Join(directives, " ")
// 		return query
// 	}

// 	func (b *{{.BuilderName}}) setQueryType(qt qtypes.QueryType) {
// 		b.ResetBuilder()
// 		b.queryType = qt
// 	}

// 	func (b *{{.BuilderName}}) Select() *{{.BuilderName}} {
// 		b.setQueryType(qtypes.SelectQuery)
// 		return b
// 	}

// 	func (b *{{.BuilderName}}) Insert(u {{.StructName}}) *{{.BuilderName}} {
// 		b.setQueryType(qtypes.InsertQuery)
// 		b.InsertParams = u.TSQBSaver()
// 		return b
// 	}

// 	func (b *{{.BuilderName}}) Update(u {{.StructName}}) *{{.BuilderName}} {
// 		b.setQueryType(qtypes.UpdateQuery)
// 		b.InsertParams = u.TSQBSaver()
// 		return b
// 	}

// 	func (b *{{.BuilderName}}) Limit(limit int) *{{.BuilderName}} {
// 		b.LimitValue = limit
// 		return b
// 	}

// 	func (b *{{.BuilderName}}) Offset(offset int) *{{.BuilderName}} {
// 		b.OffsetValue = offset
// 		return b
// 	}
// 	func (b *{{.BuilderName}}) ReturningID() *{{.BuilderName}} {
// 		b.returningID = true
// 		return b
// 	}`
// 	subQueryTemplate, err := template.New("subquery").Parse(templ)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	queryBuf := new(bytes.Buffer)
// 	err = subQueryTemplate.Execute(queryBuf, map[string]string{
// 		"BuilderName": s.getSelectBuilderName(),
// 		"InsertQuery": insertQuery,
// 		"StructName":  s.StructName,
// 	})
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	return queryBuf.String()
// }

// func (s StructMeta) genBuilderBaseQueryMethods() string {
// 	templ := `func (b *{{.BuilderName}}) Where(conditions ...{{.CondNodeTypeName}}) *{{.BuilderName}} {
// 		b.Conditions = conditions
// 		return b
// 	}

// 	func (b *{{.BuilderName}}) ComposeAnd(conditions ...{{.CondNodeTypeName}}) {{.CondNodeTypeName}} {
// 		return b.compose(qtypes.WhereAnd, conditions...)
// 	}

// 	func (b *{{.BuilderName}}) ComposeOr(conditions ...{{.CondNodeTypeName}}) {{.CondNodeTypeName}} {
// 		return b.compose(qtypes.WhereOr, conditions...)
// 	}

// 	func (b *{{.BuilderName}}) compose(w qtypes.WhereLinks, conditions ...{{.CondNodeTypeName}}) {{.CondNodeTypeName}} {
// 		cn := {{.CondNodeTypeName}}{
// 			Conditions: []{{.ConditionTypeName}}{},
// 			Nodes:      []{{.CondNodeTypeName}}{},
// 			WhereLink:  w,
// 		}
// 		for _, node := range conditions {
// 			if len(node.Conditions) == 1 {
// 				cn.Conditions = append(cn.Conditions, node.Conditions...)
// 			} else {
// 				cn.Nodes = append(cn.Nodes, node)
// 			}
// 		}
// 		return cn
// 	}

// 	func (b *{{.BuilderName}}) OrderByDesc(field {{.FieldsTypeName}}) *{{.BuilderName}} {
// 		b.order = append(b.order, {{.OrderingTypeName}}{Field: field, Direction: qtypes.OrderDesc})
// 		return b
// 	}

// 	func (b *{{.BuilderName}}) OrderBy(field {{.FieldsTypeName}}) *{{.BuilderName}} {
// 		b.order = append(b.order, {{.OrderingTypeName}}{Field: field, Direction: qtypes.OrderAsc})
// 		return b
// 	}`
// 	subQueryTemplate, err := template.New("subquery").Parse(templ)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	queryBuf := new(bytes.Buffer)
// 	err = subQueryTemplate.Execute(queryBuf, map[string]string{
// 		"OrderingTypeName":  s.getOrderTypeName(),
// 		"FieldsTypeName":    s.getFieldsTypeName(),
// 		"BuilderName":       s.getSelectBuilderName(),
// 		"CondNodeTypeName":  s.getCondNodeTypeName(),
// 		"TableTypeName":     s.getTableTypeName(),
// 		"ConditionTypeName": s.getConditionTypeName(),
// 	})
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	return queryBuf.String()
// }

func (s StructMeta) genComprationForField(condKey string, condValue string, f StructFieldMeta) string {
	templ := CONDITION_METHOD
	subQueryTemplate, err := template.New("subquery").Parse(templ)
	if err != nil {
		fmt.Println(err)
	}
	queryBuf := new(bytes.Buffer)
	err = subQueryTemplate.Execute(queryBuf, map[string]string{
		"WhereUserBuilder":  s.getWhereBuilderName(),
		"CondNodeTypeName":  s.getCondNodeTypeName(),
		"ConditionTypeName": s.getConditionTypeName(),
		"CondKey":           condKey,
		"CondValue":         condValue,
		"FieldType":         f.Type,
		"GenFieldName":      s.getFieldName(f),
		"FieldName":         f.FieldName,
	})
	if err != nil {
		fmt.Println(err)
	}
	return queryBuf.String()
}

func (s StructMeta) genComprationHelpers() string {
	funcsMap := [][]string{
		{"Eq", "qtypes.Equal"},
		{"Gt", "qtypes.GreaterThan"},
		{"Gte", "qtypes.GreaterOrEqualThan"},
		{"Lt", "qtypes.LessThan"},
		{"Lte", "qtypes.LessOrEqualThan"},
		{"Ne", "qtypes.NotEqual"},
	}
	helpers := []string{}
	for _, f := range s.Fields {
		for _, v := range funcsMap {
			helpers = append(helpers, s.genComprationForField(v[0], v[1], f))
		}
	}
	return strings.Join(helpers, "\n")
}

// func (s StructMeta) genBuilderFetch() string {
// 	templ := `func (b *{{.BuilderName}}) SetDBConnection(connection qtypes.DBConnection) *{{.BuilderName}}{
// 		b.connection = connection
// 		return b
// 	}

// 	func (b *{{.BuilderName}}) Fetch() ([]{{.StructName}}, error) {
// 		if b.connection == nil {
// 			return nil, errors.New("Required to setup (SetDBConnection) connection before fetching")
// 		}
// 		values := []{{.StructName}}{}
// 		rows, err := b.connection.Query(context.Background(), b.SQL(), b.GetStmtParams()...)
// 		if err != nil {
// 			return nil, err
// 		}
// 		for rows.Next() {
// 			v := {{.StructName}}{}
// 			err = rows.Scan(v.TSQBScanner()...)
// 			if err != nil {
// 				return nil, err
// 			}
// 			values = append(values, v)
// 		}
// 		if rows.Err() != nil {
// 			return nil, rows.Err()
// 		}
// 		return values, nil
// 	}

// 	func (b *{{.BuilderName}}) Exec() (int, error) {
// 		if b.connection == nil {
// 			return 0, errors.New("Required to setup (SetDBConnection) connection before fetching")
// 		}
// 		var ret int
// 		err := b.connection.QueryRow(context.Background(), b.SQL(), b.GetStmtParams()...).Scan(&ret)
// 		if err != nil {
// 			return 0, fmt.Errorf("insert {{.StructName}} error: %w", err)
// 		}
// 		return ret, nil
// 	}

// 	func (b *{{.BuilderName}}) UpdateExec() error {
// 		if b.connection == nil {
// 			return errors.New("Required to setup (SetDBConnection) connection before fetching")
// 		}
// 		pgTag, err := b.connection.Exec(context.Background(), b.SQL(), b.GetStmtParams()...)
// 		if err != nil {
// 			return fmt.Errorf("Update {{.StructName}} err: %w. %v", err, pgTag)
// 		}
// 		return nil
// 	}
// 	`

// 	subQueryTemplate, err := template.New("subquery").Parse(templ)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	queryBuf := new(bytes.Buffer)
// 	err = subQueryTemplate.Execute(queryBuf, map[string]string{
// 		"BuilderName": s.getSelectBuilderName(),
// 		"StructName":  s.StructName,
// 	})
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	return queryBuf.String()

// }

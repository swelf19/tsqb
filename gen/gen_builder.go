package gen

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

func (s StructMeta) getOrderTypeName() string {
	return s.StructName + "OrderCond"
}

func (s StructMeta) getBuilderName() string {
	return s.StructName + "Builder"
}

func (s StructMeta) genOrdering() string {
	templ := `
	type {{.OrderingTypeName}} struct {
		Field     {{.FieldsTypeName}}
		Direction qtypes.OrderDirection
	}

	func (o {{.OrderingTypeName}}) String() string {
		if o.Direction == qtypes.OrderAsc {
			return string(o.Field)
		} else {
			return fmt.Sprintf("%s %s", string(o.Field), string(o.Direction))
		}
	}`
	subQueryTemplate, err := template.New("subquery").Parse(templ)
	if err != nil {
		fmt.Println(err)
	}
	queryBuf := new(bytes.Buffer)
	err = subQueryTemplate.Execute(queryBuf, map[string]string{
		"OrderingTypeName": s.getOrderTypeName(),
		"FieldsTypeName":   s.getFieldsTypeName(),
	})
	if err != nil {
		fmt.Println(err)
	}
	return queryBuf.String()
}

func (s StructMeta) genBuilderStruct() string {
	templ := `type {{.BuilderName}} struct {
		queryType       qtypes.QueryType
		Fields          []{{.FieldsTypeName}}
		TableName       {{.TableTypeName}}
		Conditions      []{{.CondNodeNameType}}
		LimitValue      int
		OffsetValue     int
		SelectParams    []interface{}
		InsertParams    map[string]interface{}
		lastPlaceHolder int
		order           []{{.OrderingTypeName}}
		connection      qtypes.DBConnection
	}`
	subQueryTemplate, err := template.New("subquery").Parse(templ)
	if err != nil {
		fmt.Println(err)
	}
	queryBuf := new(bytes.Buffer)
	err = subQueryTemplate.Execute(queryBuf, map[string]string{
		"OrderingTypeName": s.getOrderTypeName(),
		"FieldsTypeName":   s.getFieldsTypeName(),
		"BuilderName":      s.getBuilderName(),
		"CondNodeNameType": s.getCondNodeTypeName(),
		"TableTypeName":    s.getTableTypeName(),
	})
	if err != nil {
		fmt.Println(err)
	}
	return queryBuf.String()
}

func (s *StructMeta) genBuilderResetMethod() string {
	fields := []string{}
	for _, f := range s.Fields {
		fields = append(fields, s.getFieldName(f))
	}
	templ := `func (b *{{.BuilderName}}) ResetBuilder() {
		b.Fields = []{{.FieldsTypeName}}{{.Fields}}
		b.TableName = {{.TableName}}
		b.Conditions = []{{.CondNodeNameType}}{}
		b.SelectParams = []interface{}{}
		b.InsertParams = map[string]interface{}{}
		b.lastPlaceHolder = 0
		b.LimitValue = 0
		b.OffsetValue = 0
		b.order = []{{.OrderingTypeName}}{}
	}`
	subQueryTemplate, err := template.New("subquery").Parse(templ)
	if err != nil {
		fmt.Println(err)
	}
	queryBuf := new(bytes.Buffer)
	err = subQueryTemplate.Execute(queryBuf, map[string]string{
		"OrderingTypeName": s.getOrderTypeName(),
		"FieldsTypeName":   s.getFieldsTypeName(),
		"BuilderName":      s.getBuilderName(),
		"CondNodeNameType": s.getCondNodeTypeName(),
		"TableName":        s.getTableVariableName(),
		"Fields":           "{" + strings.Join(fields, ",") + "}",
	})
	if err != nil {
		fmt.Println(err)
	}
	return queryBuf.String()
}

func (s StructMeta) genCreateBuilderFunction() string {
	templ := `func New{{.BuilderName}}() *{{.BuilderName}} {
		b := {{.BuilderName}}{}
		b.ResetBuilder()
		return &b
	}`
	subQueryTemplate, err := template.New("subquery").Parse(templ)
	if err != nil {
		fmt.Println(err)
	}
	queryBuf := new(bytes.Buffer)
	err = subQueryTemplate.Execute(queryBuf, map[string]string{
		"BuilderName": s.getBuilderName(),
	})
	if err != nil {
		fmt.Println(err)
	}
	return queryBuf.String()
}

func (s StructMeta) genBuilderCommonMethods() string {
	insertQuery := "insert into {{.TableName}}({{.Fields}}) values({{.Placeholders}})"
	templ := `func (b *{{.BuilderName}}) SQL() string {
		switch b.queryType {
		case qtypes.SelectQuery:
			return b.selectQuery()
		case qtypes.InsertQuery:
			return b.insertQuery()
		}
		return "You must init {{.BuilderName}} with one of initial method (Select,Insert,Update,Delete)"
	}
	
	func (b {{.BuilderName}}) getWhereString() string {
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
	
	func (b {{.BuilderName}}) getFieldsString() string {
		fields := []string{}
		for _, f := range b.Fields {
			fields = append(fields, string(f))
		}
		return strings.Join(fields, ",")
	}
	
	func (b {{.BuilderName}}) getLimitString() string {
		limitString := ""
		if b.LimitValue > 0 {
			limitString = fmt.Sprintf("limit %d", b.LimitValue)
		}
		return limitString
	}
	
	func (b {{.BuilderName}}) getOffsetString() string {
		offsetString := ""
		if b.OffsetValue > 0 {
			offsetString = fmt.Sprintf("offset %d", b.OffsetValue)
		}
		return offsetString
	}

	
	func (b {{.BuilderName}}) getOrderString() string {
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
	
	func (b {{.BuilderName}}) selectQuery() string {
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
	
	func (b {{.BuilderName}}) getInsertPlaceholders() string {
		plc := 1
		placeholders := []string{}
		for range b.InsertParams {
			placeholders = append(placeholders, fmt.Sprintf("$%d", plc))
			plc = plc + 1
		}
		return strings.Join(placeholders, ",")
	}
	
	func (b {{.BuilderName}}) insertQuery() string {
		fields := b.getFieldsString()
		subQueryTemplate, _ := template.New("subquery").Parse("{{.InsertQuery}}")
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
	
	func (b *{{.BuilderName}}) setQueryType(qt qtypes.QueryType) {
		b.ResetBuilder()
		b.queryType = qt
	}
	
	func (b *{{.BuilderName}}) Select() *{{.BuilderName}} {
		b.setQueryType(qtypes.SelectQuery)
		return b
	}
	
	func (b *{{.BuilderName}}) Insert() *{{.BuilderName}} {
		b.setQueryType(qtypes.InsertQuery)
		return b
	}
	
	func (b *{{.BuilderName}}) Values(u {{.StructName}}) *{{.BuilderName}} {
		b.InsertParams = u.TSQBSaver()
		return b
	}
	
	func (b *{{.BuilderName}}) Limit(limit int) *{{.BuilderName}} {
		b.LimitValue = limit
		return b
	}
	
	func (b *{{.BuilderName}}) Offset(offset int) *{{.BuilderName}} {
		b.OffsetValue = offset
		return b
	}`
	subQueryTemplate, err := template.New("subquery").Parse(templ)
	if err != nil {
		fmt.Println(err)
	}
	queryBuf := new(bytes.Buffer)
	err = subQueryTemplate.Execute(queryBuf, map[string]string{
		"BuilderName": s.getBuilderName(),
		"InsertQuery": insertQuery,
		"StructName":  s.StructName,
	})
	if err != nil {
		fmt.Println(err)
	}
	return queryBuf.String()
}

func (s StructMeta) genBuilderBaseQueryMethods() string {
	templ := `func (b *{{.BuilderName}}) Where(conditions ...{{.CondNodeTypeName}}) *{{.BuilderName}} {
		b.Conditions = conditions
		return b
	}
	
	func (b *{{.BuilderName}}) ComposeAnd(conditions ...{{.CondNodeTypeName}}) {{.CondNodeTypeName}} {
		return b.compose(qtypes.WhereAnd, conditions...)
	}
	
	func (b *{{.BuilderName}}) ComposeOr(conditions ...{{.CondNodeTypeName}}) {{.CondNodeTypeName}} {
		return b.compose(qtypes.WhereOr, conditions...)
	}
	
	func (b *{{.BuilderName}}) compose(w qtypes.WhereLinks, conditions ...{{.CondNodeTypeName}}) {{.CondNodeTypeName}} {
		cn := {{.CondNodeTypeName}}{
			Conditions: []{{.ConditionTypeName}}{},
			Nodes:      []{{.CondNodeTypeName}}{},
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
	
	func (b *{{.BuilderName}}) OrderByDesc(field {{.FieldsTypeName}}) *{{.BuilderName}} {
		b.order = append(b.order, {{.OrderingTypeName}}{Field: field, Direction: qtypes.OrderDesc})
		return b
	}
	
	func (b *{{.BuilderName}}) OrderBy(field {{.FieldsTypeName}}) *{{.BuilderName}} {
		b.order = append(b.order, {{.OrderingTypeName}}{Field: field, Direction: qtypes.OrderAsc})
		return b
	}`
	subQueryTemplate, err := template.New("subquery").Parse(templ)
	if err != nil {
		fmt.Println(err)
	}
	queryBuf := new(bytes.Buffer)
	err = subQueryTemplate.Execute(queryBuf, map[string]string{
		"OrderingTypeName":  s.getOrderTypeName(),
		"FieldsTypeName":    s.getFieldsTypeName(),
		"BuilderName":       s.getBuilderName(),
		"CondNodeTypeName":  s.getCondNodeTypeName(),
		"TableTypeName":     s.getTableTypeName(),
		"ConditionTypeName": s.getConditionTypeName(),
	})
	if err != nil {
		fmt.Println(err)
	}
	return queryBuf.String()
}

func (s StructMeta) genComprationForField(condKey string, condValue string, f StructFieldMeta) string {
	templ := `func (b *{{.BuilderName}}) Cond{{.CondKey}}{{.FieldName}}(compareTo {{.FieldType}}) {{.CondNodeTypeName}} {
		b.lastPlaceHolder = b.lastPlaceHolder + 1
		placeholder := fmt.Sprintf("$%d", b.lastPlaceHolder)
		c := {{.ConditionTypeName}}{
			Table: b.TableName,
			Field: {{.GenFieldName}},
			Func:  {{.CondValue}},
			Value: placeholder,
		}
		cn := {{.CondNodeTypeName}}{
			Conditions: []{{.ConditionTypeName}}{c},
		}
		b.SelectParams = append(b.SelectParams, compareTo)
		return cn
	}`
	subQueryTemplate, err := template.New("subquery").Parse(templ)
	if err != nil {
		fmt.Println(err)
	}
	queryBuf := new(bytes.Buffer)
	err = subQueryTemplate.Execute(queryBuf, map[string]string{
		"BuilderName":       s.getBuilderName(),
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
	funcsMap := map[string]string{
		"Eq":  "qtypes.Equal",
		"Gt":  "qtypes.GreaterThan",
		"Gte": "qtypes.GreaterOrEqualThan",
		"Lt":  "qtypes.LessThan",
		"Lte": "qtypes.LessOrEqualThan",
		"Ne":  "qtypes.NotEqual",
	}
	helpers := []string{}
	for _, f := range s.Fields {
		for k, v := range funcsMap {
			helpers = append(helpers, s.genComprationForField(k, v, f))
		}
	}
	return strings.Join(helpers, "\n")
}

func (s StructMeta) genBuilderFetch() string {
	templ := `func (b *{{.BuilderName}}) SetDBConnection(connection qtypes.DBConnection) *{{.BuilderName}}{
		b.connection = connection
		return b
	}
	
	func (b *{{.BuilderName}}) Fetch() ([]{{.StructName}}, error) {
		if b.connection == nil {
			return nil, errors.New("Required to setup (SetDBConnection) connection before fetching")
		}
		values := []{{.StructName}}{}
		rows, err := b.connection.Query(context.Background(), b.SQL(), b.SelectParams...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			v := {{.StructName}}{}
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
	subQueryTemplate, err := template.New("subquery").Parse(templ)
	if err != nil {
		fmt.Println(err)
	}
	queryBuf := new(bytes.Buffer)
	err = subQueryTemplate.Execute(queryBuf, map[string]string{
		"BuilderName": s.getBuilderName(),
		"StructName":  s.StructName,
	})
	if err != nil {
		fmt.Println(err)
	}
	return queryBuf.String()

}

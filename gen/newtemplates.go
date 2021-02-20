package gen

var FIELDS = `
type {{.FieldsTypeName}} struct {
	{{ range .FieldsData }}
		{{.FieldName}} {{$.StructNameLower}}{{.FieldName}}Field
	{{ end }}
}

type {{.StructNameLower}}Schema struct {
	Fields    {{.FieldsTypeName}}
	TableName string
}

var {{.FieldsTypeName}}Value = {{.FieldsTypeName}}{
	{{ range .FieldsData }}
		{{.FieldName}}: {{$.StructNameLower}}{{.FieldName}}Field{
			FieldsName: "{{.SqlFieldName}}",
			TableName:  "{{.TableName}}",
		},
	{{end}}
}

var {{.StructNameLower}}SchemaValue = {{.StructNameLower}}Schema{
	Fields:    {{.FieldsTypeName}}Value,
	TableName: "{{.StructMeta.TableName}}",
}
{{ range .FieldsData }}
type {{$.StructNameLower}}{{.FieldName}}Field field

func (f {{$.StructNameLower}}{{.FieldName}}Field) GetFieldName() string {
	return f.FieldsName
}

func (f {{$.StructNameLower}}{{.FieldName}}Field) GetTableName() string {
	return f.TableName
}

func (f {{$.StructNameLower}}{{.FieldName}}Field) GetName() string {
	return f.GetTableName() + "." + f.GetFieldName()
}
{{end}}
`

var FIELD_CONDITION = `func (f {{.StructNameLower}}{{.FieldName}}Field) {{.CondKey}}(compareTo {{.FieldType}}) qtypes.SimpleCondition {
	cn := qtypes.SimpleCondition{
		Field:     f,
		Func:      {{.CondValue}},
		CompareTo: []interface{}{compareTo},
	}
	return cn
}`

var FIELD_CONDITION_IN = `func (f {{.StructNameLower}}{{.FieldName}}Field) In(compareTo ...{{.FieldType}}) qtypes.SimpleCondition {
	comp := make([]interface{}, len(compareTo))
	for i, c := range compareTo {
		comp[i] = c
	}
	cn := qtypes.SimpleCondition{
		Field:     f,
		Func:      qtypes.IN,
		CompareTo: comp,
	}
	return cn
}`

var ALLSCHEMAS = `
func (b builders) {{.StructName}}() {{.StructNameLower}}SelectBuilder {
	return {{.StructNameLower}}SelectBuilder{
		{{.StructName}}Schema: {{.StructNameLower}}SchemaValue,
		schemaName: "{{.SQLTableName}}",
	}
}`

var BUILDER = `type {{.StructNameLower}}SelectBuilder struct {
	{{.StructName}}Schema {{.StructNameLower}}Schema
	conditions qtypes.Condition
	schemaName string
	limit      int
	offset      int
}

func (u {{.StructNameLower}}SelectBuilder) Where(conds ...qtypes.Condition) {{.StructNameLower}}SelectBuilder {
	newbuilder := u
	newbuilder.conditions = qfuncs.ComposeAnd(conds...)
	return newbuilder
}

func (u {{.StructNameLower}}SelectBuilder) Limit(limit int) {{.StructNameLower}}SelectBuilder {
	newbuilder := u
	newbuilder.limit = limit
	return newbuilder
}

func (u {{.StructNameLower}}SelectBuilder) Offset(offset int) {{.StructNameLower}}SelectBuilder {
	newbuilder := u
	newbuilder.offset = offset
	return newbuilder
}

func (u {{.StructNameLower}}SelectBuilder) Build() {{.StructNameLower}}SelectQuery {
	q := {{.StructNameLower}}SelectQuery{
		fields: []qtypes.FieldI{
			{{range .FieldsData}}
				u.{{$.StructName}}Schema.Fields.{{.FieldName}},
			{{end}}
		},
		{{.StructNameLower}}Schema:  u.{{.StructName}}Schema,
		tableName:   u.schemaName,
		limit:      u.limit,
		offset:     u.offset,
	}
	if u.conditions != nil {
		q.whereClause = u.conditions.Build()
	}
	return q
}`

var INSERTBUIDLER = `type {{.StructNameLower}}InsertBuilder struct {
	{{.StructNameLower}}Schema {{.StructNameLower}}Schema
	schemaName string
	data       {{.StructName}}
}

func (b {{.StructNameLower}}InsertBuilder) Build() {{.StructNameLower}}InsertQuery {
	params := []interface{}{}
	for _, f := range b.data.TSQBSaver() {
		params = append(params, f.Value)
	}

	return {{.StructNameLower}}InsertQuery{
		schemaName: b.schemaName,
		data:       b.data,
		params:     params,
	}
}

func (b insertBuilders) {{.StructName}}(u {{.StructName}}) {{.StructNameLower}}InsertBuilder {
	return {{.StructNameLower}}InsertBuilder{
		{{.StructNameLower}}Schema: {{.StructNameLower}}SchemaValue,
		schemaName: "{{.SQLTableName}}",
		data:       u,
	}
}`

var UPDATEBUILDER = `
func (b updateBuilders) {{.StructName}}() {{.StructNameLower}}UpdateBuilder {
	return {{.StructNameLower}}UpdateBuilder{
		{{.StructName}}Schema: {{.StructNameLower}}SchemaValue,
		schemaName: "{{.SQLTableName}}",
	}
}

type {{.StructNameLower}}UpdateBuilder struct {
	{{.StructName}}Schema   {{.StructNameLower}}Schema
	schemaName   string
	updateFields []qtypes.InsertParam
	conditions   qtypes.Condition
}

func (u {{.StructNameLower}}UpdateBuilder) Where(conds ...qtypes.Condition) {{.StructNameLower}}UpdateBuilder {
	newbuilder := u
	newbuilder.conditions = qfuncs.ComposeAnd(conds...)
	return newbuilder
}

func (b {{.StructNameLower}}UpdateBuilder) SetAllFields(u {{.StructName}}) {{.StructNameLower}}UpdateBuilder {
	b.updateFields = append(
		b.updateFields,
		[]qtypes.InsertParam{
			{{range .FieldsData}}
			{{if ne .FieldName "ID"}}
			{Name: "{{.SqlFieldName}}", Value: u.{{.FieldName}}},
			{{end}}
			{{end}}
		}...,
	)
	return b.Where(b.{{.StructName}}Schema.Fields.ID.Eq(u.ID))
}

func (b {{.StructNameLower}}UpdateBuilder) Build() {{.StructNameLower}}UpdateQuery {
	var whereClause qtypes.WhereClause
	params := []interface{}{}
	if b.conditions != nil {
		whereClause = b.conditions.Build()
		params = whereClause.StmtParams
	}
	for _, upd := range b.updateFields {
		params = append(params, upd.Value)
	}
	return {{.StructNameLower}}UpdateQuery{
		schemaName:   b.schemaName,
		params:       params,
		whereClause:  whereClause,
		updateFields: b.updateFields,
	}
}

{{range .FieldsData}}
{{if ne .FieldName "ID"}}
func (b {{$.StructNameLower}}UpdateBuilder) Set{{.FieldName}}(upd {{.Type}}) {{$.StructNameLower}}UpdateBuilder {
	b.updateFields = append(
		b.updateFields,
		[]qtypes.InsertParam{
			{Name: "{{.SqlFieldName}}", Value: upd},
		}...,
	)
	return b
}
{{end}}
{{end}}
`

var DELETEBUILDER = `func (b deleteBuilders) {{.StructName}}() {{.StructNameLower}}DeleteBuilder {
	return {{.StructNameLower}}DeleteBuilder{
		{{.StructName}}Schema: {{.StructNameLower}}SchemaValue,
		schemaName: "{{.SQLTableName}}",
	}
}

type {{.StructNameLower}}DeleteBuilder struct {
	{{.StructName}}Schema {{.StructNameLower}}Schema
	schemaName string
	conditions qtypes.Condition
}



func (b {{.StructNameLower}}DeleteBuilder) Build() {{.StructNameLower}}DeleteQuery {
	var whereClause qtypes.WhereClause
	if b.conditions != nil {
		whereClause = b.conditions.Build()
	}
	return {{.StructNameLower}}DeleteQuery{
		schemaName:  b.schemaName,
		whereClause: whereClause,
	}
}
func (u {{.StructNameLower}}DeleteBuilder) Where(conds ...qtypes.Condition) {{.StructNameLower}}DeleteBuilder {
	newbuilder := u
	newbuilder.conditions = qfuncs.ComposeAnd(conds...)
	return newbuilder
}`

var SELECTQUERY = `type {{.StructNameLower}}SelectQuery struct {
	{{.StructNameLower}}Schema  {{.StructNameLower}}Schema
	fields      []qtypes.FieldI
	tableName   string
	whereClause qtypes.WhereClause
	limit       int
	offset      int
}
func (q {{.StructNameLower}}SelectQuery) getFields() string {
	fields := []string{}
	for _, f := range q.fields {
		fields = append(fields, f.GetName())
	}
	return strings.Join(fields, ", ")
}

func (q {{.StructNameLower}}SelectQuery) SQL() string {
	sql := fmt.Sprintf(
		"select %[1]s from %[2]s",
		q.getFields(),
		q.tableName,
	)
	if q.whereClause.SqlString != "" {
		sql = sql + " where " + q.whereClause.SqlString
	}
	if q.offset > 0 {
		sql = sql + " offset " + fmt.Sprintf("%d", q.offset)
	}
	if q.limit > 0 {
		sql = sql + " limit " + fmt.Sprintf("%d", q.limit)
	}
	return sql
}
`

var INSERTQUERY = `type {{.StructNameLower}}InsertQuery struct {
	schemaName string
	data       {{.StructName}}
	params     []interface{}
}

func (q {{.StructNameLower}}InsertQuery) SQL() string {
	names := []string{}
	placeholders := []string{}
	for plc, f := range q.data.TSQBSaver() {
		placeholders = append(placeholders, fmt.Sprintf("$%d", plc+1))
		names = append(names, f.Name)
	}
	return fmt.Sprintf("insert into %s(%s) values(%s) returning id", q.schemaName, strings.Join(names, ", "), strings.Join(placeholders, ", "))
}`

var UPDATEQUERY = `func (q {{.StructNameLower}}UpdateQuery) SQL() string {
	updates := []string{}
	plc := q.whereClause.LastPlaceholder
	params := q.whereClause.StmtParams
	for _, upd := range q.updateFields {
		plc++
		params = append(params, upd.Value)
		updates = append(updates, fmt.Sprintf("%s = $%d", upd.Name, plc))
	}
	query := fmt.Sprintf("update %s set %s", q.schemaName, strings.Join(updates, ", "))
	if q.whereClause.SqlString != "" {
		query = query + " where " + q.whereClause.SqlString
	}
	return query
}

type {{.StructNameLower}}UpdateQuery struct {
	schemaName   string
	params       []interface{}
	whereClause  qtypes.WhereClause
	updateFields []qtypes.InsertParam
}`

var DELETEQUERY = `type {{.StructNameLower}}DeleteQuery struct {
	schemaName  string
	whereClause qtypes.WhereClause
}

func (q {{.StructNameLower}}DeleteQuery) SQL() string {
	query := fmt.Sprintf("delete from %s", q.schemaName)
	if q.whereClause.SqlString != "" {
		query = query + " where " + q.whereClause.SqlString
	}
	return query
}`

var DBMETHODS = `
func (q {{.StructNameLower}}SelectQuery) Fetch(ctx context.Context, connection qtypes.DBConnection) ([]{{.StructName}}, error) {
	if connection == nil {
		return nil, errors.New("Required to setup (SetDBConnection) connection before fetching")
	}
	values := []{{.StructName}}{}
	rows, err := connection.Query(ctx, q.SQL(), q.whereClause.StmtParams...)
	if err != nil {
		return nil, fmt.Errorf("fetch {{.StructName}} error: %w", err)
	}
	for rows.Next() {
		v := {{.StructName}}{}
		err = rows.Scan(v.TSQBScanner()...)
		if err != nil {
			return nil, fmt.Errorf("fetch {{.StructName}} error: %w", err)
		}
		values = append(values, v)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("fetch {{.StructName}} error: %w", rows.Err())
	}
	return values, nil
}

func (u *{{.StructName}}) TSQBScanner() []interface{} {
	return []interface{}{
		{{range .FieldsData}}
			&u.{{.FieldName}},
		{{end}}
	}
}

func (u {{.StructName}}) TSQBSaver() []qtypes.InsertParam {
	params := []qtypes.InsertParam{}
	if u.ID > 0 {
		params = append(params, qtypes.InsertParam{Name: {{.FieldsTypeName}}Value.ID.FieldsName, Value: u.ID})
	}
	params = append(params, []qtypes.InsertParam{
		{{range .FieldsData}}
		{{if ne .FieldName "ID"}}
		{Name: {{$.FieldsTypeName}}Value.{{.FieldName}}.FieldsName, Value: u.{{.FieldName}}},
		{{end}}
		{{end}}
	}...)
	return params
}

func (q {{.StructNameLower}}InsertQuery) Exec(ctx context.Context, connection qtypes.DBConnection) (int, error) {
	if connection == nil {
		return 0, errors.New("Required to setup (SetDBConnection) connection before inserting")
	}
	var ret int
	err := connection.QueryRow(ctx, q.SQL(), q.params...).Scan(&ret)
	if err != nil {
		return 0, fmt.Errorf("insert {{.StructName}} error: %w", err)
	}
	return ret, nil
}

func (q {{.StructNameLower}}UpdateQuery) Exec(ctx context.Context, connection qtypes.DBConnection) error {
	if connection == nil {
		return errors.New("Required to setup (SetDBConnection) connection before updating")
	}
	_, err := connection.Exec(ctx, q.SQL(), q.params...)
	if err != nil {
		return fmt.Errorf("update {{.StructName}} error: %w", err)
	}
	return nil
}

func (q {{.StructNameLower}}DeleteQuery) Exec(ctx context.Context, connection qtypes.DBConnection) error {
	if connection == nil {
		return errors.New("Required to setup (SetDBConnection) connection before deleteing")
	}
	_, err := connection.Exec(ctx, q.SQL(), q.whereClause.StmtParams...)
	if err != nil {
		return fmt.Errorf("delete {{.StructName}} error: %w", err)
	}
	return nil
}

`

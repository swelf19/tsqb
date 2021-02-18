package gen

var SELECT_QUERY string = `
type {{.SelectQueryName}} struct {
	stmtParams  []interface{}
	Fields      []{{.FieldsTypeName}}
	TableName   {{.TableTypeName}}
	WhereString string
	LimitValue  int
	OffsetValue int
	order       []{{.OrderingTypeName}}
}

func (q {{.SelectQueryName}}) getLimitString() string {
	limitString := ""
	if q.LimitValue > 0 {
		limitString = fmt.Sprintf("limit %d", q.LimitValue)
	}
	return limitString
}

func (q {{.SelectQueryName}}) getOffsetString() string {
	offsetString := ""
	if q.OffsetValue > 0 {
		offsetString = fmt.Sprintf("offset %d", q.OffsetValue)
	}
	return offsetString
}

func (q {{.SelectQueryName}}) getOrderString() string {
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

func (q {{.SelectQueryName}}) getFieldsString() string {
	fields := []string{}
	for _, f := range q.Fields {
		fields = append(fields, string(f))
	}
	return strings.Join(fields, ",")
}

func (b {{.SelectQueryName}}) Fetch(ctx context.Context, connection qtypes.DBConnection) ([]{{.StructName}}, error) {
	if connection == nil {
		return nil, errors.New("Required to setup (SetDBConnection) connection before fetching")
	}
	values := []{{.StructName}}{}
	rows, err := connection.Query(ctx, b.SQL(), b.stmtParams...)
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
}

func (q {{.SelectQueryName}}) SQL() string {
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
}`

var INSERT_QUERY = `type {{.InsertQueryName}} struct {
	TableName    {{.TableTypeName}}
	InsertParams []qtypes.InsertParam
}

func (q {{.InsertQueryName}}) Exec(ctx context.Context, connection qtypes.DBConnection) (int, error) {
	if connection == nil {
		return 0, errors.New("Required to setup (SetDBConnection) connection before fetching")
	}
	var ret int
	err := connection.QueryRow(ctx, q.SQL(), q.getInsertStmtParams()...).Scan(&ret)
	if err != nil {
		return 0, fmt.Errorf("insert {{.StructName}} error: %w", err)
	}
	return ret, nil
}
func (b {{.InsertQueryName}}) getInsertPlaceholders() string {
	plc := 1
	placeholders := []string{}
	for range b.InsertParams {
		placeholders = append(placeholders, fmt.Sprintf("$%d", plc))
		plc = plc + 1
	}
	return strings.Join(placeholders, ",")
}

func (b {{.InsertQueryName}}) getInsertFieldsString() string {
	fields := []string{}
	for _, f := range b.InsertParams {
		fields = append(fields, string(f.Name))
	}
	return strings.Join(fields, ",")
}

func (b {{.InsertQueryName}}) getInsertStmtParams() []interface{} {
	params := []interface{}{}
	for _, f := range b.InsertParams {
		params = append(params, f.Value)
	}
	return params
}

func (b {{.InsertQueryName}}) SQL() string {
	fields := b.getInsertFieldsString()
	subQueryTemplate, _ := template.New("subquery").Parse("{{.InsertBaseQuery}}")
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
`

var UPDATE_QUERY = `type {{.UpdateQueryName}} struct {
	TableName              {{.TableTypeName}}
	WhereString            string
	UpdateParams           []qtypes.InsertParam
	whereStmtParams        []interface{}
	startPlaceholderNumber int
}

func (q {{.UpdateQueryName}}) Exec(ctx context.Context, connection qtypes.DBConnection) error {
	if connection == nil {
		return errors.New("Required to setup (SetDBConnection) connection before fetching")
	}
	_, err := connection.Exec(ctx, q.SQL(), q.getUpdateStmtParams()...)
	if err != nil {
		return fmt.Errorf("update {{.StructName}} error: %w", err)
	}
	return nil
}

func (b {{.UpdateQueryName}}) getUpdateFields() string {
	fields := []string{}
	plc := b.startPlaceholderNumber
	for _, f := range b.UpdateParams {
		fields = append(fields, fmt.Sprintf("%s = $%d", string(f.Name), plc))
		plc = plc + 1
	}
	return strings.Join(fields, ", ")
}

func (b {{.UpdateQueryName}}) getUpdateStmtParams() []interface{} {
	params := b.whereStmtParams
	for _, f := range b.UpdateParams {
		params = append(params, f.Value)
	}
	return params
}

func (b {{.UpdateQueryName}}) SQL() string {
	subQueryTemplate, _ := template.New("subquery").Parse("{{.UpdateBaseQuery}}")
	queryBuf := new(bytes.Buffer)
	_ = subQueryTemplate.Execute(queryBuf, map[string]string{
		"TableName": string(b.TableName),
		"Updates":   b.getUpdateFields(),
		"Where":     b.WhereString,
	})
	query := strings.TrimSpace(queryBuf.String())
	return query
}
`

var SELECT_BUILDER = `type {{.SelectBuilderName}} struct {
	Fields    []{{.FieldsTypeName}}
	TableName {{.TableTypeName}}
	{{.WhereUserBuilder}}
	LimitValue         int
	OffsetValue        int
	PreparedStmtParams []interface{}
	order              []{{.OrderingTypeName}}
}

func New{{.SelectBuilderName}}() *{{.SelectBuilderName}} {
	b := {{.SelectBuilderName}}{}
	b.Fields = []{{.FieldsTypeName}}{{.FieldsList}}
	b.TableName = {{.TableName}}
	b.Conditions = []{{.CondNodeTypeName}}{}
	b.PreparedStmtParams = []interface{}{}
	b.LimitValue = 0
	b.OffsetValue = 0
	b.order = []{{.OrderingTypeName}}{}
	return &b
}

func (b *{{.SelectBuilderName}}) Build() {{.SelectQueryName}} {
	cc := b.getWhereString(b.TableName)
	return {{.SelectQueryName}}{
		Fields:      b.Fields,
		stmtParams:  cc.stmtParams,
		WhereString: cc.sqlString,
		TableName:   b.TableName,
		LimitValue:  b.LimitValue,
		OffsetValue: b.OffsetValue,
		order:       b.order,
	}
}
func (b *{{.SelectBuilderName}}) Limit(limit int) *{{.SelectBuilderName}} {
	b.LimitValue = limit
	return b
}

func (b *{{.SelectBuilderName}}) Offset(offset int) *{{.SelectBuilderName}} {
	b.OffsetValue = offset
	return b
}

func (b *{{.SelectBuilderName}}) Where(conditions ...{{.CondNodeTypeName}}) *{{.SelectBuilderName}} {
	b.Conditions = conditions
	return b
}
func (b *{{.SelectBuilderName}}) OrderByDesc(field {{.FieldsTypeName}}) *{{.SelectBuilderName}} {
	b.order = append(b.order, {{.OrderingTypeName}}{Field: field, Direction: qtypes.OrderDesc})
	return b
}

func (b *{{.SelectBuilderName}}) OrderBy(field {{.FieldsTypeName}}) *{{.SelectBuilderName}} {
	b.order = append(b.order, {{.OrderingTypeName}}{Field: field, Direction: qtypes.OrderAsc})
	return b
}`

var INSERT_BUILDER = `type {{.InsertBuilderName}} struct {
	Fields       []{{.FieldsTypeName}}
	TableName    {{.TableTypeName}}
	InsertParams []qtypes.InsertParam
}

func New{{.InsertBuilderName}}() *{{.InsertBuilderName}} {
	b := {{.InsertBuilderName}}{}
	b.Fields = []{{.FieldsTypeName}}{{.FieldsList}}
	b.TableName = {{.TableName}}
	b.InsertParams = []qtypes.InsertParam{}
	return &b
}

func (b *{{.InsertBuilderName}}) Insert(u {{.StructName}}) *{{.InsertBuilderName}} {
	b.InsertParams = u.TSQBSaver()
	return b
}

func (b *{{.InsertBuilderName}}) Build() {{.InsertQueryName}} {
	return {{.InsertQueryName}}{
		TableName:    b.TableName,
		InsertParams: b.InsertParams,
	}
}`

var UPDATE_BUILDER = `type {{.UpdateBuilderName}} struct {
	TableName    {{.TableTypeName}}
	UpdateParams []qtypes.InsertParam
	{{.WhereUserBuilder}}
}


func New{{.UpdateBuilderName}}() *{{.UpdateBuilderName}} {
	b := {{.UpdateBuilderName}}{}
	b.TableName = {{.TableName}}
	return &b
}



func (b *{{.UpdateBuilderName}}) Where(conditions ...{{.CondNodeTypeName}}) *{{.UpdateBuilderName}} {
	b.Conditions = conditions
	return b
}

func (b *{{.UpdateBuilderName}}) UpdateAllFields(u {{.StructName}}) *{{.UpdateBuilderName}} {
	b.UpdateParams = append(
		b.UpdateParams,
		[]qtypes.InsertParam{
			{{.FieldsAsParams}}
		}...,
	)
	b.Where(b.CondEqID(u.ID))
	return b
}

func (b *{{.UpdateBuilderName}}) Build() {{.UpdateQueryName}} {
	cc := b.getWhereString(b.TableName)
	return {{.UpdateQueryName}}{
		WhereString:            cc.sqlString,
		TableName:              b.TableName,
		UpdateParams:           b.UpdateParams,
		whereStmtParams:        cc.stmtParams,
		startPlaceholderNumber: cc.lastPlcNumber,
	}
}
`

var WHERE_BUILDER = `type {{.WhereUserBuilder}} struct {
	Conditions []{{.CondNodeTypeName}}
}

func (b {{.WhereUserBuilder}}) getWhereString(tableName {{.TableTypeName}}) {{.CompeleteConditionType}} {
	cc := {{.CompeleteConditionType}}{
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

func (b *{{.WhereUserBuilder}}) ComposeAnd(conditions ...{{.CondNodeTypeName}}) {{.CondNodeTypeName}} {
	return b.compose(qtypes.WhereAnd, conditions...)
}

func (b *{{.WhereUserBuilder}}) ComposeOr(conditions ...{{.CondNodeTypeName}}) {{.CondNodeTypeName}} {
	return b.compose(qtypes.WhereOr, conditions...)
}

func (b *{{.WhereUserBuilder}}) compose(w qtypes.ComposeMethod, conditions ...{{.CondNodeTypeName}}) {{.CondNodeTypeName}} {
	cn := {{.CondNodeTypeName}}{
		Conditions:    []{{.ConditionTypeName}}{},
		Nodes:         []{{.CondNodeTypeName}}{},
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
}`

var CONDITION_TYPES = `type {{.ConditionTypeName}} struct {
	Field     {{.FieldsTypeName}}
	Func      qtypes.EqualConditions
	CompareTo interface{}
}

type {{.CondNodeTypeName}} struct {
	Conditions    []{{.ConditionTypeName}}
	ComposeMethod qtypes.ComposeMethod
	Nodes         []{{.CondNodeTypeName}}
	Not           bool
}

func (c {{.ConditionTypeName}}) BuildCond(nextPlcNumber int, tableName string) {{.CompeleteConditionType}} {
	subQueryTemplate, _ := template.New("subquery").Parse("{{.ConditionQuery}}")
	queryBuf := new(bytes.Buffer)
	_ = subQueryTemplate.Execute(queryBuf, map[string]string{
		"Table": tableName,
		"Field": string(c.Field),
		"Func":  string(c.Func),
		"Value": fmt.Sprintf("%d", nextPlcNumber),
	})
	return {{.CompeleteConditionType}}{
		sqlString:     queryBuf.String(),
		lastPlcNumber: nextPlcNumber + 1,
		stmtParams:    []interface{}{c.CompareTo},
	}
}

type {{.CompeleteConditionType}} struct {
	sqlString     string
	lastPlcNumber int
	stmtParams    []interface{}
}

func (cn {{.CondNodeTypeName}}) BuildCond(nextPlcNumber int, tableName string) {{.CompeleteConditionType}} {
	conditions := []string{}
	cc := {{.CompeleteConditionType}}{}
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
}`

var CONDITION_METHOD = `func (b *{{.WhereUserBuilder}}) Cond{{.CondKey}}{{.FieldName}}(compareTo {{.FieldType}}) {{.CondNodeTypeName}} {
	c := {{.ConditionTypeName}}{
		Field:     {{.GenFieldName}},
		Func:      {{.CondValue}},
		CompareTo: compareTo,
	}
	cn := {{.CondNodeTypeName}}{
		Conditions: []{{.ConditionTypeName}}{c},
	}
	return cn
}`

var BASIC_TYPE = `type {{.FieldsTypeName}} string

var (
	{{.FieldsDeclaration}}
)

type {{.TableTypeName}} string

var (
	{{.TableName}} {{.TableTypeName}} = "{{.SQLTableName}}"
)`

var ORDER_TEMPLATE = `type {{.OrderingTypeName}} struct {
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

var ORIGINAL_STRUCT_METHODS = `func (u *{{.StructName}}) TSQBScanner() []interface{} {
	return []interface{}{
		{{.FieldsPointers}}
	}
}

func (u {{.StructName}}) TSQBSaver() []qtypes.InsertParam {
	params := []qtypes.InsertParam{}
	if u.ID > 0 {
		params = append(params, qtypes.InsertParam{{.FieldIDAsParam}})
	}
	params = append(params, []qtypes.InsertParam{
		{{.FieldsAsParams}}
	}...)
	return params
}`

var UPDATE_METHOD = `func (b *{{.UpdateBuilderName}}) {{.UpdateMethodName}}(value string) *{{.UpdateBuilderName}} {
	b.UpdateParams = append(
		b.UpdateParams,
		qtypes.InsertParam{
			Name:  string({{.FieldName}}),
			Value: value,
		},
	)
	return b
}`

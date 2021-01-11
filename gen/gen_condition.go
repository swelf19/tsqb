package gen

// func (s StructMeta) genConditions() string {
// 	conditionTemplateString := "{{.Table}}.{{.Field}} {{.Func}} {{.Value}}"
// 	templ := `type {{.ConditionTypeName}} struct {
// 		Table {{.TableTypeName}}
// 		Field {{.FieldsTypeName}}
// 		Func  qtypes.EqualConditions
// 		Value string
// 	}
// 	func (c {{.ConditionTypeName}}) String() string {
// 		subQueryTemplate, _ := template.New("subquery").Parse("{{.ConditionTemplateString}}")
// 		queryBuf := new(bytes.Buffer)
// 		_ = subQueryTemplate.Execute(queryBuf, map[string]string{
// 			"Table": string(c.Table),
// 			"Field": string(c.Field),
// 			"Func":  string(c.Func),
// 			"Value": c.Value,
// 		})
// 		return queryBuf.String()
// 	}
// 	type {{.CondNodeTypeName}} struct {
// 		Conditions []{{.ConditionTypeName}}
// 		WhereLink  qtypes.WhereLinks
// 		Nodes      []{{.CondNodeTypeName}}
// 		Not        bool
// 	}
// 	func (cn {{.CondNodeTypeName}}) String() string {
// 		conditions := []string{}
// 		for _, c := range cn.Conditions {
// 			conditions = append(conditions, c.String())
// 		}
// 		for _, n := range cn.Nodes {
// 			conditions = append(conditions, n.String())
// 		}
// 		condTemplate := "%s"
// 		if len(conditions) > 1 {
// 			condTemplate = "(%s)"
// 		}
// 		if cn.Not {
// 			condTemplate = "not " + condTemplate
// 		}
// 		return fmt.Sprintf(condTemplate, strings.Join(conditions, fmt.Sprintf(" %s ", cn.WhereLink)))
// 	}`
// 	subQueryTemplate, err := template.New("subquery").Parse(templ)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	queryBuf := new(bytes.Buffer)
// 	err = subQueryTemplate.Execute(queryBuf, map[string]string{
// 		"ConditionTypeName":       s.getConditionTypeName(),
// 		"TableTypeName":           s.getTableTypeName(),
// 		"FieldsTypeName":          s.getFieldsTypeName(),
// 		"CondNodeTypeName":        s.getCondNodeTypeName(),
// 		"ConditionTemplateString": conditionTemplateString,
// 	})
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	return queryBuf.String()
// }

package gen

// func (s StructMeta) genScannerSaverMethods() string {
// 	fields := []string{}
// 	pairFields := []string{}
// 	for _, f := range s.Fields {
// 		fields = append(fields, "&u."+f.FieldName+",")
// 		if f.FieldName != "ID" {
// 			pairFields = append(pairFields, "{Name: \""+f.SqlFieldName+"\", Value: u."+f.FieldName+"},")
// 		}
// 	}
// 	templ := `func (u *{{.StructName}}) TSQBScanner() []interface{} {
// 		return []interface{}{
// 			{{.Fields}}
// 		}
// 	}

// 	func (u {{.StructName}}) TSQBSaver() []qtypes.InsertParam {
// 		params := []qtypes.InsertParam{}
// 		if u.ID > 0 {
// 			params = append(params, qtypes.InsertParam{Name: "id", Value: u.ID})
// 		}
// 		params = append(params, []qtypes.InsertParam{
// 			{{.KeyValueFields}}
// 		}...)
// 		return params
// 	}`
// 	subQueryTemplate, err := template.New("subquery").Parse(templ)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	queryBuf := new(bytes.Buffer)
// 	err = subQueryTemplate.Execute(queryBuf, map[string]string{
// 		"StructName":     s.StructName,
// 		"Fields":         strings.Join(fields, ""),
// 		"KeyValueFields": strings.Join(pairFields, "\n"),
// 	})
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	return queryBuf.String()
// }

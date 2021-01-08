package gen

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

func (s StructMeta) genAdditionalStructMethods() string {
	fields := []string{}
	pairFields := []string{}
	for _, f := range s.Fields {
		fields = append(fields, "&u."+f.FieldName+",")
		pairFields = append(pairFields, "\""+f.SqlFieldName+"\":"+"u."+f.FieldName+",")
	}
	templ := `func (u *{{.StructName}}) TSQBScanner() []interface{} {
		return []interface{}{
			{{.Fields}}
		}
	}
	
	func (u {{.StructName}}) TSQBSaver() map[string]interface{} {
		return map[string]interface{}{
			{{.KeyValueFields}}
		}
	}`
	subQueryTemplate, err := template.New("subquery").Parse(templ)
	if err != nil {
		fmt.Println(err)
	}
	queryBuf := new(bytes.Buffer)
	err = subQueryTemplate.Execute(queryBuf, map[string]string{
		"StructName":     s.StructName,
		"Fields":         strings.Join(fields, ""),
		"KeyValueFields": strings.Join(pairFields, "\n"),
	})
	if err != nil {
		fmt.Println(err)
	}
	return queryBuf.String()
}

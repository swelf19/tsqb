package gen

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

func (s StructMeta) genComprationForField(condKey string, condValue string, f StructFieldMeta) string {
	templ := FIELD_CONDITION
	subQueryTemplate, err := template.New("subquery").Parse(templ)
	if err != nil {
		fmt.Println(err)
	}
	queryBuf := new(bytes.Buffer)
	err = subQueryTemplate.Execute(queryBuf, map[string]string{
		"StructNameLower": PascalToCamelCase(s.StructName),
		"FieldName":       f.FieldName,
		"CondKey":         condKey,
		"CondValue":       condValue,
		"FieldType":       f.Type,
	})
	if err != nil {
		fmt.Println(err)
	}
	return queryBuf.String()
}

func (s StructMeta) genInCondition(f StructFieldMeta) string {
	templ := FIELD_CONDITION_IN
	subQueryTemplate, err := template.New("subquery").Parse(templ)
	if err != nil {
		fmt.Println(err)
	}
	queryBuf := new(bytes.Buffer)
	err = subQueryTemplate.Execute(queryBuf, map[string]string{
		"StructNameLower": PascalToCamelCase(s.StructName),
		"FieldName":       f.FieldName,
		"FieldType":       f.Type,
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
		helpers = append(helpers, s.genInCondition(f))
	}
	return strings.Join(helpers, "\n")
}

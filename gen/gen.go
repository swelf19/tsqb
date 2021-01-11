package gen

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

/*

На основе структуры StructName

создаем Builder

А так еж к структуре привязываем методы для сканирования

*/

type StructFieldMeta struct {
	FieldName    string
	Type         string
	SqlFieldName string
}

type StructMeta struct {
	StructName   string
	TableName    string
	Fields       []StructFieldMeta
	PackageName  string
	ExtraImports map[string]string
}

func (s StructMeta) getFieldsList() []string {
	fields := []string{}
	for _, f := range s.Fields {
		fields = append(fields, s.getFieldName(f))
	}
	return fields
}

func (s StructMeta) getFieldPointers() []string {
	fields := []string{}
	for _, f := range s.Fields {
		fields = append(fields, "&u."+f.FieldName)
	}
	return fields
}

func (s StructMeta) getFieldsAsConstantsDeclaration() []string {
	fields := []string{}
	for _, f := range s.Fields {
		fields = append(fields,
			fmt.Sprintf(
				`%s %s = "%s"`,
				s.getFieldName(f),
				s.getFieldsTypeName(),
				f.SqlFieldName,
			),
		)
	}
	return fields
}

func (s StructMeta) getFieldsAsInsertParamsExceptID() []string {
	fields := []string{}
	for _, f := range s.Fields {
		if f.SqlFieldName == "id" {
			continue
		}
		fields = append(fields, s.getFieldAsInsertParam(f.SqlFieldName)+",")

	}
	return fields
}

func (s StructMeta) getFieldAsInsertParam(sqlfieldname string) string {
	for _, f := range s.Fields {
		if f.SqlFieldName == sqlfieldname {
			return fmt.Sprintf("{Name: string(%s), Value: u.%s}", s.getFieldName(f), f.FieldName)
		}
	}
	return ""
}

func (s StructMeta) getCompleteConditionType() string {
	return s.StructName + "CompleteCondition"
}

func (s StructMeta) getConditionTypeName() string {
	return s.StructName + "Condition"
}

func (s StructMeta) getCondNodeTypeName() string {
	return s.StructName + "CondNode"
}

func (s StructMeta) getFieldName(f StructFieldMeta) string {
	return s.StructName + "Field" + f.FieldName
}

func (s StructMeta) getFieldsTypeName() string {
	return s.StructName + "Fields"
}

func (s StructMeta) getTableTypeName() string {
	return s.StructName + "TableNameType"
}
func (s StructMeta) getTableVariableName() string {
	return s.StructName + "TableName"
}
func (s StructMeta) getOrderTypeName() string {
	return s.StructName + "OrderCond"
}

func (s StructMeta) getSelectBuilderName() string {
	return s.StructName + "SelectBuilder"
}
func (s StructMeta) getInsertBuilderName() string {
	return s.StructName + "InsertBuilder"
}
func (s StructMeta) getUpdateBuilderName() string {
	return s.StructName + "UpdateBuilder"
}
func (s StructMeta) getWhereBuilderName() string {
	return "where" + s.StructName + "Builder"
}

func (s StructMeta) getSelectQueryName() string {
	return s.StructName + "SelectQuery"
}
func (s StructMeta) getInsertQueryName() string {
	return s.StructName + "InsertQuery"
}
func (s StructMeta) getUpdateQueryName() string {
	return s.StructName + "UpdateQuery"
}

func (s StructMeta) genFieldConstantsBlock() string {
	lines := []string{"type " + s.getFieldsTypeName() + " string"}
	lines = append(lines, "var (")
	for _, f := range s.Fields {
		lines = append(lines, s.getFieldName(f)+" "+s.getFieldsTypeName()+" = "+"\""+f.SqlFieldName+"\"")
	}
	lines = append(lines, ")")
	return strings.Join(lines, "\n")
}

func (s StructMeta) genTableBlock() string {
	lines := []string{"type " + s.getTableTypeName() + " string"}
	lines = append(lines, "var (")
	lines = append(lines, s.getTableVariableName()+" "+s.getTableTypeName()+" = "+"\""+s.TableName+"\"")
	lines = append(lines, ")")
	return strings.Join(lines, "\n")
}

func (s StructMeta) getUpdateMethod(f StructFieldMeta) string {

	templ := UPDATE_METHOD
	subQueryTemplate, err := template.New("subquery").Parse(templ)
	if err != nil {
		fmt.Println(err)
	}
	queryBuf := new(bytes.Buffer)
	err = subQueryTemplate.Execute(queryBuf, map[string]string{
		"FieldName":         s.getFieldName(f),
		"UpdateBuilderName": s.getUpdateBuilderName(),
		"UpdateMethodName":  "Update" + f.FieldName,
	})
	if err != nil {
		fmt.Println(err)
	}
	return queryBuf.String()
}

func (s StructMeta) genUpdateMethods() string {
	methods := []string{}
	for _, f := range s.Fields {
		if f.SqlFieldName == "id" {
			continue
		}
		methods = append(methods, s.getUpdateMethod(f))
	}
	return strings.Join(methods, "\n")
}

func (s StructMeta) GenEntyreTemplate() string {
	templates := []string{
		BASIC_TYPE,
		CONDITION_TYPES,
		ORDER_TEMPLATE,
		SELECT_BUILDER,
		INSERT_BUILDER,
		UPDATE_BUILDER,
		s.genUpdateMethods(),
		WHERE_BUILDER,
		s.genComprationHelpers(),
		SELECT_QUERY,
		INSERT_QUERY,
		UPDATE_QUERY,
		ORIGINAL_STRUCT_METHODS,
	}
	return s.genCodeTemplate(strings.Join(templates, "\n"))
}

func (s StructMeta) genCodeTemplate(code string) string {
	subQueryTemplate, err := template.New("subquery").Parse(code)
	if err != nil {
		fmt.Println(err)
	}
	queryBuf := new(bytes.Buffer)
	err = subQueryTemplate.Execute(queryBuf, map[string]string{
		"OrderingTypeName": s.getOrderTypeName(),
		"FieldsTypeName":   s.getFieldsTypeName(),

		"SelectBuilderName": s.getSelectBuilderName(),
		"InsertBuilderName": s.getInsertBuilderName(),
		"UpdateBuilderName": s.getUpdateBuilderName(),
		"WhereUserBuilder":  s.getWhereBuilderName(),
		"SelectQueryName":   s.getSelectQueryName(),
		"InsertQueryName":   s.getInsertQueryName(),
		"UpdateQueryName":   s.getUpdateQueryName(),

		"CondNodeNameType": s.getCondNodeTypeName(),
		"TableTypeName":    s.getTableTypeName(),
		"TableName":        s.getTableVariableName(),
		// "InsertQuery":             insertQuery,
		"StructName":             s.StructName,
		"FieldsList":             "{" + strings.Join(s.getFieldsList(), ",") + "}",
		"FieldsAsParams":         strings.Join(s.getFieldsAsInsertParamsExceptID(), "\n"),
		"FieldIDAsParam":         s.getFieldAsInsertParam("id"),
		"CompeleteConditionType": s.getCompleteConditionType(),
		"CondNodeTypeName":       s.getCondNodeTypeName(),
		"ConditionTypeName":      s.getConditionTypeName(),
		// "ConditionTemplateString": conditionTemplateString,
		"Fields": strings.Join(s.getFieldsList(), ", "),
		// "KeyValueFields":    strings.Join(pairFields, "\n"),
		"InsertBaseQuery":   "insert into {{.TableName}}({{.Fields}}) values({{.Placeholders}}) returning {{.ReturningfField}}",
		"UpdateBaseQuery":   "update {{.TableName}} set {{.Updates}} {{.Where}}",
		"ConditionQuery":    "{{.Table}}.{{.Field}} {{.Func}} ${{.Value}}",
		"SQLTableName":      s.TableName,
		"FieldsDeclaration": strings.Join(s.getFieldsAsConstantsDeclaration(), "\n"),
		"FieldsPointers":    strings.Join(s.getFieldPointers(), ", ") + ",",
	})
	if err != nil {
		fmt.Println(err)
	}
	return queryBuf.String()
}

/*

package sample

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/swelf19/tsqb/qtypes"
)

BASIC_TYPE
ORDER_TEMPLATE

SELECT_QUERY
INSERT_QUERY
UPDATE_QUERY
SELECT_BUILDER
INSERT_BUILDER
UPDATE_BUILDER
WHERE_BUILDER
CONDITION_TYPES

ORIGINAL_STRUCT_METHODS

CONDITION_METHOD



*/

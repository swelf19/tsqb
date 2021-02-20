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
	FieldName        string
	Type             string
	SqlFieldName     string
	RelatedModelName string
	TableName        string
	FieldNameSpace   string
	OrigFieldName    string
}

type StructTypeEnum string

var (
	RegularStruct StructTypeEnum = "regularstruct" //Обыкновенная структура
	JoinedStruct  StructTypeEnum = "joinedstruct"  //результат джойна 2х или более структур
)

type StructMeta struct {
	StructType   StructTypeEnum
	StructName   string
	TableName    string
	Fields       []StructFieldMeta
	PackageName  string
	ExtraImports map[string]string
	JoinMembers  []StructMeta //С кем таблицу можно джойнить исходя из тегов tsqb:fk в полях структуы
	// ForeignKeyTo []StructMeta //
	SelectOnly bool
}

func (s StructMeta) getFields() []StructFieldMeta {
	return s.Fields
}

func PascalToCamelCase(psc string) string {
	return string(append([]byte(strings.ToLower(string(psc[:1]))), psc[1:]...))
}

func (s StructMeta) getFieldsTypeName() string {
	return PascalToCamelCase(s.StructName) + "Fields"
}

func (s StructMeta) GenNewEntryTemplate() string {
	templates := []string{
		FIELDS,
		s.genComprationHelpers(),
		ALLSCHEMAS,
		BUILDER,
		INSERTBUIDLER,
		UPDATEBUILDER,
		DELETEBUILDER,
		SELECTQUERY,
		INSERTQUERY,
		UPDATEQUERY,
		DELETEQUERY,
		DBMETHODS,
	}

	return s.genCodeTemplate(strings.Join(templates, "\n"))
}

func (s StructMeta) genCodeTemplate(code string) string {
	subQueryTemplate, err := template.New("subquery").Parse(code)
	if err != nil {
		fmt.Println(err)
	}
	queryBuf := new(bytes.Buffer)
	err = subQueryTemplate.Execute(queryBuf, map[string]interface{}{
		"StructMeta":      s,
		"FieldsTypeName":  s.getFieldsTypeName(),
		"FieldsData":      s.getFields(),
		"StructName":      s.StructName,
		"StructNameLower": PascalToCamelCase(s.StructName),
		"SQLTableName":    s.TableName,
	})
	if err != nil {
		fmt.Println(err)
	}
	return queryBuf.String()
}

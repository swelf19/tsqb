package gen

import (
	"strings"
)

/*

На основе структуры StructName

создаем Builder

А так еж к структуре привязываем методы для сканирования

*/

// type UserFields string
// type TableName string

// var (
// 	UserFieldID       UserFields = "id"
// 	UserFieldUserName UserFields = "username"
// 	UserFieldLastLog  UserFields = "last_log"

// 	UserTableName TableName = "users"
// )

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

/*

package sample

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/swelf19/tsqb/qtypes"
)






*/

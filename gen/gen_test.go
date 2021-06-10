package gen

import (
	"fmt"
	"go/format"
	"testing"

	"github.com/kylelemons/godebug/diff"
	"github.com/stretchr/testify/suite"
)

type GenCodeTestSuite struct {
	suite.Suite
}

func (suite *GenCodeTestSuite) SetupTest() {

}

//FormatAndCompare - Приводит код к форматированныому виду при помощи go/format,
// после чего сравнивает методом suite.Equal
func (suite *GenCodeTestSuite) FormatAndCompare(expected, actual string) bool {
	a1, err := format.Source([]byte(expected))

	suite.Assert().NoError(err)
	a2, err := format.Source([]byte(actual))
	suite.Assert().NoError(err)
	// fmt.Println(actual)
	// fmt.Println(string(a2))
	fmt.Println(diff.Diff(string(a1), string(a2)))
	return suite.Equal(a1, a2)
}

func (suite *GenCodeTestSuite) DisabledTestGenJoinedStructs() {
	roleStruct := StructMeta{
		StructType:  RegularStruct,
		StructName:  "Role",
		TableName:   "roles",
		PackageName: "gen",
		JoinMembers: []StructMeta{},
		SelectOnly:  false,
		Fields: []StructFieldMeta{
			{
				FieldName:        "ID",
				Type:             "int",
				SqlFieldName:     "id",
				RelatedModelName: "",
			},
			{
				FieldName:        "Title",
				Type:             "string",
				SqlFieldName:     "title",
				RelatedModelName: "",
			},
		},
	}

	userStruct := StructMeta{
		StructType:  RegularStruct,
		StructName:  "User",
		TableName:   "users",
		PackageName: "gen",
		JoinMembers: []StructMeta{},
		SelectOnly:  false,
		Fields: []StructFieldMeta{
			{
				FieldName:        "ID",
				Type:             "int",
				SqlFieldName:     "id",
				RelatedModelName: "",
			},
			{
				FieldName:        "Name",
				Type:             "string",
				SqlFieldName:     "name",
				RelatedModelName: "",
			},
			{
				FieldName:        "RoleID",
				Type:             "int",
				SqlFieldName:     "role_id",
				RelatedModelName: "Role",
			},
		},
	}

	joinedStruct := StructMeta{
		StructType:  JoinedStruct,
		StructName:  "UserJoinRole",
		TableName:   "users join roles on (users.role_id = roles.id)",
		PackageName: "gen",
		JoinMembers: []StructMeta{userStruct, roleStruct},
		SelectOnly:  true,
		Fields: []StructFieldMeta{
			{
				FieldName:        "UserID",
				Type:             "int",
				SqlFieldName:     "id",
				RelatedModelName: "",
			},
			{
				FieldName:        "UserName",
				Type:             "string",
				SqlFieldName:     "name",
				RelatedModelName: "",
			},
			{
				FieldName:        "UserRoleID",
				Type:             "int",
				SqlFieldName:     "role_id",
				RelatedModelName: "",
			},
			{
				FieldName:        "RoleID",
				Type:             "int",
				SqlFieldName:     "id",
				RelatedModelName: "",
			},
			{
				FieldName:        "RoleTitle",
				Type:             "string",
				SqlFieldName:     "title",
				RelatedModelName: "",
			},
		},
	}
	actual := []StructMeta{userStruct, roleStruct, joinedStruct}
	expected := JoinTables([]StructMeta{userStruct, roleStruct})

	suite.Equal(expected, actual)

}

func TestCodeGeneration(t *testing.T) {
	suite.Run(t, new(GenCodeTestSuite))
}

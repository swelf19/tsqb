package tsqbparser

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/swelf19/tsqb/gen"
)

type StructParserTestSuite struct {
	suite.Suite
	path string
}

func (suite *StructParserTestSuite) SetupTest() {
	suite.path = "test_data.txt"
}

func (suite *StructParserTestSuite) TestImportParsing() {
	expected := "pgtype"
	actual := parseImportNameFromPath(`"github.com/jackc/pgtype"`)
	suite.Equal(expected, actual)

	expected = "pgx"
	actual = parseImportNameFromPath(`"github.com/jackc/pgx/v4"`)
	suite.Equal(expected, actual)

	expected = "v4v"
	actual = parseImportNameFromPath(`"github.com/jackc/pgx/v4v"`)
	suite.Equal(expected, actual)
}

func (suite *StructParserTestSuite) TestExtraImports() {
	expected := true
	actual := isExtraImportRequered("pgtype.Timestamp")
	suite.Equal(expected, actual)
	expected = false
	actual = isExtraImportRequered("int")
	suite.Equal(expected, actual)
}

func (suite *StructParserTestSuite) TestParseAstFileToMetaStruct() {
	st := []gen.StructMeta{
		{
			StructName: "User",
			TableName:  "users",
			Fields: []gen.StructFieldMeta{
				{
					FieldName:    "ID",
					Type:         "int",
					SqlFieldName: "id",
				},
				{
					FieldName:    "UserName",
					Type:         "string",
					SqlFieldName: "username",
				},
				{
					FieldName:    "LastLog",
					Type:         "pgtype.Timestamptz",
					SqlFieldName: "last_log",
				},
			},
		},
	}
	parsed, err := ParseAST(suite.path)
	suite.Assert().NoError(err)
	suite.Equal(st, parsed.StructMetaList)
	expected := map[string]string{
		"pgtype": "github.com/jackc/pgtype",
	}
	actual := parsed.ExtraImports
	suite.Equal(expected, actual)
}

func (suite *StructParserTestSuite) TestCommentProcessors() {
	actual := extractTSQBDirective("//tsqb:gen")
	expect := TSQBCommand{
		Command: GEN_DIRECTIVE,
	}
	suite.Equal(expect, actual)

	actual = extractTSQBDirective("//tsqb:command")
	expect = TSQBCommand{
		Command: "command",
	}
	suite.Equal(expect, actual)

	actual = extractTSQBDirective("//Regular comment")
	expect = TSQBCommand{}
	suite.Equal(expect, actual)

	actual = extractTSQBDirective("//tsqb:tablename=testtable")
	expect = TSQBCommand{
		Command: "tablename",
		Value:   "testtable",
	}
	suite.Equal(expect, actual)

	suite.Equal(true, isGenSignarutePresent("//tsqb:gen"))

	suite.Equal(false, isGenSignarutePresent("//tsqb:gen1"))
}

func (suite *StructParserTestSuite) TestTagParser() {
	actual := parseStructTag("`tsqb:\"col=id\"`")
	expected := TSQBTag{ColName: "id"}
	suite.Equal(expected, actual)

	actual = parseStructTag(``)
	expected = TSQBTag{ColName: ""}
	suite.Equal(expected, actual)
}

func TestCodeGeneration(t *testing.T) {
	suite.Run(t, new(StructParserTestSuite))
}

package tsqbparser

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"reflect"
	"regexp"
	"strings"

	"github.com/swelf19/tsqb/gen"
)

var TSQB_PREFIX = "tsqb"

var GEN_DIRECTIVE = "gen"
var TABLE_NAME_FIELD = "tablename"

var COLNAME_TAG = "col"
var FOREIGNKEY_TAG = "fk"

type TSQBCommand struct {
	Command string
	Value   string
}

//extractTSQBDirective - на вход получает строку, из строки вида
//tsqb:command извлекает command, если строка не соответствует паттерну //tsqb:
//возвращает пустую строку
func extractTSQBDirective(comment string) TSQBCommand {
	command := TSQBCommand{}
	commentPrefix := fmt.Sprintf("//%s:", TSQB_PREFIX)
	if strings.HasPrefix(comment, commentPrefix) {
		strippedComment := strings.TrimPrefix(comment, commentPrefix)
		cmd := strings.Split(strippedComment, "=")[0]
		value := ""
		if len(strings.Split(strippedComment, "=")) > 1 {
			value = strings.Split(strippedComment, "=")[1]
		}
		command = TSQBCommand{
			Command: cmd,
			Value:   value,
		}
	}
	return command
}

//isGenSignarutePresent - проверяет наличие сигнатцры для генерации структуры
//tsqb:gen
func isGenSignarutePresent(comment string) bool {
	directive := extractTSQBDirective(comment)
	return directive.Command == GEN_DIRECTIVE

}

func isDeclTSQBCompatible(decl *ast.GenDecl) bool {
	if decl.Doc == nil {
		return false
	}
	for _, d := range decl.Doc.List {
		if isGenSignarutePresent(d.Text) {
			return true
		}
	}
	return false
}

func getTSQBCommands(decl *ast.GenDecl) []TSQBCommand {
	commands := []TSQBCommand{}
	for _, d := range decl.Doc.List {
		cmd := extractTSQBDirective(d.Text)
		if cmd.Command != "" {
			commands = append(commands, cmd)
		}
	}
	return commands
}

type TSQBTag struct {
	ColName string
	Related string
}

func parseStructTag(tagValue string) TSQBTag {
	tsqbtag := TSQBTag{}
	if len(tagValue) > 1 {
		// fmt.Println(tagValue)

		tag := reflect.StructTag(tagValue[1 : len(tagValue)-1])
		rawvalue := tag.Get(TSQB_PREFIX)

		for _, kv := range strings.Split(rawvalue, ",") {
			if len(strings.Split(kv, "=")) == 2 {
				if strings.Split(kv, "=")[0] == COLNAME_TAG {
					tsqbtag.ColName = strings.Split(kv, "=")[1]
				} else if strings.Split(kv, "=")[0] == FOREIGNKEY_TAG {
					tsqbtag.Related = strings.Split(kv, "=")[1]
				}
			}
		}
	}
	return tsqbtag
}

func parseTypeName(fset *token.FileSet, field *ast.Field) string {
	var typeNameBuf bytes.Buffer
	err := printer.Fprint(&typeNameBuf, fset, field.Type)
	if err != nil {
		log.Fatalf("failed printing %s", err)
		return ""
	}
	// fmt.Printf("Type:   %s\n", typeNameBuf.String())
	return typeNameBuf.String()
}

func parseField(f *ast.Field, fset *token.FileSet) gen.StructFieldMeta {
	// fmt.Println(f.Type)
	tsqbTag := parseStructTag(f.Tag.Value)
	meta := gen.StructFieldMeta{
		FieldName:        f.Names[0].Name,
		OrigFieldName:    f.Names[0].Name,
		Type:             parseTypeName(fset, f),
		SqlFieldName:     tsqbTag.ColName,
		RelatedModelName: tsqbTag.Related,
	}
	return meta
}

// func getStructMeat(s *ast.StructType) gen.StructMeta {

// }

func trimWrapQuotes(src string) string {
	dst := strings.TrimPrefix(src, `"`)
	dst = strings.TrimSuffix(dst, `"`)
	return dst
}

func parseImportNameFromPath(importPath string) string {
	importPath = trimWrapQuotes(importPath)
	name := strings.Split(importPath, "/")[len(strings.Split(importPath, "/"))-1]
	match, _ := regexp.MatchString(`^v\d+$`, name)
	if match {
		name = strings.Split(importPath, "/")[len(strings.Split(importPath, "/"))-2]
	}
	return name
}

func parseImport(i *ast.ImportSpec) (importName string, importPath string) {
	importPath = trimWrapQuotes(i.Path.Value)
	if i.Name != nil {
		importName = i.Name.Name
	} else {
		importName = parseImportNameFromPath(importPath)
	}
	return importName, importPath
}

type ParseResult struct {
	StructMetaList []gen.StructMeta
	ExtraImports   map[string]string
}

func getImports(f *ast.File) map[string]string {
	extraImports := map[string]string{}
	for _, i := range f.Imports {
		importName, importPath := parseImport(i)
		extraImports[importName] = importPath
	}
	return extraImports
}

func isExtraImportRequered(fieldType string) bool {
	return len(strings.Split(fieldType, ".")) == 2
}

func GetPackageName(path string) string {
	fset := token.NewFileSet() // positions are relative to fset
	astFile, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		log.Println(err)
		return ""
	}
	return astFile.Name.Name
}

func ParseAST(path string) (*ParseResult, error) {
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	fileImports := getImports(f)
	extraImports := map[string]string{}
	pr := ParseResult{}

	structs := []gen.StructMeta{}
	for _, decl := range f.Decls {
		switch declNode := decl.(type) {
		case *ast.GenDecl:
			if !isDeclTSQBCompatible(declNode) {
				continue
			}
			commands := getTSQBCommands(declNode)
			for _, spec := range declNode.Specs {
				switch node := spec.(type) {
				case *ast.TypeSpec:
					switch astStruct := node.Type.(type) {
					case *ast.StructType:
						s := gen.StructMeta{
							StructType: gen.RegularStruct,
							StructName: node.Name.String(),
						}
						for _, c := range commands {
							if c.Command == TABLE_NAME_FIELD {
								s.TableName = c.Value
							}
						}
						for _, field := range astStruct.Fields.List {
							m := parseField(field, fset)
							m.TableName = s.TableName
							m.FieldNameSpace = s.StructName
							s.Fields = append(s.Fields, m)
							if isExtraImportRequered(m.Type) {
								importName := strings.Split(m.Type, ".")[0]
								extraImports[importName] = fileImports[importName]
							}
						}
						structs = append(structs, s)
					}

				}
			}
		}
	}
	pr.ExtraImports = extraImports
	pr.StructMetaList = structs
	return &pr, nil
}

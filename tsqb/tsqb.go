package main

import (
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/swelf19/tsqb/gen"
	"github.com/swelf19/tsqb/tsqbparser"
)

func makefilename(path string) string {
	newpath := strings.TrimSuffix(path, ".go") + "_gen.go"
	return newpath
}

func getPackageName(path string) string {
	fset := token.NewFileSet() // positions are relative to fset
	astFile, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		log.Println(err)
		return ""
	}
	return astFile.Name.Name
}

func main() {
	err := filepath.Walk(os.Args[1],
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				pr, err := tsqbparser.ParseAST(path)
				structs := pr.StructMetaList
				if err != nil {
					log.Println(err)
					return err
				}
				if len(structs) > 0 {
					newFileName := makefilename(path)
					// fmt.Println(newFileName)
					packageName := getPackageName(path)
					filemeta := gen.NewFileMeta(structs, packageName, path, pr.ExtraImports)
					fileContent := filemeta.GenFileContent()
					err := ioutil.WriteFile(newFileName, []byte(fileContent), 0644)
					if err != nil {
						return err
					}
				}
				// for _, s := range structs {
				// 	fmt.Println(s.StructName)
				// }
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}
}

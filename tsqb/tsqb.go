package main

import (
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

func main() {
	err := filepath.Walk(os.Args[1],
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				if !strings.HasSuffix(path, ".go") {
					return nil
				}
				pr, err := tsqbparser.ParseAST(path)
				if err != nil {
					log.Println(err)
					return err
				}
				if len(pr.StructMetaList) > 0 {
					// structs := gen.JoinTables(pr.StructMetaList)
					structs := pr.StructMetaList
					if len(structs) > 0 {
						newFileName := makefilename(path)
						// fmt.Println(newFileName)
						packageName := tsqbparser.GetPackageName(path)
						filemeta := gen.NewFileMeta(structs, packageName, path, pr.ExtraImports)
						fileContent := filemeta.GenFileContent()
						err := ioutil.WriteFile(newFileName, []byte(fileContent), 0644)
						if err != nil {
							return err
						}
					}
				}
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}
}

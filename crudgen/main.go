package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

const (
	outputFilename = "z_crud2.go"
	structTagName  = "crud"
)

type StructType struct {
	TypeSpec   *ast.TypeSpec
	StructType *ast.StructType

	Name   string
	Fields []StructField
}

func (structType StructType) Metadata() string {
	if len(structType.Fields) == 0 {
		return ""
	}

	bs := &bytes.Buffer{}

	if er := structTemplate.Execute(bs, structType); er != nil {
		panic(er)
	}

	return bs.String()
}

type StructField struct {
	Name    string
	SqlName string
	Type    ast.Expr
}

func fileFilter(fi os.FileInfo) bool {
	name := fi.Name()

	if !strings.HasSuffix(name, ".go") {
		return false
	}

	return fi.Name() != outputFilename
}

func main() {
	dirPath := "."

	if len(os.Args) > 1 {
		dirPath = os.Args[1]
	}

	fset := new(token.FileSet)

	pkgs, er := parser.ParseDir(fset, dirPath, fileFilter, 0)
	if er != nil {
		log.Fatal(er)
	}

	structTypes := []*StructType{}

	if len(pkgs) > 1 {
		log.Fatal("Multiple packages found! crudgen only supports one package at a time.")
	}

	var packageName string

	// Enumerate the AST and pull out all of the struct declarations.
	for _, pkg := range pkgs {
		packageName = pkg.Name

		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				if genDecl, ok := decl.(*ast.GenDecl); ok {
					if genDecl.Tok != token.TYPE {
						continue
					}

					for _, spec := range genDecl.Specs {
						if typeSpec, ok := spec.(*ast.TypeSpec); ok {
							if structType, ok := typeSpec.Type.(*ast.StructType); ok {
								structTypes = append(structTypes, &StructType{
									TypeSpec:   typeSpec,
									StructType: structType,
									Name:       typeSpec.Name.Name,
								})
							}
						}
					}
				}
			}
		}
	}

	// Enumerate the structs we've pulled out and parse their field declarations.
	for _, structType := range structTypes {
		for _, field := range structType.StructType.Fields.List {
			if field.Tag == nil {
				continue
			}

			if len(field.Names) == 0 {
				// XXX: We may want to support anonymous fields in the future.
				continue
			}

			name := field.Names[0].Name

			tagList, er := strconv.Unquote(field.Tag.Value)
			if er != nil {
				continue
			}

			tags := reflect.StructTag(tagList)

			if tags.Get(structTagName) != "" {
				structType.Fields = append(structType.Fields, StructField{
					Name:    name,
					SqlName: tags.Get(structTagName),
					Type:    field.Type,
				})
			}
		}
	}

	filePath := filepath.Join(dirPath, outputFilename)
	f, er := os.Create(filePath)
	if er != nil {
		log.Fatal(er)
	}
	defer f.Close()

	fmt.Fprintf(f, "package %s\n// AUTOGENERATED CODE. Regenerate by running crudgen.\n\nimport (\n\t\"jonsno.ws/crud2\"\n)\n", packageName)

	for _, structType := range structTypes {
		fmt.Fprintf(f, "%s", structType.Metadata())
	}

	fmt.Fprintf(f, "\n")

	if er := f.Sync(); er != nil {
		log.Fatal(er)
	}
}

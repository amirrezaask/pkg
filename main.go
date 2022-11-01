package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

const (
	GENERATED_ANNOTATION = "GENERATED USING GOX, DONT EDIT BY HAND"
	GOX_PREFIX           = "__GOX__"
	ENUM_PREFIX          = "__ENUM__"
	ENUM_DECORATOR       = "gox:enum"
)

func genEnumName(name string) string {
	return fmt.Sprintf("%s%s%s", GOX_PREFIX, ENUM_PREFIX, name)
}

func genEnumStruct(pkg string, name string, variants []string) string {
	codename := genEnumName(name)

	var vars []string
	for _, v := range variants {
		vars = append(vars, fmt.Sprintf(`%s = %s{"%s"}`, v, codename, v))
	}

	return fmt.Sprintf(EnumTemplate, GENERATED_ANNOTATION, pkg, codename, strings.Join(vars, "\n"), name, codename, codename)
}

func main() {
	filename := os.Args[1]

	fset := token.NewFileSet()
	fast, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)

	if err != nil {
		panic(err)
	}

	// handle enums
	// for now enums can only be top level
	enumFile, err := os.Create(fmt.Sprintf("enum_%s", filename))
	if err != nil {
		panic(err)
	}
	for _, decl := range fast.Decls {
		declComment := decl.(*ast.GenDecl).Doc.Text()
		if len(declComment) > 0 && declComment[:len(ENUM_DECORATOR)] == ENUM_DECORATOR {
			name := decl.(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Name.String()
			variants := strings.Split(strings.Trim(declComment[len(ENUM_DECORATOR)+1:], " \n\t\r"), " ")
			enum := genEnumStruct(fast.Name.String(), name, variants)
			_, err := fmt.Fprint(enumFile, enum)
			if err != nil {
				panic(err)
			}
		}
	}
}

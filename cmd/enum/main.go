/*
example code:
```go
// enum: Started Arrived Finished
type RideState int
```
*/

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const usage = `usage:
	enum <input file that contains enum decorators> 
`

const (
	GENERATED_ANNOTATION = "GENERATED USING enum program, DONT EDIT BY HAND"
	ENUM_PREFIX          = "__ENUM__"
	ENUM_DECORATOR       = "enum"
)

const enumTemplate = `
// {{ .Doc }}
package {{ .Pkg }}

type {{ .EnumName }} struct {
	variant string
}

var (
	{{ range .Variants }}
	{{ . }} = {{ $.EnumName }}{"{{ . }}"}
	{{ end }}
)


func {{ .TypeName }}FromString(s string) {{ .EnumName }} {
	return {{ .EnumName }}{s}
}
`

func genEnumStruct(pkg string, name string, variants []string) string {
	codename := fmt.Sprintf("%s%s", ENUM_PREFIX, name)

	type enumStruct struct {
		Doc      string
		Pkg      string
		TypeName string
		EnumName string
		Variants []string
	}
	t := template.Must(template.New("enum").Parse(enumTemplate))
	var buff strings.Builder
	err := t.Execute(&buff, enumStruct{
		Doc:      GENERATED_ANNOTATION,
		Pkg:      pkg,
		EnumName: codename,
		TypeName: name,
		Variants: variants,
	})
	if err != nil {
		panic(err)
	}
	return buff.String()
}

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		fmt.Println("")
		return
	}
	filename := os.Args[1]

	inputFilePath, err := filepath.Abs(filename)
	if err != nil {
		panic(err)
	}
	dir, _ := filepath.Split(inputFilePath)
	fset := token.NewFileSet()
	fast, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)

	if err != nil {
		panic(err)
	}
	actualName := strings.TrimSuffix(filename, filepath.Ext(filename))
	outputFilePath := filepath.Join(dir, fmt.Sprintf("%s_enums_gen.go", actualName))
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()
	for _, decl := range fast.Decls {
		if _, ok := decl.(*ast.GenDecl); ok {
			declComment := decl.(*ast.GenDecl).Doc.Text()
			if len(declComment) > 0 && declComment[:len(ENUM_DECORATOR)] == ENUM_DECORATOR {
				name := decl.(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Name.String()
				variants := strings.Split(strings.Trim(declComment[len(ENUM_DECORATOR)+1:], " \n\t\r"), " ")
				enum := genEnumStruct(fast.Name.String(), name, variants)
				_, err := fmt.Fprint(outputFile, enum)
				if err != nil {
					panic(err)
				}
			}
		}

	}

}

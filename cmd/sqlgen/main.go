// querygen generates a complete query builder for your model from your code.
//
// Usage:
//
//	sqlgen $file
//

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

const usage = `usage:`

const _DBMODEL_DECORATOR = "sqlgen:"

func gen(pkg string, typeName string, fields map[string]string, tags []string, args map[string]string) string {
	type Struct struct {
		Pkg       string
		ModelName string
		Fields    map[string]string
	}
	t := template.Must(template.New("sqlgen").Parse(modelTemplate))
	var buff strings.Builder
	err := t.Execute(&buff, Struct{
		Pkg:       pkg,
		ModelName: typeName,
		Fields:    fields,
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
	outputFilePath := filepath.Join(dir, fmt.Sprintf("%s_sqlgen_gen.go", actualName))
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()
	for _, decl := range fast.Decls {
		if _, ok := decl.(*ast.GenDecl); ok {
			declComment := decl.(*ast.GenDecl).Doc.Text()
			if len(declComment) > 0 && declComment[:len(_DBMODEL_DECORATOR)] == _DBMODEL_DECORATOR {
				name := decl.(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Name.String()
				arguments := strings.Split(strings.Trim(declComment[len(_DBMODEL_DECORATOR)+1:], " \n\t\r"), " ")
				fields := make(map[string]string)
				for _, field := range decl.(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Type.(*ast.StructType).Fields.List {
					for _, name := range field.Names {
						fields[name.String()] = fmt.Sprint(field.Type)
					}
				}
				args := make(map[string]string)
				for _, argkv := range arguments {
					splitted := strings.Split(argkv, "=")
					args[splitted[0]] = splitted[1]
				}
				output := gen(fast.Name.String(), name, fields, nil, args)
				fmt.Fprint(outputFile, output)
			}
		}

	}

}

const modelTemplate = `
package {{ .Pkg }}

type {{ .ModelName }}WhereBuilder struct {
	{{ range $field, $type := .Fields }}
	{{$field}} *{{$type}}
	{{ end }}
}

{{ range $field, $type := .Fields }}
func (m *{{ $.ModelName }}WhereBuilder) Where{{$field}}({{ $field }} {{ $type }}) *{{ $.ModelName }}WhereBuilder {
	m.{{$field}} = &{{$field}}
	return m
}
{{ end }}

func (m *{{ $.ModelName }}WhereBuilder) String() string {
	output := ""

	{{ range $field, $type := .Fields }}
	if m.{{$field}} != nil {
		output += fmt.Sprintf("%s = %s", "{{ $field }}", m.{{$field}})
	}
	{{ end }}

	if output != "" {
		return fmt.Sprintf("WHERE %s", output)
	}
	return ""
}

type {{ .ModelName }}QueryBuilder struct {
	{{ .ModelName }}WhereBuilder
}
	
func Query{{ .ModelName }}() *{{ .ModelName }}QueryBuilder {
	return &{{ .ModelName }}QueryBuilder{}
}

func (q *{{.ModelName}}QueryBuilder) String() string {
	return ""
}
	
type {{ .ModelName }}UpdateBuilder struct {
	set struct {
		{{ range $field, $type := .Fields }}
		{{$field}} *{{$type}}
		{{ end }}
	}

	where {{ .ModelName }}WhereBuilder
}
{{ range $field, $type := .Fields }}
func (m *{{ $.ModelName }}UpdateBuilder) Set{{ $field }}({{ $field }} {{ $type }}) *{{ $.ModelName }}UpdateBuilder {
	m.set.{{$field}} = &{{ $field }}
	return m
}
{{ end }}
type {{ .ModelName }}DeleteBuilder struct {
	where {{ .ModelName }}WhereBuilder
}
`

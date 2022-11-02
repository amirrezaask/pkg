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
	pathList := filepath.SplitList(inputFilePath)
	pathList = pathList[:len(pathList)-1]
	dir := filepath.Join(pathList...)
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

import "fmt"

type __{{ .ModelName }}SQLQueryBuilder struct {
    where __{{ .ModelName }}Where
	set __{{ .ModelName }}Set
}

func {{.ModelName}}QueryBuilder() __{{ .ModelName }}SQLQueryBuilder {
	return __{{ .ModelName }}SQLQueryBuilder{}
}

type __{{ .ModelName }}Where struct {
	{{ range $field, $type := .Fields }}
	{{$field}} struct {
        argument *{{$type}}
        operator string
    }
	{{ end }}
}

type __{{ .ModelName }}Set struct {
	{{ range $field, $type := .Fields }}
	{{$field}} *{{$type}}
	{{ end }}
}

{{ range $field, $type := .Fields }}
func (m *__{{ $.ModelName }}SQLQueryBuilder) Set{{ $field }}({{ $field }} {{ $type }}) *__{{ $.ModelName }}SQLQueryBuilder {
	m.set.{{$field}} = &{{ $field }}
	return m
}

func (m *__{{ $.ModelName }}SQLQueryBuilder) Where{{$field}}Eq({{ $field }} {{ $type }}) *__{{ $.ModelName }}SQLQueryBuilder {
	m.where.{{$field}}.argument = &{{$field}}
    m.where.{{$field}}.operator = "="
	return m
}
{{ if eq $type "int" "int8" "int16" "int32" "int64" "uint8" "uint16" "uint32" "uint64" "uint" "float32" "float64"  }}
func (m *__{{$.ModelName}}SQLQueryBuilder) Where{{$field}}GE({{$field}} {{$type}}) *__{{$.ModelName}}SQLQueryBuilder {
	m.where.{{$field}}.argument = &{{$field}}
    m.where.{{$field}}.operator = ">="
	return m
}
func (m *__{{$.ModelName}}SQLQueryBuilder) Where{{$field}}GT({{$field}} {{$type}}) *__{{$.ModelName}}SQLQueryBuilder {
    m.where.{{$field}}.argument = &{{$field}}
    m.where.{{$field}}.operator = ">="
	return m
}
func (m *__{{$.ModelName}}SQLQueryBuilder) Where{{$field}}LE({{$field}} {{$type}}) *__{{$.ModelName}}SQLQueryBuilder {
    m.where.{{$field}}.argument = &{{$field}}
    m.where.{{$field}}.operator = "<="
	return m
}
func (m *__{{$.ModelName}}SQLQueryBuilder) Where{{$field}}LT({{$field}} {{$type}}) *__{{$.ModelName}}SQLQueryBuilder {
    m.where.{{$field}}.argument = &{{$field}}
    m.where.{{$field}}.operator = "<="
	return m
}
{{ end }}

{{ end }}
`

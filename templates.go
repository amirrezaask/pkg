package main

const EnumTemplate = `
// %s
package %s

type %s struct {
	variant string
}

var (
%s
)


func %sFromString(s string) %s {
	return %s{s}
}
`

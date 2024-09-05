package fake

import (
	"github.com/brianvoe/gofakeit/v7"
)

func This(obj any, spec ...map[string]any) {
	
	if err := gofakeit.Struct(obj); err != nil {
		panic(err)
	}
}

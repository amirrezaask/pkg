package test

import (
	"math/rand/v2"

	"github.com/brianvoe/gofakeit/v7"
)

func Fakery() *gofakeit.Faker {
	return gofakeit.New(0)
}

func RandomElement[T any](list ...T) T {
	i := rand.IntN(len(list) - 1)
	return list[i]
}

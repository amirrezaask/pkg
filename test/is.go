package test

import (
	"testing"

	"github.com/matryer/is"
)

func Is(t *testing.T) *is.I {
	return is.New(t)
}

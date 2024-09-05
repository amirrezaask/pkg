package test

import (
	"database/sql"
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

type fakery struct {
	*gofakeit.Faker
}

func (t *T) Fakery() *fakery {
	return t.fakery
}

func (f *fakery) FakeNullInt64() sql.NullInt64 {
	return sql.NullInt64{
		Int64: int64(f.Number(0, 1000)),
		Valid: true,
	}
}

func (f *fakery) PastTime() time.Time {
	return f.PastDate()
}

func (f *fakery) FakeInvoiceStatus() string {
	return f.RandomString([]string{"init", "paid"})
}

func (f *fakery) ID() int64 {
	return int64(f.Number(0, 1000))
}

func (f *fakery) InvoiceTargetType() string {
	return f.RandomString([]string{"lab", "home_care", "pharmacy", "product", "chat", "charge"})
}

func (f *fakery) Amount() int64 {
	return int64(f.Number(100, 100000))
}

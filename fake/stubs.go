package fake

import (
	"database/sql"
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

func FirstName() string {
	return gofakeit.FirstName()
}
func LastName() string {
	return gofakeit.LastName()
}

func Amount() int64 {
	return int64(gofakeit.Number(100, 100000))
}

func FakeNullInt64() sql.NullInt64 {
	return sql.NullInt64{
		Int64: int64(gofakeit.Number(0, 1000)),
		Valid: true,
	}
}

func PastTime() time.Time {
	return gofakeit.PastDate()
}

func UUID() string {
	return gofakeit.UUID()
}

func FakeInvoiceStatus() string {
	return gofakeit.RandomString([]string{"init", "paid"})
}

func ID() int64 {
	return int64(gofakeit.Number(0, 1000))
}

func InvoiceTargetType() string {
	return gofakeit.RandomString([]string{"lab", "home_care", "pharmacy", "product", "chat", "charge"})
}

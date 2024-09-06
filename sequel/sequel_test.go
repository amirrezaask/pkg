package sequel

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	_ "github.com/mattn/go-sqlite3"
)

type InvoiceWithPK struct {
	Id     int64
	User   int64
	Status string
}

func (i *InvoiceWithPK) SqlSchema() *Schema {
	return NewSchema("invoice",
		"id", &i.Id, PK,
		"user", &i.User,
		"status", &i.Status,
	)
}

type InvoiceInferPK struct {
	Id     int64
	User   int64
	Status string
}

func (i *InvoiceInferPK) SqlSchema() *Schema {
	return NewSchema("invoice",
		"id", &i.Id,
		"user", &i.User,
		"status", &i.Status,
	)
}

func TestSchema(t *testing.T) {
	t.Run("SqlSchema should set table correctly", func(t *testing.T) {
		s := (&InvoiceWithPK{}).SqlSchema()
		a := assert.New(t)
		a.Equal("invoice", s.table)
	})

	t.Run("SqlSchema should set columns correctly", func(t *testing.T) {
		s := (&InvoiceWithPK{}).SqlSchema()
		a := assert.New(t)
		a.Len(s.valueMap, 3)
		a.Len(s.fillable, 2)
	})

	t.Run("should infer primary key correctly", func(t *testing.T) {
		s := (&InvoiceInferPK{}).SqlSchema()
		a := assert.New(t)
		a.Len(s.valueMap, 3)
		a.Len(s.fillable, 2)
		a.NotNil(s.pk)
	})
}

type User struct {
	ID        int64
	Name      string
	LastName  string
	Credit    int
	CreatedAt time.Time
	UpdatedAt *time.Time
}

func (u *User) Schema() *Schema {
	return NewSchema("users",
		"id", &u.ID,
		"name", &u.Name,
		"last_name", &u.LastName,
		"credit", &u.Credit,
		"created_at", &u.CreatedAt,
		"updated_at", &u.UpdatedAt,
	)
}

func TestDBFunctions(t *testing.T) {
	// make this sqlite so it's simple to just run
	db, err := New("sqlite3", "file::memory:?mode=memory&cache=shared",
		"teleyare", ConnectionOptions{
			PromNS:                "test",
			MaxOpenConnections:    50,
			MaxIdleConnections:    50,
			IdleConnectionTimeout: 10 * time.Second,
			OpenConnectionTimeout: 10 * time.Second,
		})
	if err != nil {
		panic(err)
	}

	command := createMigrationCommand(sqlite, &User{})
	_, err = db.Exec(command)
	if err != nil {
		panic(err)
	}

	for i := 0; i < 3; i++ {
		Insert(db, &User{
			Name:     uuid.NewString(),
			LastName: uuid.NewString(),
			Credit:   88,
		})
	}
	user := User{
		Name:     "Amirreza",
		LastName: "Ask",
		Credit:   999999,
	}
	_, err = Insert(db, &user)
	if err != nil {
		panic(err)
	}

	fmt.Println("Amirreza ID is", user.ID)

	// Now let's do an update
	user.Credit += 1000
	err = Save(db, &user)
	if err != nil {
		panic(err)
	}

	// And now let's delete it.
	_, err = Delete(db, &user)
	if err != nil {
		panic(err)
	}

	// Now let us get a list of all our users with almost no money and give them a promotion.
	poorUsers, err := Query[User](db, "SELECT * FROM users WHERE credit < 100")
	if err != nil {
		panic(err)
	}
	for _, user := range poorUsers {
		// dump.This(user)
		user.Credit += 1000000
	}
	// and now let's do a bulk update on our sequel.
	err = Save(db, poorUsers...)
	if err != nil {
		panic(err)
	}

	tableInfo, err := ToMap(db.Query("PRAGMA table_info(users)"))
	if err != nil {
		panic(err)
	}
	for range tableInfo {
		// world is your oyster
	}
}

type Invoice struct {
	Id         int64
	User       int64
	Target     sql.NullInt64
	TargetType string
	Amount     int64
	Metadata   sql.NullString
	Status     string
	Discount   sql.NullInt64
	Voucher    sql.NullString
	Payment    string
	Debt       int
	Uuid       string
	CreatedAt  string
	UpdateAt   sql.NullString
}

// To make sure that Schema has no errors and everything is of correct type.
var _ = (&Invoice{}).Schema().Validate()

func (i *Invoice) Schema() *Schema {
	return NewSchema("invoice",
		"id", &i.Id,
		"user", &i.User,
		"target", &i.Target,
		"target_type", &i.TargetType,
		"amount", &i.Amount,
		"metadata", &i.Metadata,
		"status", &i.Status,
		"discount", &i.Discount,
		"payment", &i.Payment,
		"voucher", &i.Voucher,
		"debt", &i.Debt,
		"uuid", &i.Uuid,
		"created_at", &i.CreatedAt,
		"update_at", &i.UpdateAt,
	)
}

func TestMigrationCommand(t *testing.T) {
	t.Run("create migration command", func(t *testing.T) {
		command := createMigrationCommand("sqlite3", &Invoice{})
		db, err := sql.Open("sqlite3", "file:testdb.sqlite3?mode=memory&cache=shared")
		testNoError(t, err, "error in creating sqlite3 memory connection")
		_, err = db.Exec(command)
		testNoError(t, err, "error in running migration command: %s\n", command)

	})
}

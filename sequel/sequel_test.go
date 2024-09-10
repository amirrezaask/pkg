package sequel

import (
	"database/sql"
	"encoding/json"
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

func (i *InvoiceWithPK) SequelRecordSpec() RecordSpec {
	return RecordSpec{
		Connection: "testconnection",
		Table:      "invoice",
		Columns: []Column{
			{Name: "id", Ptr: &i.Id, Options: PK},
			{Name: "user", Ptr: &i.User},
			{Name: "status", Ptr: &i.Status},
		},
	}
}

type InvoiceInferPK struct {
	Id     int64
	User   int64
	Status string
}

func (i *InvoiceInferPK) SequelRecordSpec() RecordSpec {
	return RecordSpec{
		Connection: "testconnection",
		Table:      "invoice",
		Columns: []Column{
			{Name: "id", Ptr: &i.Id},
			{Name: "user", Ptr: &i.User},
			{Name: "status", Ptr: &i.Status},
		},
	}
}

func TestSchema(t *testing.T) {
	t.Run("SqlSchema should set table correctly", func(t *testing.T) {
		s, err := (&InvoiceWithPK{}).SequelRecordSpec().intoInternalRepr()
		a := assert.New(t)
		a.NoError(err)
		a.Equal("invoice", s.table)
	})

	t.Run("SqlSchema should set columns correctly", func(t *testing.T) {
		s, err := (&InvoiceWithPK{}).SequelRecordSpec().intoInternalRepr()
		a := assert.New(t)
		a.NoError(err)
		a.Len(s.valueMap, 3)
		a.Len(s.fillable, 2)
	})

	t.Run("should infer primary key correctly", func(t *testing.T) {
		s, err := (&InvoiceInferPK{}).SequelRecordSpec().intoInternalRepr()
		a := assert.New(t)
		a.NoError(err)
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

func (u *User) SequelRecordSpec() RecordSpec {
	return RecordSpec{
		Connection: "testingconnection",
		Table:      "users",
		Columns: []Column{
			{Name: "id", Ptr: &u.ID},
			{Name: "name", Ptr: &u.Name},
			{Name: "last_name", Ptr: &u.LastName},
			{Name: "credit", Ptr: &u.Credit},
			{Name: "created_at", Ptr: &u.CreatedAt},
			{Name: "updated_at", Ptr: &u.UpdatedAt},
		},
	}
}

func TestCreatedAtUpdatedAt(t *testing.T) {
	t.Run("when created_at and updated_at are not set they should populate automatically", func(t *testing.T) {
		db, err := New(DataSource{
			Driver:                "sqlite3",
			Name:                  "testingconnection",
			ConnectionString:      "file::memory:?mode=memory&cache=shared",
			MetricsNamespace:      "createdat",
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

		u := User{
			Name:     uuid.NewString(),
			LastName: uuid.NewString(),
			Credit:   88,
		}

		_, err = Insert(&u)
		if err != nil {
			fmt.Printf("error in saving user in %s: %s\n", t.Name(), err.Error())
			t.FailNow()
		}

	})

}

func TestDBFunctions(t *testing.T) {
	// make this sqlite so it's simple to just run
	db, err := New(DataSource{
		Driver:                "sqlite3",
		Name:                  "testingconnection",
		ConnectionString:      "file::memory:?mode=memory&cache=shared",
		MetricsNamespace:      "createdat",
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
		Insert(&User{
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
	_, err = Insert(&user)
	if err != nil {
		panic(err)
	}

	fmt.Println("Amirreza ID is", user.ID)

	// Now let's do an update
	user.Credit += 1000
	err = Save(&user)
	if err != nil {
		panic(err)
	}

	// And now let's delete it.
	_, err = Delete(&user)
	if err != nil {
		panic(err)
	}

	// Now let us get a list of all our users with almost no money and give them a promotion.
	poorUsers, err := Query[User]("SELECT * FROM users WHERE credit < 100")
	if err != nil {
		panic(err)
	}
	for _, user := range poorUsers {
		// dump.This(user)
		user.Credit += 1000000
	}
	// and now let's do a bulk update on our sequel.
	err = Save(poorUsers...)
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
	Id          int64
	User        int64
	Target      sql.NullInt64
	TargetType  string
	Amount      int64
	Metadata    sql.NullString
	MetadataMap map[string]any
	Status      string
	Discount    sql.NullInt64
	Voucher     sql.NullString
	Payment     string
	Debt        int
	Uuid        string
	CreatedAt   string
	UpdateAt    sql.NullString
}

// To make sure that Schema has no errors and everything is of correct type.
var _ = (&Invoice{}).SequelRecordSpec().Validate()

func (i *Invoice) SequelRecordSpec() RecordSpec {
	return RecordSpec{
		Connection: "financial",
		Table:      "invoice",
		Columns: []Column{
			{Name: "id", Ptr: &i.Id},
			{Name: "user", Ptr: &i.User},
			{Name: "target", Ptr: &i.Target},
			{Name: "target_type", Ptr: &i.TargetType},
			{Name: "amount", Ptr: &i.Amount},
			{Name: "metadata", Ptr: &i.Metadata},
			{Name: "status", Ptr: &i.Status},
			{Name: "discount", Ptr: &i.Discount},
			{Name: "payment", Ptr: &i.Payment},
			{Name: "voucher", Ptr: &i.Voucher},
			{Name: "debt", Ptr: &i.Debt},
			{Name: "uuid", Ptr: &i.Uuid},
			{Name: "created_at", Ptr: &i.CreatedAt},
			{Name: "update_at", Ptr: &i.UpdateAt, Options: UpdatedAt},
		},
		BeforeWrite: []BeforeWriteHook{
			func(m Record) error {
				invoice := m.(*Invoice)
				invoice.MetadataMap = map[string]any{}
				if !invoice.Metadata.Valid {
					return nil
				}
				return json.Unmarshal([]byte(invoice.Metadata.String), &invoice.MetadataMap)
			},
		},
	}
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

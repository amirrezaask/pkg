package sequel

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/amirrezaask/pkg/errors"
	gormSqlite "gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type MockDb struct {
	*DB
	FailOnTxRequest bool
}

func (m *MockDb) BeginTx(ctx context.Context, options *sql.TxOptions) (*sql.Tx, error) {
	if m.FailOnTxRequest {
		return nil, errors.New("error in getting tx object from mock database object")
	}

	return m.db.BeginTx(ctx, options)
}

func testNoError(t *testing.T, err error, msg string, args ...any) {
	if err != nil {
		t.Logf("Error: %s, expected no error but %s", fmt.Sprintf(msg, args...), err.Error())
		t.FailNow()
	}
}

func NewMockDb(t *testing.T, target **DB, models ...any) *MockDb {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", uuid.NewString())
	// dsn = fmt.Sprintf("file:%s?cache=shared", connectionName)
	db, err := sql.Open("sqlite3", dsn)
	testNoError(t, err, "cannot open sqlite3 connection from NewMockDb")
	sqlDB := &DB{
		cfg: DBConfig{
			Driver: "sqlite3",
		},
		db: db,
	}
	mockDB := &MockDb{
		DB: sqlDB,
	}

	*target = sqlDB

	gormDB, err := gorm.Open(gormSqlite.New(gormSqlite.Config{
		Conn: sqlDB,
	}))

	testNoError(t, err, "cannot create gorm object")
	testNoError(t, gormDB.AutoMigrate(models...), "cannot auto migrate")
	return mockDB
}

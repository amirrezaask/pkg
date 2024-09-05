package test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/amirrezaask/go-sith/database"
	"github.com/amirrezaask/go-sith/must"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func (t *T) NewDb(target *database.Sql, migrations ...any) *MockDb {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", uuid.NewString())
	db, err := gorm.Open(sqlite.Open(dsn))

	if err != nil {
		panic(err)
	}
	conn, err := db.DB()
	if err != nil {
		panic(err)
	}
	u := uuid.NewString()
	sqlDB := database.FromTestConnection(conn, "sqlite")
	mockDB := &MockDb{
		SqlDatabase: sqlDB,
		GormDb:      db,
	}
	*target = database.Sql(mockDB)

	(*target).SetDbName(u)
	t.dbs[u] = mockDB

	if len(migrations)%2 != 0 {
		panic("given migrations should be valid key-value pairs of tablename:model")
		t.Fatal("given migrations should be valid key-value pairs of tablename:model")
	}
	for i := 0; i < len(migrations); i++ {
		if i%2 == 0 {
			t.HasNoError(db.Table(migrations[i].(string)).AutoMigrate(migrations[i+1]))
		}
	}
	return mockDB
}

type dbAssertions struct {
	t     *T
	db    database.Sql
	query string
	args  []any
}

func (t *T) LogQueryResultAndFail(db database.Sql, query string, args ...any) {
	fmt.Printf("Result for '%s'\n", query)
	spew.Dump(must.Do(database.ToMap(db.Query(query, args...))))
	t.T.FailNow()
}

func (t *T) Db(db database.Sql) *dbAssertions {
	return &dbAssertions{
		t:  t,
		db: db,
	}
}

func (d *dbAssertions) Query(query string, args ...any) *dbAssertions {
	d.query = query
	d.args = args
	return d
}

func (d *dbAssertions) OutputCount(n int) {
	maps, err := database.ToMap(d.db.Query(d.query, d.args...))
	d.t.HasNoError(err)
	d.t.AssertEq(n, len(maps), "expected query output count %d but have %d", n, len(maps))
}

func (d *dbAssertions) OutputIsSimilar(expected []map[string]any) {
	maps, err := database.ToMap(d.db.Query(d.query, d.args...))
	d.t.HasNoError(err)
	d.t.AreSimilar(expected, maps)
}

type MockDb struct {
	*database.SqlDatabase
	GormDb          *gorm.DB
	FailOnTxRequest bool
}

func (m *MockDb) BeginTx(ctx context.Context, options *sql.TxOptions) (*sql.Tx, error) {
	if m.FailOnTxRequest {
		return nil, errors.New("error in getting tx object from mock database object")
	}

	return m.SqlDatabase.BeginTx(ctx, options)
}

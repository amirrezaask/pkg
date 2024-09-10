package sequel

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/amirrezaask/go-std/errors"

	_ "github.com/mattn/go-sqlite3"
)

type MockDb struct {
	Interface
	FailOnTxRequest bool
}

func (m *MockDb) BeginTx(ctx context.Context, options *sql.TxOptions) (*sql.Tx, error) {
	if m.FailOnTxRequest {
		return nil, errors.New("error in getting tx object from mock database object")
	}

	return m.Interface.BeginTx(ctx, options)
}

func columnInfo(dbKind string, spec columnSpec, nullables ...bool) (string, error) {
	options := ""
	var nullable bool
	var sqlType string
	switch spec.Type.Kind() {
	case reflect.Bool:
		sqlType = "tinyint(1)"
	case reflect.Int:
		sqlType = "int"

	case reflect.Int8:
		sqlType = "int"

	case reflect.Int16:
		sqlType = "int"

	case reflect.Int32:
		sqlType = "int"

	case reflect.Int64:
		sqlType = "bigint"

	case reflect.Uint:
		sqlType = "int"

	case reflect.Uint8:
		sqlType = "int"

	case reflect.Uint16:
		sqlType = "int"

	case reflect.Uint32:
		sqlType = "int"

	case reflect.Uint64:
		sqlType = "bigint"

	case reflect.Uintptr:
		return "", errors.Newf("%s unsupported column type, cannot represent %s", spec.Name, spec.Type.String())
	case reflect.Float32:
		sqlType = "double"
	case reflect.Float64:
		sqlType = "double"
	case reflect.String:
		sqlType = "varchar(500)"

	case reflect.Complex64:
		return "", errors.Newf("%s unsupported column type, cannot represent %s", spec.Name, spec.Type.String())

	case reflect.Complex128:
		return "", errors.Newf("%s unsupported column type, cannot represent %s", spec.Name, spec.Type.String())

	case reflect.Array:
		return "", errors.Newf("%s unsupported column type, cannot represent %s", spec.Name, spec.Type.String())

	case reflect.Chan:
		return "", errors.Newf("%s unsupported column type, cannot represent %s", spec.Name, spec.Type.String())

	case reflect.Func:
		return "", errors.Newf("%s unsupported column type, cannot represent %s", spec.Name, spec.Type.String())

	case reflect.Interface:
		return "", errors.Newf("%s unsupported column type, cannot represent %s", spec.Name, spec.Type.String())

	case reflect.Map:
		return "", errors.Newf("%s unsupported column type, cannot represent %s", spec.Name, spec.Type.String())

	case reflect.Pointer:
		elemType := spec.Type.Elem()
		q, err := columnInfo(dbKind, columnSpec{
			Name: spec.Name,
			Type: elemType,
			IsPK: spec.IsPK,
		}, true)
		if err != nil {
			return "", errors.Wrap(err, "%s unsupported column type, cannot represent %s", spec.Name, spec.Type.String())
		}

		return q, nil

	case reflect.Slice:
		if spec.Type.String() == "[]uint8" || spec.Type.String() == "[]byte" {
			sqlType = "varchar(500)"
		} else {
			return "", errors.Newf("%s unsupported column type, cannot represent %s", spec.Name, spec.Type.String())
		}

	case reflect.Struct:
		switch spec.Type.String() {
		case "sql.NullInt16":
			sqlType = "int"
			nullable = true
		case "sql.NullInt32":
			sqlType = "int"
			nullable = true

		case "sql.NullInt64":
			sqlType = "bigint"
			nullable = true

		case "sql.NullString":
			sqlType = "varchar(255)"
			nullable = true

		case "sql.NullByte":
			sqlType = "int"
			nullable = true

		case "sql.NullTime":
			sqlType = "DATETIME"
			nullable = true

		case "time.Time":
			sqlType = "DATETIME"

		case "sql.NullFloat64":
			sqlType = "double"
			nullable = true

		case "sql.NullBool":
			sqlType = "tinyint(1)"
			nullable = true
		default:
			return "", errors.Newf("%s unsupported column type, cannot represent %s", spec.Name, spec.Type.String())

		}

	case reflect.UnsafePointer:
		return "", errors.Newf("%s unsupported column type, cannot represent %s", spec.Name, spec.Type.String())

	}
	if spec.IsPK {
		options += "PRIMARY KEY"
		if dbKind == "mysql" {
			options += " AUTO_INCREMENT"
		} else { // others are standard
			sqlType = "SERIAL"
		}
	}

	if len(nullables) > 0 {
		nullable = nullables[0]
	}
	if nullable {
		options += " NULL "
	} else {
		options += " NOT NULL "
	}
	return fmt.Sprintf("`%s` %s %s", spec.Name, sqlType, options), nil
}

func createMigrationCommand(dbKind string, m Record) string {
	schema, err := m.SequelRecordSpec().intoInternalRepr()
	if err != nil {
		panic(err)
	}
	table, columnSpecs := schema.GetColumns()
	tableName := fmt.Sprintf("`%s`", table)
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", tableName)
	if dbKind == "mysql" {
		query += "id int auto_increment primary key,\n"
	} else if dbKind == "sqlite3" {
		query += "id integer primary key,\n"
	} else {
		panic(fmt.Sprintf("database %s is not supported for auto migration", dbKind))
	}

	for i, spec := range columnSpecs {
		q, err := columnInfo(dbKind, spec)
		if err != nil {
			panic(err)
		}
		if i == len(columnSpecs)-1 {
			query += fmt.Sprintf("%s\n", q)
		} else {
			query += fmt.Sprintf("%s,\n", q)
		}
	}

	query += ")"

	return query
}

func testNoError(t *testing.T, err error, msg string, args ...any) {
	if err != nil {
		t.Logf("Error: %s, expected no error but %s", fmt.Sprintf(msg, args...), err.Error())
		t.FailNow()
	}
}

func NewMockDb(t *testing.T, connectionName string, target *Interface, models ...Record) *MockDb {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", connectionName)
	// dsn = fmt.Sprintf("file:%s?cache=shared", connectionName)
	db, err := sql.Open("sqlite3", dsn)
	testNoError(t, err, "cannot open sqlite3 connection from NewMockDb")
	sqlDB := fromTestConnection(db, "sqlite3")
	mockDB := &MockDb{
		Interface: sqlDB,
	}
	*target = Interface(mockDB)

	connections[connectionName] = Interface(mockDB)

	for i := 0; i < len(models); i++ {
		model := models[i]
		command := createMigrationCommand("sqlite3", model)
		_, err := sqlDB.Exec(command)
		testNoError(t, err, "error in running migration command: %s", command)
	}
	return mockDB
}

type dbAssertions struct {
	t     *testing.T
	db    Interface
	query string
	args  []any
	table string
}

func AssertDb(t *testing.T, db Interface, tables ...string) *dbAssertions {
	var table string
	if len(tables) > 0 {
		table = tables[0]
	}
	return &dbAssertions{
		t:     t,
		db:    db,
		table: table,
	}
}
func (d *dbAssertions) AssertNotEmpty() {
	rows, err := d.db.Query(fmt.Sprintf("select count(*) from \"%s\"", d.table))
	testNoError(d.t, err, "error in runnig query for assert not empty")
	defer rows.Close()
	var count int
	rows.Next()
	err = rows.Scan(&count)
	testNoError(d.t, err, "error in scanning count in AssertNotEmpty")

	if count == 0 {
		d.t.Logf("expected table %s to not be empty but there are no rows", d.table)
		d.t.FailNow()
	}
}
func (d *dbAssertions) AssertEmpty() {
	rows, err := d.db.Query(fmt.Sprintf("select count(*) from \"%s\"", d.table))
	testNoError(d.t, err, "error in running query for AssertNotEmpty")
	defer rows.Close()

	var count int
	if !rows.Next() {
		d.t.Logf("no rows returned for query: %s", d.query)
		d.t.FailNow()
	}
	err = rows.Scan(&count)
	testNoError(d.t, err, "error in scanning AssertEmpty count result")

	if count != 0 {
		d.t.Logf("expected table %s to be empty but there are %d rows", d.table, count)
		d.t.FailNow()
	}
}

func (d *dbAssertions) AssertCount(expectedCount int, kvs ...any) *dbAssertions {
	query, m, count := d.countForKvs(kvs...)
	if count != expectedCount {
		d.t.Logf("For query '%s'\n%+v\nexpected row count #%d but in database there are #%d rows.", query, m, expectedCount, count)
		d.t.FailNow()
	}

	return d
}

func (d *dbAssertions) countForKvs(kvs ...any) (query string, m map[string]any, count int) {
	if d.table == "" {
		d.t.Logf("you should provide table name in test.Db(DbObject, <table>)")
		d.t.FailNow()
	}
	m = map[string]any{}
	if len(kvs)%2 != 0 {
		panic("Has input should be a valid key-value sequence")
	}

	for i := 0; i < len(kvs); i++ {
		if i%2 == 0 {
			m[kvs[i].(string)] = kvs[i+1]

		}
	}

	pairs := []string{}
	values := []any{}
	for k, v := range m {
		pairs = append(pairs, fmt.Sprintf("%s=?", k))
		values = append(values, v)
	}
	query = fmt.Sprintf("SELECT COUNT(*) FROM \"%s\" WHERE %s", d.table, strings.Join(pairs, " AND "))
	rows, err := d.db.Query(query, values...)
	testNoError(d.t, err, "error in running query for count of kvs")
	defer rows.Close()
	rows.Next()

	err = rows.Scan(&count)
	testNoError(d.t, err, "error in scanning query result for the countKvs...")

	return query, m, count
}

func (d *dbAssertions) AssertHas(kvs ...any) *dbAssertions {
	query, m, count := d.countForKvs(kvs...)
	if count < 1 {
		d.t.Logf("For query '%s'\n%+v\nexpected at least one record but there was none", query, m)
		d.t.FailNow()
	}

	return d
}

func (d *dbAssertions) Query(query string, args ...any) *dbAssertions {
	d.query = query
	d.args = args
	return d
}

func (d *dbAssertions) OutputCount(n int) {
	maps, err := ToMap(d.db.Query(d.query, d.args...))
	testNoError(d.t, err, "error in running query")
	if n != len(maps) {
		d.t.Logf("expected query output count %d but have %d", n, len(maps))
		d.t.FailNow()
	}
}

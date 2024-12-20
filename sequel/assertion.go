package sequel

import (
	"fmt"
	"strings"
	"testing"
)

type dbAssertions struct {
	t     *testing.T
	db    *DB
	query string
	args  []any
	table string
}

func AssertDb(t *testing.T, db *DB, tables ...string) *dbAssertions {
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

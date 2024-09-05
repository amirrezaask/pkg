package test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/amirrezaask/go-sith/database"
	"github.com/amirrezaask/go-sith/must"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type T struct {
	*testing.T
	dbs        map[string]*mockDB
	redisMocks map[string]redismock.ClientMock
}

func runtimeInfo() string {
	pc, file, line, _ := runtime.Caller(2)
	fName := runtime.FuncForPC(pc)
	return fmt.Sprintf("file=%s %s line=%d", file, fName, line)
}

func Test(t *testing.T) *T {
	return &T{
		T:          t,
		dbs:        map[string]*mockDB{},
		redisMocks: map[string]redismock.ClientMock{},
	}
}

func (t *T) AssertNotEq(expected any, have any, msgAndArgs ...any) {
	eql := reflect.DeepEqual(expected, have)
	if !eql {
		return
	}
	if len(msgAndArgs) < 1 {
		msgAndArgs = append(msgAndArgs, "Assertion failed due to both arguments are equal")
	}
	if _, isString := msgAndArgs[0].(string); !isString {
		msgAndArgs[0] = fmt.Sprint(msgAndArgs[0])
	}
	var args []any
	if len(msgAndArgs) > 2 {
		args = msgAndArgs[1:]
	}
	fmt.Printf("--- Failed Assert %s %s:\nExpected %v\nHave %v\n", runtimeInfo(), fmt.Sprintf(msgAndArgs[0].(string), args...), expected, have)
	t.T.FailNow()
}

func (t *T) Empty(obj any, msgAndArgs ...any) {
	v := reflect.ValueOf(obj)
	ty := reflect.TypeOf(obj)
	if ty.Kind() != reflect.Array && ty.Kind() != reflect.Slice && ty.Kind() != reflect.Map {
		panic(fmt.Sprintf("invalid argument to test.Empty, only Array|Slice|Map not %s", reflect.TypeOf(obj).Kind().String()))
		t.FailNow()
	}
	if len(msgAndArgs) < 1 {
		msgAndArgs = append(msgAndArgs, "expected empty")
	}

	t.AssertEq(0, v.Len(), msgAndArgs...)
}

func (t *T) AssertWeakEq(expected any, have any, msgAndArgs ...any) {

	if assert.ObjectsAreEqualValues(expected, have) {
		return
	}
	caller := strings.Join(callerInfo(), "\n\t")

	if len(msgAndArgs) < 1 {
		msgAndArgs = append(msgAndArgs, "Assertion failed")
	}
	if _, isString := msgAndArgs[0].(string); !isString {
		msgAndArgs[0] = fmt.Sprint(msgAndArgs[0])
	}
	var args []any
	if len(msgAndArgs) > 1 {
		args = msgAndArgs[1:]
	}
	fmt.Printf("\n\n❌❌ Failed Assert => %s\nexpected '%v'\nhave '%v'\nStack: \n\t%s\n\n\n", fmt.Sprintf(msgAndArgs[0].(string), args...), expected, have, caller)
	t.T.FailNow()
}
func (t *T) AssertEq(expected, have any, msgAndArgs ...any) {
	if len(msgAndArgs) < 1 {
		msgAndArgs = append(msgAndArgs, "Assertion failed due to type mismatch")
	}
	if reflect.TypeOf(expected) != reflect.TypeOf(have) {
		fmt.Printf("[%s] %s: expected('%T') have('%T')", t.T.Name(), fmt.Sprint(msgAndArgs[0]), expected, have)
	}
	t.AssertWeakEq(expected, have, msgAndArgs...)
}
func (t *T) NoError(err error, msgAndArgs ...any) {
	if len(msgAndArgs) < 1 {
		msgAndArgs = append(msgAndArgs, "Expected error to be nil")
	}
	t.AssertWeakEq(nil, err, msgAndArgs...)
}
func (t *T) Error(err error, msgAndArgs ...any) {
	if len(msgAndArgs) < 1 {
		msgAndArgs = append(msgAndArgs, "Expected error to not be nil")
	}
	t.AssertNotEq(nil, err, msgAndArgs...)
}

func (t *T) AssertFalse(b bool, msgAndArgs ...any) {
	t.AssertEq(false, b, msgAndArgs...)
}
func (t *T) AssertTrue(b bool, msgAndArgs ...any) {
	t.AssertEq(true, b, msgAndArgs...)
}

func (t *T) DbQueryOutputIsSimilar(db *database.SqlDatabase, expected []map[string]any, q string, args ...any) {
	have, err := database.ToMap(db.Query(q, args...))
	t.NoError(err)
	t.Similar(expected, have)
}

func (t *T) StructToMap(obj any) map[string]any {
	m := map[string]any{}
	t.NoError(mapstructure.Decode(obj, &m))

	return m
}

func (t *T) eqaulValues(expected any, have any) bool { //TODO:
	eql := reflect.DeepEqual(expected, have)

	return eql
}

type mockTransport struct {
	requestToResponseErr map[string]struct {
		Response *http.Response
		Err      error
	}
}

func (m *mockTransport) AddRequest(method string, path string, resp *http.Response, err error) {
	m.requestToResponseErr[fmt.Sprintf("%s %s", method, path)] = struct {
		Response *http.Response
		Err      error
	}{resp, err}
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	method := req.Method
	path := req.URL.Path
	host := req.URL.Host
	s := m.requestToResponseErr[fmt.Sprintf("%s %s%s", method, host, path)]
	return s.Response, s.Err
}

func (t *T) HttpClient(target *http.Client) *mockTransport {
	mock := &mockTransport{}
	hc := http.Client{Transport: mock}
	*target = hc
	return mock
}

func (t *T) Similar(expected any, have any) {
	hv := reflect.ValueOf(have)
	ev := reflect.ValueOf(expected)
	t.AssertWeakEq(ev.Kind(), hv.Kind(), "expected and have must have similar kind")

	t.AssertWeakEq(ev.Len(), hv.Len(), "expected and have must have equal len")
	if ev.Kind() == reflect.Slice {
		if ev.Type().Elem().Kind() != reflect.Map {
			m := []map[string]any{}
			t.NoError(mapstructure.Decode(expected, &m))
			expected = m
			ev = reflect.ValueOf(ev)
		}
		if hv.Type().Elem().Kind() != reflect.Map {
			m := []map[string]any{}
			t.NoError(mapstructure.Decode(have, &m))
			have = m
			hv = reflect.ValueOf(have)
		}
		for i := 0; i < ev.Len(); i++ {
			ei := ev.Index(i)
			hi := hv.Index(i)

			for _, k := range ev.Index(i).MapKeys() {
				haveValue := hi.MapIndex(k)
				t.AssertTrue(haveValue.IsValid(), "expected key '%s' exist but it wasn't:", k.Interface())
				expectedValue := ei.MapIndex(k)
				t.AssertWeakEq(expectedValue.Interface(), haveValue.Interface(), "expected same value for key '%s'", k.Interface())
			}
		}
	}

}

type mockDB struct {
	*database.SqlDatabase
	gDB             *gorm.DB
	FailOnTxRequest bool
}

func (t *T) GetDbMock(name string) *mockDB {
	return t.dbs[name]
}

func (m *mockDB) GormObject() *gorm.DB {
	return m.gDB
}

func (m *mockDB) BeginTx(ctx context.Context, options *sql.TxOptions) (*sql.Tx, error) {
	if m.FailOnTxRequest {
		return nil, errors.New("error in getting tx object from mock database object")
	}

	return m.SqlDatabase.BeginTx(ctx, options)
}

func (t *T) SqlDb(target *database.Sql, migrations ...any) *T {
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
	mockDB := &mockDB{
		SqlDatabase: sqlDB,
		gDB:         db,
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
			t.NoError(db.Table(migrations[i].(string)).AutoMigrate(migrations[i+1]))
		}
	}
	return t
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

func (t *T) AssertDb(db database.Sql) *dbAssertions {
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

func (d *dbAssertions) AssertCount(n int) {
	maps, err := database.ToMap(d.db.Query(d.query, d.args...))
	d.t.NoError(err)
	d.t.AssertEq(n, len(maps), "expected query output count %d but have %d", n, len(maps))
}

func (d *dbAssertions) AssertSimilarTo(expected []map[string]any) {
	maps, err := database.ToMap(d.db.Query(d.query, d.args...))
	d.t.NoError(err)
	d.t.Similar(expected, maps)
}

type redisMock struct {
	redismock.ClientMock
}

// TODO: a way of reporting these errors better than the current shit
func (r *redisMock) ExpectLockForKeyWithDuration(key string, dur time.Duration) {
	r.Regexp().ExpectExists(key).SetVal(0)
	r.Regexp().ExpectSet(key, 1, dur).SetVal("OK")
}
func (r *redisMock) ExpectUnlockForKey(key string) {
	r.Regexp().ExpectDel(key).SetVal(1)
}

func (t *T) Redis(target **database.Redis) *redisMock {
	client, mock := redismock.NewClientMock()
	*target = &database.Redis{client}
	return &redisMock{mock}
}

func callerInfo() []string {
	isTest := func(name, prefix string) bool {
		if !strings.HasPrefix(name, prefix) {
			return false
		}
		if len(name) == len(prefix) { // "Test" is ok
			return true
		}
		r, _ := utf8.DecodeRuneInString(name[len(prefix):])
		return !unicode.IsLower(r)
	}

	var pc uintptr
	var ok bool
	var file string
	var line int
	var name string

	callers := []string{}
	for i := 0; ; i++ {
		pc, file, line, ok = runtime.Caller(i)
		if !ok {
			// The breaks below failed to terminate the loop, and we ran off the
			// end of the call stack.
			break
		}

		// This is a huge edge case, but it will panic if this is the case, see #180
		if file == "<autogenerated>" {
			break
		}

		f := runtime.FuncForPC(pc)
		if f == nil {
			break
		}
		name = f.Name()

		// testing.tRunner is the standard library function that calls
		// tests. Subtests are called directly by tRunner, without going through
		// the Test/Benchmark/Example function that contains the t.Run calls, so
		// with subtests we should break when we hit tRunner, without adding it
		// to the list of callers.
		if name == "testing.tRunner" {
			break
		}

		parts := strings.Split(file, "/")
		if len(parts) > 1 {
			filename := parts[len(parts)-1]
			dir := parts[len(parts)-2]
			if (dir != "assert" && dir != "mock" && dir != "require") || filename == "mock_test.go" {
				callers = append(callers, fmt.Sprintf("%s:%d", file, line))
			}
		}

		// Drop the package
		segments := strings.Split(name, ".")
		name = segments[len(segments)-1]
		if isTest(name, "Test") ||
			isTest(name, "Benchmark") ||
			isTest(name, "Example") {
			break
		}
	}

	return callers
}

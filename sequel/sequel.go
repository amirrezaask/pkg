package sequel

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/amirrezaask/pkg/errors"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var connections = map[string]Interface{}

type metrics struct {
	QueryHV *prometheus.HistogramVec
	ExecHV  *prometheus.HistogramVec
}

type Interface interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, stmt string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	Exec(stmt string, args ...any) (sql.Result, error)
	BeginTx(ctx context.Context, options *sql.TxOptions) (*sql.Tx, error)
	Begin() (*sql.Tx, error)
	Driver() driver.Driver
	Close() error
	SetConnMaxLifetime(d time.Duration)
	SetConnMaxIdleTime(d time.Duration)
	SetMaxIdleConns(int)
	SetMaxOpenConns(int)
}

type DataSource struct {
	Driver                string
	Name                  string
	ConnectionString      string
	MetricsNamespace      string
	MaxOpenConnections    int
	MaxIdleConnections    int
	IdleConnectionTimeout time.Duration
	OpenConnectionTimeout time.Duration
}

const (
	sqlite   = "sqlite3"
	mysql    = "mysql"
	postgres = "postgres"
)

func getDriver(s Interface) string {
	switch fmt.Sprintf("%T", s.Driver()) {
	case "*sqlite3.SQLiteDriver":
		return sqlite
	case "*mysql.MySQLDriver":
		return mysql
	default:
		return "unknown"
	}
}

type database struct {
	Interface
	connectionName string
	metrics        *metrics
	debug          bool
}

type ConnectionOptions struct {
	PromNS                string
	MaxOpenConnections    int
	MaxIdleConnections    int
	IdleConnectionTimeout time.Duration
	OpenConnectionTimeout time.Duration
}

func isDebug(s Interface) bool {
	if os.Getenv("SEQUEL_DBG") == "true" {
		return true
	}
	if our, isOurSql := s.(*database); isOurSql {
		return our.debug
	}

	return false
}

func fromTestConnection(conn *sql.DB, kind string) *database {
	return &database{
		Interface: conn,
	}
}

func Open(driver string, connectionString string) (Interface, error) {
	return sql.Open(driver, connectionString)
}

func New(ds DataSource) (Interface, error) {
	db, err := sql.Open(ds.Driver, ds.ConnectionString)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(ds.OpenConnectionTimeout)
	db.SetConnMaxIdleTime(ds.IdleConnectionTimeout)
	db.SetMaxIdleConns(ds.MaxIdleConnections)
	db.SetMaxOpenConns(ds.MaxOpenConnections)

	var hist *prometheus.HistogramVec
	if !testing.Testing() {
		hist = promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: ds.MetricsNamespace,
			Name:      fmt.Sprintf("%s_db_query_duration_seconds", ds.Name),
			Help:      "Database query durations by [dbName] [query]",
			Buckets: []float64{
				0.0005,
				0.001, // 1ms
				0.002,
				0.005,
				0.01, // 10ms
				0.02,
				0.05,
				0.1, // 100 ms
				0.2,
				0.5,
				1.0, // 1s
				2.0,
				5.0,
				10.0, // 10s
				15.0,
				20.0,
				30.0,
			},
		}, []string{"dbName", "goCall", "type", "table"})
	}

	d := &database{
		Interface:      db,
		connectionName: ds.Name,
		metrics: &metrics{
			QueryHV: hist,
		},
	}

	connections[ds.Name] = d

	return d, nil
}

func extractQueryInfo(query string) (queryType, tableName string) {
	// Normalize the query by removing extra spaces and converting to uppercase.
	query = strings.TrimSpace(strings.ToUpper(query))

	// Define regex patterns for different query types.
	selectPattern := `^\s*SELECT\s+.*\s+FROM\s+(\w+)\s*.*$`
	insertPattern := `^\s*INSERT\s+INTO\s+(\w+)\s*.*$`
	updatePattern := `^\s*UPDATE\s+(\w+)\s*SET\s+.*$`
	deletePattern := `^\s*DELETE\s+FROM\s+(\w+)\s*.*$`

	// Compile regex patterns.
	selectRegex, _ := regexp.Compile(selectPattern)
	insertRegex, _ := regexp.Compile(insertPattern)
	updateRegex, _ := regexp.Compile(updatePattern)
	deleteRegex, _ := regexp.Compile(deletePattern)

	// Check the query type and extract the table name.
	switch {
	case selectRegex.MatchString(query):
		queryType = "SELECT"
		matches := selectRegex.FindStringSubmatch(query)
		tableName = matches[1]
	case insertRegex.MatchString(query):
		queryType = "INSERT"
		matches := insertRegex.FindStringSubmatch(query)
		tableName = matches[1]
	case updateRegex.MatchString(query):
		queryType = "UPDATE"
		matches := updateRegex.FindStringSubmatch(query)
		tableName = matches[1]
	case deleteRegex.MatchString(query):
		queryType = "DELETE"
		matches := deleteRegex.FindStringSubmatch(query)
		tableName = matches[1]
	}

	if queryType == "" {
		queryType = "unknown"
	}
	if tableName == "" {
		tableName = "unknown"
	}

	return queryType, tableName
}

func Debug(i Interface) Interface {
	return &database{
		Interface: i,
		debug:     true,
	}
}

func callerInfo() []string {
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
			rest := strings.Join(parts[:len(parts)-1], "/")
			if (!strings.Contains(rest, "pkg/sequel")) && filename != "sql.go" {
				callers = append(callers, fmt.Sprintf("%s:%d", file, line))
			}
		}
	}
	return callers
}

func debugLog(query string) {
	stack := callerInfo()
	caller := "No Caller Found"
	if len(stack) > 0 {
		caller = stack[len(stack)-1]
	}
	if strings.Contains(caller, "_test.go") && len(stack) > 2 && !strings.Contains(stack[len(stack)-2], "pkg/sequel") {
		caller = stack[len(stack)-2]
	}
	fmt.Printf("%s => %s\n\n", caller, query)
}

func (db *database) SetConnMaxLifetime(d time.Duration) {
	db.Interface.SetConnMaxLifetime(d)
	return
}
func (db *database) SetConnMaxIdleTime(d time.Duration) {
	db.Interface.SetConnMaxIdleTime(d)
	return
}
func (db *database) SetMaxIdleConns(n int) {
	db.Interface.SetMaxIdleConns(n)
	return
}
func (db *database) SetMaxOpenConns(n int) {
	db.Interface.SetMaxOpenConns(n)
	return
}

func (db *database) Exec(query string, args ...any) (sql.Result, error) {

	return db.ExecContext(context.Background(), query, args...)
}

func (db *database) Query(query string, args ...any) (*sql.Rows, error) {

	return db.QueryContext(context.Background(), query, args...)
}

func (db *database) IsDebug() bool {
	return db.debug || testing.Verbose()
}

func (db *database) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if isDebug(db) {
		debugLog(query)
		// dump.This(args)
	}
	if db.metrics == nil || db.metrics.QueryHV == nil {
		return db.Interface.ExecContext(ctx, query, args...)
	}
	queryType, table := extractQueryInfo(query)
	timer := prometheus.NewTimer(db.metrics.QueryHV.WithLabelValues(db.connectionName, strings.ToLower("ExecContext"), strings.ToLower(queryType), strings.ToLower(table)))
	res, err := db.Interface.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	timer.ObserveDuration()
	return res, nil
}
func ToMap(rows *sql.Rows, err error) ([]map[string]interface{}, error) {
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	columnNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Result set
	var result []map[string]interface{}

	for rows.Next() {
		columns := make([]interface{}, len(columnNames))
		columnPointers := make([]interface{}, len(columnNames))
		for i, _ := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		m := make(map[string]interface{})
		for i, colName := range columnNames {
			val := columnPointers[i].(*interface{})
			m[colName] = *val
		}
		result = append(result, m)
	}

	return result, nil
}

func (db *database) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if isDebug(db) {
		debugLog(query)
		// dump.This(args)
	}
	if db.metrics == nil || db.metrics.QueryHV == nil {
		return db.Interface.QueryContext(ctx, query, args...)
	}
	queryType, table := extractQueryInfo(query)
	timer := prometheus.NewTimer(db.metrics.QueryHV.WithLabelValues(db.connectionName, strings.ToLower("QueryContext"), strings.ToLower(queryType), strings.ToLower(table)))
	rows, err := db.Interface.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	timer.ObserveDuration()
	return rows, nil
}

func Insert(obj Record) (sql.Result, error) {
	schema, err := obj.SequelRecordSpec().intoInternalRepr()
	if err != nil {
		return nil, err
	}
	db := connections[schema.connectionName]
	if db == nil {
		return nil, errors.Newf("no connection %s found for model %T", schema.connectionName, obj)
	}
	for _, bi := range schema.beforeWrite {
		err := bi(obj)
		if err != nil {
			return nil, errors.Wrap(err, "error in running before insert hook %T", obj)
		}
	}
	schema, _ = obj.SequelRecordSpec().intoInternalRepr()
	table := schema.table
	columns := schema.fillable
	values := []any{}
	for _, col := range columns {
		values = append(values, getColumnWriteValue(schema, col))
	}
	placeholders := strings.Repeat("?,", len(columns))
	placeholders = placeholders[:len(placeholders)-1]
	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)`,
		table,
		strings.Join(columns, ","),
		placeholders,
	)
	if isDebug(db) {
		debugLog(query)
		// dump.This(values)
		// dump.This(obj)
	}
	return db.Exec(
		query,
		values...,
	)
}

func getColumnWriteValue(schema *recordSpec, col string) any {
	value := schema.valueMap[col]
	if col == schema.createdAtName {
		if schema.valueMap[col] == nil {
			value = "CURRENT_TIMESTAMP"
		} else {
			t, isTime := schema.valueMap[col].(time.Time)
			if !isTime || t.IsZero() {
				value = "CURRENT_TIMESTAMP"
			} else {
				value = t
			}

		}
	}
	if col == schema.updatedAtName {
		if schema.valueMap[col] == nil {
			value = "CURRENT_TIMESTAMP"
		} else {
			t, isTime := schema.valueMap[col].(time.Time)
			if !isTime || t.IsZero() {
				value = "CURRENT_TIMESTAMP"
			} else {
				value = t
			}
		}
	}

	return value
}

func Save[T Record](objs ...T) error {
	if len(objs) < 1 {
		return errors.New("you need to pass at least one model to save.")
	}
	schema, err := objs[0].SequelRecordSpec().intoInternalRepr()
	if err != nil {
		return errors.Wrap(err, "cannot get internal schema from record")
	}
	db := connections[schema.connectionName]
	if db == nil {
		return errors.Newf("no connection %s found for model %T", schema.connectionName, objs[0])
	}

	for _, obj := range objs {
		for _, bi := range schema.beforeWrite {
			err := bi(obj)
			if err != nil {
				return errors.Wrap(err, "error in running before insert hook %T", obj)
			}
		}
	}

	schema, _ = objs[0].SequelRecordSpec().intoInternalRepr()
	table := schema.table
	columns := schema.fillable
	values := []any{}
	updatePairs := []string{}
	for _, col := range columns {
		if getDriver(db) == "mysql" {
			updatePairs = append(updatePairs, fmt.Sprintf("%s=VALUES(%s)", col, col))
		} else if getDriver(db) == "postgres" || getDriver(db) == sqlite {
			updatePairs = append(updatePairs, fmt.Sprintf("%s=excluded.%s", col, col))
		} else {
			return fmt.Errorf("unsupported database '%s' for generating save query", getDriver(db))
		}
	}
	if schema.updatedAtName != "" {
		updatePairs = append(updatePairs, fmt.Sprintf("%s=CURRENT_TIMESTAMP", schema.updatedAtName))
	}

	var valuePlaceholders []string
	for _, obj := range objs {
		thisSchema, err := obj.SequelRecordSpec().intoInternalRepr()
		if err != nil {
			return errors.Wrap(err, "error in turning RecordSpec into internal representation for table: %s", obj.SequelRecordSpec().Table)
		}
		for _, col := range schema.fillable {
			values = append(values, getColumnWriteValue(thisSchema, col))
		}
		placeholders := strings.Repeat("?,", len(schema.fillable))
		placeholders = placeholders[:len(placeholders)-1]
		valuePlaceholders = append(valuePlaceholders, fmt.Sprintf("(%s)", placeholders))
	}
	if getDriver(db) == sqlite || getDriver(db) == "postgres" {
		values = append(values, schema.pk)
	}
	var res sql.Result
	if getDriver(db) == "mysql" {
		query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES %s ON DUPLICATE KEY UPDATE %s`,
			table,
			strings.Join(columns, ","),
			strings.Join(valuePlaceholders, ","),
			strings.Join(updatePairs, ","),
		)
		if isDebug(db) {
			debugLog(query)
			// for _, obj := range objs {
			// 	dump.This(obj)
			// }
		}
		res, err = db.Exec(
			query,
			values...,
		)
	} else if getDriver(db) == "postgres" || getDriver(db) == sqlite {
		query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES %s ON CONFLICT (id) DO UPDATE SET %s WHERE ID=?`,
			table,
			strings.Join(columns, ","),
			strings.Join(valuePlaceholders, ","),
			strings.Join(updatePairs, ","),
		)
		if isDebug(db) {
			debugLog(query)
			for _, obj := range objs {
				fmt.Printf(">>>>> %+v\n", obj)
				// dump.This(obj)
			}
		}
		res, err = db.Exec(
			query,
			values...,
		)
	} else {
		return fmt.Errorf("unsupported database '%s' for generating save query", getDriver(db))
	}

	if err != nil {
		return errors.Wrap(err, "cannot save on table %s", table)
	}

	id, err := res.LastInsertId()
	if err == nil {
		if schema.pk != nil {
			*schema.pk = id
		}
	}

	return nil
}

func Delete(m Record) (sql.Result, error) {
	schema, err := m.SequelRecordSpec().intoInternalRepr()
	if err != nil {
		return nil, err
	}
	db := connections[schema.connectionName]
	if db == nil {
		return nil, errors.Newf("no connection %s found for model %T", schema.connectionName, m)
	}
	pkName := schema.pkName
	pkValue := schema.valueMap[pkName]
	return db.Exec(fmt.Sprintf("DELETE FROM %s WHERE %s=?", schema.table, pkName), pkValue)
}

type pointer[T any] interface {
	SequelRecordSpec() RecordSpec
	*T
}

func Scan[M any, T pointer[M]](rows *sql.Rows, err error) ([]T, error) {
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("cannot scan from rows since we have an error: %w", err)
	}
	columns, err := rows.Columns()

	records := []T{}

	for rows.Next() {
		var m M
		s := T(&m)
		values := []any{}
		for _, col := range columns {
			internal, err := s.SequelRecordSpec().intoInternalRepr()
			if err != nil {
				return nil, errors.Wrap(err, "error in converting to internal repr")
			}
			values = append(values, internal.valueMap[col])
		}
		err := rows.Scan(values...)
		if err != nil {
			return nil, errors.Wrap(err, "error in scanning to model of table %s", s.SequelRecordSpec().Table)
		}
		thisSchema, err := s.SequelRecordSpec().intoInternalRepr()
		if err != nil {
			return nil, errors.Wrap(err, "error in turning into internal repr")
		}
		for _, ar := range thisSchema.afterRead {
			err := ar(s)
			if err != nil {
				return nil, errors.Wrap(err, "error in processing after read for model of table %s", thisSchema.table)
			}
		}
		records = append(records, s)
	}

	return records, nil
}

func Query[M any, T pointer[M]](q string, args ...any) ([]T, error) {
	var m M
	s := T(&m)
	internal, err := s.SequelRecordSpec().intoInternalRepr()
	if err != nil {
		return nil, err
	}
	connectionName := internal.connectionName
	db := connections[connectionName]
	if db == nil {
		return nil, errors.Newf("no connection %s found for model %T", connectionName, m)
	}
	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return Scan[M, T](rows, err)
}

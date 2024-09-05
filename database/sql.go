package database

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/amirrezaask/go-sith/errors"
	"reflect"
	"regexp"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	QueryHV *prometheus.HistogramVec
	ExecHV  *prometheus.HistogramVec
}

type Sql interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, stmt string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	Exec(stmt string, args ...any) (sql.Result, error)
	BeginTx(ctx context.Context, options *sql.TxOptions) (*sql.Tx, error)
	Save(obj Model) error
	NOW() string
	SetDbName(s string)
	GetDbName() string
}

type SqlDatabase struct {
	*sql.DB
	DbName  string
	kind    string
	metrics *Metrics
}

type SqlConnectionOptions struct {
	PromNS                string
	MaxOpenConnections    int
	MaxIdleConnections    int
	IdleConnectionTimeout time.Duration
	OpenConnectionTimeout time.Duration
}

func FromTestConnection(conn *sql.DB, kind string) *SqlDatabase {
	return &SqlDatabase{
		kind: kind,
		DB:   conn,
	}
}

func NewMysql(connectionString string, dbName string, options SqlConnectionOptions) (*SqlDatabase, error) {
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(options.OpenConnectionTimeout)
	db.SetConnMaxIdleTime(options.IdleConnectionTimeout)
	db.SetMaxIdleConns(options.MaxIdleConnections)
	db.SetMaxOpenConns(options.MaxOpenConnections)

	hist := promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: options.PromNS,
		Name:      fmt.Sprintf("%s_db_query_duration_seconds", dbName),
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

	return &SqlDatabase{
		kind:   "mysql",
		DB:     db,
		DbName: dbName,
		metrics: &Metrics{
			QueryHV: hist,
		},
	}, nil
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
func (db *SqlDatabase) Exec(query string, args ...any) (sql.Result, error) {
	return db.ExecContext(context.Background(), query, args...)
}

func (db *SqlDatabase) Query(query string, args ...any) (*sql.Rows, error) {
	return db.QueryContext(context.Background(), query, args...)
}

func (db *SqlDatabase) SetDbName(s string) {
	db.DbName = s
}

func (db *SqlDatabase) GetDbName() string { return db.DbName }

func (db *SqlDatabase) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if db.metrics == nil {
		return db.DB.ExecContext(ctx, query, args...)
	}
	queryType, table := extractQueryInfo(query)
	timer := prometheus.NewTimer(db.metrics.QueryHV.WithLabelValues(db.DbName, "ExecContext", queryType, table))
	res, err := db.DB.ExecContext(ctx, query, args...)
	timer.ObserveDuration()

	return res, err
}
func ToMap(rows *sql.Rows, err error) ([]map[string]interface{}, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (db *SqlDatabase) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if db.metrics == nil {
		return db.DB.QueryContext(ctx, query, args...)
	}
	queryType, table := extractQueryInfo(query)
	timer := prometheus.NewTimer(db.metrics.QueryHV.WithLabelValues(db.DbName, "QueryContext", queryType, table))
	rows, err := db.DB.QueryContext(ctx, query, args...)
	timer.ObserveDuration()

	return rows, err
}

type Schema struct {
	pk           *int64
	fillable     []string
	table        string
	valueMap     map[string]any
	beforeInsert []func(m Model) error
}

func MakeSchema() *Schema {
	return &Schema{
		valueMap: map[string]any{},
	}
}

func (s *Schema) Table(t string) *Schema {
	s.table = t
	return s
}
func (s *Schema) Fillables(fs ...string) *Schema {
	s.fillable = append(s.fillable, fs...)
	return s
}
func (s *Schema) Fields(kvs ...any) *Schema {
	if len(kvs)%2 != 0 {
		panic("Fields input should be even number to be a correct key-value pair")
	}
	for i := 0; i < len(kvs); i++ {
		if i%2 == 0 {
			s.MapField(fmt.Sprint(kvs[i]), kvs[i+1])
		}
	}

	return s
}
func (s *Schema) MapField(key string, value any) *Schema {
	if reflect.ValueOf(value).Kind() != reflect.Pointer {
		// If it's not a pointer, it's not addressable in memory later when we want to set in it.
		panic("field map value should always be a pointer")
	}
	s.valueMap[key] = value
	return s
}
func (s *Schema) PrimaryKey(val *int64) *Schema {
	s.pk = val
	return s
}

func (s *Schema) BeforeInsert(fs ...func(m Model) error) *Schema {
	s.beforeInsert = append(s.beforeInsert, fs...)
	return s
}

func (s *Schema) validate() error {
	if s.table == "" {
		return errors.Newf("No table has been defined")
	}
	if len(s.fillable) < 1 {
		return errors.Newf("No fillables defined for model of table %s", s.table)
	}
	if s.pk == nil {
		return errors.Newf("No primary key pointer was set for model of table %s", s.table)
	}

	if len(s.valueMap) == 0 {
		return errors.Newf("no field mapping defined for model of table %s", s.table)
	}
	//TODO(amirreza): optional reflection check for validity of types.

	return nil
}

type Model interface {
	Schema() *Schema
}

func Insert[I Model](db *SqlDatabase, obj I) (sql.Result, error) {
	schema := obj.Schema()
	if err := schema.validate(); err != nil {
		return nil, errors.Wrap(err, "invalid schema")
	}
	for _, bi := range schema.beforeInsert {
		err := bi(obj)
		if err != nil {
			return nil, errors.Wrap(err, "error in running before insert hook %T", obj)
		}
	}
	schema = obj.Schema()
	table := schema.table
	columns := schema.fillable
	values := []any{}
	for _, col := range columns {
		values = append(values, schema.valueMap[col])
	}
	placeholders := strings.Repeat("?,", len(columns))
	placeholders = placeholders[:len(placeholders)-1]
	return db.Exec(
		fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)`,
			table,
			strings.Join(columns, ","),
			placeholders,
		),
		values...,
	)
}

func (db *SqlDatabase) Save(obj Model) error {
	if db.kind == "mysql" {
		schema := obj.Schema()
		if err := schema.validate(); err != nil {
			return errors.Wrap(err, "invalid schema")
		}
		for _, bi := range schema.beforeInsert {
			err := bi(obj)
			if err != nil {
				return errors.Wrap(err, "error in running before insert hook %T", obj)
			}
		}
		schema = obj.Schema()
		table := schema.table
		columns := schema.fillable
		values := []any{}
		updatePairs := []string{}
		for _, col := range columns {
			values = append(values, schema.valueMap[col], schema.valueMap[col])
			updatePairs = append(updatePairs, fmt.Sprintf("%s=?", col))
		}
		placeholders := strings.Repeat("?,", len(columns))
		placeholders = placeholders[:len(placeholders)-1]
		res, err := db.Exec(
			fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) ON DUPLICATE KEY UPDATE %s`,
				table,
				strings.Join(columns, ","),
				placeholders,
				strings.Join(updatePairs, ","),
			),
			values...,
		)

		if err != nil {
			return errors.Wrap(err, "cannot save")
		}

		id, err := res.LastInsertId()
		if err == nil {

			if schema.pk != nil {
				*schema.pk = id
			}
		}

		return nil
	} else if db.kind == "sqlite" {
		schema := obj.Schema()
		if err := schema.validate(); err != nil {
			return errors.Wrap(err, "invalid schema")
		}
		for _, bi := range schema.beforeInsert {
			err := bi(obj)
			if err != nil {
				return errors.Wrap(err, "error in running before insert hook %T", obj)
			}
		}
		schema = obj.Schema()
		table := schema.table
		columns := schema.fillable
		values := []any{}
		updatePairs := []string{}
		updateValues := []any{}
		for _, col := range columns {
			values = append(values, schema.valueMap[col])
			updateValues = append(updateValues, schema.valueMap[col])
			updatePairs = append(updatePairs, fmt.Sprintf("%s=?", col))
		}
		updateValues = append(updateValues, schema.pk)

		placeholders := strings.Repeat("?,", len(columns))
		placeholders = placeholders[:len(placeholders)-1]
		query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (id) DO UPDATE SET %s WHERE ID=?`,
			table,
			strings.Join(columns, ","),
			placeholders,
			strings.Join(updatePairs, ","),
		)

		res, err := db.Exec(
			query,
			append(values, updateValues...)...,
		)

		if err != nil {
			return errors.Wrap(err, "cannot save")
		}

		id, err := res.LastInsertId()
		if err == nil {
			if schema.pk != nil {
				*schema.pk = id
			}
		}

		return nil
	} else {
		return errors.Newf("save function does not support %s databases yet.", db.kind)
	}

	return nil
}

func (db *SqlDatabase) NOW() string {
	if db.kind == "mysql" {
		return "NOW()"
	} else if db.kind == "sqlite" {
		return "datetime()"
	}

	return "NOW()"
}

func Query[T Model](db *SqlDatabase, q string, args ...any) ([]T, error) {
	schema := (*new(T)).Schema()
	q = strings.Replace(q, "^Table", schema.table, -1)
	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	columns, err := rows.Columns()

	records := []T{}

	for rows.Next() {
		s := *new(T)
		values := []any{}
		for _, col := range columns {
			values = append(values, s.Schema().valueMap[col])
		}
		rows.Scan(values...)
		records = append(records, s)
	}

	return records, nil
}

package sequel

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/amirrezaask/pkg/env"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type QueryExecerContext interface {
	Driver() driver.Driver
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

type DBConfig struct {
	Driver                 string
	Username               string
	Password               string
	Host                   string
	Port                   int
	DBName                 string
	MetricsNamespace       string
	MaxOpenConnections     int
	MaxIdleConnections     int
	IdleConnectionTimeout  time.Duration
	OpenConnectionTimeout  time.Duration
	PrometheusHistogram    bool
	PrometheusErrorCounter bool
	ParseTime              bool
}

type DB struct {
	db           *sql.DB
	cfg          DBConfig
	hist         *prometheus.HistogramVec
	errorCounter *prometheus.CounterVec
	isDebug      bool
}

var _ QueryExecerContext = &DB{}

type Tx struct {
	tx           *sql.Tx
	cfg          DBConfig
	startTimer   *prometheus.Timer
	hist         *prometheus.HistogramVec
	errorCounter *prometheus.CounterVec
	isDebug      bool
}

func (cfg *DBConfig) OpenArgs() (string, string, DBConfig) {
	return cfg.Driver, cfg.ConnectionString(), *cfg
}

func (cfg *DBConfig) ConnectionString() string {
	url := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	if cfg.ParseTime {
		url += "?parseTime=true"
	}
	return url
}

// returns connection string uri with masked password (useful for logging)
func (cfg *DBConfig) ConnectionStringMaskedPassword() string {
	url := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", cfg.Username, strings.Repeat("*", len(cfg.Password)), cfg.Host, cfg.Port, cfg.DBName)
	if cfg.ParseTime {
		url += "?parseTime=true"
	}
	return url
}

func Open(driver string, connectionString string, cfgs ...DBConfig) (*DB, error) {
	db, err := sql.Open(driver, connectionString)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	var hist *prometheus.HistogramVec
	var counter *prometheus.CounterVec

	if len(cfgs) > 0 {
		cfg := cfgs[0]
		db.SetMaxIdleConns(cfg.MaxIdleConnections)
		db.SetMaxOpenConns(cfg.MaxOpenConnections)
		db.SetConnMaxIdleTime(cfg.IdleConnectionTimeout)
		db.SetConnMaxLifetime(cfg.OpenConnectionTimeout)

		if !testing.Testing() {
			if cfg.PrometheusHistogram {
				hist = promauto.NewHistogramVec(prometheus.HistogramOpts{
					Namespace: cfg.MetricsNamespace,
					Name:      fmt.Sprintf("%s_db_query_duration_seconds", cfg.DBName),
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
			if cfg.PrometheusErrorCounter {
				counter = promauto.NewCounterVec(
					prometheus.CounterOpts{
						Namespace: cfg.MetricsNamespace,
						Name:      fmt.Sprintf("%s_db_query_failure_count", cfg.DBName),
						Help:      "Database query failures by [dbName] [query]",
					},
					[]string{"dbName", "goCall", "type", "table", "error"},
				)
			}
		}

	}

	return &DB{db: db, hist: hist, errorCounter: counter, cfg: cfgs[0]}, nil
}

// NewFromEnv creates `DBConfig` from env. It will fetch env values for keys `uppercase(prefix)`+["_"]+<"DATABASE_NAME"|"DATABASE_USER"|"DATABASE_PASS"|"DATABASE_HOST"|"DATABASE_PORT"> if `maybeDefaults` is not set or contains no value for those keys.
func NewConfigFromEnv(appName string, prefix string, maybeDefaults *DBConfig) DBConfig {
	if prefix != "" && !strings.HasSuffix(prefix, "_") {
		prefix += "_"
	}
	prefix = strings.ToUpper(prefix)
	cfg := DBConfig{
		Driver:                 "mysql",
		MaxOpenConnections:     50,
		MaxIdleConnections:     50,
		IdleConnectionTimeout:  10 * time.Second,
		OpenConnectionTimeout:  10 * time.Second,
		PrometheusHistogram:    true,
		PrometheusErrorCounter: true,
	}
	if maybeDefaults != nil {
		// update `cfg` from `maybeDefaults`
		if maybeDefaults.Driver != "" {
			cfg.Driver = maybeDefaults.Driver
		}
		if maybeDefaults.MaxOpenConnections != 0 {
			cfg.MaxOpenConnections = maybeDefaults.MaxOpenConnections
		}
		if maybeDefaults.MaxIdleConnections != 0 {
			cfg.MaxIdleConnections = maybeDefaults.MaxIdleConnections
		}
		if maybeDefaults.OpenConnectionTimeout != 0 {
			cfg.OpenConnectionTimeout = maybeDefaults.OpenConnectionTimeout
		}
		if maybeDefaults.IdleConnectionTimeout != 0 {
			cfg.IdleConnectionTimeout = maybeDefaults.IdleConnectionTimeout
		}
		cfg.DBName = maybeDefaults.DBName
		cfg.Username = maybeDefaults.Username
		cfg.Password = maybeDefaults.Password
		cfg.Host = maybeDefaults.Host
		cfg.Port = maybeDefaults.Port
		cfg.ParseTime = maybeDefaults.ParseTime
		cfg.PrometheusHistogram = maybeDefaults.PrometheusHistogram
		cfg.PrometheusErrorCounter = maybeDefaults.PrometheusErrorCounter
	}
	cfg.MetricsNamespace = appName
	cfg.Driver = env.Default(prefix+"DATABASE_DRIVER", cfg.Driver)
	// if some values are not set in `maybeDefaults`, fetch them from env:
	if cfg.DBName == "" {
		cfg.DBName = env.RequiredNotEmpty(prefix + "DATABASE_NAME")
	}
	if cfg.Username == "" {
		cfg.Username = env.RequiredNotEmpty(prefix + "DATABASE_USER")
	}
	if cfg.Password == "" {
		cfg.Password = env.Required(prefix + "DATABASE_PASS")
	}
	if cfg.Host == "" {
		cfg.Host = env.RequiredNotEmpty(prefix + "DATABASE_HOST")
	}
	if cfg.Port == 0 {
		portStr := env.RequiredNotEmpty(prefix + "DATABASE_PORT")
		port, err := strconv.ParseInt(portStr, 10, 64)
		if err != nil || port == 0 {
			err = fmt.Errorf(prefix + "DATABASE_PORT is set to `" + portStr + "`")
			panic(err)
		}
		cfg.Port = int(port)
	}
	return cfg
}

func New(ds DBConfig) (*DB, error) { return Open(ds.Driver, ds.ConnectionString(), ds) }

// NewFromEnv creates `DBConfig` from env and calls `New(DBConfig)`. It will fetch env values for keys `uppercase(prefix)`+["_"]+<"DATABASE_NAME"|"DATABASE_USER"|"DATABASE_PASS"|"DATABASE_HOST"|"DATABASE_PORT"> if `maybeDefaults` is not set or contains no value for those keys.
func NewFromEnv(appName string, prefix string, maybeDefaults *DBConfig) (*DB, error) {
	cfg := NewConfigFromEnv(appName, prefix, maybeDefaults)
	slog.Debug("created database configuration from env", "url", cfg.ConnectionStringMaskedPassword())
	return New(cfg)
}

func (db *DB) Debug() *DB {
	return &DB{
		db:           db.db,
		cfg:          db.cfg,
		hist:         db.hist,
		errorCounter: db.errorCounter,
		isDebug:      true,
	}
}

func (db *DB) maybeStartTimer(functionName, queryType, tableName string) *prometheus.Timer {
	if db.hist != nil {
		return prometheus.NewTimer(
			db.hist.WithLabelValues(
				db.cfg.DBName,
				strings.ToLower(functionName),
				strings.ToLower(queryType),
				strings.ToLower(tableName),
			),
		)
	} else {
		return nil
	}
}

func (db *DB) maybeObserveDuration(maybeTimer *prometheus.Timer) {
	if maybeTimer != nil {
		maybeTimer.ObserveDuration()
	}
}

func (db *DB) maybeIncreaseErrorCounter(functionName, queryType, tableName string, err error) {
	if db.errorCounter != nil {
		db.errorCounter.WithLabelValues(
			db.cfg.DBName,
			strings.ToLower(functionName),
			strings.ToLower(queryType),
			strings.ToLower(tableName),
			dbErrText(err),
		).Inc()
	}
}

func (db *DB) Stats() sql.DBStats {
	return db.db.Stats()
}

func (db *DB) Exec(query string, args ...any) (sql.Result, error) {
	return db.ExecContext(context.Background(), query, args...)
}

func (db *DB) Query(query string, args ...any) (*sql.Rows, error) {
	return db.QueryContext(context.Background(), query, args...)
}

func (db *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if isDebug(db) {
		debugLog(query)
		// dump.This(args)
	}
	queryType, table := extractQueryInfo(query)
	maybeTimer := db.maybeStartTimer("Exec", queryType, table)
	res, err := db.db.ExecContext(ctx, query, args...)
	if err != nil {
		db.maybeIncreaseErrorCounter("Exec", queryType, table, err)
		return nil, err
	}
	db.maybeObserveDuration(maybeTimer)
	return res, nil
}

func (db *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if isDebug(db) {
		debugLog(query)
		// dump.This(args)
	}
	queryType, table := extractQueryInfo(query)
	maybeTimer := db.maybeStartTimer("Query", queryType, table)
	rows, err := db.db.QueryContext(ctx, query, args...)
	if err != nil {
		db.maybeIncreaseErrorCounter("Query", queryType, table, err)
		return nil, err
	}
	db.maybeObserveDuration(maybeTimer)
	return rows, nil
}

func (db *DB) Ping() error {
	return db.PingContext(context.Background())
}

func (db *DB) PingContext(ctx context.Context) error {
	maybeTimer := db.maybeStartTimer("Ping", "ping", "undefined")
	err := db.db.PingContext(ctx)
	if err != nil {
		db.maybeIncreaseErrorCounter("Ping", "ping", "undefined", err)
		return err
	}
	db.maybeObserveDuration(maybeTimer)
	return nil
}

func (db *DB) Prepare(query string) (*sql.Stmt, error) {
	return db.PrepareContext(context.Background(), query)
}

func (db *DB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	queryType, table := extractQueryInfo(query)
	maybeTimer := db.maybeStartTimer("Prepare", queryType, table)
	stmt, err := db.db.PrepareContext(ctx, query)
	if err != nil {
		db.maybeIncreaseErrorCounter("Prepare", queryType, table, err)
		return nil, err
	}
	db.maybeObserveDuration(maybeTimer)
	return stmt, nil
}

func (db *DB) QueryRow(query string, args ...any) *sql.Row {
	return db.QueryRowContext(context.Background(), query, args...)
}

func (db *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	queryType, table := extractQueryInfo(query)
	maybeTimer := db.maybeStartTimer("QueryRow", queryType, table)
	row := db.db.QueryRowContext(ctx, query, args...)
	if row == nil {
		db.maybeIncreaseErrorCounter("QueryRow", queryType, table, sql.ErrNoRows)
	}
	db.maybeObserveDuration(maybeTimer)
	return row
}

func (db *DB) Begin() (*Tx, error) {
	return db.BeginTx(context.Background(), nil)
}

func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	queryType, table := "begin", "undefined"
	maybeTimer := db.maybeStartTimer("Begin", queryType, table)
	tx, err := db.db.BeginTx(ctx, opts)
	if err != nil {
		db.maybeIncreaseErrorCounter("Begin", queryType, table, err)
		return nil, err
	}
	db.maybeObserveDuration(maybeTimer)
	txTimer := db.maybeStartTimer("tx", queryType, table)
	return &Tx{tx: tx, cfg: db.cfg, isDebug: db.isDebug, startTimer: txTimer, hist: db.hist, errorCounter: db.errorCounter}, nil
}

func (db *DB) debug() bool { return db.isDebug }

func (db *DB) Close() error { return db.db.Close() }

func (db *DB) Conn(ctx context.Context) (*sql.Conn, error) {
	return db.db.Conn(ctx)
}

func (db *DB) SetConnMaxLifetime(d time.Duration) {
	db.db.SetConnMaxLifetime(d)
	return
}

func (db *DB) SetConnMaxIdleTime(d time.Duration) {
	db.db.SetConnMaxIdleTime(d)
	return
}

func (db *DB) Driver() driver.Driver {
	return db.db.Driver()
}

func (db *DB) SetMaxIdleConns(n int) {
	db.db.SetMaxIdleConns(n)
	return
}

func (db *DB) SetMaxOpenConns(n int) {
	db.db.SetMaxOpenConns(n)
	return
}

func (tx *Tx) debug() bool { return tx.isDebug }

func (tx *Tx) maybeStartTimer(functionName, queryType, tableName string) *prometheus.Timer {
	if tx.hist != nil {
		return prometheus.NewTimer(
			tx.hist.WithLabelValues(
				tx.cfg.DBName,
				"tx_"+strings.ToLower(functionName),
				"tx_"+strings.ToLower(queryType),
				strings.ToLower(tableName),
			),
		)
	} else {
		return nil
	}
}

func (tx *Tx) maybeObserveDuration(maybeTimer *prometheus.Timer) {
	if maybeTimer != nil {
		maybeTimer.ObserveDuration()
	}
}

func (tx *Tx) maybeIncreaseErrorCounter(functionName, queryType, tableName string, err error) {
	if tx.errorCounter != nil {
		tx.errorCounter.WithLabelValues(
			tx.cfg.DBName,
			"tx_"+strings.ToLower(functionName),
			"tx_"+strings.ToLower(queryType),
			strings.ToLower(tableName),
			dbErrText(err),
		).Inc()
	}
}

func (tx *Tx) Commit() error {
	queryType, tableName := "commit", "undefined"
	maybeTimer := tx.maybeStartTimer("Commit", queryType, tableName)
	err := tx.tx.Commit()
	if tx.startTimer != nil { // this is for how much we hold `tx` before commit
		tx.startTimer.ObserveDuration()
	}
	if err != nil {
		tx.maybeIncreaseErrorCounter("Commit", queryType, tableName, err)
		return err
	}
	// this is for commit operation itself
	tx.maybeObserveDuration(maybeTimer)
	return nil
}

func (tx *Tx) Rollback() error {
	queryType, tableName := "rollback", "undefined"
	maybeTimer := tx.maybeStartTimer("Rollback", queryType, tableName)
	err := tx.tx.Rollback()
	if tx.startTimer != nil { // this is for how much we hold `tx` before rollback
		tx.startTimer.ObserveDuration()
	}
	if err != nil {
		tx.maybeIncreaseErrorCounter("Rollback", queryType, tableName, err)
		return err
	}
	// this is for rollback operation itself
	tx.maybeObserveDuration(maybeTimer)
	return nil
}

func (tx *Tx) Exec(query string, args ...any) (sql.Result, error) {
	return tx.ExecContext(context.Background(), query, args...)
}

func (tx *Tx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if tx.isDebug {
		debugLog(query)
		// dump.This(args)
	}
	queryType, table := extractQueryInfo(query)
	maybeTimer := tx.maybeStartTimer("Exec", queryType, table)
	res, err := tx.tx.ExecContext(ctx, query, args...)
	if err != nil {
		tx.maybeIncreaseErrorCounter("Exec", queryType, table, err)
		return nil, err
	}
	tx.maybeObserveDuration(maybeTimer)
	return res, nil
}

func (tx *Tx) Prepare(query string) (*sql.Stmt, error) { return tx.tx.Prepare(query) }

func (tx *Tx) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	queryType, table := extractQueryInfo(query)
	maybeTimer := tx.maybeStartTimer("Prepare", queryType, table)
	stmt, err := tx.tx.PrepareContext(ctx, query)
	if err != nil {
		tx.maybeIncreaseErrorCounter("Prepare", queryType, table, err)
		return nil, err
	}
	tx.maybeObserveDuration(maybeTimer)
	return stmt, nil
}

func (tx *Tx) Query(query string, args ...any) (*sql.Rows, error) {
	return tx.QueryContext(context.Background(), query, args...)
}

func (tx *Tx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if tx.isDebug {
		debugLog(query)
		// dump.This(args)
	}
	queryType, table := extractQueryInfo(query)
	maybeTimer := tx.maybeStartTimer("Query", queryType, table)
	rows, err := tx.tx.QueryContext(ctx, query, args...)
	if err != nil {
		tx.maybeIncreaseErrorCounter("Query", queryType, table, err)
		return nil, err
	}
	tx.maybeObserveDuration(maybeTimer)
	return rows, nil
}

func (tx *Tx) QueryRow(query string, args ...any) *sql.Row {
	return tx.QueryRowContext(context.Background(), query, args...)
}

func (tx *Tx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	queryType, table := extractQueryInfo(query)
	maybeTimer := tx.maybeStartTimer("QueryRow", queryType, table)
	row := tx.tx.QueryRowContext(ctx, query, args...)
	if row == nil {
		tx.maybeIncreaseErrorCounter("QueryRow", queryType, table, sql.ErrNoRows)
	}
	tx.maybeObserveDuration(maybeTimer)
	return row
}

func (tx *Tx) Stmt(stmt *sql.Stmt) *sql.Stmt {
	return tx.StmtContext(context.Background(), stmt)
}

func (tx *Tx) StmtContext(ctx context.Context, stmt *sql.Stmt) *sql.Stmt {
	queryType, table := "statement", "undefined"
	maybeTimer := tx.maybeStartTimer("Stmt", queryType, table)
	newStmt := tx.tx.StmtContext(ctx, stmt)
	if newStmt == nil {
		tx.maybeIncreaseErrorCounter("Stmt", queryType, table, sql.ErrNoRows)
		return nil
	}
	tx.maybeObserveDuration(maybeTimer)
	return newStmt
}

func (tx *Tx) GetConfig() DBConfig { return tx.cfg }

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

func isDebug(db *DB) bool {
	return db.isDebug || os.Getenv("SEQUEL_DBG") == "true"
}

var replaceParamsRe = regexp.MustCompile(`\{[a-zA-Z_][a-zA-Z0-9_]*\}`)

func ReplaceNamedParamsWithPositional(query string) string {
	return replaceParamsRe.ReplaceAllString(query, "?")
}

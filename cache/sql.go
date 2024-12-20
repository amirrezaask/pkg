package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/amirrezaask/pkg/errors"
	"github.com/amirrezaask/pkg/sequel"
)

type sqlCacher struct {
	*sequel.DB
	namespace string
	driver    string
}

func NewSqlCacher(ctx context.Context, db *sequel.DB, cacheNamespace string) (Cacher, error) {
	//migrate database if needed.
	driver := fmt.Sprintf("%T", db.Driver()) // to not import sql drivers here as well.

	var createTable string
	switch driver {
	case "*mysql.MysqlDriver":
		createTable = "CREATE TABLE sqlcacher_store (" +
			"id INT AUTO_INCREMENT PRIMARY KEY," +
			"namespace VARCHAR(255) NOT NULL," +
			"`key` VARCHAR(255) NOT NULL," +
			"`value` TEXT NOT NULL," +
			"expires_at DATETIME," +
			"UNIQUE INDEX idx_namespace_key (namespace, `key`)" +
			");"

	case "*sqlite3.SQLiteDriver":
		createTable = "CREATE TABLE sqlcacher_store (" +
			"id INTEGER PRIMARY KEY AUTOINCREMENT," +
			"namespace TEXT NOT NULL," +
			"`key` TEXT NOT NULL," +
			"`value` TEXT NOT NULL," +
			"expires_at DATETIME," +
			"UNIQUE(namespace, `key`)" +
			");"

	default:
		return nil, errors.Newf("error in creating sql cacher, unsupported database driver: %s", driver)
	}

	_, err := db.ExecContext(ctx, createTable)
	if err != nil {
		return nil, errors.Wrap(err, "error in creating table sqlcacher_store")
	}

	return &sqlCacher{
		DB:        db,
		namespace: cacheNamespace,
		driver:    driver,
	}, nil
}

func (s *sqlCacher) Remember(ctx context.Context, key string, value any, ttl time.Duration) error {
	bs, err := json.Marshal(value)
	if err != nil {
		return errors.Wrap(err, "error in marshaling value into json string")
	}
	switch s.driver {
	case "*mysql.MysqlDriver":
		_, err := s.ExecContext(ctx, "INSERT INTO sqlcacher_store (namespace, `key`, `value`, expires_at) VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE `value` = VALUES(`value`), expires_at = VALUES(expires_at);", key, bs, time.Now().Add(ttl))

		if err != nil {
			return errors.Wrap(err, "error in inserting new cache entry")
		}

		return nil
	case "*sqlite3.SQLiteDriver":
		_, err := s.ExecContext(ctx, "INSERT INTO sqlcacher_store (namespace, `key`, `value`, expires_at) VALUES (?,?,?,?) ON CONFLICT(namespace, `key`) DO UPDATE SET `value` = excluded.`value`, expires_at = excluded.expires_at;", key, bs, time.Now().Add(ttl))
		if err != nil {
			return errors.Wrap(err, "error in inserting new cache entry")
		}
		return nil

	default:
		return errors.Newf("unsupported driver '%s'", s.driver)
	}

}

func (s *sqlCacher) Get(ctx context.Context, key string) (any, error) {
	rows, err := s.QueryContext(ctx, "SELECT id, `value`, expires_at FROM sqlcacher_store WHERE namespace= ? AND `key` = ?", s.namespace, key)
	if err != nil {
		return nil, errors.Wrap(err, "error in querying for sqlcacher get")
	}

	if !rows.Next() {
		return nil, ErrNoEntry
	}

	var id int64
	var value string
	var expiresAt time.Time

	err = rows.Scan(&id, &value, &expiresAt)
	if err != nil {
		return nil, errors.Wrap(err, "cannot scan from values returned by database in sqlcacher_Get")
	}

	if time.Now().After(expiresAt) {
		s.ExecContext(ctx, "DELETE FROM sqlcacher_store WHERE id=?", id)
		return nil, ErrEntryExpired
	}

	return value, nil
}

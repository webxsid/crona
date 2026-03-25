package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

type Store struct {
	sqlDB *sql.DB
	db    *bun.DB
}

func Open(dbPath string) (*Store, error) {
	sqlDB, err := sql.Open(sqliteshim.ShimName, fmt.Sprintf("file:%s?_busy_timeout=5000&_journal_mode=WAL", dbPath))
	if err != nil {
		return nil, err
	}
	// SQLite behaves much more predictably with a single shared connection for
	// this local kernel process; leaving the default pool enabled can surface
	// SQLITE_BUSY during overlapping RPCs and health checks.
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	db := bun.NewDB(sqlDB, sqlitedialect.New())
	return &Store{
		sqlDB: sqlDB,
		db:    db,
	}, nil
}

func (s *Store) Close() error {
	if s.db != nil {
		_ = s.db.Close()
	}
	if s.sqlDB != nil {
		return s.sqlDB.Close()
	}
	return nil
}

func (s *Store) DB() *bun.DB {
	return s.db
}

func (s *Store) Ping(ctx context.Context) error {
	return s.sqlDB.PingContext(ctx)
}

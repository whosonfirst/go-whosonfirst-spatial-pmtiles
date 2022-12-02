package database

import (
	"context"
	"database/sql"
	"github.com/aaronland/go-sqlite/v2"
	"log"
	"sync"
)

type AnyDatabase struct {
	sqlite.Database
	conn   *sql.DB
	dsn    string
	mu     *sync.Mutex
	logger *log.Logger
}

func NewAnyDatabase(ctx context.Context, dsn string, conn *sql.DB) (sqlite.Database, error) {

	mu := new(sync.Mutex)

	logger := log.Default()

	db := AnyDatabase{
		conn:   conn,
		dsn:    dsn,
		mu:     mu,
		logger: logger,
	}

	return &db, nil
}

func (db *AnyDatabase) Lock(ctx context.Context) error {
	db.mu.Lock()
	return nil
}

func (db *AnyDatabase) Unlock(ctx context.Context) error {
	db.mu.Unlock()
	return nil
}

func (db *AnyDatabase) Conn(ctx context.Context) (*sql.DB, error) {
	return db.conn, nil
}

func (db *AnyDatabase) Close(ctx context.Context) error {
	return db.conn.Close()
}

func (db *AnyDatabase) DSN(ctx context.Context) string {
	return db.dsn
}

func (db *AnyDatabase) SetLogger(ctx context.Context, logger *log.Logger) error {
	db.logger = logger
	return nil
}

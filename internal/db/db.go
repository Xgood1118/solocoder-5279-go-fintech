package db

import (
	"context"
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

func New(path string) (*DB, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	dsn := path + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout=5000&_pragma=foreign_keys(ON)"
	sqldb, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	sqldb.SetMaxOpenConns(1)
	sqldb.SetMaxIdleConns(1)

	db := &DB{sqldb}

	if err := db.Migrate(); err != nil {
		sqldb.Close()
		return nil, err
	}

	return db, nil
}

func (db *DB) Migrate() error {
	log.Println("Running database migrations...")

	migrations := []string{
		migration001,
		migration002,
		migration003,
		migration004,
		migration005,
		migration006,
		migration007,
	}

	for i, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			log.Printf("Migration %d failed: %v", i+1, err)
			return err
		}
		log.Printf("Migration %d applied", i+1)
	}

	return nil
}

func (db *DB) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("rollback error: %v", rbErr)
		}
		return err
	}

	return tx.Commit()
}

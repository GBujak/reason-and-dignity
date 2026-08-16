package db

import (
	"fmt"
	"net/url"
	"time"

	"github.com/jmoiron/sqlx"

	_ "github.com/mattn/go-sqlite3"
)

type SqliteConnection struct {
	Writer *sqlx.DB
	Reader *sqlx.DB
}

func (c *SqliteConnection) Close() error {
	_, _ = c.Writer.Exec("PRAGMA optimize;")
	_, _ = c.Writer.Exec("PRAGMA incremental_vacuum;")

	if err := c.Reader.Close(); err != nil {
		return err
	}
	return c.Writer.Close()
}

func createClient(dbPath string) (*SqliteConnection, error) {
	writerDSN := buildDSN(dbPath, false)
	writerDB, err := sqlx.Connect("sqlite3", writerDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite writer: %w", err)
	}

	writerDB.SetMaxOpenConns(1)
	writerDB.SetMaxIdleConns(1)
	writerDB.SetConnMaxLifetime(12 * time.Hour)

	if err := writerDB.Ping(); err != nil {
		_ = writerDB.Close()
		return nil, fmt.Errorf("failed to initialize SQLite writer for %s: %w", dbPath, err)
	}

	err = migrate(dbPath, writerDB)
	if err != nil {
		return nil, fmt.Errorf("Could not migrate database %v: %w", dbPath, err)
	}

	readerDSN := buildDSN(dbPath, true)
	readerDB, err := sqlx.Connect("sqlite3", readerDSN)
	if err != nil {
		_ = writerDB.Close()
		return nil, fmt.Errorf("failed to open sqlite reader: %w", err)
	}

	readerDB.SetMaxOpenConns(4)
	readerDB.SetMaxIdleConns(4)
	readerDB.SetConnMaxLifetime(12 * time.Hour)

	return &SqliteConnection{
		Writer: writerDB,
		Reader: readerDB,
	}, nil
}

func buildDSN(path string, readOnly bool) string {
	q := url.Values{}
	q.Set("_journal_mode", "WAL")
	q.Set("_sync", "NORMAL")
	q.Set("_busy_timeout", "5000")
	q.Set("_foreign_keys", "ON")

	q.Add("_pragma", "temp_store(MEMORY)")

	// Lean private cache (4 MB per conn) - let mmap handle reads
	q.Add("_pragma", "cache_size(-4000)")

	// 128 MB virtual address space ceiling per database
	q.Add("_pragma", "mmap_size(134217728)")

	if readOnly {
		q.Set("mode", "ro")
		q.Add("_pragma", "query_only(TRUE)")
	}

	return fmt.Sprintf("file:%s?%s", path, q.Encode())
}

package db

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"
)

func noDDLAuthorizer(action int, _, _, _ string) int {
	switch action {
	case
		sqlite3.SQLITE_SELECT,
		sqlite3.SQLITE_READ,
		sqlite3.SQLITE_INSERT,
		sqlite3.SQLITE_UPDATE,
		sqlite3.SQLITE_DELETE,
		sqlite3.SQLITE_FUNCTION,
		sqlite3.SQLITE_TRANSACTION,
		sqlite3.SQLITE_SAVEPOINT:

		return sqlite3.SQLITE_OK

	default:
		return sqlite3.SQLITE_DENY
	}
}

func RestrictedSchemaChangesConnection(ctx context.Context, pool *sqlx.DB) (*sqlx.Conn, error) {
	conn, err := pool.Connx(ctx)
	if err != nil {
		return nil, err
	}

	err = conn.Raw(func(driverConn any) error {
		sqliteConn := driverConn.(*sqlite3.SQLiteConn)
		sqliteConn.RegisterAuthorizer(noDDLAuthorizer)
		return nil
	})
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to enable authorizer: %w", err)
	}

	return conn, nil
}

func ClearSchemaChangesRestrictionAndClose(conn *sqlx.Conn) {
	conn.Raw(func(driverConn any) error {
		sqliteConn := driverConn.(*sqlite3.SQLiteConn)
		sqliteConn.RegisterAuthorizer(nil)
		return nil
	})
	conn.Close()
}

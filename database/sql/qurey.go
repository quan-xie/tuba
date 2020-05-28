package sql

import (
	"context"
	"database/sql"
)

// Qurey is wrap mysql qurey
func (db *DB) Qurey(c context.Context, query string, args ...interface{}) (rows *sql.Rows, err error) {
	idx := db.readIndex()
	if len(db.read) > idx {
		if rows, err = db.read[(idx)%len(db.read)].QueryContext(c, query, args); err != nil {
			return
		}
	}
	return db.write.QueryContext(c, query, args...)
}

// QureyRow is wrap mysql qureyrow
func (db *DB) QureyRow(c context.Context, query string, args ...interface{}) (row *sql.Row) {
	idx := db.readIndex()
	if len(db.read) > idx {
		return db.read[(idx)%len(db.read)].QueryRowContext(c, query, args)
	}
	return db.write.QueryRowContext(c, query, args)
}

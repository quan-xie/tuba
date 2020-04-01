package mysql

import (
	"database/sql"
	"time"

	"github.com/pkg/errors"

	"github.com/quan-xie/tuba/util/xtime"
)

// conn database connection
type conn struct {
	*sql.DB
	conf *Config
}

// DB database.
type DB struct {
	write  *conn
	read   []*conn
	idx    int64
	master *DB
}

// Config mysql config.
type Config struct {
	Addr         string
	DSN          string
	ReadDSN      []string
	Active       int
	Idle         int
	IdleTimeout  xtime.Duration // connect max life time.
	QueryTimeout xtime.Duration // query sql timeout
	ExecTimeout  xtime.Duration // execute sql timeout
	TranTimeout  xtime.Duration // transaction sql timeout
}

// NewMySQL new db and retry connection when has error.
func NewMySQL(c *Config) (db *DB) {
	if c.QueryTimeout == 0 || c.ExecTimeout == 0 || c.TranTimeout == 0 {
		panic("mysql must be set query/execute/transction timeout")
	}
	db, err := Open(c)
	if err != nil {
		panic(err)
	}
	return
}

func Open(c *Config) (*DB, error) {
	db := new(DB)
	d, err := connect(c, c.DSN)
	if err != nil {
		return nil, err
	}
	w := &conn{DB: d, conf: c}
	rs := make([]*conn, 0, len(c.ReadDSN))
	for _, rd := range c.ReadDSN {
		d, err := connect(c, rd)
		if err != nil {
			return nil, err
		}
		r := &conn{DB: d, conf: c}
		rs = append(rs, r)
	}
	db.write = w
	db.read = rs
	db.master = &DB{write: db.write}
	return db, nil
}

func connect(c *Config, dataSourceName string) (*sql.DB, error) {
	d, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		err = errors.WithStack(err)
		return nil, err
	}
	d.SetMaxOpenConns(c.Active)
	d.SetMaxIdleConns(c.Idle)
	d.SetConnMaxLifetime(time.Duration(c.IdleTimeout))
	return d, nil
}

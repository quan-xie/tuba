package sql

import "database/sql"

// DB is mysql database .
type DB struct {
}

// conn is database connection .
type conn struct {
}

func Open(c *Config) (*DB, error) {
	db := new(DB)
	return
}

func connect(c *Config, dsn string) (*sql.DB, error) {

}

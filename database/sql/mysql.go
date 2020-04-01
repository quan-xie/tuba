package sql

import (
	"github.com/quan-xie/tuba/log"
	"github.com/quan-xie/tuba/util/xtime"
)

const ()

type Config struct {
	DSNConfig    *DSNConfig //
	DSNConfig    []*DSNConfigs
	DSN          string         // write data source name.
	ReadDSN      []string       // read data source name.
	Active       int            // pool
	Idle         int            // pool
	IdleTimeout  xtime.Duration // connect max life time.
	QueryTimeout xtime.Duration // query sql timeout
	ExecTimeout  xtime.Duration // execute sql timeout
	TranTimeout  xtime.Duration // transaction sql timeout
}

// NewMySQL new db instance .
func NewMySQL(c *Config) (db *DataBase) {
	if c.QueryTimeout == 0 {
		log.Warn("NewMySQL QueryTimeout is c.QueryTimeout=%d", c.QueryTimeout)
		c.QueryTimeout = 20000
	}
	if c.ExecTimeout == 0 {
		log.Warn("NewMySQL ExecTimeout is c.ExecTimeout=%d", c.ExecTimeout)
		c.ExecTimeout = 20000
	}
	if c.TranTimeout == 0 {
		log.Warn("NewMySQL TranTimeout is c.TranTimeout=%d", c.TranTimeout)
		c.TranTimeout = 2000
	}
	return db
}

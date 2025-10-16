package mariadb

import (
	"database/sql"

	"github.com/cloudoperators/heureka/internal/util"
	"github.com/jmoiron/sqlx"
)

type Stmt interface {
	Close() error
	Queryx(args ...interface{}) (*sqlx.Rows, error)
}

type NamedStmt interface {
	Close() error
	Exec(arg interface{}) (sql.Result, error)
}

type SqlRows interface {
	Close() error
	Err() error
	Next() bool
	Scan(dest ...any) error
}

type Db interface {
	Close() error
	Exec(query string, args ...interface{}) (sql.Result, error)
	Get(dest interface{}, query string, args ...interface{}) error
	GetDbInstance() *sql.DB
	Preparex(query string) (Stmt, error)
	PrepareNamed(query string) (NamedStmt, error)
	Select(dest interface{}, query string, args ...interface{}) error
	Query(query string, args ...interface{}) (SqlRows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

func NewDb(cfg util.Config) (Db, error) {
	db, err := Connect(cfg)
	if err != nil {
		return nil, err
	}
	if cfg.DBTrace == true {
		return &TraceDb{db: db}, nil
	}
	return &QuietDb{db: db}, nil
}

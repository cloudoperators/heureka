// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"database/sql"

	"github.com/cloudoperators/heureka/internal/util"
	"github.com/jmoiron/sqlx"
)

type Stmt interface {
	Close() error
	Queryx(args ...any) (*sqlx.Rows, error)
}

type NamedStmt interface {
	Close() error
	Exec(arg any) (sql.Result, error)
}

type SqlRows interface {
	Close() error
	Err() error
	Next() bool
	Scan(dest ...any) error
}

type Db interface {
	Close() error
	Exec(query string, args ...any) (sql.Result, error)
	Get(dest any, query string, args ...any) error
	GetDbInstance() *sql.DB
	Preparex(query string) (Stmt, error)
	PrepareNamed(query string) (NamedStmt, error)
	Select(dest any, query string, args ...any) error
	Query(query string, args ...any) (SqlRows, error)
	QueryRow(query string, args ...any) *sql.Row
}

func NewDb(cfg util.Config) (Db, error) {
	db, err := Connect(cfg)
	if err != nil {
		return nil, err
	}

	if cfg.DBTrace {
		return &TraceDb{db: db}, nil
	}

	return &QuietDb{db: db}, nil
}

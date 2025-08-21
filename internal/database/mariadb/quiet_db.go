// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
)

type QuietDb struct {
	db *sqlx.DB
}

func (qdb *QuietDb) Close() error {
	return qdb.db.Close()
}

func (qdb *QuietDb) Exec(query string, args ...interface{}) (sql.Result, error) {
	return qdb.db.Exec(query, args...)
}

func (qdb *QuietDb) Get(dest interface{}, query string, args ...interface{}) error {
	return qdb.db.Get(dest, query, args...)
}

func (qdb *QuietDb) GetDbInstance() *sql.DB {
	return qdb.db.DB
}

func (qdb *QuietDb) PrepareNamed(query string) (*sqlx.NamedStmt, error) {
	return qdb.db.PrepareNamed(query)
}

func (qdb *QuietDb) Preparex(query string) (*sqlx.Stmt, error) {
	return qdb.db.Preparex(query)
}

func (qdb *QuietDb) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return qdb.db.Query(query, args...)
}

func (qdb *QuietDb) QueryRow(query string, args ...interface{}) *sql.Row {
	return qdb.db.QueryRow(query, args...)
}

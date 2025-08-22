// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"time"
)

type TraceDb struct {
	db *sqlx.DB
}

func enterTrace() time.Time {
	return time.Now()
}

func exitTrace(fnName string, query string, t0 time.Time) {
	logrus.WithFields(logrus.Fields{
		"executionTime": time.Since(t0).String(),
		"query": query,
	}).Printf("%s finished", fnName)
}

func (tdb *TraceDb) Close() error {
	return tdb.db.Close()
}

func (tdb *TraceDb) Exec(query string, args ...interface{}) (sql.Result, error) {
	t0 := enterTrace()
	defer exitTrace("Exec", query, t0)
	return tdb.db.Exec(query, args...)
}

func (tdb *TraceDb) Get(dest interface{}, query string, args ...interface{}) error {
	t0 := enterTrace()
	defer exitTrace("Get", query, t0)
	return tdb.db.Get(dest, query, args...)
}

func (tdb *TraceDb) GetDbInstance() *sql.DB {
	return tdb.db.DB
}

func (tdb *TraceDb) PrepareNamed(query string) (*sqlx.NamedStmt, error) {
	t0 := enterTrace()
	defer exitTrace("PrepareNamed", query, t0)
	return tdb.db.PrepareNamed(query)
}

func (tdb *TraceDb) Preparex(query string) (*sqlx.Stmt, error) {
	t0 := enterTrace()
	defer exitTrace("Preparex", query, t0)
	return tdb.db.Preparex(query)
}

func (tdb *TraceDb) Query(query string, args ...interface{}) (*sql.Rows, error) {
	t0 := enterTrace()
	defer exitTrace("Query", query, t0)
	return tdb.db.Query(query, args...)
}

func (tdb *TraceDb) QueryRow(query string, args ...interface{}) *sql.Row {
	t0 := enterTrace()
	defer exitTrace("QueryRow", query, t0)
	return tdb.db.QueryRow(query, args...)
}

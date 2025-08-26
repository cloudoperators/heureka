// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"time"
)

// Trace
type Trace struct {
	fnName string
	query  string
	t0     time.Time
}

func NewTrace(fnName string, query string) *Trace {
	return &Trace{fnName: fnName, query: query, t0: time.Now()}
}

func (t Trace) exitTrace() {
	logrus.WithFields(logrus.Fields{
		"executionTime": time.Since(t.t0).String(),
		"query":         t.query,
	}).Printf("%s finished", t.fnName)
}

func (t *Trace) errorTrace() {
	t.fnName += " (ERROR)"
	t.exitTrace()
}

// TraceSqlRows
type TraceSqlRows struct {
	trace *Trace
	rows  *sql.Rows
}

func (tsr *TraceSqlRows) Close() error {
	defer tsr.trace.exitTrace()
	return tsr.rows.Close()
}

func (tsr *TraceSqlRows) Err() error {
	return tsr.rows.Err()
}

func (tsr *TraceSqlRows) Next() bool {
	return tsr.rows.Next()
}

func (tsr *TraceSqlRows) Scan(dest ...any) error {
	return tsr.rows.Scan(dest...)
}

// TraceStmt
type TraceStmt struct {
	trace *Trace
	stmt  *sqlx.Stmt
}

func (ts *TraceStmt) Close() error {
	defer ts.trace.exitTrace()
	return ts.stmt.Close()
}

func (ts *TraceStmt) Queryx(args ...interface{}) (*sqlx.Rows, error) {
	return ts.stmt.Queryx(args...)
}

// TraceNamedStmt
type TraceNamedStmt struct {
	trace     *Trace
	namedStmt *sqlx.NamedStmt
}

func (tns *TraceNamedStmt) Close() error {
	defer tns.trace.exitTrace()
	return tns.namedStmt.Close()
}

func (tns *TraceNamedStmt) Exec(arg interface{}) (sql.Result, error) {
	return tns.namedStmt.Exec(arg)
}

// TraceDb
type TraceDb struct {
	db *sqlx.DB
}

func (tdb *TraceDb) Close() error {
	return tdb.db.Close()
}

func (tdb *TraceDb) Exec(query string, args ...interface{}) (sql.Result, error) {
	defer NewTrace("Exec", query).exitTrace()
	return tdb.db.Exec(query, args...)
}

func (tdb *TraceDb) Get(dest interface{}, query string, args ...interface{}) error {
	defer NewTrace("Get", query).exitTrace()
	return tdb.db.Get(dest, query, args...)
}

func (tdb *TraceDb) GetDbInstance() *sql.DB {
	return tdb.db.DB
}

func (tdb *TraceDb) PrepareNamed(query string) (NamedStmt, error) {
	trace := NewTrace("PrepareNamed", query)
	namedStmt, err := tdb.db.PrepareNamed(query)
	if err != nil {
		trace.errorTrace()
		return namedStmt, err
	}
	return &TraceNamedStmt{namedStmt: namedStmt, trace: trace}, nil
}

func (tdb *TraceDb) Preparex(query string) (Stmt, error) {
	trace := NewTrace("Preparex", query)
	stmt, err := tdb.db.Preparex(query)
	if err != nil {
		trace.errorTrace()
		return stmt, err
	}
	return &TraceStmt{stmt: stmt, trace: trace}, nil
}

func (tdb *TraceDb) Query(query string, args ...interface{}) (SqlRows, error) {
	trace := NewTrace("Query", query)
	sqlRows, err := tdb.db.Query(query, args...)
	if err != nil {
		trace.errorTrace()
		return sqlRows, err
	}
	return &TraceSqlRows{rows: sqlRows, trace: trace}, nil
}

func (tdb *TraceDb) QueryRow(query string, args ...interface{}) *sql.Row {
	defer NewTrace("QueryRow", query).exitTrace()
	return tdb.db.QueryRow(query, args...)
}

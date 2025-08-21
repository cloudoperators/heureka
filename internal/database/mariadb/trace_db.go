// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"log"
	"time"
)

type TraceDb struct {
	db *sqlx.DB
}

type TraceObj struct {
	t0  time.Time
	uid uuid.UUID
}

func enterTrace(fnName string, query string) TraceObj {
	uid := uuid.New()
	msg := fmt.Sprintf("[%s] Starting %s with query: \n--- QUERY: ---\n%s\n--------------\n", uid, fnName, query)
	log.Print(msg)
	return TraceObj{t0: time.Now(), uid: uid}
}

func exitTrace(fnName string, to TraceObj) {
	msg := fmt.Sprintf("[%s] %s finished, Execution time: %s", to.uid, fnName, time.Since(to.t0))
	log.Print(msg)
}

func (tdb *TraceDb) Close() error {
	return tdb.db.Close()
}

func (tdb *TraceDb) Exec(query string, args ...interface{}) (sql.Result, error) {
	t0 := enterTrace("Exec", query)
	defer exitTrace("Exec", t0)
	return tdb.db.Exec(query, args...)
}

func (tdb *TraceDb) Get(dest interface{}, query string, args ...interface{}) error {
	t0 := enterTrace("Get", query)
	defer exitTrace("Get", t0)
	return tdb.db.Get(dest, query, args...)
}

func (tdb *TraceDb) GetDbInstance() *sql.DB {
	return tdb.db.DB
}

func (tdb *TraceDb) PrepareNamed(query string) (*sqlx.NamedStmt, error) {
	t0 := enterTrace("PrepareNamed", query)
	defer exitTrace("PrepareNamed", t0)
	return tdb.db.PrepareNamed(query)
}

func (tdb *TraceDb) Preparex(query string) (*sqlx.Stmt, error) {
	t0 := enterTrace("Preparex", query)
	defer exitTrace("Preparex", t0)
	return tdb.db.Preparex(query)
}

func (tdb *TraceDb) Query(query string, args ...interface{}) (*sql.Rows, error) {
	t0 := enterTrace("Query", query)
	defer exitTrace("Query", t0)
	return tdb.db.Query(query, args...)
}

func (tdb *TraceDb) QueryRow(query string, args ...interface{}) *sql.Row {
	t0 := enterTrace("QueryRow", query)
	defer exitTrace("QueryRow", t0)
	return tdb.db.QueryRow(query, args...)
}

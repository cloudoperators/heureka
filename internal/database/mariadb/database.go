// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"database/sql"
	"fmt"
	"io"
	"time"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/util"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

const (
	FILTER_FORMAT_STR = " %s %s"
	OP_AND            = "AND"
	OP_OR             = "OR"

	ERROR_MSG_PREPARED_STMT = "Error while creating prepared statement."
)

type SqlDatabase struct {
	db                    *sqlx.DB
	defaultIssuePriority  int64
	defaultRepositoryName string
	dbName                string
}

func (s *SqlDatabase) CloseConnection() error {
	return s.db.Close()
}

func buildDSN(cfg util.Config) string {
	if cfg.DBAddress == "/var/run/mysqld/mysqld.sock" {
		return fmt.Sprintf("%s:%s@unix(%s)/%s?multiStatements=true&parseTime=true", cfg.DBUser, cfg.DBPassword, cfg.DBAddress, cfg.DBName)
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?multiStatements=true&parseTime=true", cfg.DBUser, cfg.DBPassword, cfg.DBAddress, cfg.DBPort, cfg.DBName)
}

func TestConnection(cfg util.Config, backOff int) error {
	if cfg.DBAddress == "/var/run/mysqld/mysqld.sock" {
		// No need to test local socket connection
		return nil
	}
	if backOff <= 0 {
		return fmt.Errorf("Unable to connect to Database, exceeded backoffs...")
	}

	db, err := getSqlxConnection(cfg)
	if err != nil {
		fmt.Printf("Error connecting to DB: %s\n", err)
		return TestConnection(cfg, backOff-1)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		//before next try wait 100 milliseconds
		time.Sleep(100 * time.Millisecond)
		return TestConnection(cfg, backOff-1)
	}
	return nil
}

func getSqlxConnection(cfg util.Config) (*sqlx.DB, error) {
	connectionString := buildDSN(cfg)
	return sqlx.Connect("mysql", connectionString)
}

func buildDSN(cfg util.Config) string {
	if cfg.DBAddress == "/var/run/mysqld/mysqld.sock" {
		return fmt.Sprintf("%s:%s@unix(%s)/%s?multiStatements=true&parseTime=true", cfg.DBUser, cfg.DBPassword, cfg.DBAddress, cfg.DBName)
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?multiStatements=true&parseTime=true", cfg.DBUser, cfg.DBPassword, cfg.DBAddress, cfg.DBPort, cfg.DBName)
}

func Connect(cfg util.Config) (*sqlx.DB, error) {
	db, err := getSqlxConnection(cfg)
	if err != nil {
		logrus.WithError(err).Error(err)
		return nil, err
	}

	db.SetConnMaxLifetime(time.Second * 5)
	db.SetMaxIdleConns(cfg.DBMaxIdleConnections)
	db.SetMaxOpenConns(cfg.DBMaxOpenConnections)
	return db, nil
}

func NewSqlDatabase(cfg util.Config) (*SqlDatabase, error) {
	db, err := Connect(cfg)
	if err != nil {
		return nil, err
	}
	db.Exec(fmt.Sprintf("USE %s", cfg.DBName))
	return &SqlDatabase{
		db:                    db,
		defaultIssuePriority:  cfg.DefaultIssuePriority,
		defaultRepositoryName: cfg.DefaultRepositoryName,
		dbName:                cfg.DBName,
	}, nil
}

func (s *SqlDatabase) DropSchema(name string) error {
	_, err := s.db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s", name))
	return err
}

func (s *SqlDatabase) ConnectDB(dbName string) error {
	s.dbName = dbName
	return s.connectDB()
}

func (s *SqlDatabase) connectDB() error {
	_, err := s.db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", s.dbName))
	if err != nil {
		return fmt.Errorf("Could not create database '%s'. %w", s.dbName, err)
	}

	_, err = s.db.Exec(fmt.Sprintf("USE %s", s.dbName))
	if err != nil {
		return fmt.Errorf("Could not use database '%s'. %w", s.dbName, err)
	}
	return nil
}

func (s *SqlDatabase) GrantAccess(username string, database string, host string) error {
	_, err := s.db.Exec(fmt.Sprintf("GRANT ALL ON %s TO '%s'@'%s';", database, username, host))
	if err != nil {
		return err
	}
	_, err = s.db.Exec(fmt.Sprintf("GRANT ALL ON %s.* TO '%s'@'%s';", database, username, host))
	return err
}

// GetDefaultIssuePriority ...
func (s *SqlDatabase) GetDefaultIssuePriority() int64 {
	return s.defaultIssuePriority
}

func (s *SqlDatabase) GetDefaultRepositoryName() string {
	return s.defaultRepositoryName
}

func (s *SqlDatabase) GetVersion() (string, error) {
	m, err := s.openMigration()
	if err != nil {
		return "", fmt.Errorf("Could not open migration without source: %w", err)
	}
	defer m.Close()

	v, d, err := getMigrationVersion(m)
	if err != nil {
		return "", fmt.Errorf("Could not get migration version: %w", err)
	}

	return versionToString(v, d), nil
}

func GetVersion(cfg util.Config) (string, error) {
	db, err := NewSqlDatabase(cfg)
	if err != nil {
		return "", fmt.Errorf("Error while Creating Db")
	}

	v, err := db.GetVersion()
	if err != nil {
		return "", fmt.Errorf("Error while checking Db migration: %w", err)
	}
	return v, nil
}

func (s *SqlDatabase) RunMigrations() error {
	m, err := s.openMigration()
	if err != nil {
		return err
	}
	defer m.Close()

	err = m.Up()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange && err != io.EOF {
		return err
	}

	return nil
}

func versionToString(v uint, dirty bool) string {
	var dirtyStr string
	if dirty {
		dirtyStr = " (DIRTY)"
	}
	return fmt.Sprintf("%d%s", v, dirtyStr)
}

func getMigrationVersion(m *migrate.Migrate) (uint, bool, error) {
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return 0, false, err
	}
	return version, dirty, nil
}

func (s *SqlDatabase) openMigration() (*migrate.Migrate, error) {
	err := s.connectDB()
	if err != nil {
		return nil, fmt.Errorf("Could not connect DB: %w", err)
	}

	d, err := iofs.New(Migration, "migrations")
	if err != nil {
		return nil, err
	}

	driver, err := mysql.WithInstance(s.db.DB, &mysql.Config{DatabaseName: s.dbName})
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithInstance(
		"iofs", d,
		"mysql", driver)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func combineFilterQueries(filterQueries []string, op string) string {
	filterStr := ""
	i := 0
	for _, f := range filterQueries {
		if f == "" {
			continue
		}
		if i > 0 {
			filterStr = fmt.Sprintf(FILTER_FORMAT_STR, filterStr, op)
		}
		filterStr = fmt.Sprintf(FILTER_FORMAT_STR, filterStr, f)
		i += 1
	}

	//encapsulate in brackets
	if filterStr != "" {
		filterStr = fmt.Sprintf("( %s )", filterStr)
	}

	return filterStr
}

func buildFilterQuery[T any](filter []T, expr string, op string) string {
	filterStr := ""
	for i := range filter {
		if i > 0 {
			filterStr = fmt.Sprintf(FILTER_FORMAT_STR, filterStr, op)
		}
		filterStr = fmt.Sprintf(FILTER_FORMAT_STR, filterStr, expr)
	}

	//encapsulate in brackets
	if filterStr != "" {
		filterStr = fmt.Sprintf("( %s )", filterStr)
	}

	return filterStr
}

func buildQueryParameters[T any](params []interface{}, filter []T) []interface{} {
	return buildQueryParametersCount(params, filter, 1)
}

func buildQueryParametersCount[T any](params []interface{}, filter []T, count int) []interface{} {
	for _, item := range filter {
		for i := 0; i < count; i++ {
			params = append(params, item)
		}
	}
	return params
}

func performInsert[T any](s *SqlDatabase, query string, item T, l *logrus.Entry) (int64, error) {
	res, err := performExec(s, query, item, l)

	if err != nil {
		return -1, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		msg := "Error while getting last insert id"
		l.WithFields(
			logrus.Fields{
				"error": err,
			}).Error(msg)
		return -1, fmt.Errorf("%s", msg)
	}

	l.WithFields(
		logrus.Fields{
			"id": id,
		}).Debug("Successfully performed insert")

	return id, nil
}

func performExec[T any](s *SqlDatabase, query string, item T, l *logrus.Entry) (sql.Result, error) {
	stmt, err := s.db.PrepareNamed(query)

	if err != nil {
		msg := ERROR_MSG_PREPARED_STMT
		l.WithFields(
			logrus.Fields{
				"error": err,
				"query": query,
			}).Error(msg)
		return nil, fmt.Errorf("%s", msg)
	}

	defer stmt.Close()
	res, err := stmt.Exec(item)
	if err != nil {
		msg := err.Error()
		l.WithFields(
			logrus.Fields{
				"error": err,
			}).Error(msg)
		return nil, fmt.Errorf("%s", msg)
	}
	return res, nil
}

func performListScan[T DatabaseRow, E entity.HeurekaEntity | DatabaseRow](stmt *sqlx.Stmt, filterParameters []interface{}, l *logrus.Entry, listBuilder func([]E, T) []E) ([]E, error) {
	rows, err := stmt.Queryx(filterParameters...)
	if err != nil {
		msg := "Error while performing Query from prepared Statement"
		l.WithFields(
			logrus.Fields{
				"error":      err.Error(),
				"parameters": filterParameters,
			}).Error(msg)
		return nil, fmt.Errorf("%s", msg)
	}

	defer rows.Close()
	var listEntries []E
	for rows.Next() {
		var row T
		err := rows.StructScan(&row)
		if err != nil {
			msg := "Error while scanning struct"
			cols, _ := rows.Columns()
			l.WithFields(
				logrus.Fields{
					"error":           err.Error(),
					"returnedColumns": cols,
				}).Error(msg)
			return nil, fmt.Errorf("%s", msg)
		}

		listEntries = listBuilder(listEntries, row)
	}

	rows.Close()
	l.WithFields(
		logrus.Fields{
			"count": len(listEntries),
		}).Debug("Successfully performed list scan")

	return listEntries, nil
}

func performIdScan(stmt *sqlx.Stmt, filterParameters []interface{}, l *logrus.Entry) ([]int64, error) {
	var rows *sqlx.Rows

	rows, err := stmt.Queryx(filterParameters...)
	if err != nil {
		msg := "Error while performing query with prepared Statement"
		l.WithFields(
			logrus.Fields{
				"error":             err,
				"preparedStatement": stmt,
				"parameters":        filterParameters,
			}).Error(msg)

		return make([]int64, 0), fmt.Errorf("%s", msg)
	}
	defer rows.Close()

	var listEntries []int64
	for rows.Next() {
		var row int64
		err = rows.Scan(&row)
		if err != nil {
			msg := "Error while scanning into in64"
			cols, _ := rows.Columns()
			l.WithFields(
				logrus.Fields{
					"error":           err,
					"returnedColumns": cols,
				}).Error(msg)
			return make([]int64, 0), fmt.Errorf("%s", msg)
		}

		listEntries = append(listEntries, row)
	}
	l.WithFields(
		logrus.Fields{
			"id_count": len(listEntries),
		}).Debug("Successfully performed id scan")

	return listEntries, nil
}

func performCountScan(stmt *sqlx.Stmt, filterParameters []interface{}, l *logrus.Entry) (int64, error) {
	var rows *sqlx.Rows

	rows, err := stmt.Queryx(filterParameters...)
	if err != nil {
		msg := "Error while performing query with prepared Statement"
		l.WithFields(
			logrus.Fields{
				"error":             err,
				"preparedStatement": stmt,
				"parameters":        filterParameters,
			}).Error(msg)

		return -1, fmt.Errorf("%s", msg)
	}
	defer rows.Close()

	rows.Next()
	var row int64
	err = rows.Scan(&row)

	if err != nil {
		msg := "Error while scanning into in64"
		cols, _ := rows.Columns()
		l.WithFields(
			logrus.Fields{
				"error":           err,
				"returnedColumns": cols,
			}).Error(msg)
		return -1, fmt.Errorf("%s", msg)
	}

	l.WithFields(
		logrus.Fields{
			"count": row,
		}).Debug("Successfully performed count scan")

	return row, nil
}

func getCursor(p entity.Paginated, filterStr string, stmt string) entity.Cursor {
	prependAnd := ""
	if filterStr != "" {
		prependAnd = OP_AND
	}
	stmt = fmt.Sprintf("%s %s", prependAnd, stmt)

	var cursorValue int64 = 0
	if p.After != nil {
		cursorValue = *p.After
	}

	limit := 1000
	if p.First != nil {
		limit = *p.First
	}

	return entity.Cursor{
		Statement: stmt,
		Value:     cursorValue,
		Limit:     limit,
	}
}

func GetDefaultOrder(order []entity.Order, by entity.OrderByField, direction entity.OrderDirection) []entity.Order {
	if len(order) == 0 {
		order = append([]entity.Order{{By: by, Direction: direction}}, order...)
	}

	return order
}

func buildStateFilterQuery(state []entity.StateFilterType, prefix string) string {
	stateQueries := []string{}
	if len(state) < 1 {
		stateQueries = append(stateQueries, fmt.Sprintf("%s_deleted_at IS NULL", prefix))
	} else {
		for i := range state {
			if state[i] == entity.Active {
				stateQueries = append(stateQueries, fmt.Sprintf("%s_deleted_at IS NULL", prefix))
			} else if state[i] == entity.Deleted {
				stateQueries = append(stateQueries, fmt.Sprintf("%s_deleted_at IS NOT NULL", prefix))
			}
		}
	}
	return combineFilterQueries(stateQueries, OP_OR)
}

func buildJsonFilterQuery(filter []*entity.Json, column string, op string) string {
	var conFilQueries []string
	for _, conFil := range filter {
		attrs := util.SeparateJsonAttributes(*conFil)
		var queries []string
		for _, conAttr := range attrs {
			queries = append(queries, fmt.Sprintf("JSON_VALUE(%s, '$.%s') = ?", column, conAttr.Key))
		}
		conFilQueries = append(conFilQueries, combineFilterQueries(queries, OP_AND))
	}
	return combineFilterQueries(conFilQueries, op)
}

func buildJsonQueryParameters(params []interface{}, filter []*entity.Json) []interface{} {
	var conQueryParams []interface{}
	for _, conFil := range filter {
		attrs := util.SeparateJsonAttributes(*conFil)
		for _, conAttr := range attrs {
			conQueryParams = append(conQueryParams, conAttr.Attr)
		}
	}
	return buildQueryParameters(params, conQueryParams)
}

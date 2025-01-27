// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/util"
	util2 "github.com/cloudoperators/heureka/pkg/util"
	_ "github.com/go-sql-driver/mysql"
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
	config                *util.Config
}

func (s *SqlDatabase) CloseConnection() error {
	return s.db.Close()
}
func TestConnection(cfg util.Config, backOff int) error {

	if backOff <= 0 {
		return fmt.Errorf("Unable to connect to Database, exceeded backoffs...")
	}

	//before each try wait 1 Second
	time.Sleep(1 * time.Second)
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?multiStatements=true&parseTime=true", cfg.DBUser, cfg.DBPassword, cfg.DBAddress, cfg.DBPort, cfg.DBName)
	db, err := sqlx.Connect("mysql", connectionString)
	if err != nil {
		return TestConnection(cfg, backOff-1)
	}
	defer db.Close()
	//do an actual ping to check if not only the handshake works but the db schema is as well ready to operate on
	err = db.Ping()
	if err != nil {
		return TestConnection(cfg, backOff-1)
	}
	return nil
}

func Connect(cfg util.Config) (*sqlx.DB, error) {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?multiStatements=true&parseTime=true", cfg.DBUser, cfg.DBPassword, cfg.DBAddress, cfg.DBPort, cfg.DBName)

	db, err := sqlx.Connect("mysql", connectionString)
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
	}, nil
}

func (s *SqlDatabase) DropSchemaByName(name string) error {
	_, err := s.db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s", name))
	return err
}

func (s *SqlDatabase) DropSchema() error {
	return s.DropSchemaByName("heureka")
}

func (s *SqlDatabase) GrantAccess(username string, database string, host string) error {
	_, err := s.db.Exec(fmt.Sprintf("GRANT ALL ON %s TO '%s'@'%s';", database, username, host))
	if err != nil {
		return err
	}
	_, err = s.db.Exec(fmt.Sprintf("GRANT ALL ON %s.* TO '%s'@'%s';", database, username, host))
	return err
}

func (s *SqlDatabase) SetupSchema(cfg util.Config) error {
	var sf string
	if strings.HasPrefix(cfg.DBSchema, "/") {
		sf = cfg.DBSchema
	} else {
		pr, err := util2.GetProjectRoot()
		if err != nil {
			logrus.WithError(err).Fatalln(err)
			return err
		}
		sf = fmt.Sprintf("%s/%s", pr, cfg.DBSchema)
	}
	file, err := os.ReadFile(sf)
	if err != nil {
		logrus.WithError(err).Fatalln(err)
		return err
	}

	schema := string(file)

	schema = strings.Replace(schema, "heureka", cfg.DBName, 2)
	_, err = s.db.Exec(schema)

	if err != nil {
		logrus.WithError(err).Fatalln(err)
		return err
	}
	return nil
}

// GetDefaultIssuePriority ...
func (s *SqlDatabase) GetDefaultIssuePriority() int64 {
	return s.defaultIssuePriority
}

func (s *SqlDatabase) GetDefaultRepositoryName() string {
	return s.defaultRepositoryName
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

func GetDefaultOrder(order []entity.Order, by entity.DbColumnName, direction entity.OrderDirection) []entity.Order {
	if len(order) == 0 {
		order = append([]entity.Order{{By: by, Direction: direction}}, order...)
	}

	return order
}

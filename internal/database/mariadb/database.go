// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"database/sql"
	"fmt"
	"os"
	"reflect"
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

	ERROR_MSG_PREPARED_STMT           = "Error while creating prepared statement."
	ERROR_MSG_INVALID_OR_UNSET_CURSOR = "Cursor field %s is invalid or unset"
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

// buildCursorStmtRecursive builds the where statement for the cursor based pagination recursively
// it builds the required where statement for the given order slice. This is done by iterating over the order slice and
// building the where statement for the first order element and then recursively calling the function with the remaining
// order elements (removing the last order element) and concatenating  the where statement with the previous one
// the cursorFieldGetter function is used to get the database field for the given cursor field.
//
// An example for the order slice ["target_remediation_date", "name", "id"] should result in the following where statement:
// (
//
//	 (target_remediation_date > ?) OR
//	 (target_remediation_date = ? AND name > ?) OR
//	 (target_remediation_date = ? AND name = ? AND id > ?)
//	)
//
// This translates to:
// - Elements need to be bigger than the first order element OR
// - If they are equal to the first order element they need to be bigger than the second order element OR
// - If they are equal to the first and second order element they need to be bigger than the third order element
func buildCursorStmtRecursive(where string, order []entity.Order, cursorFieldGetter func(string) (string, error)) (string, error) {

	if len(order) == 0 {
		//ensure paratheneses around second overall cursor ordering where statment
		return fmt.Sprintf("(%s)", where), nil
	}

	subWhere := ""
	for i, o := range order {
		cursorField, err := cursorFieldGetter(string(o.By))
		if err != nil {
			return "", err
		}
		if i >= len(order)-1 {
			subWhere = fmt.Sprintf("%s %s > ?", subWhere, cursorField)
		} else {
			subWhere = fmt.Sprintf("%s %s = ? AND", subWhere, cursorField)
		}
	}

	// ensure paratheses around subWhere
	subWhere = fmt.Sprintf("( %s )", subWhere)
	if where != "" {
		subWhere = fmt.Sprintf("%s OR %s", subWhere, where)
	}

	return buildCursorStmtRecursive(subWhere, order[:len(order)-1], cursorFieldGetter)
}

func buildCursorParameterForOrderRecursive(parsed interface{}, o entity.Order) ([]interface{}, error) {
	params := []interface{}{}
	reflectedType := reflect.TypeOf(parsed)
	reflectedValue := reflect.ValueOf(parsed)

	for i := 0; i < reflectedType.NumField(); i++ {
		structField := reflectedType.Field(i)
		structVal := reflectedValue.Field(i)

		if !structField.IsExported() {
			continue
		}

		if structVal.Kind() == reflect.Ptr {
			if structVal.IsNil() {
				continue
			}
			structVal = structVal.Elem()
		}

		cursorTag := structField.Tag.Get("cursor")
		structFieldVal := structVal.Interface()
		if cursorTag == string(o.By) {
			// check the type of the row field and set the cursor parameter accordingly
			switch v := structFieldVal.(type) {
			case sql.NullString:
				if !v.Valid {
					return nil, fmt.Errorf(ERROR_MSG_INVALID_OR_UNSET_CURSOR, o.By)
				}
				params = append(params, v.String)
			case sql.NullInt64:
				if !v.Valid {
					return nil, fmt.Errorf(ERROR_MSG_INVALID_OR_UNSET_CURSOR, o.By)
				}
				params = append(params, v.Int64)
			case sql.NullTime:
				if !v.Valid {
					return nil, fmt.Errorf(ERROR_MSG_INVALID_OR_UNSET_CURSOR, o.By)
				}
				params = append(params, v.Time)
			case sql.NullBool:
				if !v.Valid {
					return nil, fmt.Errorf(ERROR_MSG_INVALID_OR_UNSET_CURSOR, o.By)
				}
				params = append(params, v.Bool)
			//other types currently unused and not properly handled
			default:
				return nil, fmt.Errorf("Cursor field %s has invalid type %T", o.By, structFieldVal)
			}
			break

		} else if structVal.Kind() == reflect.Struct {

			// recursively call the function to handle nested structs
			nestedParams, err := buildCursorParameterForOrderRecursive(structFieldVal, o)
			if err != nil {
				return nil, err
			}
			if len(nestedParams) > 0 {
				params = append(nestedParams, params...)
			}
		}
	}
	return params, nil
}

// buildCursorParameters constructs a map of cursor parameters for pagination based on the provided struct and order slice.
//
// It iterates over the fields of the struct, checking for a "cursor" tag that matches the order slice elements.
// For each matching field, it extracts the value and adds it to the params map with a key formatted as "cursor_<field>".
//
// The function handles various nullable SQL types (e.g., sql.NullString, sql.NullInt64, sql.NullTime, etc.)
// by checking their validity and extracting the underlying value if valid. If the field type is not a recognized
// nullable type, it returns an error.
//
// Parameters:
//   - parsed: A struct of type T that implements the DatabaseRow interface. This struct contains the data from which
//     cursor parameters are extracted.
//   - order: A slice of entity.Order that specifies the order of fields for cursor-based pagination.
//
// Returns:
//   - A map[string]interface{} containing the cursor parameters with keys formatted as "cursor_<field>" and their
//     corresponding values extracted from the parsed struct.
//   - An error if a field type is not recognized or if any other issue occurs during the extraction process.
func buildCursorParametersRecursive(parsed interface{}, order []entity.Order) ([]any, error) {
	if len(order) == 0 {
		return []interface{}{}, nil
	}
	params := []interface{}{}

	for _, o := range order {
		newParams, err := buildCursorParameterForOrderRecursive(parsed, o)
		if err != nil {
			return nil, err
		}
		params = append(params, newParams...)
	}

	recursiveParams, err := buildCursorParametersRecursive(parsed, order[:len(order)-1])
	if err != nil {
		return nil, err
	}
	params = append(recursiveParams, params...)

	return params, nil
}

// getCursorField retrieves the database field name corresponding to the given cursor field.
//
// This should be the default CursorField Getter function used in the buildCursor function for most cases.
// Only implement a specialied version when you have match ordering cases or other special cases that need different
// field mapping.
//
// Parameters:
// - field: The cursor field name to look for.
// - row: A struct of type DatabaseRow that contains the data from which the database field name is extracted.
// - tableAlias: A string to prepend to the database field name. This should be the Table alias appended with a dot. e.g. "IM."
//
// Returns:
// - A string containing the database field name with the specified prefix.
// - An error if the cursor field is not found in the row struct.
func getCursorField[R DatabaseRow](field string, row R, tableAlias string) (string, error) {
	imrType := reflect.TypeOf(row)
	for i := 0; i < imrType.NumField(); i++ {
		structField := imrType.Field(i)
		cursorTag := structField.Tag.Get("cursor")
		if cursorTag == field {
			dbTag := structField.Tag.Get("db")
			if dbTag != "" {
				return fmt.Sprintf("%s%s", tableAlias, dbTag), nil
			}
		}
	}
	return "", fmt.Errorf("field %s not found", field)
}

// buildCursor constructs a DatabaseCursor for pagination based on the provided filter, order, and row.
//
// It orchestrates the individual steps which are implemented in the respective sub functions for constructing a
// DatabaseCursor by unmarshalling the cursor, building the statement, extracting the cursor parameters, and setting the limit.
//
// Parameters:
// - c: is the cursor string that is unmarshalled to extract the cursor parameters.
// - limit: A pointer to an integer that specifies the maximum number of results to return.
// - order: A slice of entity.Order[O] that specifies the order of fields for cursor-based pagination.
// - row: A struct of type T that implements the DatabaseRow interface. This struct contains the data from which cursor parameters are extracted.
// - cfg: This is the cursor field getter function, if you don't have any special cases such as case matching use getCursorField
//
// Returns:
// - A pointer to a DatabaseCursor containing the constructed statement, parameter values, and limit.
// - An error if any issue occurs during the cursor construction process.
func buildCursor[T DatabaseRow](c *string, limit *int, order []entity.Order, row T, cfg func(field string) (string, error)) (*DatabaseCursor, error) {

	if c == nil {
		return nil, fmt.Errorf("Cursor is empty")
	}

	err := UnmarshalCursor(*c, &row)

	if err != nil {
		return nil, err
	}

	stmt, err := buildCursorStmtRecursive("", order, cfg)

	if err != nil {
		return nil, err
	}

	parameters, err := buildCursorParametersRecursive(row, order)
	if err != nil {
		return nil, err
	}

	// set limit to 1000 if not set
	l := 1000
	if limit != nil {
		l = *limit
	}

	databaseCursor := DatabaseCursor{
		Statement:       stmt,
		ParameterValues: parameters,
		Limit:           l,
	}

	return &databaseCursor, nil
}

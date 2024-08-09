// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

// GetScans ...
func (s *SqlDatabase) GetScans(filter *entity.ScanFilter) ([]entity.Scan, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetComponents",
	})

	baseQuery := `
		SELECT S.* FROM Scans S
		%s
		%s
		%s GROUP BY C.scan_id ORDER BY C.scan_id LIMIT ?
	`
	filter = s.ensureScanFilter(filter)

	stmt, filterParameters, err := s.buildScanStatement(baseQuery, filter, true, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.Scan, e ScanRow) []entity.Scan {
			return append(l, e.AsScan())
		},
	)

}

// ensureScanFilter ...
func (s *SqlDatabase) ensureScanFilter(f *entity.ScanFilter) *entity.ScanFilter {
	var first = 1000
	var after int64 = 0
	if f == nil {
		return &entity.ScanFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
			Id:    nil,
			Scope: nil,
		}
	}
	if f.After == nil {
		f.After = &after
	}
	if f.First == nil {
		f.First = &first
	}
	return f
}

// getScanFilterString ...
func (s *SqlDatabase) getScanFilterString(filter *entity.ScanFilter) string {
	var fl []string
	fl = append(fl, buildFilterQuery(filter.Id, "S.scan_id = ?", OP_OR))
	fl = append(fl, "S.scan_finished_at IS NULL")

	return combineFilterQueries(fl, OP_AND)
}

func (s *SqlDatabase) getScanJoins(filter *entity.ScanFilter) string {
	joins := ""
	if len(filter.Id) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, "LEFT JOIN Users U on S.scan_id = U.user_id")
	}
	return joins
}

// buildScanStatement ...
func (s *SqlDatabase) buildScanStatement(baseQuery string, filter *entity.ScanFilter, withCursor bool, l *logrus.Entry) (*sqlx.Stmt, []interface{}, error) {
	var query string
	filter = s.ensureScanFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	filterStr := s.getScanFilterString(filter)
	joins := s.getScanJoins(filter)
	cursor := getCursor(filter.Paginated, filterStr, "S.scan_id > ?")

	whereClause := ""
	if filterStr != "" || withCursor {
		whereClause = fmt.Sprintf("WHERE %s", filterStr)
	}

	// construct final query
	if withCursor {
		query = fmt.Sprintf(baseQuery, joins, whereClause, cursor.Statement)
	} else {
		query = fmt.Sprintf(baseQuery, joins, whereClause)
	}

	//construct prepared statement and if where clause does exist add parameters
	var stmt *sqlx.Stmt
	var err error

	stmt, err = s.db.Preparex(query)
	if err != nil {
		msg := ERROR_MSG_PREPARED_STMT
		l.WithFields(
			logrus.Fields{
				"error": err,
				"query": query,
				"stmt":  stmt,
			}).Error(msg)
		return nil, nil, fmt.Errorf("%s", msg)
	}

	//adding parameters
	var filterParameters []interface{}
	filterParameters = buildQueryParameters(filterParameters, filter.Id)
	if withCursor {
		filterParameters = append(filterParameters, cursor.Value)
		filterParameters = append(filterParameters, cursor.Limit)
	}

	return stmt, filterParameters, nil

}

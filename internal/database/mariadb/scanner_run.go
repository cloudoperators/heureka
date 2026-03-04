// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

func (s *SqlDatabase) CreateScannerRun(scannerRun *entity.ScannerRun) (bool, error) {
	l := logrus.WithFields(logrus.Fields{
		"scannerrun": scannerRun,
		"event":      "database.CreateScannerRun",
	})

	query := `
		INSERT INTO ScannerRun (
			scannerrun_uuid,
			scannerrun_tag,
			scannerrun_start_run,
			scannerrun_end_run,
			scannerrun_is_completed,
			scannerrun_created_by,
			scannerrun_updated_by
		) VALUES (
			:scannerrun_uuid,
			:scannerrun_tag,
			:scannerrun_start_run,
			:scannerrun_end_run,
			:scannerrun_is_completed,
			:scannerrun_created_by,
			:scannerrun_updated_by
		)
	`

	srr := ScannerRunRow{}
	srr.FromScannerRun(scannerRun)

	id, err := performInsert(s, query, srr, l)
	if err != nil {
		return false, err
	}

	scannerRun.RunID = id

	return true, nil
}

func (s *SqlDatabase) CompleteScannerRun(uuid string) (bool, error) {
	updateQuery := `UPDATE ScannerRun 
					SET 
						scannerrun_is_completed = TRUE,
						scannerrun_end_run = current_timestamp()
					WHERE 
						scannerrun_uuid = ? AND
						scannerrun_is_completed = FALSE`

	_, err := s.db.Exec(updateQuery, uuid)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *SqlDatabase) FailScannerRun(uuid string, message string) (bool, error) {
	updateScannerRunQuery := `UPDATE ScannerRun 
					SET 
						scannerrun_is_completed = FALSE,
						scannerrun_end_run = current_timestamp()
					WHERE 
						scannerrun_uuid = ? AND
						scannerrun_is_completed = TRUE`

	insertScannerRunErrorQuery := `INSERT INTO ScannerRunError 
										(scannerrunerror_scannerrun_run_id, error) 
								   VALUES (
								   		(SELECT scannerrun_run_id FROM ScannerRun WHERE scannerrun_uuid = ?),
										?)`

	_, err := s.db.Exec(updateScannerRunQuery, uuid)

	if err != nil {
		return false, err
	} else {
		_, err = s.db.Exec(insertScannerRunErrorQuery, uuid, message)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

func (s *SqlDatabase) ScannerRunByUUID(uuid string) (*entity.ScannerRun, error) {
	query := `SELECT 
				* 
			  FROM ScannerRun 
			  WHERE scannerrun_uuid = ?`

	srr := ScannerRunRow{}
	err := s.db.Get(&srr, query, uuid)
	if err != nil {
		return nil, err
	}

	sr := srr.AsScannerRun()
	return &sr, nil
}

func (s *SqlDatabase) GetScannerRuns(filter *entity.ScannerRunFilter) ([]entity.ScannerRun, error) {
	filter = ensureScannerRunFilter(filter)

	baseQuery := `
		SELECT * FROM ScannerRun
    `
	queryArgs, baseQuery := applyScannerRunFilter(baseQuery, filter)
	rows, err := s.db.Query(baseQuery, queryArgs...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	result := []entity.ScannerRun{}

	for rows.Next() {
		srr := ScannerRunRow{}
		err = rows.Scan(&srr.RunID, &srr.UUID, &srr.Tag, &srr.StartRun, &srr.EndRun, &srr.IsCompleted, &srr.CreatedAt, &srr.CreatedBy, &srr.DeletedAt, &srr.UpdatedAt, &srr.UpdatedBy)
		if err != nil {
			return nil, err
		}

		result = append(result, srr.AsScannerRun())
	}

	return result, nil
}

func applyScannerRunFilter(baseQuery string, filter *entity.ScannerRunFilter) ([]any, string) {
	queryArgs := []any{}

	baseQuery += " WHERE"

	for i := 0; filter.Tag != nil && i < len(filter.Tag); i++ {
		baseQuery += " scannerrun_tag = ?"
		queryArgs = append(queryArgs, filter.Tag[i])
		if i < len(filter.Tag)-1 {
			baseQuery += " OR"
		}
	}

	if filter.Completed {
		if len(filter.Tag) > 0 {
			baseQuery += " AND"
		}
		baseQuery += " scannerrun_is_completed = TRUE"
	}

	if filter.HasArgs() {
		baseQuery += " AND"
	}

	if filter.After != nil {
		baseQuery += " scannerrun_run_id > ?"
		queryArgs = append(queryArgs, *filter.After)
	}

	baseQuery += " ORDER BY scannerrun_run_id"

	queryArgs = append(queryArgs, *filter.First)
	baseQuery += " LIMIT ?"
	return queryArgs, baseQuery
}

func ensureScannerRunFilter(f *entity.ScannerRunFilter) *entity.ScannerRunFilter {
	var first int = 100
	var after string
	if f == nil {
		return &entity.ScannerRunFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
	}
	if f.First == nil {
		f.First = &first
	}
	if f.After == nil {
		f.After = &after
	}
	return f
}

func (s *SqlDatabase) GetScannerRunTags() ([]string, error) {
	query := `SELECT DISTINCT
				scannerrun_tag 
			  FROM ScannerRun`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	res := []string{}

	for rows.Next() {

		var tag string
		err = rows.Scan(&tag)
		if err != nil {
			return nil, err
		}

		res = append(res, tag)
	}

	return res, nil
}

func (s *SqlDatabase) CountScannerRuns(filter *entity.ScannerRunFilter) (int, error) {
	query := `SELECT COUNT(*) AS ScannerRunCount 
			  FROM ScannerRun`

	args, query := applyScannerRunFilter(query, filter)
	row := s.db.QueryRow(query, args...)

	if row.Err() != nil {
		return -1, row.Err()
	}

	var res int

	row.Scan(&res)

	return res, nil
}

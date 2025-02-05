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
			scannerrun_is_completed
		) VALUES (
			:scannerrun_uuid,
			:scannerrun_tag,
			:scannerrun_start_run,
			:scannerrun_end_run,
			:scannerrun_is_completed
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
	filter = s.ensureScannerRunFilter(filter)

	baseQuery := `
		SELECT * FROM ScannerRun
    `
	queryArgs := []interface{}{}

	if filter.HasArgs() {
		baseQuery += " WHERE"
	}

	for i := 0; filter.Tag != nil && i < len(filter.Tag); i++ {
		baseQuery += " scannerrun_tag = ?"
		queryArgs = append(queryArgs, filter.Tag[i])
		if i < len(filter.Tag)-1 {
			baseQuery += " OR"
		}
	}

	if filter.Completed {
		baseQuery += " scannerrun_is_completed = TRUE"
	}

	baseQuery += " ORDER BY scannerrun_run_id"
	rows, err := s.db.Query(baseQuery, queryArgs...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	result := []entity.ScannerRun{}

	for rows.Next() {
		srr := ScannerRunRow{}
		err = rows.Scan(&srr.RunID, &srr.UUID, &srr.Tag, &srr.StartRun, &srr.EndRun, &srr.IsCompleted)

		if err != nil {
			return nil, err
		}

		result = append(result, srr.AsScannerRun())
	}

	return result, nil
}

func (s *SqlDatabase) ensureScannerRunFilter(f *entity.ScannerRunFilter) *entity.ScannerRunFilter {
	var first int = 1000
	var after int64 = 0
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

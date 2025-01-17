// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

func (s *SqlDatabase) CreateScannerRun(scannerRun *entity.ScannerRun) (*entity.ScannerRun, error) {
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
		return nil, err
	}

	scannerRun.RunID = id

	return scannerRun, nil
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

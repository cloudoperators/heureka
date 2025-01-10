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

func (s *SqlDatabase) CompleteScannerRun(scannerRun *entity.ScannerRun) error {
	updateQuery := `UPDATE ScannerRun 
					SET 
						scannerrun_is_completed = TRUE,
						scannerrun_end_run = current_timestamp()
					WHERE 
						scannerrun_run_id = :scannerrun_run_id AND
						scannerrun_is_completed = FALSE`

	srr := ScannerRunRow{}
	srr.FromScannerRun(scannerRun)

	tx, err := s.db.Beginx()

	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.NamedExec(updateQuery, srr)

	if err != nil {
		return err
	}

	newSRR := ScannerRunRow{}
	err = tx.Get(&newSRR,
		`SELECT * FROM ScannerRun WHERE scannerrun_run_id = ?`, scannerRun.RunID)

	if err != nil {
		return err
	}

	updatedScannerRun := newSRR.AsScannerRun()

	scannerRun.StartRun = updatedScannerRun.StartRun
	scannerRun.EndRun = updatedScannerRun.EndRun
	scannerRun.Completed = updatedScannerRun.Completed

	return tx.Commit()
}

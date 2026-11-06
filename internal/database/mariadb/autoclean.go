// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/cloudoperators/heureka/internal/util"
	"github.com/sirupsen/logrus"
)

func (s *SqlDatabase) Autoclean(ctx context.Context, keepNRuns uint64) error {
	if keepNRuns == 0 {
		return nil
	}

	runs, err := s.fetchCompletedRunsWithNewestFirst(ctx)
	if err != nil {
		return err
	}

	return s.processAutocleanOnCompletedRuns(ctx, runs, keepNRuns)
}

func (s *SqlDatabase) processAutocleanOnCompletedRuns(ctx context.Context, runs map[string][]int64, keepNRuns uint64) error {
	runsToDelete := getOnlyRunsToDelete(runs, keepNRuns)
	runIdsToDelete := getRunIds(runsToDelete)

	if err := s.deleteErrorsByRunIds(ctx, runIdsToDelete); err != nil {
		return err
	}

	if err := s.deleteTrackersByRunIds(ctx, runIdsToDelete); err != nil {
		return err
	}

	if err := s.deleteRunsByRunIds(ctx, runIdsToDelete); err != nil {
		return err
	}

	return nil
}

func getOnlyRunsToDelete(runs map[string][]int64, keepNRuns uint64) map[string][]int64 {
	result := map[string][]int64{}

	for tag, tagRuns := range runs {
		if keepNRuns >= uint64(len(tagRuns)) {
			continue
		} else {
			result[tag] = tagRuns[keepNRuns:]
		}
	}

	return result
}

func getRunIds(runs map[string][]int64) []int64 {
	result := []int64{}

	for _, r := range runs {
		result = append(result, r...)
	}

	return result
}

func (s *SqlDatabase) deleteErrorsByRunIds(ctx context.Context, runIds []int64) error {
	query, args, err := sq.Delete("ScannerRunError").
		Where(sq.Eq{"scannerrunerror_scannerrun_run_id": runIds}).
		ToSql()
	if err != nil {
		return err
	}

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	logrus.Infof("Deleted %d Scanner Run Errors", rowsAffected)

	return nil
}

func (s *SqlDatabase) deleteTrackersByRunIds(ctx context.Context, runIds []int64) error {
	query, args, err := sq.Delete("ScannerRunComponentInstanceTracker").
		Where(sq.Eq{"scannerruncomponentinstancetracker_scannerrun_run_id": runIds}).
		ToSql()
	if err != nil {
		return err
	}

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	logrus.Infof("Deleted %d Scanner Run Trackers", rowsAffected)

	return nil
}

func (s *SqlDatabase) deleteRunsByRunIds(ctx context.Context, runIds []int64) error {
	qb := sq.
		Update("ScannerRun").
		Set("scannerrun_deleted_at", sq.Expr("NOW()")).
		Set("scannerrun_updated_by", util.SystemUserId).
		Where(sq.Eq{"scannerrun_run_id": runIds})

	query, args, err := qb.ToSql()
	if err != nil {
		return err
	}

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	logrus.Infof("Soft-Deleted %d Scanner Runs", rowsAffected)

	return nil
}

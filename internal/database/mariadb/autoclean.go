// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/cloudoperators/heureka/internal/util"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

func (s *SqlDatabase) Autoclean(ctx context.Context, keepNRuns uint64) error {
	if keepNRuns == 0 {
		return nil
	}

	runsWithStatus, err := s.fetchRunsWithNewestFirst(ctx)
	if err != nil {
		return err
	}

	return s.processAutocleanOnCompletedRuns(ctx, runsWithStatus, int(keepNRuns))
}

func (s *SqlDatabase) processAutocleanOnCompletedRuns(ctx context.Context, runs map[string][]scannerRunWithStatus, keepNRuns int) error {
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

func getOnlyRunsToDelete(runsWithStatus map[string][]scannerRunWithStatus, keepNRuns int) map[string][]int64 {
	return lo.MapValues(runsWithStatus, func(v []scannerRunWithStatus, tag string) []int64 {
		completedCnt := 0

		return lo.FilterMap(v, func(v0 scannerRunWithStatus, n int) (int64, bool) {
			if v0.completed {
				if n < keepNRuns || completedCnt < 2 {
					completedCnt += 1
					return 0, false
				}
			} else {
				if n < keepNRuns {
					return 0, false
				}
			}

			return v0.id, true
		})
	})
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

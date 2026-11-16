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

func (s *SqlDatabase) Autoclean(ctx context.Context, keepNRuns int) error {
	if keepNRuns == 0 {
		return nil
	}

	if err := s.acquireScannerRunAccess(defaultTimeoutForScannerRunLock); err != nil {
		return err
	}
	defer s.releaseScannerRunAccess()

	runsWithStatus, err := s.fetchRunsWithNewestFirst(ctx)
	if err != nil {
		return err
	}

	return s.processAutocleanScannerRuns(ctx, runsWithStatus, int(keepNRuns))
}

type scannerRunWithStatus struct {
	id        int64
	completed bool
}

func (s *SqlDatabase) fetchRunsWithNewestFirst(ctx context.Context) (map[string][]scannerRunWithStatus, error) {
	query, args, err := sq.Select("scannerrun_tag", "scannerrun_run_id", "scannerrun_is_completed").
		From("ScannerRun").
		Where("scannerrun_deleted_at IS NULL").
		OrderBy("scannerrun_tag, scannerrun_run_id DESC").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logrus.Warnf("error during close rows: %s", err)
		}
	}()

	// tag -> list of runs (newest first)
	runs := map[string][]scannerRunWithStatus{}

	for rows.Next() {
		var (
			tag       string
			id        int64
			completed bool
		)

		if err := rows.Scan(&tag, &id, &completed); err != nil {
			return nil, err
		}

		runs[tag] = append(runs[tag], scannerRunWithStatus{id: id, completed: completed})
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return runs, nil
}

func (s *SqlDatabase) processAutocleanScannerRuns(ctx context.Context, runs map[string][]scannerRunWithStatus, keepNRuns int) error {
	runsToKeep := getOnlyRunsToKeep(runs, keepNRuns)
	runIdsToKeep := getRunIds(runsToKeep)

	if err := s.deleteErrorsFromNotListedRuns(ctx, runIdsToKeep); err != nil {
		return err
	}

	if err := s.deleteTrackersFromNotListedRuns(ctx, runIdsToKeep); err != nil {
		return err
	}

	if err := s.deleteNotListedRuns(ctx, runIdsToKeep); err != nil {
		return err
	}

	return nil
}

func getOnlyRunsToKeep(runsWithStatus map[string][]scannerRunWithStatus, keepNRuns int) map[string][]int64 {
	return lo.MapValues(runsWithStatus, func(v []scannerRunWithStatus, tag string) []int64 {
		completedCnt := 0

		return lo.FilterMap(v, func(v0 scannerRunWithStatus, n int) (int64, bool) {
			if v0.completed {
				if n < keepNRuns || completedCnt < 2 {
					completedCnt += 1
					return v0.id, true
				}
			} else {
				if n < keepNRuns {
					return v0.id, true
				}
			}

			return 0, false
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

func (s *SqlDatabase) deleteErrorsFromNotListedRuns(ctx context.Context, runIds []int64) error {
	query, args, err := sq.Delete("ScannerRunError").
		Where(sq.NotEq{"scannerrunerror_scannerrun_run_id": runIds}).
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

func (s *SqlDatabase) deleteTrackersFromNotListedRuns(ctx context.Context, runIds []int64) error {
	query, args, err := sq.Delete("ScannerRunComponentInstanceTracker").
		Where(sq.NotEq{"scannerruncomponentinstancetracker_scannerrun_run_id": runIds}).
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

func (s *SqlDatabase) deleteNotListedRuns(ctx context.Context, runIds []int64) error {
	qb := sq.
		Update("ScannerRun").
		Set("scannerrun_deleted_at", sq.Expr("NOW()")).
		Set("scannerrun_updated_by", util.SystemUserId).
		Where(sq.NotEq{"scannerrun_run_id": runIds})

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

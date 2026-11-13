// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/util"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

func (s *SqlDatabase) Autopatch(ctx context.Context) (bool, error) {
	runs, err := s.fetchCompletedRunsWithNewestFirst(ctx)
	if err != nil {
		return false, err
	}

	return s.processAutopatchOnCompletedRuns(ctx, runs)
}

type scannerRunWithStatus struct {
	id        int64
	completed bool
}

func (s *SqlDatabase) fetchRunsWithNewestFirst(ctx context.Context) (map[string][]scannerRunWithStatus, error) {
	query, args, err := sq.Select("scannerrun_tag", "scannerrun_run_id", "scannerrun_is_completed").
		From("ScannerRun").
		OrderBy("scannerrun_tag, scannerrun_run_id DESC").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

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

func (s *SqlDatabase) fetchCompletedRunsWithNewestFirst(ctx context.Context) (map[string][]int64, error) {
	runsWithStatus, err := s.fetchRunsWithNewestFirst(ctx)
	if err != nil {
		return nil, err
	}

	return lo.MapValues(runsWithStatus, func(v []scannerRunWithStatus, tag string) []int64 {
		return lo.FilterMap(v, func(v0 scannerRunWithStatus, _ int) (int64, bool) {
			if !v0.completed {
				return 0, false
			}

			return v0.id, true
		})
	}), nil
}

func (s *SqlDatabase) processAutopatchOnCompletedRuns(ctx context.Context, runs map[string][]int64) (bool, error) {
	autopatched := false

	for _, tagRuns := range runs {
		// Need at least two completed runs
		if len(tagRuns) < 2 {
			continue
		}

		patchedForTag, err := s.processAutopatchForSingleTag(ctx, tagRuns)
		if err != nil {
			return false, err
		}

		if patchedForTag {
			autopatched = true
		}
	}

	return autopatched, nil
}

func (s *SqlDatabase) processAutopatchForSingleTag(ctx context.Context, tagRuns []int64) (bool, error) {
	latest := tagRuns[0]
	secondLatest := tagRuns[1]

	// Fetch ComponentInstances for each run
	latestInstances, err := s.fetchComponentInstancesForRun(latest)
	if err != nil {
		return false, err
	}

	secondLatestInstances, err := s.fetchComponentInstancesForRun(secondLatest)
	if err != nil {
		return false, err
	}

	disappearedInstances := getDisappearedInstances(latestInstances, secondLatestInstances)
	if len(disappearedInstances) == 0 {
		return false, nil
	}

	patches, err := s.getPatches(disappearedInstances)
	if err != nil {
		return false, err
	}

	err = s.deleteIssueMatchesOfDisappearedInstances(ctx, disappearedInstances)
	if err != nil {
		return false, err
	}

	versionsOfDisappearedInstances, err := s.getVersionIdsOfDisappearedInstances(
		ctx,
		disappearedInstances,
	)
	if err != nil {
		return false, err
	}

	err = s.deleteDisappearedInstances(disappearedInstances)
	if err != nil {
		return false, err
	}

	componentsOfDisappearedInstances, err := s.getComponentIdsOfDisappearedInstances(
		ctx,
		versionsOfDisappearedInstances,
	)
	if err != nil {
		return false, err
	}

	err = s.deleteVersionsOfDisappearedInstances(ctx, versionsOfDisappearedInstances)
	if err != nil {
		return false, err
	}

	err = s.deleteComponentsOfDisappearedInstances(ctx, componentsOfDisappearedInstances)
	if err != nil {
		return false, err
	}

	err = s.insertPatches(patches)
	if err != nil {
		return false, err
	}

	return true, nil
}

func getDisappearedInstances(
	latestInstances map[int64]struct{},
	secondLatestInstances map[int64]struct{},
) []int64 {
	// Compute disappeared instances
	var disappeared []int64

	for inst := range secondLatestInstances {
		if _, stillThere := latestInstances[inst]; !stillThere {
			disappeared = append(disappeared, inst)
		}
	}

	return disappeared
}

func (s *SqlDatabase) getPatches(disappearedInstances []int64) (map[patchInfo]struct{}, error) {
	patches := make(map[patchInfo]struct{})

	for _, inst := range disappearedInstances {
		patchInfo, err := s.fetchServiceAndVersionForInstance(inst)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue // skip instances with no rows
			}

			return nil, err
		}

		patches[patchInfo] = struct{}{}
	}

	return patches, nil
}

// Insert patches only for service/version which does not reflect any existing component instance
// (for removed instances)
func (s *SqlDatabase) insertPatches(patches map[patchInfo]struct{}) error {
	for patch := range patches {
		if err := s.insertPatchIfNoInstanceExists(patch); err != nil {
			return err
		}
	}

	return nil
}

func (s *SqlDatabase) deleteIssueMatchesOfDisappearedInstances(ctx context.Context, disappearedInstances []int64) error {
	for _, di := range disappearedInstances {
		issueMatchFilter := entity.IssueMatchFilter{ComponentInstanceId: []*int64{new(int64(di))}}

		issueMatchIds, err := s.GetAllIssueMatchIds(ctx, &issueMatchFilter)
		if err != nil {
			return err
		}

		for _, issueMatchId := range issueMatchIds {
			if err := s.DeleteIssueMatch(issueMatchId, util.SystemUserId); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *SqlDatabase) getVersionIdsOfDisappearedInstances(
	ctx context.Context,
	disappearedInstances []int64,
) (map[int64]struct{}, error) {
	idsDisappeared := lo.Map(disappearedInstances, func(v int64, _ int) *int64 {
		vv := int64(v)
		return &vv
	})

	cif := entity.ComponentInstanceFilter{Id: idsDisappeared}

	res, err := s.GetComponentInstances(ctx, &cif, nil)
	if err != nil {
		return nil, err
	}

	versionIdsOfDisappearedInstances := make(map[int64]struct{})
	for _, ci := range res {
		versionIdsOfDisappearedInstances[ci.ComponentVersionId] = struct{}{}
	}

	return versionIdsOfDisappearedInstances, nil
}

func (s *SqlDatabase) getComponentIdsOfDisappearedInstances(
	ctx context.Context,
	versions map[int64]struct{},
) (map[int64]struct{}, error) {
	var versionsToFilter []*int64
	for v := range versions {
		versionsToFilter = append(versionsToFilter, &v)
	}

	cvf := entity.ComponentVersionFilter{Id: versionsToFilter}

	res, err := s.GetComponentVersions(ctx, &cvf, nil)
	if err != nil {
		return nil, err
	}

	componentIdsOfDisappearedInstances := make(map[int64]struct{})
	for _, cv := range res {
		componentIdsOfDisappearedInstances[cv.ComponentId] = struct{}{}
	}

	return componentIdsOfDisappearedInstances, nil
}

func (s *SqlDatabase) deleteVersionsOfDisappearedInstances(
	ctx context.Context,
	versionIdsOfDisappearedInstances map[int64]struct{},
) error {
	for vIdDi := range versionIdsOfDisappearedInstances {
		cif := entity.ComponentInstanceFilter{ComponentVersionId: []*int64{&vIdDi}}

		res, err := s.GetComponentInstances(ctx, &cif, nil)
		if err != nil {
			return err
		}

		if len(res) == 0 {
			for vIdDi := range versionIdsOfDisappearedInstances {
				if err := s.RemoveAllIssuesFromComponentVersion(vIdDi); err != nil {
					return err
				}
			}

			if err := s.DeleteComponentVersion(vIdDi, util.SystemUserId); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *SqlDatabase) deleteComponentsOfDisappearedInstances(
	ctx context.Context,
	componentIdsOfDisappearedInstances map[int64]struct{},
) error {
	for cIdDi := range componentIdsOfDisappearedInstances {
		cvf := entity.ComponentVersionFilter{ComponentId: []*int64{&cIdDi}}

		res, err := s.GetComponentVersions(ctx, &cvf, nil)
		if err != nil {
			return err
		}

		if len(res) == 0 {
			if err := s.DeleteComponent(cIdDi, util.SystemUserId); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *SqlDatabase) deleteDisappearedInstances(disappearedInstances []int64) error {
	for _, di := range disappearedInstances {
		if err := s.DeleteComponentInstance(int64(di), util.SystemUserId); err != nil {
			return err
		}
	}

	return nil
}

type patchInfo struct {
	serviceId            int64
	serviceName          string
	componentVersionId   int64
	componentVersionName string
}

func (s *SqlDatabase) fetchComponentInstancesForRun(scannerRunId int64) (map[int64]struct{}, error) {
	rows, err := s.db.Query(`
        SELECT scannerruncomponentinstancetracker_component_instance_id
        FROM ScannerRunComponentInstanceTracker
        WHERE scannerruncomponentinstancetracker_scannerrun_run_id = ?
    `, scannerRunId)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			logrus.Warnf("error during close rows: %s", err)
		}
	}()

	instances := map[int64]struct{}{}

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		instances[id] = struct{}{}
	}

	return instances, rows.Err()
}

func (s *SqlDatabase) fetchServiceAndVersionForInstance(instanceID int64) (patchInfo, error) {
	query := `
        SELECT
            ci.componentinstance_service_id,
            s.service_ccrn,
            ci.componentinstance_component_version_id,
            cv.componentversion_version
        FROM ComponentInstance ci
        INNER JOIN Service s
            ON ci.componentinstance_service_id = s.service_id AND s.service_deleted_at IS NULL
        INNER JOIN ComponentVersion cv
            ON ci.componentinstance_component_version_id = cv.componentversion_id AND cv.componentversion_deleted_at IS NULL
        WHERE ci.componentinstance_id = ? AND ci.componentinstance_deleted_at IS NULL`

	row := s.db.QueryRow(query, instanceID)

	var pInfo patchInfo

	err := row.Scan(
		&pInfo.serviceId,
		&pInfo.serviceName,
		&pInfo.componentVersionId,
		&pInfo.componentVersionName,
	)

	return pInfo, err
}

func (s *SqlDatabase) insertPatchIfNoInstanceExists(patch patchInfo) error {
	query := `
		INSERT INTO Patch (
			patch_service_id,
			patch_service_name,
			patch_component_version_id,
			patch_component_version_name
		)
		SELECT ?, ?, ?, ?
		WHERE NOT EXISTS (
			SELECT 1
			FROM ComponentInstance ci
			WHERE ci.componentinstance_service_id = ?
			  AND ci.componentinstance_component_version_id = ?
              AND ci.componentinstance_deleted_at IS NULL
		)
	`

	_, err := s.db.Exec(
		query,
		patch.serviceId,
		patch.serviceName,
		patch.componentVersionId,
		patch.componentVersionName,
		patch.serviceId,
		patch.componentVersionId,
	)
	if err != nil {
		return err
	}

	return nil
}

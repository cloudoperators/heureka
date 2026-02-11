// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/util"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

func (s *SqlDatabase) Autopatch() (bool, error) {
	runs, err := s.fetchCompletedRunsWithNewestFirst()
	if err != nil {
		return false, err
	}

	return s.processAutopatchOnCompletedRuns(runs)
}

func (s *SqlDatabase) fetchCompletedRunsWithNewestFirst() (map[string][]int, error) {
	rows, err := s.db.Query(`
        SELECT scannerrun_tag, scannerrun_run_id
        FROM ScannerRun
        WHERE scannerrun_is_completed = TRUE AND scannerrun_deleted_at IS NULL
        ORDER BY scannerrun_tag, scannerrun_run_id DESC
    `)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logrus.Warnf("error during closing rows: %s", err)
		}
	}()

	// tag -> list of runs (newest first)
	runs := map[string][]int{}

	for rows.Next() {
		var tag string
		var id int
		if err := rows.Scan(&tag, &id); err != nil {
			return nil, err
		}
		runs[tag] = append(runs[tag], id)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return runs, nil
}

func (s *SqlDatabase) processAutopatchOnCompletedRuns(runs map[string][]int) (bool, error) {
	autopatched := false
	for _, tagRuns := range runs {

		// Need at least two completed runs
		if len(tagRuns) < 2 {
			continue
		}

		patchedForTag, err := s.processAutopatchForSingleTag(tagRuns)
		if err != nil {
			return false, err
		}
		if patchedForTag {
			autopatched = true
		}
	}

	return autopatched, nil
}

func (s *SqlDatabase) processAutopatchForSingleTag(tagRuns []int) (bool, error) {
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

	err = s.deleteIssueMatchesOfDisappearedInstances(disappearedInstances)
	if err != nil {
		return false, err
	}

	versionsOfDisappearedInstances, err := s.getVersionIdsOfDisappearedInstances(disappearedInstances)
	if err != nil {
		return false, err
	}

	err = s.deleteDisappearedInstances(disappearedInstances)
	if err != nil {
		return false, err
	}

	err = s.deleteVersionIssuesOfDisappearedInstances(versionsOfDisappearedInstances)
	if err != nil {
		return false, err
	}

	componentsOfDisappearedInstances, err := s.getComponentIdsOfDisappearedInstances(versionsOfDisappearedInstances)
	if err != nil {
		return false, err
	}

	err = s.deleteVersionsOfDisappearedInstances(versionsOfDisappearedInstances)
	if err != nil {
		return false, err
	}

	err = s.deleteComponentsOfDisappearedInstances(componentsOfDisappearedInstances)
	if err != nil {
		return false, err
	}

	err = s.insertPatches(patches)
	if err != nil {
		return false, err
	}

	return true, nil
}

func getDisappearedInstances(latestInstances map[int]struct{}, secondLatestInstances map[int]struct{}) []int {
	// Compute disappeared instances
	var disappeared []int
	for inst := range secondLatestInstances {
		if _, stillThere := latestInstances[inst]; !stillThere {
			disappeared = append(disappeared, inst)
		}
	}
	return disappeared
}

func (s *SqlDatabase) getPatches(disappearedInstances []int) (map[patchInfo]struct{}, error) {
	patches := make(map[patchInfo]struct{})
	for _, inst := range disappearedInstances {
		patchInfo, err := s.fetchServiceAndVersionForInstance(inst)
		if err != nil {
			return nil, err
		}
		patches[patchInfo] = struct{}{}
	}
	return patches, nil
}

// Insert patches only for service/version which does not reflect any existing component instance (for removed instances)
func (s *SqlDatabase) insertPatches(patches map[patchInfo]struct{}) error {
	for patch := range patches {
		if err := s.insertPatchIfNoInstanceExists(patch); err != nil {
			return err
		}
	}
	return nil
}

func (s *SqlDatabase) deleteIssueMatchesOfDisappearedInstances(disappearedInstances []int) error {
	for _, di := range disappearedInstances {
		issueMatchFilter := entity.IssueMatchFilter{ComponentInstanceId: []*int64{lo.ToPtr(int64(di))}}
		issueMatchIds, err := s.GetAllIssueMatchIds(&issueMatchFilter)
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

func (s *SqlDatabase) getVersionIdsOfDisappearedInstances(disappearedInstances []int) (map[int64]struct{}, error) {
	idsDisappeared := lo.Map(disappearedInstances, func(v int, _ int) *int64 {
		vv := int64(v)
		return &vv
	})

	cif := entity.ComponentInstanceFilter{Id: idsDisappeared}
	res, err := s.GetComponentInstances(&cif, nil)
	if err != nil {
		return nil, err
	}

	versionIdsOfDisappearedInstances := make(map[int64]struct{})
	for _, ci := range res {
		versionIdsOfDisappearedInstances[ci.ComponentVersionId] = struct{}{}
	}
	return versionIdsOfDisappearedInstances, nil
}

func (s *SqlDatabase) getComponentIdsOfDisappearedInstances(versions map[int64]struct{}) (map[int64]struct{}, error) {
	var versionsToFilter []*int64
	for v := range versions {
		versionsToFilter = append(versionsToFilter, &v)
	}

	cvf := entity.ComponentVersionFilter{Id: versionsToFilter}
	res, err := s.GetComponentVersions(&cvf, nil)
	if err != nil {
		return nil, err
	}

	componentIdsOfDisappearedInstances := make(map[int64]struct{})
	for _, cv := range res {
		componentIdsOfDisappearedInstances[cv.ComponentId] = struct{}{}
	}
	return componentIdsOfDisappearedInstances, nil
}

func (s *SqlDatabase) deleteVersionIssuesOfDisappearedInstances(versionIdsOfDisappearedInstances map[int64]struct{}) error {
	for vIdDi := range versionIdsOfDisappearedInstances {
		if err := s.RemoveAllIssuesFromComponentVersion(vIdDi); err != nil {
			return err
		}
	}

	return nil
}

func (s *SqlDatabase) deleteVersionsOfDisappearedInstances(versionIdsOfDisappearedInstances map[int64]struct{}) error {
	for vIdDi := range versionIdsOfDisappearedInstances {
		cif := entity.ComponentInstanceFilter{ComponentVersionId: []*int64{&vIdDi}}
		res, err := s.GetComponentInstances(&cif, nil)
		if err != nil {
			return err
		}
		if len(res) == 0 {
			if err := s.DeleteComponentVersion(vIdDi, util.SystemUserId); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *SqlDatabase) deleteComponentsOfDisappearedInstances(componentIdsOfDisappearedInstances map[int64]struct{}) error {
	for cIdDi := range componentIdsOfDisappearedInstances {
		cvf := entity.ComponentVersionFilter{ComponentId: []*int64{&cIdDi}}
		res, err := s.GetComponentVersions(&cvf, nil)
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

func (s *SqlDatabase) deleteDisappearedInstances(disappearedInstances []int) error {
	for _, di := range disappearedInstances {
		if err := s.DeleteComponentInstance(int64(di), util.SystemUserId); err != nil {
			return err
		}
	}
	return nil
}

type patchInfo struct {
	serviceId            int
	serviceName          string
	componentVersionId   int
	componentVersionName string
}

func (s *SqlDatabase) fetchComponentInstancesForRun(scannerRunId int) (map[int]struct{}, error) {
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
			logrus.Warnf("error during closing rows: %s", err)
		}
	}()

	instances := map[int]struct{}{}
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		instances[id] = struct{}{}
	}
	return instances, rows.Err()
}

func (s *SqlDatabase) fetchServiceAndVersionForInstance(instanceID int) (patchInfo, error) {
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

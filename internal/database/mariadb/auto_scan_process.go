// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"strings"
)

func (s *SqlDatabase) Autopatch() (bool, error) {
	runs, err := s.fetchCompletedRunsWithNewestFirst()
	if err != nil {
		return false, err
	}

	return s.processAutopatchOnCompletedRuns(runs)
}

func (s *SqlDatabase) Autoclose() (bool, error) {
	runs, err := s.fetchCompletedRunsWithNewestFirst()
	if err != nil {
		return false, err
	}

	return s.processAutocloseOnCompletedRuns(runs)
}

func (s *SqlDatabase) fetchCompletedRunsWithNewestFirst() (map[string][]int, error) {
	rows, err := s.db.Query(`
        SELECT scannerrun_tag, scannerrun_run_id
        FROM ScannerRun
        WHERE scannerrun_is_completed = TRUE
        ORDER BY scannerrun_tag, scannerrun_run_id DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

type disappearedInstance struct {
	instId int
	runId  int
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

	// Compute disappeared instances
	var disappeared []disappearedInstance
	for inst := range secondLatestInstances {
		if _, stillThere := latestInstances[inst]; !stillThere {
			disappeared = append(disappeared, disappearedInstance{instId: inst, runId: latest})
		}
	}

	if len(disappeared) == 0 {
		return false, nil
	}

	//TODO: get disappearedServiceVersion
	// iterate through patches related to service/version instead of componentinstance

	// Create a patch for each disappeared instance
	for _, di := range disappeared {
		if err := s.insertPatch(di); err != nil {
			return false, err
		}
	}

	//TODO: remove IssueMatches iterating through disappearedComponentInstances and listing issueMatches (remove autoclosing because remove of issueMatches replaces mittigated state)
	//TODO: remove componentinstances iterating through disappearedComponentInstances
	return true, nil
}

func (s *SqlDatabase) processAutocloseOnCompletedRuns(runs map[string][]int) (bool, error) {
	autoclosed := false

	// For each tag, process only latest + second-latest
	for _, tagRuns := range runs {

		// Ensure it has at least 2 completed runs
		if len(tagRuns) < 2 {
			continue
		}

		closedForTag, err := s.processAutocloseForSingleTag(tagRuns)
		if err != nil {
			return false, err
		}
		if closedForTag {
			autoclosed = true
		}
	}
	return autoclosed, nil
}

func (s *SqlDatabase) processAutocloseForSingleTag(tagRuns []int) (bool, error) {
	latest := tagRuns[0]
	secondLatest := tagRuns[1]

	// fetch issues for each run
	latestIssues, err := s.fetchIssuesForRun(latest)
	if err != nil {
		return false, err
	}

	secondIssues, err := s.fetchIssuesForRun(secondLatest)
	if err != nil {
		return false, err
	}

	// Compute which issues disappeared
	var missing []int
	for issue := range secondIssues {
		if _, stillThere := latestIssues[issue]; !stillThere {
			missing = append(missing, issue)
		}
	}

	// Mark as mitigated
	if len(missing) > 0 {
		if err := s.markIssuesMitigated(missing); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
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
	defer rows.Close()

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

func (s *SqlDatabase) fetchIssuesForRun(scannerRunId int) (map[int]struct{}, error) {
	rows, err := s.db.Query(`
        SELECT scannerrunissuetracker_issue_id
        FROM ScannerRunIssueTracker
        WHERE scannerrunissuetracker_scannerrun_run_id = ?
    `, scannerRunId)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	issues := map[int]struct{}{}
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		issues[id] = struct{}{}
	}
	return issues, rows.Err()
}

type patchInfo struct {
	serviceId            int
	serviceName          string
	componentVersionId   int
	componentVersionName string
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
            ON ci.componentinstance_service_id = s.service_id
        INNER JOIN ComponentVersion cv
            ON ci.componentinstance_component_version_id = cv.componentversion_id
        WHERE ci.componentinstance_id = ?`

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

func (s *SqlDatabase) insertPatch(di disappearedInstance) error {
	pInfo, err := s.fetchServiceAndVersionForInstance(di.instId)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
        INSERT INTO Patch (patch_service_id, patch_service_name, patch_component_version_id, patch_component_version_name)
        VALUES (?, ?, ?, ?)
    `, pInfo.serviceId, pInfo.serviceName, pInfo.componentVersionId, pInfo.componentVersionName)
	return err
}

func (s *SqlDatabase) markIssuesMitigated(issueIDs []int) error {
	if len(issueIDs) == 0 {
		return nil
	}

	placeholders := make([]string, len(issueIDs))
	args := make([]interface{}, len(issueIDs))

	for i, id := range issueIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	q := `
        UPDATE IssueMatch
        SET issuematch_status = 'mitigated'
        WHERE issuematch_issue_id IN (` + strings.Join(placeholders, ",") + `)
    `
	_, err := s.db.Exec(q, args...)
	return err
}

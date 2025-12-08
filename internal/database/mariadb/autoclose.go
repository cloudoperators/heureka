// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"strings"
)

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

func (s *SqlDatabase) fetchIssuesForRun(runID int) (map[int]struct{}, error) {
	rows, err := s.db.Query(`
        SELECT scannerrunissuetracker_issue_id
        FROM ScannerRunIssueTracker
        WHERE scannerrunissuetracker_scannerrun_run_id = ?
    `, runID)

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

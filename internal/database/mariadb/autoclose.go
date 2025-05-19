// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

// autoCloseComponents is a SQL query that updates the IssueMatch table, setting the issuematch_status to 'mitigated'
// for issues that are present in the second to last completed scanner run but not in the last completed scanner run.
// It identifies these issues by comparing the component instances associated with each issue across the two most recent scanner runs.
var autoCloseComponents = `
	UPDATE IssueMatch 
	SET
		issuematch_status = 'mitigated'
	WHERE 
		issuematch_id IN (
			SELECT DISTINCT issuematch_id FROM IssueMatch WHERE issuematch_component_instance_id NOT IN (
				SELECT DISTINCT scannerruncomponentinstance_component_instance_id 
				FROM  
					ScannerRunComponentInstanceTracker
				WHERE
					scannerruncomponentinstance_scannerrun_run_id IN (
					SELECT scannerrun_run_id 
						FROM ScannerRun 
						WHERE scannerrun_run_id IN (
							SELECT scannerrun_run_id 
								FROM	(
									SELECT 
										scannerrun_run_id, 
										ROW_NUMBER() OVER (PARTITION BY scannerrun_tag ORDER BY scannerrun_run_id DESC) AS row_num
									FROM 
										ScannerRun
									WHERE 
										scannerrun_is_completed = TRUE
								) AS before_last
								WHERE row_num = 2
							)
						)
					)
			) AND
		issuematch_id NOT IN(
			SELECT DISTINCT issuematch_id FROM IssueMatch WHERE issuematch_component_instance_id NOT IN (
				SELECT DISTINCT scannerruncomponentinstance_component_instance_id 
				FROM  
					ScannerRunComponentInstanceTracker
				WHERE
					scannerruncomponentinstance_scannerrun_run_id IN (
					SELECT scannerrun_run_id 
						FROM ScannerRun 
						WHERE scannerrun_run_id IN (
							SELECT scannerrun_run_id 
								FROM	(
									SELECT 
										scannerrun_run_id, 
										ROW_NUMBER() OVER (PARTITION BY scannerrun_tag ORDER BY scannerrun_run_id DESC) AS row_num
									FROM 
										ScannerRun
									WHERE 
										scannerrun_is_completed = TRUE
								) AS last
								WHERE row_num = 1
							)
						)
					)
			)
			`

func (s *SqlDatabase) Autoclose() (bool, error) {
	var err error
	var autoclosed bool

	rows, err := s.db.Query(`
		SELECT
		 	DISTINCT scannerrun_tag AS Tag,
			COUNT(*) AS Count 
			FROM ScannerRun 
			WHERE 
				scannerrun_is_completed = TRUE
			GROUP BY scannerrun_tag`)

	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var count int
		var tag string
		err = rows.Scan(&tag, &count)

		if err != nil {
			return autoclosed, err
		}

		if count >= 2 {
			var id1, id2 int
			{
				rows, err := s.db.Query(`
					SELECT scannerrun_run_id AS ID 
					FROM ScannerRun 
					WHERE scannerrun_tag=? 
					ORDER BY scannerrun_run_id DESC LIMIT 2`, tag)
				if err != nil {
					return autoclosed, err
				}
				defer rows.Close()
				rows.Next()

				if rows.Err() != nil {
					return autoclosed, rows.Err()
				}

				err = rows.Scan(&id1)

				if err != nil {
					return autoclosed, err
				}

				rows.Next()

				if rows.Err() != nil {
					return autoclosed, rows.Err()
				}

				err = rows.Scan(&id2)

				if err != nil {
					return autoclosed, err
				}
			}

			row := s.db.QueryRow(`
				SELECT COUNT(DISTINCT scannerrunissuetracker_issue_id) 
				FROM ScannerRunIssueTracker WHERE 
					(scannerrunissuetracker_issue_id NOT IN 
						(SELECT scannerrunissuetracker_issue_id FROM ScannerRunIssueTracker WHERE scannerrunissuetracker_scannerrun_run_id = ?)) AND 
					(scannerrunissuetracker_issue_id IN 
						(SELECT scannerrunissuetracker_issue_id FROM ScannerRunIssueTracker WHERE scannerrunissuetracker_scannerrun_run_id = ?))`, id1, id2)

			var issueCount int
			err = row.Scan(&issueCount)
			if err != nil {
				return autoclosed, err
			}
			autoclosed = autoclosed || (issueCount > 0)
		}
	}

	if rows.Err() != nil {
		return autoclosed, rows.Err()
	}

	if res, err := s.db.Exec(autoCloseComponents); err != nil {
		return autoclosed, err
	} else {
		if rowsAffected, err := res.RowsAffected(); err != nil {
			return autoclosed, err
		} else if rowsAffected > 0 {
			autoclosed = true
		}
	}
	return autoclosed, nil
}

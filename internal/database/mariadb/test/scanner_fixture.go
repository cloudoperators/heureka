// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

type ScannerRunDef struct {
	Tag                  string
	IsCompleted          bool
	Timestamp            time.Time
	Issues               []string
	Components           []string
	IssueMatchComponents []string // WARNING: This needs pairs of Issue name and compoenent name
}

func (s *DatabaseSeeder) SeedScannerRuns(scannerRunDefs ...ScannerRunDef) error {
	knownIssues := make(map[string]int)
	knownComponentInstance := make(map[string]int)
	serviceCounter := 0
	componentCounter := 0
	componentVersionCounter := 0

	for _, srd := range scannerRunDefs {
		res, err := s.insertScannerRunInstance(gofakeit.UUID(), srd.Tag, srd.Timestamp, srd.Timestamp, srd.IsCompleted)
		if err != nil {
			return err

		}

		scannerRunId, err := res.LastInsertId()
		if err != nil {
			return err
		}

		for _, issue := range srd.Issues {

			if _, ok := knownIssues[issue]; !ok {
				res, err := s.insertIssue(issue)
				if err != nil {
					return err

				}
				issueId, err := res.LastInsertId()
				if err != nil {
					return err
				}

				knownIssues[issue] = int(issueId)
			}

			if err != nil {
				return err
			}

			if err := s.insertScannerRunIssueTracker(scannerRunId, knownIssues[issue]); err != nil {
				return err
			}
		}

		if len(srd.Components) > 0 {
			if err := s.insertService(fmt.Sprintf("service-%d", serviceCounter)); err != nil {
				return fmt.Errorf("InsertIntoService failed: %v", err)
			}
			serviceCounter++
			if err := s.insertComponent(fmt.Sprintf("component-%d", componentCounter)); err != nil {
				return fmt.Errorf("InsertIntoComponent failed: %v", err)
			}
			componentCounter++
			if err := s.insertComponentVersion(fmt.Sprintf("version-%d", componentVersionCounter)); err != nil {
				return fmt.Errorf("InsertIntoComponentVersion failed: %v", err)
			}
			componentVersionCounter++
			for _, component := range srd.Components {
				if _, ok := knownComponentInstance[component]; ok {
					continue
				}
				res, err = s.insertComponentInstance(component)
				if err != nil {
					return fmt.Errorf("bad things insertintocomponentinstance: %v", err)
				}
				if resId, err := res.LastInsertId(); err != nil {
					return fmt.Errorf("bad things insertintocomponentinstance get lastInsertId %v", err)
				} else {
					knownComponentInstance[component] = int(resId)
				}
			}
		}

		if len(srd.IssueMatchComponents) > 0 {
			for i, _ := range srd.IssueMatchComponents {
				if i%2 != 0 {
					continue
				}

				issueName := srd.IssueMatchComponents[i]
				componentName := srd.IssueMatchComponents[i+1]
				if err := s.insertIssueMatchComponent(knownIssues[issueName], knownComponentInstance[componentName]); err != nil {
					return fmt.Errorf("InsertIntoIssueMatchComponent failed: %v", err)
				}
			}
		}
	}
	return nil
}

func (s *DatabaseSeeder) insertScannerRunInstance(uuid string, tag string, startRun time.Time, endRun time.Time, isCompleted bool) (sql.Result, error) {
	insertScannerRun := `
		INSERT INTO ScannerRun (
			scannerrun_uuid,
			scannerrun_tag,
			scannerrun_start_run,
			scannerrun_end_run,
			scannerrun_is_completed
		) VALUES (
			?,
			?,
			?,
			?,
			?
		)
	`
	return s.db.Exec(insertScannerRun, uuid, tag, startRun, endRun, isCompleted)
}

func (s *DatabaseSeeder) insertIssue(issue string) (sql.Result, error) {
	insertIssue := `
		INSERT INTO Issue (
			issue_type,
			issue_primary_name,
			issue_description
		) VALUES (
			'Vulnerability',
			?,
			?
		)
	`
	return s.db.Exec(insertIssue, issue, issue)
}

func (s *DatabaseSeeder) insertScannerRunIssueTracker(scannerRunId int64, issueId int) error {
	insertScannerRunIssueTracker := `
		INSERT INTO ScannerRunIssueTracker (
			scannerrunissuetracker_scannerrun_run_id,
			scannerrunissuetracker_issue_id
		) VALUES (
			?,
			?
		)
	`
	_, err := s.db.Exec(insertScannerRunIssueTracker, scannerRunId, issueId)
	return err
}

func (s *DatabaseSeeder) insertService(ccrn string) error {
	insertService := `
	INSERT INTO Service (
			service_ccrn
		) VALUES (
			?
		)
	`
	_, err := s.db.Exec(insertService, ccrn)
	return err
}

func (s *DatabaseSeeder) insertComponent(ccrn string) error {
	insertComponent := `
	INSERT INTO Component (
			component_ccrn,
			component_type
		) VALUES (
			?,
			'floopy disk'
		)
	`
	_, err := s.db.Exec(insertComponent, ccrn)
	return err
}

func (s *DatabaseSeeder) insertComponentVersion(version string) error {
	insertComponentVersion := `
		INSERT INTO ComponentVersion (
			componentversion_version,
			componentversion_component_id,
			componentversion_created_by
		) VALUES (
			?,
			1,
			1
		)
	`
	_, err := s.db.Exec(insertComponentVersion, version)
	return err
}

func (s *DatabaseSeeder) insertComponentInstance(ccrn string) (sql.Result, error) {
	insertComponentInstance := `
		INSERT INTO ComponentInstance (
			componentinstance_ccrn,
			componentinstance_count,
			componentinstance_component_version_id,
			componentinstance_service_id,
			componentinstance_created_by
		) VALUES (
			?,
			1,
			1,
			1,
			1
		)
	`

	return s.db.Exec(insertComponentInstance, ccrn)
}

func (s *DatabaseSeeder) insertIssueMatchComponent(issueId int, componentInstanceId int) error {
	insertIssueMatchComponent := `
		INSERT INTO IssueMatch (
			issuematch_status,
			issuematch_rating,
			issuematch_target_remediation_date,
			issuematch_user_id,
			issuematch_issue_id,
			issuematch_component_instance_id
		) VALUES (
			'new',
			'CRITICAL',
			current_timestamp(),
			1,
			?,
			?
		)
	`
	var err error
	_, err = s.db.Exec(insertIssueMatchComponent, issueId, componentInstanceId)
	return err
}

func (s *DatabaseSeeder) SeedScannerRunInstances(uuids ...string) error {
	for _, uuid := range uuids {
		if _, err := s.insertScannerRunInstance(uuid, "tag", time.Now(), time.Now(), false); err != nil {
			return err
		}
	}
	return nil
}

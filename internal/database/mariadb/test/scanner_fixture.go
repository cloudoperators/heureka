// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/cloudoperators/heureka/internal/database/mariadb"

	"github.com/brianvoe/gofakeit/v7"
)

type IssueMatchComponent struct {
	Issue     string
	Component string
}

type Component struct {
	Name    string
	Version string
	Service string
}

type ScannerRunDef struct {
	Tag                  string
	IsCompleted          bool
	Timestamp            time.Time
	Issues               []string
	Components           []Component
	IssueMatchComponents []IssueMatchComponent
}

type ScannerRunsSeeder struct {
	serviceCounter          int
	componentCounter        int
	componentVersionCounter int

	dbSeeder *DatabaseSeeder

	knownComponentInstances map[string]int
	knownIssues             map[string]int
	knownServices           map[string]int64
	knownVersions           map[string]int64
	knownComponents         map[string]int64
}

func NewScannerRunsSeeder(dbSeeder *DatabaseSeeder) *ScannerRunsSeeder {
	return &ScannerRunsSeeder{
		dbSeeder:                dbSeeder,
		knownComponentInstances: make(map[string]int),
		knownIssues:             make(map[string]int),
		knownServices:           make(map[string]int64),
		knownVersions:           make(map[string]int64),
		knownComponents:         make(map[string]int64),
	}
}

func (srs *ScannerRunsSeeder) Seed(srd ScannerRunDef) error {
	if err := srs.processScannerRunInstance(srd); err != nil {
		return err
	}
	return nil
}

func (srs *ScannerRunsSeeder) processScannerRunInstance(srd ScannerRunDef) error {
	res, err := srs.dbSeeder.insertScannerRunInstance(gofakeit.UUID(), srd.Tag, srd.Timestamp, srd.Timestamp, srd.IsCompleted)
	if err != nil {
		return err
	}
	scannerRunId, err := res.LastInsertId()
	if err != nil {
		return err
	}
	if err := srs.processIssues(srd.Issues, scannerRunId); err != nil {
		return err
	}
	if err := srs.processComponents(srd.Components, scannerRunId); err != nil {
		return err
	}
	if err := srs.processIssueMatchComponents(srd.IssueMatchComponents); err != nil {
		return err
	}
	return nil
}

func (srs *ScannerRunsSeeder) processIssues(issues []string, scannerRunId int64) error {
	for _, issue := range issues {
		if _, ok := srs.knownIssues[issue]; !ok {
			issueId, err := srs.dbSeeder.insertIssue(issue)
			if err != nil {
				return err
			}

			srs.knownIssues[issue] = int(issueId)
		}
	}
	return nil
}

type dataFeeds struct {
	serviceId    int64
	componentId  int64
	versionId    int64
	scannerRunId int64
}

func (srs *ScannerRunsSeeder) processComponents(components []Component, scannerRunId int64) error {
	if len(components) > 0 {
		defaultFeeds, err := srs.createDefaultDataFeeds(scannerRunId)
		if err != nil {
			return err
		}

		srs.processComponentInstances(components, defaultFeeds)
		if err != nil {
			return err
		}
	}
	return nil
}

func (srs *ScannerRunsSeeder) createDefaultDataFeeds(scannerRunId int64) (*dataFeeds, error) {
	serviceId, err := srs.insertNextService()
	if err != nil {
		return nil, err
	}

	componentId, err := srs.insertNextComponent()
	if err != nil {
		return nil, err
	}

	versionId, err := srs.insertNextComponentVersion(componentId)
	if err != nil {
		return nil, err
	}

	return &dataFeeds{serviceId: serviceId, componentId: componentId, versionId: versionId, scannerRunId: scannerRunId}, nil
}

func (srs *ScannerRunsSeeder) insertNextService() (int64, error) {
	serviceCcrn := fmt.Sprintf("service-%d", srs.serviceCounter)
	srs.serviceCounter++
	return srs.insertService(serviceCcrn)
}

func (srs *ScannerRunsSeeder) insertNextComponent() (int64, error) {
	componentCcrn := fmt.Sprintf("component-%d", srs.componentCounter)
	srs.componentCounter++
	return srs.insertComponent(componentCcrn)
}

func (srs *ScannerRunsSeeder) insertNextComponentVersion(componentId int64) (int64, error) {
	versionName := fmt.Sprintf("version-%d", srs.componentVersionCounter)
	srs.componentVersionCounter++
	return srs.insertComponentVersion(versionName, componentId)
}

func (srs *ScannerRunsSeeder) insertService(serviceCcrn string) (int64, error) {
	if serviceId, ok := srs.knownServices[serviceCcrn]; ok {
		return serviceId, nil
	}
	serviceId, err := srs.dbSeeder.insertService(serviceCcrn)
	if err != nil {
		return 0, err
	}
	srs.knownServices[serviceCcrn] = serviceId
	return serviceId, nil
}

func (srs *ScannerRunsSeeder) insertComponent(componentCcrn string) (int64, error) {
	componentId, err := srs.dbSeeder.getOrCreateComponent(componentCcrn)
	if componentId, ok := srs.knownComponents[componentCcrn]; ok {
		return componentId, nil
	}
	if err != nil {
		return 0, err
	}
	srs.knownComponents[componentCcrn] = componentId
	return componentId, nil
}

func (srs *ScannerRunsSeeder) insertComponentVersion(versionName string, componentId int64) (int64, error) {
	if versionId, ok := srs.knownVersions[versionName]; ok {
		return versionId, nil
	}
	versionId, err := srs.dbSeeder.insertComponentVersion(versionName, componentId)
	if err != nil {
		return 0, err
	}
	srs.knownVersions[versionName] = versionId
	return versionId, nil
}

func (srs *ScannerRunsSeeder) processComponentInstances(components []Component, defaultFeeds *dataFeeds) error {
	for _, component := range components {
		if err := srs.processComponentInstance(component, defaultFeeds); err != nil {
			return err
		}
	}
	return nil
}

func (srs *ScannerRunsSeeder) processComponentInstance(component Component, defaultFeeds *dataFeeds) error {
	if componentId, ok := srs.knownComponentInstances[component.Name]; ok {
		return srs.dbSeeder.insertScannerRunComponentInstanceTracker(defaultFeeds.scannerRunId, componentId)
	}
	return srs.insertNewComponentInstance(component, defaultFeeds)
}

func (srs *ScannerRunsSeeder) insertNewComponentInstance(component Component, defaultFeeds *dataFeeds) error {
	var err error
	versionId := defaultFeeds.versionId
	serviceId := defaultFeeds.serviceId
	if component.Version != "" {
		versionId, err = srs.insertComponentVersion(component.Version, defaultFeeds.componentId)
		if err != nil {
			return err
		}
	}
	if component.Service != "" {
		serviceId, err = srs.insertService(component.Service)
		if err != nil {
			return err
		}
	}
	res, err := srs.dbSeeder.insertComponentInstance(component.Name, versionId, serviceId)
	if err != nil {
		return fmt.Errorf("bad things insertintocomponentinstance: %v", err)
	}
	resId, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("bad things insertintocomponentinstance get lastInsertId %v", err)
	}
	componentId := int(resId)
	srs.knownComponentInstances[component.Name] = componentId
	if err := srs.dbSeeder.insertScannerRunComponentInstanceTracker(defaultFeeds.scannerRunId, componentId); err != nil {
		return err
	}
	return nil
}

func (srs *ScannerRunsSeeder) processIssueMatchComponents(imc []IssueMatchComponent) error {
	for _, issueMatch := range imc {
		issue := srs.knownIssues[issueMatch.Issue]
		componentId := srs.knownComponentInstances[issueMatch.Component]
		if err := srs.dbSeeder.insertIssueMatchComponent(issue, componentId); err != nil {
			return fmt.Errorf("InsertIntoIssueMatchComponent failed: %v", err)
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

func (s *DatabaseSeeder) insertIssue(issue string) (int64, error) {
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
	res, err := s.db.Exec(insertIssue, issue, issue)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *DatabaseSeeder) insertScannerRunComponentInstanceTracker(scannerRunId int64, componentId int) error {
	insertScannerRunComponentInstanceTracker := `
		INSERT INTO ScannerRunComponentInstanceTracker (
			scannerruncomponentinstancetracker_scannerrun_run_id,
			scannerruncomponentinstancetracker_component_instance_id
		) VALUES (
			?,
			?
		)
	`
	_, err := s.db.Exec(insertScannerRunComponentInstanceTracker, scannerRunId, componentId)
	return err
}

func (s *DatabaseSeeder) insertService(ccrn string) (int64, error) {
	insertService := `
	INSERT INTO Service (
			service_ccrn
		) VALUES (
			?
		)
	`
	res, err := s.db.Exec(insertService, ccrn)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *DatabaseSeeder) getOrCreateComponent(ccrn string) (int64, error) {
	query := `
		INSERT INTO Component (component_ccrn, component_type)
		VALUES (?, 'floopy disk')
		ON DUPLICATE KEY UPDATE component_id = LAST_INSERT_ID(component_id)
	`

	res, err := s.db.Exec(query, ccrn)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	return id, err
}

func (s *DatabaseSeeder) insertComponentVersion(version string, componentId int64) (int64, error) {
	insertComponentVersion := `
		INSERT INTO ComponentVersion (
			componentversion_version,
			componentversion_component_id,
			componentversion_created_by
		) VALUES (
			?,
			?,
			1
		)
	`
	res, err := s.db.Exec(insertComponentVersion, version, componentId)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *DatabaseSeeder) insertComponentInstance(ccrn string, versionId int64, serviceId int64) (sql.Result, error) {
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
			?,
			?,
			1
		)
	`

	return s.db.Exec(insertComponentInstance, ccrn, versionId, serviceId)
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

func (s *DatabaseSeeder) CleanupScannerRuns() error {
	cleanupQuery := `
	DELETE FROM ScannerRunComponentInstanceTracker;
	DELETE FROM ScannerRun;
	DELETE FROM IssueMatch;
	DELETE FROM Issue;
	DELETE FROM ComponentInstance;
	DELETE FROM Service;
	DELETE FROM ComponentVersion;
	DELETE FROM Component;
	`
	var err error
	_, err = s.db.Exec(cleanupQuery)
	return err
}

func (s *DatabaseSeeder) FetchPatchesByComponentInstanceCCRN(
	ccrn string,
) ([]mariadb.PatchRow, error) {
	query := `
        SELECT
            p.patch_id,
            p.patch_service_id,
            p.patch_service_name,
            p.patch_component_version_id,
            p.patch_component_version_name,
            p.patch_created_at
        FROM Patch p
        INNER JOIN ComponentInstance ci
            ON p.patch_service_id = ci.componentinstance_service_id
            AND p.patch_component_version_id = ci.componentinstance_component_version_id
        WHERE ci.componentinstance_ccrn = ?
    `

	rows, err := s.db.Query(query, ccrn)
	if err != nil {
		return nil, fmt.Errorf("failed to query patches by ccrn: %w", err)
	}
	defer rows.Close()

	var patches []mariadb.PatchRow

	for rows.Next() {
		var p mariadb.PatchRow
		if err := rows.Scan(
			&p.Id,
			&p.ServiceId,
			&p.ServiceName,
			&p.ComponentVersionId,
			&p.ComponentVersionName,
			&p.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan patch row: %w", err)
		}
		patches = append(patches, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	// Optional: return empty slice instead of nil
	if patches == nil {
		return []mariadb.PatchRow{}, nil
	}

	return patches, nil
}

func (s *DatabaseSeeder) FetchAllNamesOfDeletedIssueMatches() ([]string, error) {
	query := `
        SELECT
            i.issue_primary_name
        FROM IssueMatch im
        INNER JOIN Issue i
            ON im.issuematch_issue_id = i.issue_id
        WHERE im.issuematch_deleted_at IS NOT NULL
    `

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query issue matches by id: %w", err)
	}
	defer rows.Close()

	var issueNames []string

	for rows.Next() {
		var in string
		if err := rows.Scan(
			&in,
		); err != nil {
			return nil, fmt.Errorf("failed to scan issue match row: %w", err)
		}
		issueNames = append(issueNames, in)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	// Optional: return empty slice instead of nil
	if issueNames == nil {
		return []string{}, nil
	}

	return issueNames, nil
}

func (s *DatabaseSeeder) FetchAllNamesOfDeletedVersions() ([]string, error) {
	query := `
        SELECT
            cv.componentversion_version
        FROM ComponentVersion cv
        WHERE cv.componentversion_deleted_at IS NOT NULL
    `

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query versions: %w", err)
	}
	defer rows.Close()

	var versions []string

	for rows.Next() {
		var in string
		if err := rows.Scan(
			&in,
		); err != nil {
			return nil, fmt.Errorf("failed to scan component version row: %w", err)
		}
		versions = append(versions, in)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	// Optional: return empty slice instead of nil
	if versions == nil {
		return []string{}, nil
	}

	return versions, nil
}

func (s *DatabaseSeeder) GetCountOfPatches() (int64, error) {
	const query = `
		SELECT COUNT(*)
		FROM Patch
	`

	var count int64
	err := s.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *DatabaseSeeder) FetchComponentInstanceByCCRN(ccrn string) (*mariadb.ComponentInstanceRow, error) {
	query := `
        SELECT
            ci.componentinstance_id,
            ci.componentinstance_deleted_at
        FROM ComponentInstance ci
        WHERE ci.componentinstance_ccrn = ?
        LIMIT 1`

	var ci mariadb.ComponentInstanceRow
	err := s.db.QueryRow(query, ccrn).Scan(
		&ci.Id,
		&ci.DeletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("component instance not found for component: '%s' %w", ccrn, err)
		}
		return nil, fmt.Errorf("failed to fetch component instance by ccrn: %w", err)
	}

	return &ci, nil
}

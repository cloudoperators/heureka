// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/util"

	"github.com/brianvoe/gofakeit/v7"
)

type DbSeeds struct {
	IssueMatchComponents   []IssueMatchComponent
	Issues                 []string
	Components             []ComponentData
	ComponentVersionIssues []ComponentVersionIssue
}

type IssueMatchComponent struct {
	Issue             string
	ComponentInstance string
}

type ComponentData struct {
	Name    string
	Version string
	Service string
}

type ComponentVersionIssue struct {
	Issue            string
	ComponentVersion string
}

type ScannerRunDef struct {
	Tag                        string
	IsCompleted                bool
	Timestamp                  time.Time
	DetectedComponentInstances []string
}

type ScannerRunsSeeder struct {
	serviceCounter          int
	componentCounter        int
	componentVersionCounter int

	dbSeeder *DatabaseSeeder

	knownComponentInstances     map[string]int64
	knownIssues                 map[string]int64
	knownServices               map[string]int64
	knownVersions               map[string]int64
	knownComponents             map[string]int64
	knownComponentInstancesData map[string]componentInstanceData
	knownIssuesMatchComponents  map[string][]int64
	knownComponentVersionIssues map[ComponentVersionIssue]int64
}

type componentInstanceData struct {
	componentVersionId int64
	serviceId          int64
	issueMatches       []int64
}

func NewScannerRunsSeeder(dbSeeder *DatabaseSeeder) *ScannerRunsSeeder {
	return &ScannerRunsSeeder{
		dbSeeder:                    dbSeeder,
		knownComponentInstances:     make(map[string]int64),
		knownIssues:                 make(map[string]int64),
		knownServices:               make(map[string]int64),
		knownVersions:               make(map[string]int64),
		knownComponents:             make(map[string]int64),
		knownComponentInstancesData: make(map[string]componentInstanceData),
		knownIssuesMatchComponents:  make(map[string][]int64),
		knownComponentVersionIssues: make(map[ComponentVersionIssue]int64)
	}
}

func (srs *ScannerRunsSeeder) SeedDbSeeds(dbs DbSeeds) error {
	if err := srs.processDbSeeds(dbs); err != nil {
		return err
	}
	return nil
}

func (srs *ScannerRunsSeeder) SeedScannerRun(srd ScannerRunDef) error {
	if err := srs.processScannerRunInstance(srd); err != nil {
		return err
	}
	return nil
}

func (srs *ScannerRunsSeeder) processDbSeeds(dbs DbSeeds) error {
	if err := srs.seedIssues(dbs.Issues); err != nil {
		return err
	}
	if err := srs.seedComponentsData(dbs.Components); err != nil {
		return err
	}
	if err := srs.storeIssueMatchComponents(dbs.IssueMatchComponents); err != nil {
		return err
	}
	if err := srs.seedComponentVersionIssues(dbs.ComponentVersionIssues); err != nil {
		return err
	}
	return nil
}

func (srs *ScannerRunsSeeder) seedIssues(issues []string) error {
	for _, issue := range issues {
		if _, ok := srs.knownIssues[issue]; ok {
			return fmt.Errorf("Trying to seed Issue that already exists: '%s'.", issue)
		}
		issueId, err := srs.dbSeeder.insertIssue(issue)
		if err != nil {
			return err

		}

		srs.knownIssues[issue] = issueId
	}
	return nil
}

func (srs *ScannerRunsSeeder) seedComponentVersionIssues(componentVersionIssues []ComponentVersionIssue) error {
	for _, cvi := range componentVersionIssues {
		if _, ok := srs.knownComponentVersionIssues[cvi]; ok {
			return fmt.Errorf("Trying to seed ComponentVersionIssue that already exists: 'i: %s, cv: %s'.", cvi.Issue, cvi.ComponentVersion)
		}
		issueId, ok := srs.knownIssues[cvi.Issue]
		if !ok {
			return fmt.Errorf("Trying to seed ComponentVersionIssue but issue does not exist: 'i: %s, cv: %s'.", cvi.Issue, cvi.ComponentVersion)
		}
		componentVersionId, ok := srs.knownVersions[cvi.ComponentVersion]
		if !ok {
			return fmt.Errorf("Trying to seed ComponentVersionIssue but component version does not exist: 'i: %s, cv: %s'.", cvi.Issue, cvi.ComponentVersion)
		}
		cviId, err := srs.dbSeeder.insertComponentVersionIssue(componentVersionId, issueId)
		if err != nil {
			return err

		}

		srs.knownComponentVersionIssues[cvi] = cviId
	}
	return nil
}

func (srs *ScannerRunsSeeder) seedComponentsData(components []ComponentData) error {
	for _, component := range components {
		_, ok := srs.knownComponentInstances[component.Name]
		if ok {
			return fmt.Errorf("ComponentInstance: '%s' already exists.", component.Name)
		}
		_, err := srs.seedComponentData(component)
		if err != nil {
			return err
		}
	}
	return nil
}

func (srs *ScannerRunsSeeder) seedComponentData(component ComponentData) (componentInstanceData, error) {
	componentId, err := srs.seedOrGetComponent(component.Name)
	if err != nil {
		return componentInstanceData{}, err
	}

	componentVersionId, err := srs.seedOrGetComponentVersion(component.Version, componentId)
	if err != nil {
		return componentInstanceData{}, err
	}

	serviceId, err := srs.seedOrGetService(component.Service)
	if err != nil {
		return componentInstanceData{}, err
	}

	ciData := componentInstanceData{componentVersionId: componentVersionId, serviceId: serviceId}
	srs.knownComponentInstancesData[component.Name] = ciData
	return ciData, err
}

func (srs *ScannerRunsSeeder) seedComponentInstance(componentInstanceName string) (int64, error) {
	ciData, ok := srs.knownComponentInstancesData[componentInstanceName]
	if !ok {
		var err error
		ciData, err = srs.seedComponentData(ComponentData{Name: componentInstanceName})
		if err != nil {
			return 0, err
		}
	}
	res, err := srs.dbSeeder.insertComponentInstance(componentInstanceName, ciData.componentVersionId, ciData.serviceId)
	if err != nil {
		return 0, err
	}
	componentInstanceId, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Insert IssueMatchComponent for created ComponentInstance
	for _, issueId := range srs.knownIssuesMatchComponents[componentInstanceName] {
		if err := srs.dbSeeder.insertIssueMatchComponent(issueId, componentInstanceId); err != nil {
			return 0, fmt.Errorf("InsertIntoIssueMatchComponent failed: %v", err)
		}
	}

	srs.knownComponentInstances[componentInstanceName] = componentInstanceId
	return componentInstanceId, nil
}

func (srs *ScannerRunsSeeder) seedOrGetComponent(componentCcrn string) (int64, error) {
	if componentCcrn != "" {
		if componentId, ok := srs.knownComponents[componentCcrn]; ok {
			return componentId, nil
		}
		return srs.insertComponent(componentCcrn)
	}
	return srs.insertNextComponent()
}

func (srs *ScannerRunsSeeder) insertNextComponent() (int64, error) {
	componentCcrn := fmt.Sprintf("component-%d", srs.componentCounter)
	srs.componentCounter++
	if _, ok := srs.knownComponents[componentCcrn]; ok {
		return 0, fmt.Errorf("Trying to insert Component which already exists: '%s'", componentCcrn)
	}
	return srs.insertComponent(componentCcrn)
}

func (srs *ScannerRunsSeeder) insertComponent(componentCcrn string) (int64, error) {
	componentId, err := srs.dbSeeder.insertComponent(componentCcrn)
	if err != nil {
		return 0, err
	}
	srs.knownComponents[componentCcrn] = componentId
	return componentId, nil
}

func (srs *ScannerRunsSeeder) seedOrGetService(serviceCcrn string) (int64, error) {
	if serviceCcrn != "" {
		if serviceId, ok := srs.knownServices[serviceCcrn]; ok {
			return serviceId, nil
		}
		return srs.insertService(serviceCcrn)
	}
	return srs.insertNextService()
}

func (srs *ScannerRunsSeeder) insertNextService() (int64, error) {
	serviceCcrn := fmt.Sprintf("service-%d", srs.serviceCounter)
	srs.serviceCounter++
	if _, ok := srs.knownServices[serviceCcrn]; ok {
		return 0, fmt.Errorf("Trying to insert Service which already exists: '%s'", serviceCcrn)
	}
	return srs.insertService(serviceCcrn)
}

func (srs *ScannerRunsSeeder) insertService(serviceCcrn string) (int64, error) {
	serviceId, err := srs.dbSeeder.insertService(serviceCcrn)
	if err != nil {
		return 0, err
	}
	srs.knownServices[serviceCcrn] = serviceId
	return serviceId, nil
}

func (srs *ScannerRunsSeeder) seedOrGetComponentVersion(versionName string, componentId int64) (int64, error) {
	if versionName != "" {
		if componentVersionId, ok := srs.knownVersions[versionName]; ok {
			return componentVersionId, nil
		}
		return srs.insertComponentVersion(versionName, componentId)
	}
	return srs.insertNextComponentVersion(componentId)
}

func (srs *ScannerRunsSeeder) insertNextComponentVersion(componentId int64) (int64, error) {
	versionName := fmt.Sprintf("version-%d", srs.componentVersionCounter)
	srs.componentVersionCounter++
	if _, ok := srs.knownVersions[versionName]; ok {
		return 0, fmt.Errorf("Trying to insert ComponentVersion which already exists: '%s'", versionName)
	}
	return srs.insertComponentVersion(versionName, componentId)
}

func (srs *ScannerRunsSeeder) insertComponentVersion(versionName string, componentId int64) (int64, error) {
	versionId, err := srs.dbSeeder.insertComponentVersion(versionName, componentId)
	if err != nil {
		return 0, err
	}
	srs.knownVersions[versionName] = versionId
	return versionId, nil
}

func (srs *ScannerRunsSeeder) storeIssueMatchComponents(issueMatchComponents []IssueMatchComponent) error {
	for _, imc := range issueMatchComponents {
		issue, ok := srs.knownIssues[imc.Issue]
		if !ok {
			return fmt.Errorf("Issue from IssueMatchComponent not found in known Issues: '%s'", imc.Issue)
		}
		srs.knownIssuesMatchComponents[imc.ComponentInstance] = append(srs.knownIssuesMatchComponents[imc.ComponentInstance], issue)
	}
	return nil
}

// -------------------------------------------

func (srs *ScannerRunsSeeder) processScannerRunInstance(srd ScannerRunDef) error {
	res, err := srs.dbSeeder.insertScannerRunInstance(gofakeit.UUID(), srd.Tag, srd.Timestamp, srd.Timestamp, srd.IsCompleted)
	if err != nil {
		return err

	}
	scannerRunId, err := res.LastInsertId()
	if err != nil {
		return err
	}

	if err := srs.processDetectedComponentInstances(srd.DetectedComponentInstances, scannerRunId); err != nil {
		return err
	}
	return nil
}

func (srs *ScannerRunsSeeder) processDetectedComponentInstances(detectedComponentInstances []string, scannerRunId int64) error {
	for _, detectedCI := range detectedComponentInstances {
		componentInstanceId, ok := srs.knownComponentInstances[detectedCI]
		if !ok {
			var err error
			componentInstanceId, err = srs.seedComponentInstance(detectedCI)
			if err != nil {
				return err
			}
		}
		err := srs.dbSeeder.insertScannerRunComponentInstanceTracker(scannerRunId, componentInstanceId)
		if err != nil {
			return err
		}
	}
	return nil
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

func (s *DatabaseSeeder) insertScannerRunComponentInstanceTracker(scannerRunId int64, componentId int64) error {
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

func (s *DatabaseSeeder) insertComponent(ccrn string) (int64, error) {
	query := `
		INSERT INTO Component (component_ccrn, component_type)
		VALUES (?, 'floopy disk')
	`

	res, err := s.db.Exec(query, ccrn)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
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
			?
		)
	`
	res, err := s.db.Exec(insertComponentVersion, version, componentId, util.SystemUserId)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *DatabaseSeeder) insertComponentVersionIssue(versionId int64, issueId int64) (int64, error) {
	insertComponentVersionIssue := `
		INSERT INTO ComponentVersionIssue (
			componentversionissue_component_version_id,
			componentversionissue_issue_id
		) VALUES (
			?,
			?
		)
	`
	res, err := s.db.Exec(insertComponentVersionIssue, versionId, issueId)
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
			?
		)
	`

	return s.db.Exec(insertComponentInstance, ccrn, versionId, serviceId, util.SystemUserId)
}

func (s *DatabaseSeeder) insertIssueMatchComponent(issueId int64, componentInstanceId int64) error {
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
			?,
			?,
			?
		)
	`
	var err error
	_, err = s.db.Exec(insertIssueMatchComponent, util.SystemUserId, issueId, componentInstanceId)
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
	DELETE FROM ComponentVersionIssue;
	DELETE FROM Patch;
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

func (s *DatabaseSeeder) FetchAllNamesOfComponentVersionIssues() ([]ComponentVersionIssue, error) {
	query := `
        SELECT
			issue_primary_name,
            componentversion_version
        FROM ComponentVersionIssue cvi
        INNER JOIN ComponentVersion cv ON cvi.componentversionissue_component_version_id = cv.componentversion_id
        INNER JOIN Issue i ON cvi.componentversionissue_issue_id = i.issue_id
    `

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query versions: %w", err)
	}
	defer rows.Close()

	var result []ComponentVersionIssue

	for rows.Next() {
		var i, cv string
		if err := rows.Scan(
			&i, &cv,
		); err != nil {
			return nil, fmt.Errorf("failed to scan component version issue row: %w", err)
		}
		result = append(result, ComponentVersionIssue{Issue: i, ComponentVersion: cv})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	// Optional: return empty slice instead of nil
	if result == nil {
		return []ComponentVersionIssue{}, nil
	}

	return result, nil
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

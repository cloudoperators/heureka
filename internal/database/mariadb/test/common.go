// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/entity"
	. "github.com/onsi/gomega"
)

// Temporary used until order is used in all entities
func TestPaginationOfListWithOrder[F entity.HeurekaFilter, E entity.HeurekaEntity](
	listFunction func(*F, []entity.Order) ([]E, error),
	filterFunction func(*int, *int64, *string) *F,
	order []entity.Order,
	getAfterFunction func([]E) string,
	elementCount int,
	pageSize int,
) {
	quotient, remainder := elementCount/pageSize, elementCount%pageSize
	expectedPages := quotient
	if remainder > 0 {
		expectedPages = expectedPages + 1
	}

	var after *int64
	var afterS string
	for i := expectedPages; i > 0; i-- {
		entries, err := listFunction(filterFunction(&pageSize, after, &afterS), order)

		Expect(err).To(BeNil())

		if i == 1 && remainder > 0 {
			Expect(len(entries)).To(BeEquivalentTo(remainder), "on the last page we expect")
		} else {
			if pageSize > elementCount {
				Expect(len(entries)).To(BeEquivalentTo(elementCount), "on a page with a higher pageSize then element count we expect")
			} else {
				Expect(len(entries)).To(BeEquivalentTo(pageSize), "on a normal page we expect the element count to be equal to the page size")

			}
		}
		afterS = getAfterFunction(entries)
	}
}

func TestPaginationOfList[F entity.HeurekaFilter, E entity.HeurekaEntity](
	listFunction func(*F) ([]E, error),
	filterFunction func(*int, *int64) *F,
	getAfterFunction func([]E) *int64,
	elementCount int,
	pageSize int,
) {
	quotient, remainder := elementCount/pageSize, elementCount%pageSize
	expectedPages := quotient
	if remainder > 0 {
		expectedPages = expectedPages + 1
	}

	var after *int64
	for i := expectedPages; i > 0; i-- {
		entries, err := listFunction(filterFunction(&pageSize, after))

		Expect(err).To(BeNil())

		if i == 1 && remainder > 0 {
			Expect(len(entries)).To(BeEquivalentTo(remainder), "on the last page we expect")
		} else {
			if pageSize > elementCount {
				Expect(len(entries)).To(BeEquivalentTo(elementCount), "on a page with a higher pageSize then element count we expect")
			} else {
				Expect(len(entries)).To(BeEquivalentTo(pageSize), "on a normal page we expect the element count to be equal to the page size")

			}
		}
		after = getAfterFunction(entries)

	}
}

// DB stores rating as enum
// entity.Severity.Score is based on CVSS vector and has a range between x and y
// This means a rating "Low" can have a Score 3.1, 3.3, ...
// Ordering is done based on enum on DB layer, so Score can't be used for checking order
// and needs a numerical translation
func SeverityToNumerical(s string) int {
	rating := map[string]int{
		"None":     0,
		"Low":      1,
		"Medium":   2,
		"High":     3,
		"Critical": 4,
	}
	if val, ok := rating[s]; ok {
		return val
	} else {
		return -1
	}
}

// getTestDataPath returns the path to the test data directory relative to the calling file
func GetTestDataPath(path string) string {
	// Get the current file path
	_, currentFilename, _, _ := runtime.Caller(1)
	// Get the directory containing the current file
	currentDir := filepath.Dir(currentFilename)
	// Return path to test data directory (adjust the relative path as needed)
	return filepath.Join(currentDir, path)
}

// LoadIssueMatches loads issue matches from JSON file
func LoadIssueMatches(filename string) ([]mariadb.IssueMatchRow, error) {
	// Read JSON file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	// Parse JSON into temporary struct that matches the JSON format
	type tempIssueMatch struct {
		Status                string    `json:"status"`
		Rating                string    `json:"rating"`
		Vector                string    `json:"vector"`
		UserID                int64     `json:"user_id"`
		ComponentInstanceID   int64     `json:"component_instance_id"`
		IssueID               int64     `json:"issue_id"`
		TargetRemediationDate time.Time `json:"target_remediation_date"`
	}
	var tempMatches []tempIssueMatch
	if err := json.Unmarshal(data, &tempMatches); err != nil {
		return nil, err
	}
	// Convert to IssueMatchRow format
	matches := make([]mariadb.IssueMatchRow, len(tempMatches))
	for i, tm := range tempMatches {
		matches[i] = mariadb.IssueMatchRow{
			Status:                sql.NullString{String: tm.Status, Valid: true},
			Rating:                sql.NullString{String: tm.Rating, Valid: true},
			Vector:                sql.NullString{String: tm.Vector, Valid: true},
			UserId:                sql.NullInt64{Int64: tm.UserID, Valid: true},
			ComponentInstanceId:   sql.NullInt64{Int64: tm.ComponentInstanceID, Valid: true},
			IssueId:               sql.NullInt64{Int64: tm.IssueID, Valid: true},
			TargetRemediationDate: sql.NullTime{Time: tm.TargetRemediationDate, Valid: true},
		}
	}
	return matches, nil
}

// LoadIssues loads issues from JSON file
func LoadIssues(filename string) ([]mariadb.IssueRow, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	type tempIssue struct {
		Type        string `json:"type"`
		PrimaryName string `json:"primary_name"`
		Description string `json:"description"`
	}
	var tempIssues []tempIssue
	if err := json.Unmarshal(data, &tempIssues); err != nil {
		return nil, err
	}
	issues := make([]mariadb.IssueRow, len(tempIssues))
	for i, ti := range tempIssues {
		issues[i] = mariadb.IssueRow{
			Type:        sql.NullString{String: ti.Type, Valid: true},
			PrimaryName: sql.NullString{String: ti.PrimaryName, Valid: true},
			Description: sql.NullString{String: ti.Description, Valid: true},
		}
	}
	return issues, nil
}

// LoadComponentInstances loads component instances from JSON file
func LoadComponentInstances(filename string) ([]mariadb.ComponentInstanceRow, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	type tempComponentInstance struct {
		CCRN               string `json:"ccrn"`
		Count              int16  `json:"count"`
		ComponentVersionID int64  `json:"component_version_id"`
		ServiceID          int64  `json:"service_id"`
	}
	var tempComponents []tempComponentInstance
	if err := json.Unmarshal(data, &tempComponents); err != nil {
		return nil, err
	}
	components := make([]mariadb.ComponentInstanceRow, len(tempComponents))
	for i, tc := range tempComponents {
		components[i] = mariadb.ComponentInstanceRow{
			CCRN:               sql.NullString{String: tc.CCRN, Valid: true},
			Count:              sql.NullInt16{Int16: tc.Count, Valid: true},
			ComponentVersionId: sql.NullInt64{Int64: tc.ComponentVersionID, Valid: true},
			ServiceId:          sql.NullInt64{Int64: tc.ServiceID, Valid: true},
		}
	}
	return components, nil
}

func LoadIssueVariants(filename string) ([]mariadb.IssueVariantRow, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	type tempIssueVariant struct {
		IssueId           int64  `json:"issue_id"`
		IssueRepositoryId int64  `json:"issue_repository_id"`
		Rating            string `json:"rating"`
		SecondaryName     string `json:"secondary_name"`
		Description       string `json:"description"`
	}
	var tempIssueVariants []tempIssueVariant
	if err := json.Unmarshal(data, &tempIssueVariants); err != nil {
		return nil, err
	}
	issueVariants := make([]mariadb.IssueVariantRow, len(tempIssueVariants))
	for i, tiv := range tempIssueVariants {
		issueVariants[i] = mariadb.IssueVariantRow{
			IssueId:           sql.NullInt64{Int64: tiv.IssueId, Valid: true},
			IssueRepositoryId: sql.NullInt64{Int64: tiv.IssueRepositoryId, Valid: true},
			Rating:            sql.NullString{String: tiv.Rating, Valid: true},
			SecondaryName:     sql.NullString{String: tiv.SecondaryName, Valid: true},
			Description:       sql.NullString{String: tiv.Description, Valid: true},
		}
	}
	return issueVariants, nil
}

func LoadComponentVersionIssues(filename string) ([]mariadb.ComponentVersionIssueRow, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	type tempComponentVersionIssue struct {
		IssueId            int64 `json:"issue_id"`
		ComponentVersionId int64 `json:"component_version_id"`
	}
	var tempComponentVersionIssues []tempComponentVersionIssue
	if err := json.Unmarshal(data, &tempComponentVersionIssues); err != nil {
		return nil, err
	}
	componentVersionIssues := make([]mariadb.ComponentVersionIssueRow, len(tempComponentVersionIssues))
	for i, tcvi := range tempComponentVersionIssues {
		componentVersionIssues[i] = mariadb.ComponentVersionIssueRow{
			IssueId:            sql.NullInt64{Int64: tcvi.IssueId, Valid: true},
			ComponentVersionId: sql.NullInt64{Int64: tcvi.ComponentVersionId, Valid: true},
		}
	}
	return componentVersionIssues, nil
}

func LoadServiceIssueCounts(filename string) (map[string]entity.IssueSeverityCounts, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	type tempIssueCount struct {
		Critical  int64 `json:"critical"`
		High      int64 `json:"high"`
		Medium    int64 `json:"medium"`
		Low       int64 `json:"low"`
		None      int64 `json:"none"`
		Total     int64 `json:"total"`
		ServiceId int64 `json:"service_id"`
	}
	var tempIssueCounts []tempIssueCount
	if err := json.Unmarshal(data, &tempIssueCounts); err != nil {
		return nil, err
	}
	issueCounts := make(map[string]entity.IssueSeverityCounts, len(tempIssueCounts))
	for _, tic := range tempIssueCounts {
		serviceId := fmt.Sprintf("%d", tic.ServiceId)
		issueCounts[serviceId] = entity.IssueSeverityCounts{
			Critical: tic.Critical,
			High:     tic.High,
			Medium:   tic.Medium,
			Low:      tic.Low,
			None:     tic.None,
			Total:    tic.Total,
		}
	}
	return issueCounts, nil
}

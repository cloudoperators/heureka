// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"
	"strconv"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)



// Modified issue methods that use the new error system

// GetIssue fetches a single issue by ID
func (s *SqlDatabase) GetIssue(id int64) (*entity.Issue, error) {
	l := logrus.WithFields(logrus.Fields{
		"id":    id,
		"event": "database.GetIssue",
	})

	// Get issue by ID
	lo := entity.IssueListOptions{
		ListOptions: *entity.NewListOptions(),
	}
	issues, err := s.GetIssues(&entity.IssueFilter{Id: []*int64{&id}}, []entity.Order{})

	if err != nil {
		return nil, err
	}

	if len(issues) == 0 {
		return nil, fmt.Errorf("record not found")
	}

	return issues[0].Issue, nil
}

// CreateIssue creates a new issue
func (s *SqlDatabase) CreateIssue(issue *entity.Issue) (*entity.Issue, error) {
	l := logrus.WithFields(logrus.Fields{
		"issue": issue,
		"event": "database.CreateIssue",
	})

	query := `
		INSERT INTO Issue (
			issue_primary_name,
			issue_type,
			issue_description,
			issue_created_by,
			issue_updated_by
		) VALUES (
			:issue_primary_name,
			:issue_type,
			:issue_description,
			:issue_created_by,
			:issue_updated_by
		)
	`

	issueRow := IssueRow{}
	issueRow.FromIssue(issue)

	id, err := performInsert(s, query, issueRow, l)

	if err != nil {
		return nil, err
	}

	issue.Id = id

	return issue, nil
}

// UpdateIssue updates an existing issue
func (s *SqlDatabase) UpdateIssue(issue *entity.Issue) error {
	l := logrus.WithFields(logrus.Fields{
		"issue": issue,
		"event": "database.UpdateIssue",
	})

	// First check if issue exists
	_, err := s.GetIssue(issue.Id)
	if err != nil {
		return err // Pass through the raw error
	}

	baseQuery := `
		UPDATE Issue SET
		%s
		WHERE issue_id = :issue_id
	`

	updateFields := s.getIssueUpdateFields(issue)

	query := fmt.Sprintf(baseQuery, updateFields)

	issueRow := IssueRow{}
	issueRow.FromIssue(issue)

	_, err = performExec(s, query, issueRow, l)

	return err // Return raw error
}

// DeleteIssue soft-deletes an issue by ID
func (s *SqlDatabase) DeleteIssue(id int64, userId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"id":    id,
		"event": "database.DeleteIssue",
	})

	// First check if issue exists
	_, err := s.GetIssue(id)
	if err != nil {
		return err // Pass through the raw error
	}

	query := `
		UPDATE Issue SET
		issue_deleted_at = NOW(),
		issue_updated_by = :userId
		WHERE issue_id = :id
	`

	args := map[string]interface{}{
		"userId": userId,
		"id":     id,
	}

	_, err = performExec(s, query, args, l)

	return err // Return raw error
}

// AddComponentVersionToIssue associates a component version with an issue
func (s *SqlDatabase) AddComponentVersionToIssue(issueId int64, componentVersionId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"issueId":            issueId,
		"componentVersionId": componentVersionId,
		"event":              "database.AddComponentVersionToIssue",
	})

	// Check if issue exists
	_, err := s.GetIssue(issueId)
	if err != nil {
		return err // Pass through the raw error
	}

	query := `
		INSERT INTO ComponentVersionIssue (
			componentversionissue_issue_id,
			componentversionissue_component_version_id
		) VALUES (
			:issue_id,
			:component_version_id
		)
	`

	args := map[string]interface{}{
		"issue_id":             issueId,
		"component_version_id": componentVersionId,
	}

	_, err = performExec(s, query, args, l)

	return err // Return raw error
}

// RemoveComponentVersionFromIssue removes association between component version and issue
func (s *SqlDatabase) RemoveComponentVersionFromIssue(issueId int64, componentVersionId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"issueId":            issueId,
		"componentVersionId": componentVersionId,
		"event":              "database.RemoveComponentVersionFromIssue",
	})

	// Check if issue exists
	_, err := s.GetIssue(issueId)
	if err != nil {
		return err // Pass through the raw error
	}

	query := `
		DELETE FROM ComponentVersionIssue
		WHERE
			componentversionissue_issue_id = :issue_id
			AND componentversionissue_component_version_id = :component_version_id
	`

	args := map[string]interface{}{
		"issue_id":             issueId,
		"component_version_id": componentVersionId,
	}

	_, err = performExec(s, query, args, l)

	return err // Return raw error
}

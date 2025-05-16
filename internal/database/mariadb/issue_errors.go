// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/sirupsen/logrus"
)

// FromDatabaseError converts database errors to application errors
func FromDatabaseError(op string, entity string, id string, err error) error {
	if err == nil {
		return nil
	}

	// Check for no rows error
	if errors.Is(err, sql.ErrNoRows) {
		return appErrors.NotFoundError(op, entity, id)
	}

	// Check for duplicate entry error
	var dupErr *database.DuplicateEntryDatabaseError
	if errors.As(err, &dupErr) {
		return appErrors.AlreadyExistsError(op, entity, id)
	}

	// Check for specific MySQL/MariaDB errors
	if strings.Contains(err.Error(), "Error 1062") || // Duplicate entry
		strings.Contains(err.Error(), "Duplicate entry") {
		return appErrors.AlreadyExistsError(op, entity, id)
	}

	// Default to internal error
	return appErrors.InternalError(op, entity, id, err)
}

// Modified issue methods that use the new error system

// GetIssue fetches a single issue by ID
func (s *SqlDatabase) GetIssue(id int64) (*entity.Issue, error) {
	const op = "mariadb.GetIssue"
	
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
		return nil, appErrors.InternalError(op, "Issue", "", err)
	}

	if len(issues) == 0 {
		return nil, appErrors.NotFoundError(op, "Issue", strconv.FormatInt(id, 10))
	}

	return issues[0].Issue, nil
}

// CreateIssue creates a new issue
func (s *SqlDatabase) CreateIssue(issue *entity.Issue) (*entity.Issue, error) {
	const op = "mariadb.CreateIssue"
	
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
		return nil, FromDatabaseError(op, "Issue", "", err)
	}

	issue.Id = id

	return issue, nil
}

// UpdateIssue updates an existing issue
func (s *SqlDatabase) UpdateIssue(issue *entity.Issue) error {
	const op = "mariadb.UpdateIssue"
	
	l := logrus.WithFields(logrus.Fields{
		"issue": issue,
		"event": "database.UpdateIssue",
	})

	// First check if issue exists
	_, err := s.GetIssue(issue.Id)
	if err != nil {
		return err // Already handled by GetIssue
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

	if err != nil {
		return FromDatabaseError(op, "Issue", strconv.FormatInt(issue.Id, 10), err)
	}

	return nil
}

// DeleteIssue soft-deletes an issue by ID
func (s *SqlDatabase) DeleteIssue(id int64, userId int64) error {
	const op = "mariadb.DeleteIssue"
	
	l := logrus.WithFields(logrus.Fields{
		"id":    id,
		"event": "database.DeleteIssue",
	})

	// First check if issue exists
	_, err := s.GetIssue(id)
	if err != nil {
		return err // Already handled by GetIssue
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

	if err != nil {
		return FromDatabaseError(op, "Issue", strconv.FormatInt(id, 10), err)
	}

	return nil
}

// AddComponentVersionToIssue associates a component version with an issue
func (s *SqlDatabase) AddComponentVersionToIssue(issueId int64, componentVersionId int64) error {
	const op = "mariadb.AddComponentVersionToIssue"
	
	l := logrus.WithFields(logrus.Fields{
		"issueId":            issueId,
		"componentVersionId": componentVersionId,
		"event":              "database.AddComponentVersionToIssue",
	})

	// Check if issue exists
	_, err := s.GetIssue(issueId)
	if err != nil {
		return err // Already handled by GetIssue
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

	if err != nil {
		return FromDatabaseError(op, "Issue", strconv.FormatInt(issueId, 10), err)
	}

	return nil
}

// RemoveComponentVersionFromIssue removes association between component version and issue
func (s *SqlDatabase) RemoveComponentVersionFromIssue(issueId int64, componentVersionId int64) error {
	const op = "mariadb.RemoveComponentVersionFromIssue"
	
	l := logrus.WithFields(logrus.Fields{
		"issueId":            issueId,
		"componentVersionId": componentVersionId,
		"event":              "database.RemoveComponentVersionFromIssue",
	})

	// Check if issue exists
	_, err := s.GetIssue(issueId)
	if err != nil {
		return err // Already handled by GetIssue
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

	if err != nil {
		return FromDatabaseError(op, "Issue", strconv.FormatInt(issueId, 10), err)
	}

	return nil
}

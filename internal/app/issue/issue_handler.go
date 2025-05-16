// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0
package issue

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

type issueHandler struct {
	database      database.Database
	eventRegistry event.EventRegistry
	logger        *logrus.Logger
}

func NewIssueHandler(db database.Database, er event.EventRegistry) IssueHandler {
	return &issueHandler{
		database:      db,
		eventRegistry: er,
		logger:        logrus.StandardLogger(),
	}
}

// LogError logs an error with contextual information
func (is *issueHandler) logError(err error, fields logrus.Fields) {
	// Extract operation, entity, and code from the error
	var appErr *appErrors.Error
	if !errors.As(err, &appErr) {
		is.logger.WithError(err).WithFields(fields).Error("Unknown error")
		return
	}

	errorFields := logrus.Fields{
		"error_code": string(appErr.Code),
	}

	// Add error-specific fields
	if appErr.Entity != "" {
		errorFields["entity"] = appErr.Entity
	}
	if appErr.ID != "" {
		errorFields["entity_id"] = appErr.ID
	}
	if appErr.Op != "" {
		errorFields["operation"] = appErr.Op
	}

	// Add any additional fields from the error
	for k, v := range appErr.Fields {
		errorFields[k] = v
	}

	// Add any passed-in fields
	for k, v := range fields {
		errorFields[k] = v
	}

	// Log with context
	is.logger.WithFields(errorFields).WithError(appErr.Err).Error(appErr.Error())
}

func (is *issueHandler) GetIssue(id int64) (*entity.Issue, error) {
	const op = "issueHandler.GetIssue"
	
	fields := logrus.Fields{
		"event": GetIssueEventName,
		"id":    id,
	}

	// Input validation
	if id <= 0 {
		err := appErrors.InvalidArgumentError(op, "Issue", "id must be positive")
		is.logError(err, fields)
		return nil, err
	}

	lo := entity.IssueListOptions{
		ListOptions: *entity.NewListOptions(),
	}
	issues, err := is.ListIssues(&entity.IssueFilter{Id: []*int64{&id}}, &lo)

	if err != nil {
		// Error already logged in ListIssues
		return nil, appErrors.InternalError(op, "Issue", strconv.FormatInt(id, 10), err)
	}

	if len(issues.Elements) != 1 {
		err := appErrors.NotFoundError(op, "Issue", strconv.FormatInt(id, 10))
		is.logError(err, fields)
		return nil, err
	}

	issue := issues.Elements[0].Issue
	is.eventRegistry.PushEvent(&GetIssueEvent{IssueID: id, Issue: issue})
	return issue, nil
}

func (is *issueHandler) ListIssues(filter *entity.IssueFilter, options *entity.IssueListOptions) (*entity.IssueList, error) {
	const op = "issueHandler.ListIssues"
	
	var pageInfo *entity.PageInfo
	var res []entity.IssueResult
	var err error
	issueList := entity.IssueList{
		List: &entity.List[entity.IssueResult]{},
	}

	fields := logrus.Fields{
		"event":  ListIssuesEventName,
		"filter": filter,
	}

	// Input validation
	if filter == nil {
		err := appErrors.InvalidArgumentError(op, "Issue", "filter cannot be nil")
		is.logError(err, fields)
		return nil, err
	}

	common.EnsurePaginatedX(&filter.PaginatedX)

	if options.IncludeAggregations {
		res, err = is.database.GetIssuesWithAggregations(filter, options.Order)
		if err != nil {
			appErr := appErrors.InternalError(op, "Issue", "", err)
			is.logError(appErr, fields)
			return nil, appErr
		}
	} else {
		res, err = is.database.GetIssues(filter, options.Order)
		if err != nil {
			appErr := appErrors.InternalError(op, "Issue", "", err)
			is.logError(appErr, fields)
			return nil, appErr
		}
	}

	issueList.Elements = res

	if options.ShowPageInfo {
		if len(res) > 0 {
			cursors, err := is.database.GetAllIssueCursors(filter, options.Order)
			if err != nil {
				appErr := appErrors.InternalError(op, "Issue", "", err)
				is.logError(appErr, fields)
				return nil, appErr
			}
			pageInfo = common.GetPageInfoX(res, cursors, *filter.First, filter.After)
			issueList.PageInfo = pageInfo
		}
	}
	if options.ShowPageInfo || options.ShowTotalCount || options.ShowIssueTypeCounts {
		counts, err := is.database.CountIssueTypes(filter)
		if err != nil {
			appErr := appErrors.InternalError(op, "Issue", "", err)
			is.logError(appErr, fields)
			return nil, appErr
		}
		tc := counts.TotalIssueCount()
		issueList.PolicyViolationCount = &counts.PolicyViolationCount
		issueList.SecurityEventCount = &counts.SecurityEventCount
		issueList.VulnerabilityCount = &counts.VulnerabilityCount
		issueList.TotalCount = &tc
	}

	is.eventRegistry.PushEvent(&ListIssuesEvent{
		Filter:  filter,
		Options: options,
		Issues:  &issueList,
	})

	return &issueList, nil
}

func (is *issueHandler) CreateIssue(issue *entity.Issue) (*entity.Issue, error) {
	const op = "issueHandler.CreateIssue"
	
	fields := logrus.Fields{
		"event":  CreateIssueEventName,
		"object": issue,
	}

	// Input validation
	if issue == nil {
		err := appErrors.InvalidArgumentError(op, "Issue", "issue cannot be nil")
		is.logError(err, fields)
		return nil, err
	}

	if issue.PrimaryName == "" {
		err := appErrors.InvalidArgumentError(op, "Issue", "primary name is required")
		is.logError(err, fields)
		return nil, err
	}

	// Check for existing issue with the same name
	f := &entity.IssueFilter{
		PrimaryName: []*string{&issue.PrimaryName},
	}
	fields["filter"] = f

	var err error
	issue.CreatedBy, err = common.GetCurrentUserId(is.database)
	if err != nil {
		appErr := appErrors.InternalError(op, "Issue", "", err)
		is.logError(appErr, fields)
		return nil, appErr
	}
	issue.UpdatedBy = issue.CreatedBy

	lo := entity.IssueListOptions{
		ListOptions: *entity.NewListOptions(),
	}
	issues, err := is.ListIssues(f, &lo)

	if err != nil {
		// Error already logged in ListIssues
		return nil, appErrors.InternalError(op, "Issue", "", err)
	}

	if len(issues.Elements) > 0 {
		err := appErrors.AlreadyExistsError(op, "Issue", issue.PrimaryName)
		is.logError(err, fields)
		return nil, err
	}

	newIssue, err := is.database.CreateIssue(issue)

	if err != nil {
		appErr := appErrors.InternalError(op, "Issue", "", err)
		is.logError(appErr, fields)
		return nil, appErr
	}

	is.eventRegistry.PushEvent(&CreateIssueEvent{Issue: newIssue})
	return newIssue, nil
}

func (is *issueHandler) UpdateIssue(issue *entity.Issue) (*entity.Issue, error) {
	const op = "issueHandler.UpdateIssue"
	
	fields := logrus.Fields{
		"event":  UpdateIssueEventName,
		"object": issue,
	}

	// Input validation
	if issue == nil {
		err := appErrors.InvalidArgumentError(op, "Issue", "issue cannot be nil")
		is.logError(err, fields)
		return nil, err
	}
	
	if issue.Id <= 0 {
		err := appErrors.InvalidArgumentError(op, "Issue", "id must be positive")
		is.logError(err, fields)
		return nil, err
	}

	// Check if issue exists
	_, err := is.GetIssue(issue.Id)
	if err != nil {
		// Error already logged in GetIssue
		return nil, err
	}

	var userId int64
	userId, err = common.GetCurrentUserId(is.database)
	if err != nil {
		appErr := appErrors.InternalError(op, "Issue", strconv.FormatInt(issue.Id, 10), err)
		is.logError(appErr, fields)
		return nil, appErr
	}
	issue.UpdatedBy = userId

	err = is.database.UpdateIssue(issue)

	if err != nil {
		appErr := appErrors.InternalError(op, "Issue", strconv.FormatInt(issue.Id, 10), err)
		is.logError(appErr, fields)
		return nil, appErr
	}

	// Get the updated issue
	lo := entity.IssueListOptions{
		ListOptions: *entity.NewListOptions(),
	}
	issueResult, err := is.ListIssues(&entity.IssueFilter{Id: []*int64{&issue.Id}}, &lo)

	if err != nil {
		// Error already logged in ListIssues
		return nil, appErrors.InternalError(op, "Issue", strconv.FormatInt(issue.Id, 10), err)
	}

	if len(issueResult.Elements) != 1 {
		err := appErrors.InternalError(op, "Issue", strconv.FormatInt(issue.Id, 10), 
			fmt.Errorf("expected 1 issue, found %d", len(issueResult.Elements)))
		is.logError(err, fields)
		return nil, err
	}

	is.eventRegistry.PushEvent(&UpdateIssueEvent{Issue: issue})
	return issueResult.Elements[0].Issue, nil
}

func (is *issueHandler) DeleteIssue(id int64) error {
	const op = "issueHandler.DeleteIssue"
	
	fields := logrus.Fields{
		"event": DeleteIssueEventName,
		"id":    id,
	}

	// Input validation
	if id <= 0 {
		err := appErrors.InvalidArgumentError(op, "Issue", "id must be positive")
		is.logError(err, fields)
		return err
	}
	
	// Check if issue exists
	_, err := is.GetIssue(id)
	if err != nil {
		// Error already logged in GetIssue
		return err
	}

	userId, err := common.GetCurrentUserId(is.database)
	if err != nil {
		appErr := appErrors.InternalError(op, "Issue", strconv.FormatInt(id, 10), err)
		is.logError(appErr, fields)
		return appErr
	}

	err = is.database.DeleteIssue(id, userId)

	if err != nil {
		appErr := appErrors.InternalError(op, "Issue", strconv.FormatInt(id, 10), err)
		is.logError(appErr, fields)
		return appErr
	}

	is.eventRegistry.PushEvent(&DeleteIssueEvent{IssueID: id})
	return nil
}

func (is *issueHandler) AddComponentVersionToIssue(issueId, componentVersionId int64) (*entity.Issue, error) {
	const op = "issueHandler.AddComponentVersionToIssue"
	
	fields := logrus.Fields{
		"event": AddComponentVersionToIssueEventName,
		"id":    issueId,
		"componentVersionId": componentVersionId,
	}

	// Input validation
	if issueId <= 0 {
		err := appErrors.InvalidArgumentError(op, "Issue", "issue id must be positive")
		is.logError(err, fields)
		return nil, err
	}

	if componentVersionId <= 0 {
		err := appErrors.InvalidArgumentError(op, "ComponentVersion", "component version id must be positive")
		is.logError(err, fields)
		return nil, err
	}

	// Check if issue exists
	_, err := is.GetIssue(issueId)
	if err != nil {
		// Error already logged in GetIssue
		return nil, err
	}

	err = is.database.AddComponentVersionToIssue(issueId, componentVersionId)

	if err != nil {
		// Check for specific known errors
		var dupErr *database.DuplicateEntryDatabaseError
		if errors.As(err, &dupErr) {
			appErr := appErrors.AlreadyExistsError(op, "ComponentVersionIssue", 
				fmt.Sprintf("issue %d, component version %d", issueId, componentVersionId))
			is.logError(appErr, fields)
			return nil, appErr
		}
		
		// Handle other errors
		appErr := appErrors.InternalError(op, "Issue", strconv.FormatInt(issueId, 10), err)
		is.logError(appErr, fields)
		return nil, appErr
	}

	is.eventRegistry.PushEvent(&AddComponentVersionToIssueEvent{
		IssueID:            issueId,
		ComponentVersionID: componentVersionId,
	})

	return is.GetIssue(issueId)
}

func (is *issueHandler) RemoveComponentVersionFromIssue(issueId, componentVersionId int64) (*entity.Issue, error) {
	const op = "issueHandler.RemoveComponentVersionFromIssue"
	
	fields := logrus.Fields{
		"event": RemoveComponentVersionFromIssueEventName,
		"id":    issueId,
		"componentVersionId": componentVersionId,
	}

	// Input validation
	if issueId <= 0 {
		err := appErrors.InvalidArgumentError(op, "Issue", "issue id must be positive")
		is.logError(err, fields)
		return nil, err
	}

	if componentVersionId <= 0 {
		err := appErrors.InvalidArgumentError(op, "ComponentVersion", "component version id must be positive")
		is.logError(err, fields)
		return nil, err
	}
	
	// Check if issue exists
	_, err := is.GetIssue(issueId)
	if err != nil {
		// Error already logged in GetIssue
		return nil, err
	}

	err = is.database.RemoveComponentVersionFromIssue(issueId, componentVersionId)

	if err != nil {
		appErr := appErrors.InternalError(op, "Issue", strconv.FormatInt(issueId, 10), err)
		is.logError(appErr, fields)
		return nil, appErr
	}

	is.eventRegistry.PushEvent(&RemoveComponentVersionFromIssueEvent{
		IssueID:            issueId,
		ComponentVersionID: componentVersionId,
	})

	return is.GetIssue(issueId)
}

func (is *issueHandler) ListIssueNames(filter *entity.IssueFilter, options *entity.ListOptions) ([]string, error) {
	const op = "issueHandler.ListIssueNames"
	
	fields := logrus.Fields{
		"event":  ListIssueNamesEventName,
		"filter": filter,
	}

	// Input validation
	if filter == nil {
		err := appErrors.InvalidArgumentError(op, "Issue", "filter cannot be nil")
		is.logError(err, fields)
		return nil, err
	}

	issueNames, err := is.database.GetIssueNames(filter)

	if err != nil {
		appErr := appErrors.InternalError(op, "Issue", "", err)
		is.logError(appErr, fields)
		return nil, appErr
	}

	is.eventRegistry.PushEvent(&ListIssueNamesEvent{
		Filter:  filter,
		Options: options,
		Names:   issueNames,
	})

	return issueNames, nil
}

func (is *issueHandler) GetIssueSeverityCounts(filter *entity.IssueFilter) (*entity.IssueSeverityCounts, error) {
	const op = "issueHandler.GetIssueSeverityCounts"
	
	fields := logrus.Fields{
		"event":  "GetIssueSeverityCounts",
		"filter": filter,
	}

	// Input validation
	if filter == nil {
		err := appErrors.InvalidArgumentError(op, "Issue", "filter cannot be nil")
		is.logError(err, fields)
		return nil, err
	}

	counts, err := is.database.CountIssueRatings(filter)

	if err != nil {
		appErr := appErrors.InternalError(op, "Issue", "", err)
		is.logError(appErr, fields)
		return nil, appErr
	}

	is.eventRegistry.PushEvent(&GetIssueSeverityCountsEvent{
		Filter: filter,
		Counts: counts,
	})

	return counts, nil
}

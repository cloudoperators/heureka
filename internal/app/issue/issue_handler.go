// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue

import (
	"errors"
	"fmt"
	"time"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/cache"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

var CacheTtlGetIssuesWithAggregations = 12 * time.Hour
var CacheTtlGetIssues = 12 * time.Hour
var CacheTtlGetAllIssueCursors = 12 * time.Hour
var CacheTtlCountIssueTypes = 12 * time.Hour
var CacheTtlGetIssueNames = 12 * time.Hour

type issueHandler struct {
	database      database.Database
	eventRegistry event.EventRegistry
	cache         cache.Cache
}

func NewIssueHandler(db database.Database, er event.EventRegistry, cache cache.Cache) IssueHandler {
	return &issueHandler{
		database:      db,
		eventRegistry: er,
		cache:         cache,
	}
}

type IssueHandlerError struct {
	msg string
}

func (e *IssueHandlerError) Error() string {
	return fmt.Sprintf("IssueHandlerError: %s", e.msg)
}

func NewIssueHandlerError(msg string) *IssueHandlerError {
	return &IssueHandlerError{msg: msg}
}

func (is *issueHandler) GetIssue(id int64) (*entity.Issue, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": GetIssueEventName,
		"id":    id,
	})

	lo := entity.IssueListOptions{
		ListOptions: *entity.NewListOptions(),
	}
	issues, err := is.ListIssues(&entity.IssueFilter{Id: []*int64{&id}}, &lo)

	if err != nil {
		l.Error(err)
		return nil, NewIssueHandlerError("Internal error while retrieving issue.")
	}

	if len(issues.Elements) != 1 {
		return nil, NewIssueHandlerError(fmt.Sprintf("Issue %d not found.", id))
	}

	issue := issues.Elements[0].Issue
	is.eventRegistry.PushEvent(&GetIssueEvent{IssueID: id, Issue: issue})
	return issue, nil
}

func (is *issueHandler) ListIssues(filter *entity.IssueFilter, options *entity.IssueListOptions) (*entity.IssueList, error) {
	var pageInfo *entity.PageInfo
	var res []entity.IssueResult
	var err error
	issueList := entity.IssueList{
		List: &entity.List[entity.IssueResult]{},
	}

	l := logrus.WithFields(logrus.Fields{
		"event":  ListIssuesEventName,
		"filter": filter,
	})

	common.EnsurePaginatedX(&filter.PaginatedX)

	if options.IncludeAggregations {
		res, err = cache.CallCached[[]entity.IssueResult](
			is.cache,
			CacheTtlGetIssuesWithAggregations,
			"GetIssuesWithAggregations",
			is.database.GetIssuesWithAggregations,
			filter,
			options.Order,
		)
		if err != nil {
			l.Error(err)
			return nil, NewIssueHandlerError("Internal error while retrieving list results witis aggregations")
		}
	} else {
		res, err = cache.CallCached[[]entity.IssueResult](
			is.cache,
			CacheTtlGetIssues,
			"GetIssues",
			is.database.GetIssues,
			filter,
			options.Order,
		)
		if err != nil {
			l.Error(err)
			return nil, NewIssueHandlerError("Internal error while retrieving list results.")
		}
	}

	issueList.Elements = res

	if options.ShowPageInfo {
		if len(res) > 0 {
			cursors, err := cache.CallCached[[]string](
				is.cache,
				CacheTtlGetAllIssueCursors,
				"GetAllIssueCursors",
				is.database.GetAllIssueCursors,
				filter,
				options.Order,
			)
			if err != nil {
				l.Error(err)
				return nil, NewIssueHandlerError("Error while getting all cursors")
			}
			pageInfo = common.GetPageInfoX(res, cursors, *filter.First, filter.After)
			issueList.PageInfo = pageInfo
		}
	}
	if options.ShowPageInfo || options.ShowTotalCount || options.ShowIssueTypeCounts {
		counts, err := cache.CallCached[*entity.IssueTypeCounts](
			is.cache,
			CacheTtlCountIssueTypes,
			"CountIssueTypes",
			is.database.CountIssueTypes,
			filter,
		)
		if err != nil {
			l.Error(err)
			return nil, NewIssueHandlerError("Error while count of issues")
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
	f := &entity.IssueFilter{
		PrimaryName: []*string{&issue.PrimaryName},
	}

	l := logrus.WithFields(logrus.Fields{
		"event":  CreateIssueEventName,
		"object": issue,
		"filter": f,
	})

	var err error
	issue.CreatedBy, err = common.GetCurrentUserId(is.database)
	if err != nil {
		l.Error(err)
		return nil, NewIssueHandlerError("Internal error while creating issue (GetUserId).")
	}
	issue.UpdatedBy = issue.CreatedBy

	lo := entity.IssueListOptions{
		ListOptions: *entity.NewListOptions(),
	}
	issues, err := is.ListIssues(f, &lo)

	if err != nil {
		l.Error(err)
		return nil, NewIssueHandlerError("Internal error while creating issue.")
	}

	if len(issues.Elements) > 0 {
		return nil, NewIssueHandlerError(fmt.Sprintf("Duplicated entry %s for primaryName.", issue.PrimaryName))
	}

	newIssue, err := is.database.CreateIssue(issue)

	if err != nil {
		l.Error(err)
		return nil, NewIssueHandlerError("Internal error while creating issue.")
	}

	is.eventRegistry.PushEvent(&CreateIssueEvent{Issue: newIssue})
	return newIssue, nil
}

func (is *issueHandler) UpdateIssue(issue *entity.Issue) (*entity.Issue, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  UpdateIssueEventName,
		"object": issue,
	})

	var err error
	issue.UpdatedBy, err = common.GetCurrentUserId(is.database)
	if err != nil {
		l.Error(err)
		return nil, NewIssueHandlerError("Internal error while updating issue (GetUserId).")
	}

	err = is.database.UpdateIssue(issue)

	if err != nil {
		l.Error(err)
		return nil, NewIssueHandlerError("Internal error while updating issue.")
	}

	lo := entity.IssueListOptions{
		ListOptions: *entity.NewListOptions(),
	}
	issueResult, err := is.ListIssues(&entity.IssueFilter{Id: []*int64{&issue.Id}}, &lo)

	if err != nil {
		l.Error(err)
		return nil, NewIssueHandlerError("Internal error while retrieving updated issue.")
	}

	if len(issueResult.Elements) != 1 {
		l.Error(err)
		return nil, NewIssueHandlerError("Multiple issues found.")
	}

	is.eventRegistry.PushEvent(&UpdateIssueEvent{Issue: issue})
	return issueResult.Elements[0].Issue, nil
}

func (is *issueHandler) DeleteIssue(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": DeleteIssueEventName,
		"id":    id,
	})

	userId, err := common.GetCurrentUserId(is.database)
	if err != nil {
		l.Error(err)
		return NewIssueHandlerError("Internal error while deleting issue (GetUserId).")
	}

	err = is.database.DeleteIssue(id, userId)

	if err != nil {
		l.Error(err)
		return NewIssueHandlerError("Internal error while deleting issue.")
	}

	is.eventRegistry.PushEvent(&DeleteIssueEvent{IssueID: id})
	return nil
}

func (is *issueHandler) AddComponentVersionToIssue(issueId, componentVersionId int64) (*entity.Issue, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": AddComponentVersionToIssueEventName,
		"id":    issueId,
	})

	err := is.database.AddComponentVersionToIssue(issueId, componentVersionId)

	if err != nil {
		l.Error(err)
		duplicateEntryError := &database.DuplicateEntryDatabaseError{}
		if errors.As(err, &duplicateEntryError) {
			return nil, NewIssueHandlerError("Entry already Exists")
		}
		return nil, NewIssueHandlerError("Internal error while adding component version to issue.")
	}

	is.eventRegistry.PushEvent(&AddComponentVersionToIssueEvent{
		IssueID:            issueId,
		ComponentVersionID: componentVersionId,
	})

	return is.GetIssue(issueId)
}

func (is *issueHandler) RemoveComponentVersionFromIssue(issueId, componentVersionId int64) (*entity.Issue, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": RemoveComponentVersionFromIssueEventName,
		"id":    issueId,
	})

	err := is.database.RemoveComponentVersionFromIssue(issueId, componentVersionId)

	if err != nil {
		l.Error(err)
		return nil, NewIssueHandlerError("Internal error while removing component version from issue.")
	}

	is.eventRegistry.PushEvent(&RemoveComponentVersionFromIssueEvent{
		IssueID:            issueId,
		ComponentVersionID: componentVersionId,
	})

	return is.GetIssue(issueId)
}

func (is *issueHandler) ListIssueNames(filter *entity.IssueFilter, options *entity.ListOptions) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  ListIssueNamesEventName,
		"filter": filter,
	})

	issueNames, err := cache.CallCached[[]string](
		is.cache,
		CacheTtlGetIssueNames,
		"GetIssueNames",
		is.database.GetIssueNames,
		filter,
	)

	if err != nil {
		l.Error(err)
		return nil, NewIssueHandlerError("Internal error while retrieving issueNames.")
	}

	is.eventRegistry.PushEvent(&ListIssueNamesEvent{
		Filter:  filter,
		Options: options,
		Names:   issueNames,
	})

	return issueNames, nil
}

func (is *issueHandler) GetIssueSeverityCounts(filter *entity.IssueFilter) (*entity.IssueSeverityCounts, error) {
	counts, err := is.database.CountIssueRatings(filter)

	if err != nil {
		return nil, err
	}

	is.eventRegistry.PushEvent(&GetIssueSeverityCountsEvent{
		Filter: filter,
		Counts: counts,
	})

	return counts, nil
}

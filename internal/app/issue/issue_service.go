// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0
package issue

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/app/common"
	"github.wdf.sap.corp/cc/heureka/internal/app/event"
	"github.wdf.sap.corp/cc/heureka/internal/database"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
	"github.wdf.sap.corp/cc/heureka/pkg/util"
)

type issueService struct {
	database      database.Database
	eventRegistry event.EventRegistry
}

func NewIssueService(db database.Database, er event.EventRegistry) IssueService {
	return &issueService{
		database:      db,
		eventRegistry: er,
	}
}

type IssueServiceError struct {
	msg string
}

func (e *IssueServiceError) Error() string {
	return fmt.Sprintf("IssueServiceError: %s", e.msg)
}

func NewIssueServiceError(msg string) *IssueServiceError {
	return &IssueServiceError{msg: msg}
}

func (is *issueService) getIssueResultsWithAggregations(filter *entity.IssueFilter) ([]entity.IssueResult, error) {
	var issueResults []entity.IssueResult
	issues, err := is.database.GetIssuesWithAggregations(filter)
	if err != nil {
		return nil, err
	}

	for _, issue := range issues {
		cursor := fmt.Sprintf("%d", issue.Id)
		issueResults = append(issueResults, entity.IssueResult{
			WithCursor:        entity.WithCursor{Value: cursor},
			IssueAggregations: util.Ptr(issue.IssueAggregations),
			Issue:             util.Ptr(issue.Issue),
		})
	}

	return issueResults, nil
}

func (is *issueService) getIssueResults(filter *entity.IssueFilter) ([]entity.IssueResult, error) {
	var issueResults []entity.IssueResult
	issues, err := is.database.GetIssues(filter)
	if err != nil {
		return nil, err
	}
	for _, issue := range issues {
		cursor := fmt.Sprintf("%d", issue.Id)
		issueResults = append(issueResults, entity.IssueResult{
			WithCursor:        entity.WithCursor{Value: cursor},
			IssueAggregations: nil,
			Issue:             util.Ptr(issue),
		})
	}
	return issueResults, nil
}

func (is *issueService) GetIssue(id int64) (*entity.Issue, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": GetIssueEventName,
		"id":    id,
	})

	issues, err := is.ListIssues(&entity.IssueFilter{Id: []*int64{&id}}, &entity.IssueListOptions{})

	if err != nil {
		l.Error(err)
		return nil, NewIssueServiceError("Internal error while retrieving issue.")
	}

	if len(issues.Elements) != 1 {
		return nil, NewIssueServiceError(fmt.Sprintf("Issue %d not found.", id))
	}

	issue := issues.Elements[0].Issue
	is.eventRegistry.PushEvent(&GetIssueEvent{IssueID: id, Issue: issue})
	return issue, nil
}

func (is *issueService) ListIssues(filter *entity.IssueFilter, options *entity.IssueListOptions) (*entity.IssueList, error) {
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

	common.EnsurePaginated(&filter.Paginated)

	if options.IncludeAggregations {
		res, err = is.getIssueResultsWithAggregations(filter)
		if err != nil {
			l.Error(err)
			return nil, NewIssueServiceError("Internal error while retrieving list results witis aggregations")
		}
	} else {
		res, err = is.getIssueResults(filter)
		if err != nil {
			l.Error(err)
			return nil, NewIssueServiceError("Internal error while retrieving list results.")
		}
	}

	issueList.Elements = res

	if options.ShowPageInfo {
		if len(res) > 0 {
			ids, err := is.database.GetAllIssueIds(filter)
			if err != nil {
				l.Error(err)
				return nil, NewIssueServiceError("Error while getting all Ids")
			}
			pageInfo = common.GetPageInfo(res, ids, *filter.First, *filter.After)
			issueList.PageInfo = pageInfo
		}
	}
	if options.ShowPageInfo || options.ShowTotalCount || options.ShowIssueTypeCounts {
		counts, err := is.database.CountIssueTypes(filter)
		if err != nil {
			l.Error(err)
			return nil, NewIssueServiceError("Error while count of issues")
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

func (is *issueService) CreateIssue(issue *entity.Issue) (*entity.Issue, error) {
	f := &entity.IssueFilter{
		PrimaryName: []*string{&issue.PrimaryName},
	}

	l := logrus.WithFields(logrus.Fields{
		"event":  CreateIssueEventName,
		"object": issue,
		"filter": f,
	})

	issues, err := is.ListIssues(f, &entity.IssueListOptions{})

	if err != nil {
		l.Error(err)
		return nil, NewIssueServiceError("Internal error while creating issue.")
	}

	if len(issues.Elements) > 0 {
		return nil, NewIssueServiceError(fmt.Sprintf("Duplicated entry %s for primaryName.", issue.PrimaryName))
	}

	newIssue, err := is.database.CreateIssue(issue)

	if err != nil {
		l.Error(err)
		return nil, NewIssueServiceError("Internal error while creating issue.")
	}

	is.eventRegistry.PushEvent(&CreateIssueEvent{Issue: newIssue})
	return newIssue, nil
}

func (is *issueService) UpdateIssue(issue *entity.Issue) (*entity.Issue, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  UpdateIssueEventName,
		"object": issue,
	})

	err := is.database.UpdateIssue(issue)

	if err != nil {
		l.Error(err)
		return nil, NewIssueServiceError("Internal error while updating issue.")
	}

	issueResult, err := is.ListIssues(&entity.IssueFilter{Id: []*int64{&issue.Id}}, &entity.IssueListOptions{})

	if err != nil {
		l.Error(err)
		return nil, NewIssueServiceError("Internal error while retrieving updated issue.")
	}

	if len(issueResult.Elements) != 1 {
		l.Error(err)
		return nil, NewIssueServiceError("Multiple issues found.")
	}

	is.eventRegistry.PushEvent(&UpdateIssueEvent{Issue: issue})
	return issueResult.Elements[0].Issue, nil
}

func (is *issueService) DeleteIssue(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": DeleteIssueEventName,
		"id":    id,
	})

	err := is.database.DeleteIssue(id)

	if err != nil {
		l.Error(err)
		return NewIssueServiceError("Internal error while deleting issue.")
	}

	is.eventRegistry.PushEvent(&DeleteIssueEvent{IssueID: id})
	return nil
}

func (is *issueService) AddComponentVersionToIssue(issueId, componentVersionId int64) (*entity.Issue, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": AddComponentVersionToIssueEventName,
		"id":    issueId,
	})

	err := is.database.AddComponentVersionToIssue(issueId, componentVersionId)

	if err != nil {
		l.Error(err)
		return nil, NewIssueServiceError("Internal error while adding component version to issue.")
	}

	is.eventRegistry.PushEvent(&AddComponentVersionToIssueEvent{
		IssueID:            issueId,
		ComponentVersionID: componentVersionId,
	})

	return is.GetIssue(issueId)
}

func (is *issueService) RemoveComponentVersionFromIssue(issueId, componentVersionId int64) (*entity.Issue, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": RemoveComponentVersionFromIssueEventName,
		"id":    issueId,
	})

	err := is.database.RemoveComponentVersionFromIssue(issueId, componentVersionId)

	if err != nil {
		l.Error(err)
		return nil, NewIssueServiceError("Internal error while removing component version from issue.")
	}

	is.eventRegistry.PushEvent(&RemoveComponentVersionFromIssueEvent{
		IssueID:            issueId,
		ComponentVersionID: componentVersionId,
	})

	return is.GetIssue(issueId)
}

func (is *issueService) ListIssueNames(filter *entity.IssueFilter, options *entity.ListOptions) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  ListIssueNamesEventName,
		"filter": filter,
	})

	issueNames, err := is.database.GetIssueNames(filter)

	if err != nil {
		l.Error(err)
		return nil, NewIssueServiceError("Internal error while retrieving issueNames.")
	}

	is.eventRegistry.PushEvent(&ListIssueNamesEvent{
		Filter:  filter,
		Options: options,
		Names:   issueNames,
	})

	return issueNames, nil
}

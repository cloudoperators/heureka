// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/cloudoperators/heureka/internal/app/common"
	applog "github.com/cloudoperators/heureka/internal/app/logging"
	"github.com/cloudoperators/heureka/internal/cache"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/sirupsen/logrus"
)

type issueHandler struct {
	common.BaseHandler[entity.IssueResult, *entity.IssueFilter]
}

func NewIssueHandler(handlerContext common.HandlerContext) IssueHandler {
	return &issueHandler{
		BaseHandler: common.NewBaseHandler(handlerContext, common.BaseConfig[entity.IssueResult, *entity.IssueFilter]{
			Op:        appErrors.Op("issueHandler"),
			Entity:    "Issues",
			GetFn:     handlerContext.DB.GetIssues,
			CursorsFn: handlerContext.DB.GetAllIssueCursors,
		}),
	}
}

func (is *issueHandler) GetIssue(ctx context.Context, id int64) (*entity.Issue, error) {
	op := appErrors.CallerOp()

	if id <= 0 {
		return nil, appErrors.E(op, "Issue", appErrors.InvalidArgument, fmt.Sprintf("invalid ID: %d", id))
	}

	lo := entity.IssueListOptions{ListOptions: *entity.NewListOptions()}

	issues, err := is.ListIssues(ctx, &entity.IssueFilter{Id: []*int64{&id}}, &lo)
	if err != nil {
		return nil, appErrors.E(op, "Issue", strconv.FormatInt(id, 10), appErrors.Internal, err)
	}

	if len(issues.Elements) == 0 {
		return nil, appErrors.E(op, "Issue", strconv.FormatInt(id, 10), appErrors.NotFound)
	}

	if len(issues.Elements) > 1 {
		return nil, appErrors.E(op, "Issue", strconv.FormatInt(id, 10), appErrors.Internal,
			fmt.Sprintf("found %d issues with ID %d, expected 1", len(issues.Elements), id))
	}

	issue := issues.Elements[0].Issue
	is.PushEvent(&GetIssueEvent{IssueID: id, Issue: issue})

	return issue, nil
}

func (is *issueHandler) ListIssues(
	ctx context.Context,
	filter *entity.IssueFilter,
	options *entity.IssueListOptions,
) (*entity.IssueList, error) {
	op := appErrors.CallerOp()

	issueList := entity.IssueList{List: &entity.List[entity.IssueResult]{}}

	common.EnsurePaginated(&filter.Paginated)

	options = ensureIssueListOptions(options)

	if options.IncludeAggregations {
		res, err := cache.CallCached[[]entity.IssueResult](
			is.Cache(),
			cache.NewCacheCallParams(
				common.DefaultCacheTTL, ctx, "GetIssuesWithAggregations",
				is.DB().GetIssuesWithAggregations, filter, options.Order,
			),
		)
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), "Issues", "", err)
			applog.LogError(logrus.StandardLogger(), wrappedErr, logrus.Fields{"filter": filter})

			return nil, wrappedErr
		}

		issueList.Elements = res
	} else {
		res, err := cache.CallCached[[]entity.IssueResult](
			is.Cache(),
			cache.NewCacheCallParams(
				common.DefaultCacheTTL, ctx, "GetIssues",
				is.DB().GetIssues, filter, options.Order,
			),
		)
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), "Issues", "", err)
			applog.LogError(logrus.StandardLogger(), wrappedErr, logrus.Fields{"filter": filter})

			return nil, wrappedErr
		}

		issueList.Elements = res
	}

	if options.ShowPageInfo && len(issueList.Elements) > 0 {
		cursors, err := cache.CallCached[[]string](
			is.Cache(),
			cache.NewCacheCallParams(
				common.DefaultCacheTTL, ctx, "GetAllIssueCursors",
				is.DB().GetAllIssueCursors, filter, options.Order,
			),
		)
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), "IssueCursors", "", err)
			applog.LogError(logrus.StandardLogger(), wrappedErr, logrus.Fields{"filter": filter})

			return nil, wrappedErr
		}

		issueList.PageInfo = common.GetPageInfo(issueList.Elements, cursors, *filter.First, filter.After)
	}

	if options.ShowPageInfo || options.ShowTotalCount || options.ShowIssueTypeCounts {
		counts, err := cache.CallCached[*entity.IssueTypeCounts](
			is.Cache(),
			cache.NewCacheCallParams(
				common.DefaultCacheTTL, ctx, "CountIssueTypes",
				is.DB().CountIssueTypes, filter,
			),
		)
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), "IssueTypeCounts", "", err)
			applog.LogError(logrus.StandardLogger(), wrappedErr, logrus.Fields{"filter": filter})

			return nil, wrappedErr
		}

		tc := counts.TotalIssueCount()
		issueList.PolicyViolationCount = &counts.PolicyViolationCount
		issueList.SecurityEventCount = &counts.SecurityEventCount
		issueList.VulnerabilityCount = &counts.VulnerabilityCount
		issueList.TotalCount = &tc
	}

	is.PushEvent(&ListIssuesEvent{Filter: filter, Options: options, Issues: &issueList})

	return &issueList, nil
}

func (is *issueHandler) CreateIssue(ctx context.Context, issue *entity.Issue) (*entity.Issue, error) {
	op := appErrors.CallerOp()

	var err error

	issue.CreatedBy, err = common.GetCurrentUserId(ctx, is.DB())
	if err != nil {
		return nil, appErrors.InternalError(string(op), "Issue", "", err)
	}

	issue.UpdatedBy = issue.CreatedBy

	lo := entity.IssueListOptions{ListOptions: *entity.NewListOptions()}

	existing, err := is.ListIssues(ctx, &entity.IssueFilter{PrimaryName: []*string{&issue.PrimaryName}}, &lo)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "Issue", "", err)
	}

	if len(existing.Elements) > 0 {
		return nil, appErrors.AlreadyExistsError(string(op), "Issue", issue.PrimaryName)
	}

	newIssue, err := is.DB().CreateIssue(issue)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "Issue", "", err)
	}

	is.PushEvent(&CreateIssueEvent{Issue: newIssue})

	return newIssue, nil
}

func (is *issueHandler) UpdateIssue(ctx context.Context, issue *entity.Issue) (*entity.Issue, error) {
	op := appErrors.CallerOp()
	id := strconv.FormatInt(issue.Id, 10)

	var err error

	issue.UpdatedBy, err = common.GetCurrentUserId(ctx, is.DB())
	if err != nil {
		return nil, appErrors.InternalError(string(op), "Issue", id, err)
	}

	if err = is.DB().UpdateIssue(issue); err != nil {
		return nil, appErrors.InternalError(string(op), "Issue", id, err)
	}

	lo := entity.IssueListOptions{ListOptions: *entity.NewListOptions()}

	result, err := is.ListIssues(ctx, &entity.IssueFilter{Id: []*int64{&issue.Id}}, &lo)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "Issue", id, err)
	}

	if len(result.Elements) != 1 {
		return nil, appErrors.E(op, "Issue", id, appErrors.Internal,
			fmt.Sprintf("unexpected result count: %d", len(result.Elements)))
	}

	updatedIssue := result.Elements[0].Issue
	is.PushEvent(&UpdateIssueEvent{Issue: updatedIssue})

	return updatedIssue, nil
}

func (is *issueHandler) DeleteIssue(ctx context.Context, id int64) error {
	op := appErrors.CallerOp()
	idStr := strconv.FormatInt(id, 10)

	userId, err := common.GetCurrentUserId(ctx, is.DB())
	if err != nil {
		return appErrors.InternalError(string(op), "Issue", idStr, err)
	}

	if err = is.DB().DeleteIssue(id, userId); err != nil {
		return appErrors.InternalError(string(op), "Issue", idStr, err)
	}

	is.PushEvent(&DeleteIssueEvent{IssueID: id})

	return nil
}

func (is *issueHandler) AddComponentVersionToIssue(
	ctx context.Context,
	issueId, componentVersionId int64,
) (*entity.Issue, error) {
	op := appErrors.CallerOp()

	if err := is.DB().AddComponentVersionToIssue(issueId, componentVersionId); err != nil {
		duplicateEntryError := &database.DuplicateEntryDatabaseError{}
		if errors.As(err, &duplicateEntryError) {
			return nil, appErrors.AlreadyExistsError(string(op), "ComponentVersionIssue",
				fmt.Sprintf("issue:%d-componentVersion:%d", issueId, componentVersionId))
		}

		return nil, appErrors.InternalError(string(op), "ComponentVersionIssue",
			fmt.Sprintf("issue:%d-componentVersion:%d", issueId, componentVersionId), err)
	}

	is.PushEvent(&AddComponentVersionToIssueEvent{
		IssueID:            issueId,
		ComponentVersionID: componentVersionId,
	})

	return is.GetIssue(ctx, issueId)
}

func (is *issueHandler) RemoveComponentVersionFromIssue(
	ctx context.Context,
	issueId, componentVersionId int64,
) (*entity.Issue, error) {
	op := appErrors.CallerOp()

	if err := is.DB().RemoveComponentVersionFromIssue(issueId, componentVersionId); err != nil {
		return nil, appErrors.InternalError(string(op), "ComponentVersionIssue",
			fmt.Sprintf("issue:%d-componentVersion:%d", issueId, componentVersionId), err)
	}

	is.PushEvent(&RemoveComponentVersionFromIssueEvent{
		IssueID:            issueId,
		ComponentVersionID: componentVersionId,
	})

	return is.GetIssue(ctx, issueId)
}

func (is *issueHandler) ListIssueNames(
	ctx context.Context,
	filter *entity.IssueFilter,
	options *entity.ListOptions,
) ([]string, error) {
	op := appErrors.CallerOp()

	issueNames, err := cache.CallCached[[]string](
		is.Cache(),
		cache.NewCacheCallParams(
			common.DefaultCacheTTL, ctx, "GetIssueNames",
			is.DB().GetIssueNames, filter,
		),
	)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "IssueNames", "", err)
	}

	is.PushEvent(&ListIssueNamesEvent{Filter: filter, Options: options, Names: issueNames})

	return issueNames, nil
}

func (is *issueHandler) GetIssueSeverityCounts(
	ctx context.Context,
	filter *entity.IssueFilter,
) (*entity.IssueSeverityCounts, error) {
	op := appErrors.CallerOp()

	counts, err := cache.CallCached[*entity.IssueSeverityCounts](
		is.Cache(),
		cache.NewCacheCallParams(
			common.DefaultCacheTTL, ctx, "CountIssueRatings",
			is.DB().CountIssueRatings, filter,
		),
	)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "IssueSeverityCounts", "", err)
	}

	is.PushEvent(&GetIssueSeverityCountsEvent{Filter: filter, Counts: counts})

	return counts, nil
}

func ensureIssueListOptions(options *entity.IssueListOptions) *entity.IssueListOptions {
	if options == nil {
		return &entity.IssueListOptions{
			ListOptions: *common.EnsureListOptions(nil),
		}
	}

	return options
}

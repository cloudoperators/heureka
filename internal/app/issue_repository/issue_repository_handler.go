// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue_repository

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/entity"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
)

type issueRepositoryHandler struct {
	common.BaseHandler[entity.IssueRepositoryResult, *entity.IssueRepositoryFilter]
}

func NewIssueRepositoryHandler(handlerContext common.HandlerContext) IssueRepositoryHandler {
	return &issueRepositoryHandler{
		BaseHandler: common.NewBaseHandler(handlerContext, common.BaseConfig[entity.IssueRepositoryResult, *entity.IssueRepositoryFilter]{
			Op:           appErrors.Op("issueRepositoryHandler"),
			Entity:       "IssueRepositories",
			CursorEntity: "IssueRepositoryCursors",
			CountEntity:  "IssueRepositoryCount",
			GetFn:        handlerContext.DB.GetIssueRepositories,
			CursorsFn:    handlerContext.DB.GetAllIssueRepositoryCursors,
			CountFn:      handlerContext.DB.CountIssueRepositories,
			ListEventFn: func(f *entity.IssueRepositoryFilter, o *entity.ListOptions, r *entity.List[entity.IssueRepositoryResult]) any {
				return &ListIssueRepositoriesEvent{Filter: f, Options: o, Results: r}
			},
			DeleteFn:      handlerContext.DB.DeleteIssueRepository,
			DeleteEventFn: func(id int64) any { return &DeleteIssueRepositoryEvent{IssueRepositoryID: id} },
		}),
	}
}

func (ir *issueRepositoryHandler) ListIssueRepositories(
	ctx context.Context,
	filter *entity.IssueRepositoryFilter,
	options *entity.ListOptions,
) (*entity.List[entity.IssueRepositoryResult], error) {
	return ir.List(ctx, appErrors.CallerOp(), filter, options)
}

func (ir *issueRepositoryHandler) CreateIssueRepository(
	ctx context.Context,
	issueRepository *entity.IssueRepository,
) (*entity.IssueRepository, error) {
	op := appErrors.CallerOp()

	var err error

	issueRepository.BaseIssueRepository.CreatedBy, err = common.GetCurrentUserId(ctx, ir.DB())
	if err != nil {
		return nil, appErrors.InternalError(string(op), "IssueRepository", "", err)
	}

	issueRepository.BaseIssueRepository.UpdatedBy = issueRepository.BaseIssueRepository.CreatedBy

	existing, err := ir.ListIssueRepositories(
		ctx,
		&entity.IssueRepositoryFilter{Name: []*string{&issueRepository.Name}},
		entity.NewListOptions(),
	)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "IssueRepository", "", err)
	}

	if len(existing.Elements) > 0 {
		return nil, appErrors.AlreadyExistsError(string(op), "IssueRepository", issueRepository.Name)
	}

	newIssueRepository, err := ir.DB().CreateIssueRepository(issueRepository)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "IssueRepository", "", err)
	}

	ir.PushEvent(&CreateIssueRepositoryEvent{IssueRepository: newIssueRepository})

	return newIssueRepository, nil
}

func (ir *issueRepositoryHandler) UpdateIssueRepository(
	ctx context.Context,
	issueRepository *entity.IssueRepository,
) (*entity.IssueRepository, error) {
	op := appErrors.CallerOp()
	id := strconv.FormatInt(issueRepository.Id, 10)

	var err error

	issueRepository.BaseIssueRepository.UpdatedBy, err = common.GetCurrentUserId(ctx, ir.DB())
	if err != nil {
		return nil, appErrors.InternalError(string(op), "IssueRepository", id, err)
	}

	if err = ir.DB().UpdateIssueRepository(issueRepository); err != nil {
		return nil, appErrors.InternalError(string(op), "IssueRepository", id, err)
	}

	result, err := ir.ListIssueRepositories(
		ctx,
		&entity.IssueRepositoryFilter{Id: []*int64{&issueRepository.Id}},
		entity.NewListOptions(),
	)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "IssueRepository", id, err)
	}

	if len(result.Elements) != 1 {
		return nil, appErrors.E(op, "IssueRepository", id, appErrors.Internal,
			fmt.Sprintf("unexpected result count: %d", len(result.Elements)))
	}

	ir.PushEvent(&UpdateIssueRepositoryEvent{IssueRepository: issueRepository})

	return result.Elements[0].IssueRepository, nil
}

func (ir *issueRepositoryHandler) DeleteIssueRepository(ctx context.Context, id int64) error {
	return ir.Delete(ctx, id)
}

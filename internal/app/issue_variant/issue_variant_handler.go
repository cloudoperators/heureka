// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue_variant

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/issue_repository"
	"github.com/cloudoperators/heureka/internal/entity"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/samber/lo"
)

type issueVariantHandler struct {
	common.BaseHandler[entity.IssueVariantResult, *entity.IssueVariantFilter]
	repositoryService issue_repository.IssueRepositoryHandler
}

func NewIssueVariantHandler(
	handlerContext common.HandlerContext,
	rs issue_repository.IssueRepositoryHandler,
) IssueVariantHandler {
	return &issueVariantHandler{
		BaseHandler: common.NewBaseHandler(handlerContext, common.BaseConfig[entity.IssueVariantResult, *entity.IssueVariantFilter]{
			Op:           appErrors.Op("issueVariantHandler"),
			Entity:       "IssueVariants",
			CursorEntity: "IssueVariantCursors",
			CountEntity:  "IssueVariantCount",
			GetFn:        handlerContext.DB.GetIssueVariants,
			CursorsFn:    handlerContext.DB.GetAllIssueVariantCursors,
			CountFn:      handlerContext.DB.CountIssueVariants,
			ListEventFn: func(f *entity.IssueVariantFilter, o *entity.ListOptions, r *entity.List[entity.IssueVariantResult]) any {
				return &ListIssueVariantsEvent{Filter: f, Options: o, Results: r}
			},
			DeleteFn:      handlerContext.DB.DeleteIssueVariant,
			DeleteEventFn: func(id int64) any { return &DeleteIssueVariantEvent{IssueVariantID: id} },
		}),
		repositoryService: rs,
	}
}

func (iv *issueVariantHandler) ListIssueVariants(
	ctx context.Context,
	filter *entity.IssueVariantFilter,
	options *entity.ListOptions,
) (*entity.List[entity.IssueVariantResult], error) {
	return iv.List(ctx, appErrors.CallerOp(), filter, options)
}

func (iv *issueVariantHandler) ListEffectiveIssueVariants(
	ctx context.Context,
	filter *entity.IssueVariantFilter,
	options *entity.ListOptions,
) (*entity.List[entity.IssueVariantResult], error) {
	op := appErrors.CallerOp()

	issueVariants, err := iv.ListIssueVariants(ctx, filter, options)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "IssueVariants", "", err)
	}

	repositoryIds := lo.Map(
		issueVariants.Elements,
		func(item entity.IssueVariantResult, _ int) *int64 {
			return &item.IssueRepositoryId
		},
	)

	repositories, err := iv.repositoryService.ListIssueRepositories(
		ctx,
		&entity.IssueRepositoryFilter{Id: repositoryIds},
		entity.NewListOptions(),
	)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "IssueRepositories", "", err)
	}

	maxPriorityIr := lo.MaxBy(
		repositories.Elements,
		func(item entity.IssueRepositoryResult, max entity.IssueRepositoryResult) bool {
			return item.Priority > max.Priority
		},
	)

	maxRepositoryIds := lo.FilterMap(
		repositories.Elements,
		func(item entity.IssueRepositoryResult, index int) (int64, bool) {
			if item.Priority == maxPriorityIr.Priority {
				return item.Id, true
			}

			return 0, false
		},
	)

	maxPriorityVariants := lo.Filter(
		issueVariants.Elements,
		func(item entity.IssueVariantResult, _ int) bool {
			return lo.Contains(maxRepositoryIds, item.IssueRepositoryId)
		},
	)

	ret := &entity.List[entity.IssueVariantResult]{Elements: maxPriorityVariants}
	iv.PushEvent(&ListEffectiveIssueVariantsEvent{Filter: filter, Options: options, Results: ret})

	return ret, nil
}

func (iv *issueVariantHandler) CreateIssueVariant(
	ctx context.Context,
	issueVariant *entity.IssueVariant,
) (*entity.IssueVariant, error) {
	op := appErrors.CallerOp()

	var err error

	issueVariant.CreatedBy, err = common.GetCurrentUserId(ctx, iv.DB())
	if err != nil {
		return nil, appErrors.InternalError(string(op), "IssueVariant", "", err)
	}

	issueVariant.UpdatedBy = issueVariant.CreatedBy

	existing, err := iv.ListIssueVariants(
		ctx,
		&entity.IssueVariantFilter{SecondaryName: []*string{&issueVariant.SecondaryName}},
		entity.NewListOptions(),
	)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "IssueVariant", "", err)
	}

	if len(existing.Elements) > 0 {
		return nil, appErrors.AlreadyExistsError(string(op), "IssueVariant", issueVariant.SecondaryName)
	}

	newIv, err := iv.DB().CreateIssueVariant(issueVariant)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "IssueVariant", "", err)
	}

	iv.PushEvent(&CreateIssueVariantEvent{IssueVariant: newIv})

	return newIv, nil
}

func (iv *issueVariantHandler) UpdateIssueVariant(
	ctx context.Context,
	issueVariant *entity.IssueVariant,
) (*entity.IssueVariant, error) {
	op := appErrors.CallerOp()
	id := strconv.FormatInt(issueVariant.Id, 10)

	var err error

	issueVariant.UpdatedBy, err = common.GetCurrentUserId(ctx, iv.DB())
	if err != nil {
		return nil, appErrors.InternalError(string(op), "IssueVariant", id, err)
	}

	if err = iv.DB().UpdateIssueVariant(issueVariant); err != nil {
		return nil, appErrors.InternalError(string(op), "IssueVariant", id, err)
	}

	result, err := iv.ListIssueVariants(
		ctx,
		&entity.IssueVariantFilter{Id: []*int64{&issueVariant.Id}},
		entity.NewListOptions(),
	)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "IssueVariant", id, err)
	}

	if len(result.Elements) != 1 {
		return nil, appErrors.E(op, "IssueVariant", id, appErrors.Internal,
			fmt.Sprintf("unexpected result count: %d", len(result.Elements)))
	}

	iv.PushEvent(&UpdateIssueVariantEvent{IssueVariant: result.Elements[0].IssueVariant})

	return result.Elements[0].IssueVariant, nil
}

func (iv *issueVariantHandler) DeleteIssueVariant(ctx context.Context, id int64) error {
	return iv.Delete(ctx, id)
}

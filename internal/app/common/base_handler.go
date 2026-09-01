// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/cloudoperators/heureka/internal/app/event"
	applog "github.com/cloudoperators/heureka/internal/app/logging"
	"github.com/cloudoperators/heureka/internal/cache"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/cloudoperators/heureka/internal/openfga"
	"github.com/sirupsen/logrus"
)

var DefaultCacheTTL = 12 * time.Hour

// BaseConfig holds the DB functions and metadata required for a standard List operation.
// Entity-specific handlers fill this at construction time; zero TTLs fall back to DefaultCacheTTL.
type BaseConfig[R entity.HasCursor, F entity.Filter] struct {
	// Op is the operation name embedded in structured errors, e.g. "commentHandler.ListComments".
	Op appErrors.Op
	// Entity is the entity name for error context, e.g. "Comments".
	Entity string
	// CursorEntity is the entity name used when the cursor fetch fails, e.g. "CommentCursors".
	CursorEntity string
	// CountEntity is the entity name used when the count fetch fails, e.g. "CommentCount".
	CountEntity string

	GetFn     func(context.Context, F, []entity.Order) ([]R, error)
	CursorsFn func(context.Context, F, []entity.Order) ([]string, error)
	CountFn   func(context.Context, F) (int64, error)

	GetTTL     time.Duration
	CursorsTTL time.Duration
	CountTTL   time.Duration

	ListEventFn   func(filter F, options *entity.ListOptions, result *entity.List[R]) any
	DeleteFn      func(id, userId int64) error
	DeleteEventFn func(id int64) any

	Authz                 openfga.Authorization
	AuthzObjectType       openfga.ObjectType
	AuthzApplyFn          func(filter F, accessibleIds []*int64)
	GetWithAggregationsFn func(context.Context, F, []entity.Order) ([]R, error)
}

// BaseHandler provides the canonical List implementation shared by all entity handlers.
// Entity-specific handlers embed BaseHandler and add only their unique methods
// (relationship operations, custom list variants, entity-specific authz scoping).
type BaseHandler[R entity.HasCursor, F entity.Filter] struct {
	db            database.Database
	eventRegistry event.EventRegistry
	cache         cache.Cache
	logger        *logrus.Logger
	cfg           BaseConfig[R, F]
}

func NewBaseHandler[R entity.HasCursor, F entity.Filter](hc HandlerContext, cfg BaseConfig[R, F]) BaseHandler[R, F] {
	if cfg.GetTTL == 0 {
		cfg.GetTTL = DefaultCacheTTL
	}

	if cfg.CursorsTTL == 0 {
		cfg.CursorsTTL = DefaultCacheTTL
	}

	if cfg.CountTTL == 0 {
		cfg.CountTTL = DefaultCacheTTL
	}

	return BaseHandler[R, F]{
		db:            hc.DB,
		eventRegistry: hc.EventReg,
		cache:         hc.Cache,
		logger:        logrus.StandardLogger(),
		cfg:           cfg,
	}
}

// DB exposes the database for entity-specific methods that need direct DB access.
func (b *BaseHandler[R, F]) DB() database.Database { return b.db }

// Cache exposes the cache for entity-specific methods that perform cached calls.
func (b *BaseHandler[R, F]) Cache() cache.Cache { return b.cache }

// Authz exposes the configured authorization for entity-specific per-item permission checks.
func (b *BaseHandler[R, F]) Authz() openfga.Authorization { return b.cfg.Authz }

// PushEvent forwards an event to the registry.
func (b *BaseHandler[R, F]) PushEvent(e event.Event) { b.eventRegistry.PushEvent(e) }

// List is the canonical paginated list implementation for any entity.
// Entity-specific ListXxx methods call this, passing appErrors.CallerOp(), and then push their domain event.
func (b *BaseHandler[R, F]) List(
	ctx context.Context, op appErrors.Op, filter F, options *entity.ListOptions,
) (*entity.List[R], error) {
	options = EnsureListOptions(options)

	EnsurePaginated(filter.GetPaginated())

	if b.cfg.Authz != nil && b.cfg.AuthzApplyFn != nil {
		currentUserId, err := GetCurrentUserId(ctx, b.db)
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), b.cfg.Entity, "", err)
			applog.LogError(b.logger, wrappedErr, logrus.Fields{"filter": filter})

			return nil, wrappedErr
		}

		accessibleIds, err := b.cfg.Authz.GetListOfAccessibleObjectIds(
			openfga.UserId(fmt.Sprint(currentUserId)),
			b.cfg.AuthzObjectType,
		)
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), b.cfg.Entity, "", err)
			applog.LogError(b.logger, wrappedErr, logrus.Fields{"filter": filter})

			return nil, wrappedErr
		}

		b.cfg.AuthzApplyFn(filter, accessibleIds)
	}

	getFn := b.cfg.GetFn

	cacheKey := "Get" + b.cfg.Entity
	if options.IncludeAggregations && b.cfg.GetWithAggregationsFn != nil {
		getFn = b.cfg.GetWithAggregationsFn
		cacheKey = "Get" + b.cfg.Entity + "WithAggregations"
	}

	res, err := cache.CallCached[[]R](b.cache, cache.NewCacheCallParams(
		b.cfg.GetTTL, ctx, cacheKey, getFn, filter, options.Order,
	))
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), b.cfg.Entity, "", err)
		applog.LogError(b.logger, wrappedErr, logrus.Fields{"filter": filter})

		return nil, wrappedErr
	}

	var (
		count    int64
		pageInfo *entity.PageInfo
	)

	if options.ShowPageInfo && len(res) > 0 {
		cursors, err := cache.CallCached[[]string](b.cache, cache.NewCacheCallParams(
			b.cfg.CursorsTTL, ctx, "GetAll"+b.cfg.CursorEntity, b.cfg.CursorsFn, filter, options.Order,
		))
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), b.cfg.CursorEntity, "", err)
			applog.LogError(b.logger, wrappedErr, logrus.Fields{"filter": filter})

			return nil, wrappedErr
		}

		pageInfo = GetPageInfo(res, cursors, *filter.GetPaginated().First, filter.GetPaginated().After)
		count = int64(len(cursors))
	} else if options.ShowTotalCount {
		count, err = cache.CallCached[int64](b.cache, cache.NewCacheCallParams(
			b.cfg.CountTTL, ctx, "Count"+b.cfg.Entity, b.cfg.CountFn, filter,
		))
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), b.cfg.CountEntity, "", err)
			applog.LogError(b.logger, wrappedErr, logrus.Fields{"filter": filter})

			return nil, wrappedErr
		}
	}

	result := &entity.List[R]{TotalCount: &count, PageInfo: pageInfo, Elements: res}

	if b.cfg.ListEventFn != nil {
		if e, ok := b.cfg.ListEventFn(filter, options, result).(event.Event); ok {
			b.PushEvent(e)
		}
	}

	return result, nil
}

// Delete is the canonical delete implementation: gets the current user, calls DeleteFn, and pushes DeleteEventFn.
func (b *BaseHandler[R, F]) Delete(ctx context.Context, id int64) error {
	op := b.cfg.Op
	idStr := strconv.FormatInt(id, 10)

	userId, err := GetCurrentUserId(ctx, b.db)
	if err != nil {
		return appErrors.InternalError(string(op), b.cfg.Entity, idStr, err)
	}

	if err = b.cfg.DeleteFn(id, userId); err != nil {
		return appErrors.InternalError(string(op), b.cfg.Entity, idStr, err)
	}

	if b.cfg.DeleteEventFn != nil {
		if e, ok := b.cfg.DeleteEventFn(id).(event.Event); ok {
			b.PushEvent(e)
		}
	}

	return nil
}

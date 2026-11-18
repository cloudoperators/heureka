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
	AuthzWriteObjectType  openfga.ObjectType
	AuthzSystemAdminRel   openfga.RelationType
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

// CheckWritePermission verifies the current user has can_write on the given object. objectEntity
// is used for error context, e.g. "Service" when writing to a service, or "Component" when
// writing to a parent component.
func (b *BaseHandler[R, F]) CheckWritePermission(
	ctx context.Context, op appErrors.Op, objectEntity string, objectType openfga.ObjectType, objectId int64,
) error {
	return b.CheckPermission(ctx, op, objectEntity, openfga.RelCanWrite, objectType, objectId)
}

// CheckPermission checks whether the current user has the given relation on (objectType, objectId).
// Falls back from the numeric DB id to the unique user id the same way CheckWritePermission does,
// so entity-specific handlers get the same scanner/service account support instead of
// re-implementing the check ad hoc.
func (b *BaseHandler[R, F]) CheckPermission(
	ctx context.Context, op appErrors.Op, objectEntity string, relation openfga.RelationType, objectType openfga.ObjectType, objectId int64,
) error {
	idStr := strconv.FormatInt(objectId, 10)

	hasPermission, err := b.checkPermission(ctx, op, objectEntity, relation, objectType, openfga.ObjectId(idStr))
	if err != nil {
		return err
	}

	if !hasPermission {
		return appErrors.PermissionDeniedError(string(op), objectEntity, idStr)
	}

	return nil
}

// checkRelation performs a single OpenFGA check for a [user] subject, wrapping any transport
// error consistently. This is the one place a RelationInput for a user subject is built.
func (b *BaseHandler[R, F]) checkRelation(
	op appErrors.Op, objectEntity string, userId openfga.UserId, relation openfga.RelationType, objectType openfga.ObjectType, objectId openfga.ObjectId,
) (bool, error) {
	hasPermission, err := b.cfg.Authz.CheckPermission(openfga.RelationInput{
		UserType:   openfga.TypeUser,
		UserId:     userId,
		Relation:   relation,
		ObjectType: objectType,
		ObjectId:   objectId,
	})
	if err != nil {
		return false, appErrors.InternalError(string(op), objectEntity, string(objectId), err)
	}

	return hasPermission, nil
}

// checkPermission checks relation on (objectType, objectId) using the current user's numeric DB
// id first (users such as owners/members), but will fall back to their unique user id
// for cases where permissions are granted via roles or service accounts.
func (b *BaseHandler[R, F]) checkPermission(
	ctx context.Context, op appErrors.Op, objectEntity string, relation openfga.RelationType, objectType openfga.ObjectType, objectId openfga.ObjectId,
) (bool, error) {
	currentUserId, err := GetCurrentUserId(ctx, b.db)
	if err != nil {
		return false, appErrors.InternalError(string(op), objectEntity, string(objectId), err)
	}

	hasPermission, err := b.checkRelation(op, objectEntity, openfga.UserId(fmt.Sprint(currentUserId)), relation, objectType, objectId)
	if err != nil {
		return false, err
	}

	if hasPermission {
		return true, nil
	}

	uniqueUserId, err := GetCurrentUniqueUserId(ctx)
	if err != nil {
		return false, appErrors.InternalError(string(op), objectEntity, string(objectId), err)
	}

	return b.checkRelation(op, objectEntity, openfga.UserId(uniqueUserId), relation, objectType, objectId)
}

// CheckSystemPermission verifies the current user has the given relation on the system role,
// e.g. RelAdmin for full system admins, or a model-defined relation covering additional
// service accounts/roles permitted to perform the same root-level operation.
func (b *BaseHandler[R, F]) CheckSystemPermission(ctx context.Context, op appErrors.Op, relation openfga.RelationType) error {
	hasPermission, err := b.checkPermission(ctx, op, b.cfg.Entity, relation, openfga.TypeRole, openfga.SystemRoleId)
	if err != nil {
		return err
	}

	if !hasPermission {
		return appErrors.PermissionDeniedError(string(op), b.cfg.Entity, "")
	}

	return nil
}

// checkSystemRolePermission verifies the given OpenFGA user id has the given relation on the
// system role.
func (b *BaseHandler[R, F]) checkSystemRolePermission(op appErrors.Op, userId openfga.UserId, relation openfga.RelationType) error {
	hasPermission, err := b.checkRelation(op, b.cfg.Entity, userId, relation, openfga.TypeRole, openfga.SystemRoleId)
	if err != nil {
		return err
	}

	if !hasPermission {
		return appErrors.PermissionDeniedError(string(op), b.cfg.Entity, "")
	}

	return nil
}

// CheckSystemAdminPermission checks if user is admin derived from the relation in the role tied to a tuple.
// This is used for creating root-level entities.
func (b *BaseHandler[R, F]) CheckSystemAdminPermission(ctx context.Context, op appErrors.Op) error {
	relation := b.cfg.AuthzSystemAdminRel
	if relation == "" {
		relation = openfga.RelAdmin
	}

	// Check if the user has the system-admin relation on the system role first.
	// Uses GetCurrentUserId to determine the current user
	if err := b.CheckSystemPermission(ctx, op, relation); err == nil {
		return nil
	} else if !appErrors.Is(err, appErrors.PermissionDenied) {
		return err
	}

	uniqueUserId, err := GetCurrentUniqueUserId(ctx)
	if err != nil {
		return appErrors.InternalError(string(op), b.cfg.Entity, "", err)
	}

	// Check if the user has the system-admin relation on the system role.
	// Uses GetCurrentUniqueUserId from token context to determine the current user/scanner identity.
	if err := b.checkSystemRolePermission(op, openfga.UserId(uniqueUserId), openfga.RelationType(relation)); err != nil {
		return err
	}

	return nil
}

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

		// Role-based grants (admin/scanner/service accounts) are keyed on the unique user id
		// rather than the numeric DB id, so a user can have access under either or both - union
		// the two result sets rather than only falling back when the numeric id has none, since
		// partial direct access under the numeric id would otherwise mask role-based access.
		uniqueUserId, err := GetCurrentUniqueUserId(ctx)
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), b.cfg.Entity, "", err)
			applog.LogError(b.logger, wrappedErr, logrus.Fields{"filter": filter})

			return nil, wrappedErr
		}

		if uniqueUserId != fmt.Sprint(currentUserId) {
			uniqueAccessibleIds, err := b.cfg.Authz.GetListOfAccessibleObjectIds(
				openfga.UserId(uniqueUserId),
				b.cfg.AuthzObjectType,
			)
			if err != nil {
				wrappedErr := appErrors.InternalError(string(op), b.cfg.Entity, "", err)
				applog.LogError(b.logger, wrappedErr, logrus.Fields{"filter": filter})

				return nil, wrappedErr
			}

			accessibleIds = unionAccessibleIds(accessibleIds, uniqueAccessibleIds)
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

	if b.cfg.Authz != nil && b.cfg.AuthzWriteObjectType != "" {
		if err := b.CheckWritePermission(ctx, op, b.cfg.Entity, b.cfg.AuthzWriteObjectType, id); err != nil {
			return err
		}
	}

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

// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"sync"
	"time"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/component"
	"github.com/cloudoperators/heureka/internal/app/component_instance"
	"github.com/cloudoperators/heureka/internal/app/component_version"
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/app/issue"
	"github.com/cloudoperators/heureka/internal/app/issue_match"
	"github.com/cloudoperators/heureka/internal/app/issue_repository"
	"github.com/cloudoperators/heureka/internal/app/issue_variant"
	"github.com/cloudoperators/heureka/internal/app/patch"
	"github.com/cloudoperators/heureka/internal/app/profiler"
	"github.com/cloudoperators/heureka/internal/app/remediation"
	"github.com/cloudoperators/heureka/internal/app/scanner_run"
	"github.com/cloudoperators/heureka/internal/app/service"
	"github.com/cloudoperators/heureka/internal/app/severity"
	"github.com/cloudoperators/heureka/internal/app/support_group"
	"github.com/cloudoperators/heureka/internal/app/user"
	"github.com/cloudoperators/heureka/internal/cache"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/openfga"
	"github.com/cloudoperators/heureka/internal/util"
)

type HeurekaApp struct {
	component_instance.ComponentInstanceHandler
	component_version.ComponentVersionHandler
	component.ComponentHandler
	issue_match.IssueMatchHandler
	issue_repository.IssueRepositoryHandler
	issue_variant.IssueVariantHandler
	issue.IssueHandler
	scanner_run.ScannerRunHandler
	service.ServiceHandler
	severity.SeverityHandler
	support_group.SupportGroupHandler
	user.UserHandler
	remediation.RemediationHandler
	patch.PatchHandler

	authz openfga.Authorization

	eventRegistry event.EventRegistry
	database      database.Database

	cache cache.Cache

	ctx context.Context
	wg  *sync.WaitGroup

	profiler *profiler.Profiler
}

func NewHeurekaApp(ctx context.Context, wg *sync.WaitGroup, db database.Database, cfg util.Config) *HeurekaApp {
	cache := NewAppCache(ctx, wg, cfg)
	enableLogs := true

	authz := openfga.NewAuthorizationHandler(&cfg, enableLogs)

	profiler := profiler.NewProfiler(cfg.CpuProfilerFilePath)
	profiler.Start()

	er := event.NewEventRegistry(db, authz)

	handlerContext := common.HandlerContext{
		DB:       db,
		EventReg: er,
		Cache:    cache,
		Authz:    authz,
	}

	rh := issue_repository.NewIssueRepositoryHandler(handlerContext)
	ivh := issue_variant.NewIssueVariantHandler(handlerContext, rh)
	sh := severity.NewSeverityHandler(handlerContext, ivh)

	er.Run(ctx)

	heureka := &HeurekaApp{
		ComponentHandler:         component.NewComponentHandler(handlerContext),
		ComponentInstanceHandler: component_instance.NewComponentInstanceHandler(handlerContext),
		ComponentVersionHandler:  component_version.NewComponentVersionHandler(handlerContext),
		IssueHandler:             issue.NewIssueHandler(handlerContext),
		IssueMatchHandler:        issue_match.NewIssueMatchHandler(handlerContext, sh),
		IssueRepositoryHandler:   rh,
		IssueVariantHandler:      ivh,
		ScannerRunHandler:        scanner_run.NewScannerRunHandler(handlerContext),
		ServiceHandler:           service.NewServiceHandler(handlerContext),
		SeverityHandler:          sh,
		SupportGroupHandler:      support_group.NewSupportGroupHandler(handlerContext),
		UserHandler:              user.NewUserHandler(handlerContext),
		RemediationHandler:       remediation.NewRemediationHandler(handlerContext),
		PatchHandler:             patch.NewPatchHandler(handlerContext),
		eventRegistry:            handlerContext.EventReg,
		database:                 handlerContext.DB,
		cache:                    handlerContext.Cache,
		ctx:                      ctx,
		authz:                    handlerContext.Authz,
		wg:                       wg,
		profiler:                 profiler,
	}

	heureka.SubscribeHandlers()
	heureka.SubscribeAuthzHandlers()
	return heureka
}

func NewAppCache(ctx context.Context, wg *sync.WaitGroup, cfg util.Config) cache.Cache {
	var cacheConfig interface{}
	if cfg.CacheEnable == true {
		cacheBaseConfig := cache.CacheConfig{
			MonitorInterval:          time.Duration(cfg.CacheMonitorMSec) * time.Millisecond,
			MaxDbConcurrentRefreshes: cfg.CacheMaxDbConcurrentRefreshes,
			ThrottleInterval:         time.Duration(cfg.CacheThrottleIntervalMSec) * time.Millisecond,
			ThrottlePerInterval:      cfg.CacheThrottlePerInterval,
		}
		if cfg.CacheValkeyUrl != "" {
			cacheConfig = cache.ValkeyCacheConfig{
				Url:         cfg.CacheValkeyUrl,
				Username:    cfg.CacheValkeyUsername,
				Password:    cfg.CacheValkeyPassword,
				ClientName:  cfg.CacheValkeyClientName,
				CacheConfig: cacheBaseConfig,
			}
		} else {
			cacheConfig = cache.InMemoryCacheConfig{CacheConfig: cacheBaseConfig}
		}
	}
	return cache.NewCache(ctx, wg, cacheConfig)
}

func (h *HeurekaApp) SubscribeHandlers() {
	// Register handlers to follow up on create events and link entities together
	handlers := []struct {
		eventName event.EventName
		handler   event.EventHandlerFunc
	}{
		{component_instance.CreateComponentInstanceEventName, event.EventHandlerFunc(issue_match.OnComponentInstanceCreate)},
		{service.CreateServiceEventName, event.EventHandlerFunc(service.OnServiceCreate)},
		{issue_repository.CreateIssueRepositoryEventName, event.EventHandlerFunc(issue_repository.OnIssueRepositoryCreate)},
		{issue.AddComponentVersionToIssueEventName, event.EventHandlerFunc(issue.OnComponentVersionAttachmentToIssue)},
	}

	for _, hdl := range handlers {
		h.eventRegistry.RegisterEventHandler(hdl.eventName, hdl.handler)
	}
}

func (h *HeurekaApp) SubscribeAuthzHandlers() {
	// Register handlers to update, create, and delete authz relations in openfga
	authzHandlers := []struct {
		eventName event.EventName
		handler   event.EventHandlerFunc
	}{
		// Create events
		{service.CreateServiceEventName, event.EventHandlerFunc(service.OnServiceCreateAuthz)},
		{component_instance.CreateComponentInstanceEventName, event.EventHandlerFunc(component_instance.OnComponentInstanceCreateAuthz)},
		{component_version.CreateComponentVersionEventName, event.EventHandlerFunc(component_version.OnComponentVersionCreateAuthz)},
		{support_group.CreateSupportGroupEventName, event.EventHandlerFunc(support_group.OnSupportGroupCreateAuthz)},
		{component.CreateComponentEventName, event.EventHandlerFunc(component.OnComponentCreateAuthz)},
		{issue_match.CreateIssueMatchEventName, event.EventHandlerFunc(issue_match.OnIssueMatchCreateAuthz)},
		// Delete events
		{user.DeleteUserEventName, event.EventHandlerFunc(user.OnUserDeleteAuthz)},
		{service.DeleteServiceEventName, event.EventHandlerFunc(service.OnServiceDeleteAuthz)},
		{component_instance.DeleteComponentInstanceEventName, event.EventHandlerFunc(component_instance.OnComponentInstanceDeleteAuthz)},
		{component_version.DeleteComponentVersionEventName, event.EventHandlerFunc(component_version.OnComponentVersionDeleteAuthz)},
		{support_group.DeleteSupportGroupEventName, event.EventHandlerFunc(support_group.OnSupportGroupDeleteAuthz)},
		{component.DeleteComponentEventName, event.EventHandlerFunc(component.OnComponentDeleteAuthz)},
		{issue_match.DeleteIssueMatchEventName, event.EventHandlerFunc(issue_match.OnIssueMatchDeleteAuthz)},
		// Update events
		{component_version.UpdateComponentVersionEventName, event.EventHandlerFunc(component_version.OnComponentVersionUpdateAuthz)},
		{issue_match.UpdateIssueMatchEventName, event.EventHandlerFunc(issue_match.OnIssueMatchUpdateAuthz)},
		{component_instance.UpdateComponentInstanceEventName, event.EventHandlerFunc(component_instance.OnComponentInstanceUpdateAuthz)},
		{support_group.AddServiceToSupportGroupEventName, event.EventHandlerFunc(support_group.OnAddServiceToSupportGroup)},
		{support_group.RemoveServiceFromSupportGroupEventName, event.EventHandlerFunc(support_group.OnRemoveServiceFromSupportGroup)},
		{support_group.AddUserToSupportGroupEventName, event.EventHandlerFunc(support_group.OnAddUserToSupportGroup)},
		{support_group.RemoveUserFromSupportGroupEventName, event.EventHandlerFunc(support_group.OnRemoveUserFromSupportGroup)},
		{service.AddOwnerToServiceEventName, event.EventHandlerFunc(service.OnAddOwnerToService)},
		{service.RemoveOwnerFromServiceEventName, event.EventHandlerFunc(service.OnRemoveOwnerFromService)},
	}

	for _, handler := range authzHandlers {
		h.eventRegistry.RegisterEventHandler(handler.eventName, handler.handler)
	}
}

func (h *HeurekaApp) Shutdown() error {
	h.profiler.Stop()
	return h.database.CloseConnection()
}

func (h HeurekaApp) GetCache() cache.Cache {
	return h.cache
}

func (h HeurekaApp) WaitPostMigrations() error {
	return h.database.WaitPostMigrations()
}

// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"sync"
	"time"

	"github.com/cloudoperators/heureka/internal/app/activity"
	"github.com/cloudoperators/heureka/internal/app/component"
	"github.com/cloudoperators/heureka/internal/app/component_instance"
	"github.com/cloudoperators/heureka/internal/app/component_version"
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/app/evidence"
	"github.com/cloudoperators/heureka/internal/app/issue"
	"github.com/cloudoperators/heureka/internal/app/issue_match"
	"github.com/cloudoperators/heureka/internal/app/issue_match_change"
	"github.com/cloudoperators/heureka/internal/app/issue_repository"
	"github.com/cloudoperators/heureka/internal/app/issue_variant"
	"github.com/cloudoperators/heureka/internal/app/profiler"
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
	activity.ActivityHandler
	component_instance.ComponentInstanceHandler
	component_version.ComponentVersionHandler
	component.ComponentHandler
	evidence.EvidenceHandler
	issue_match_change.IssueMatchChangeHandler
	issue_match.IssueMatchHandler
	issue_repository.IssueRepositoryHandler
	issue_variant.IssueVariantHandler
	issue.IssueHandler
	scanner_run.ScannerRunHandler
	service.ServiceHandler
	severity.SeverityHandler
	support_group.SupportGroupHandler
	user.UserHandler

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
	rh := issue_repository.NewIssueRepositoryHandler(db, er, authz)
	ivh := issue_variant.NewIssueVariantHandler(db, er, rh, cache, authz)
	sh := severity.NewSeverityHandler(db, er, ivh, authz)

	er.Run(ctx)

	heureka := &HeurekaApp{
		ActivityHandler:          activity.NewActivityHandler(db, er, authz),
		ComponentHandler:         component.NewComponentHandler(db, er, cache, authz),
		ComponentInstanceHandler: component_instance.NewComponentInstanceHandler(db, er, cache, authz),
		ComponentVersionHandler:  component_version.NewComponentVersionHandler(db, er, cache, authz),
		EvidenceHandler:          evidence.NewEvidenceHandler(db, er, authz),
		IssueHandler:             issue.NewIssueHandler(db, er, cache, authz),
		IssueMatchChangeHandler:  issue_match_change.NewIssueMatchChangeHandler(db, er, authz),
		IssueMatchHandler:        issue_match.NewIssueMatchHandler(db, er, sh, cache, authz),
		IssueRepositoryHandler:   rh,
		IssueVariantHandler:      ivh,
		ScannerRunHandler:        scanner_run.NewScannerRunHandler(db, er),
		ServiceHandler:           service.NewServiceHandler(db, er, cache, authz),
		SeverityHandler:          sh,
		SupportGroupHandler:      support_group.NewSupportGroupHandler(db, er, authz),
		UserHandler:              user.NewUserHandler(db, er, authz),
		eventRegistry:            er,
		database:                 db,
		cache:                    cache,
		ctx:                      ctx,
		authz:                    authz,
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
	// Event handlers for Components
	h.eventRegistry.RegisterEventHandler(
		component_instance.CreateComponentInstanceEventName,
		event.EventHandlerFunc(issue_match.OnComponentInstanceCreate),
	)

	// Event handlers for Services
	h.eventRegistry.RegisterEventHandler(
		service.CreateServiceEventName,
		event.EventHandlerFunc(service.OnServiceCreate),
	)

	// Event handlers for IssueRepositories
	h.eventRegistry.RegisterEventHandler(
		issue_repository.CreateIssueRepositoryEventName,
		event.EventHandlerFunc(issue_repository.OnIssueRepositoryCreate),
	)

	// Event handlers for ComponentVersion attachments to Issues
	h.eventRegistry.RegisterEventHandler(
		issue.AddComponentVersionToIssueEventName,
		event.EventHandlerFunc(issue.OnComponentVersionAttachmentToIssue),
	)
}

func (h *HeurekaApp) SubscribeAuthzHandlers() {
	// Authz event handlers for Services
	h.eventRegistry.RegisterEventHandler(
		service.CreateServiceEventName,
		event.EventHandlerFunc(service.OnServiceCreateAuthz),
	)
	// Authz event handlers for ComponentInstances
	h.eventRegistry.RegisterEventHandler(
		component_instance.CreateComponentInstanceEventName,
		event.EventHandlerFunc(component_instance.OnComponentInstanceCreateAuthz),
	)
	// Authz event handlers for ComponentVersions
	h.eventRegistry.RegisterEventHandler(
		component_version.CreateComponentVersionEventName,
		event.EventHandlerFunc(component_version.OnComponentVersionCreateAuthz),
	)
	// Authz event handlers for SupporGroups
	h.eventRegistry.RegisterEventHandler(
		support_group.CreateSupportGroupEventName,
		event.EventHandlerFunc(support_group.OnSupportGroupCreateAuthz),
	)
	// Authz event handlers for Components
	h.eventRegistry.RegisterEventHandler(
		component.CreateComponentEventName,
		event.EventHandlerFunc(component.OnComponentCreateAuthz),
	)
	// Authz event handlers for IssueMatches
	h.eventRegistry.RegisterEventHandler(
		issue_match.CreateIssueMatchEventName,
		event.EventHandlerFunc(issue_match.OnIssueMatchCreateAuthz),
	)
}

func (h *HeurekaApp) Shutdown() error {
	h.profiler.Stop()
	return h.database.CloseConnection()
}

func (h HeurekaApp) GetCache() cache.Cache {
	return h.cache
}

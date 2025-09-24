// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"sync"
	"time"

	"github.com/cloudoperators/heureka/internal/app/activity"
	"github.com/cloudoperators/heureka/internal/app/common"
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

	er := event.NewEventRegistry(db)

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
		ActivityHandler:          activity.NewActivityHandler(handlerContext),
		ComponentHandler:         component.NewComponentHandler(handlerContext),
		ComponentInstanceHandler: component_instance.NewComponentInstanceHandler(handlerContext),
		ComponentVersionHandler:  component_version.NewComponentVersionHandler(handlerContext),
		EvidenceHandler:          evidence.NewEvidenceHandler(handlerContext),
		IssueHandler:             issue.NewIssueHandler(handlerContext),
		IssueMatchChangeHandler:  issue_match_change.NewIssueMatchChangeHandler(handlerContext),
		IssueMatchHandler:        issue_match.NewIssueMatchHandler(handlerContext, sh),
		IssueRepositoryHandler:   rh,
		IssueVariantHandler:      ivh,
		ScannerRunHandler:        scanner_run.NewScannerRunHandler(handlerContext),
		ServiceHandler:           service.NewServiceHandler(handlerContext),
		SeverityHandler:          sh,
		SupportGroupHandler:      support_group.NewSupportGroupHandler(handlerContext),
		UserHandler:              user.NewUserHandler(handlerContext),
		eventRegistry:            handlerContext.EventReg,
		database:                 handlerContext.DB,
		cache:                    handlerContext.Cache,
		ctx:                      ctx,
		authz:                    handlerContext.Authz,
		wg:                       wg,
		profiler:                 profiler,
	}

	heureka.SubscribeHandlers()
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

func (h *HeurekaApp) Shutdown() error {
	h.profiler.Stop()
	return h.database.CloseConnection()
}

func (h HeurekaApp) GetCache() cache.Cache {
	return h.cache
}

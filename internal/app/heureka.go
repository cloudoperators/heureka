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
	"github.com/cloudoperators/heureka/internal/app/scanner_run"
	"github.com/cloudoperators/heureka/internal/app/service"
	"github.com/cloudoperators/heureka/internal/app/severity"
	"github.com/cloudoperators/heureka/internal/app/support_group"
	"github.com/cloudoperators/heureka/internal/app/user"
	"github.com/cloudoperators/heureka/internal/cache"
	"github.com/cloudoperators/heureka/internal/database"
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

	eventRegistry event.EventRegistry
	database      database.Database

	cache cache.Cache

	ctx context.Context
	wg  *sync.WaitGroup
}

func NewHeurekaApp(ctx context.Context, wg *sync.WaitGroup, db database.Database, cfg util.Config) *HeurekaApp {
	er := event.NewEventRegistry(db)
	rh := issue_repository.NewIssueRepositoryHandler(db, er)
	ivh := issue_variant.NewIssueVariantHandler(db, er, rh)
	sh := severity.NewSeverityHandler(db, er, ivh)

	cache := NewAppCache(ctx, wg, cfg)
	er.Run(ctx)

	heureka := &HeurekaApp{
		ActivityHandler:          activity.NewActivityHandler(db, er),
		ComponentHandler:         component.NewComponentHandler(db, er),
		ComponentInstanceHandler: component_instance.NewComponentInstanceHandler(db, er),
		ComponentVersionHandler:  component_version.NewComponentVersionHandler(db, er),
		EvidenceHandler:          evidence.NewEvidenceHandler(db, er),
		IssueHandler:             issue.NewIssueHandler(db, er),
		IssueMatchChangeHandler:  issue_match_change.NewIssueMatchChangeHandler(db, er),
		IssueMatchHandler:        issue_match.NewIssueMatchHandler(db, er, sh),
		IssueRepositoryHandler:   rh,
		IssueVariantHandler:      ivh,
		ScannerRunHandler:        scanner_run.NewScannerRunHandler(db, er),
		ServiceHandler:           service.NewServiceHandler(db, er, cache),
		SeverityHandler:          sh,
		SupportGroupHandler:      support_group.NewSupportGroupHandler(db, er),
		UserHandler:              user.NewUserHandler(db, er),
		eventRegistry:            er,
		database:                 db,
		cache:                    cache,
		ctx:                      ctx,
		wg:                       wg,
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
	return h.database.CloseConnection()
}

func (h HeurekaApp) GetCache() cache.Cache {
	return h.cache
}

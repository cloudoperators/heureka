// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	events "github.com/cloudoperators/heureka/internal/event"
	log "github.com/sirupsen/logrus"

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
	"github.com/cloudoperators/heureka/internal/database"
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
}

// todo: inject event registry
func NewHeurekaApp(db database.Database, er event.EventRegistry) *HeurekaApp {
	rh := issue_repository.NewIssueRepositoryHandler(db, er)
	ivh := issue_variant.NewIssueVariantHandler(db, er, rh)
	sh := severity.NewSeverityHandler(db, er, ivh)
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
		ServiceHandler:           service.NewServiceHandler(db, er),
		SeverityHandler:          sh,
		SupportGroupHandler:      support_group.NewSupportGroupHandler(db, er),
		UserHandler:              user.NewUserHandler(db, er),
		eventRegistry:            er,
		database:                 db,
	}
	return heureka
}

// todo: move to event package of app package
func (h *HeurekaApp) SubscribeHandlers() {

	h.eventRegistry.RegisterEventHandler(
		component_instance.ListComponentInstancesEventName,
		events.EventHandler{
			func(db database.Database, event events.Event) {
				//do nothing
				log.Info("Received ListComponentInstancesEvent and calling handler....")
				if listEvent, ok := event.(*component_instance.ListComponentInstancesEvent); ok {
					log.WithField("event", listEvent).Infof("Marshalled event")
				}
				return
			},
			component_instance.ListComponentInstancesEvent{}.Unmarshal,
		},
	)

	// Event handlers for Components
	h.eventRegistry.RegisterEventHandler(
		component_instance.CreateComponentInstanceEventName,
		events.EventHandler{
			//todo: move handler to component_instance?
			issue_match.OnComponentInstanceCreate,
			component_instance.CreateComponentInstanceEvent{}.Unmarshal,
		},
	)

	// Event handlers for Services
	h.eventRegistry.RegisterEventHandler(
		service.CreateServiceEventName,
		events.EventHandler{
			service.OnServiceCreate,
			service.CreateServiceEvent{}.Unmarshal,
		},
	)

	// Event handlers for IssueRepositories
	h.eventRegistry.RegisterEventHandler(
		issue_repository.CreateIssueRepositoryEventName,
		events.EventHandler{
			issue_repository.OnIssueRepositoryCreate,
			issue_repository.CreateIssueRepositoryEvent{}.Unmarshal,
		},
	)

	// Event handlers for ComponentVersion attachments to Issues
	h.eventRegistry.RegisterEventHandler(
		issue.AddComponentVersionToIssueEventName,
		events.EventHandler{
			issue.OnComponentVersionAttachmentToIssue,
			issue.AddComponentVersionToIssueEvent{}.Unmarshal,
		},
	)
}

func (h *HeurekaApp) Shutdown() error {
	return h.database.CloseConnection()
}

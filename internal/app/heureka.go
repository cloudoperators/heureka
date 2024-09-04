// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"github.wdf.sap.corp/cc/heureka/internal/app/activity"
	"github.wdf.sap.corp/cc/heureka/internal/app/component"
	"github.wdf.sap.corp/cc/heureka/internal/app/component_instance"
	"github.wdf.sap.corp/cc/heureka/internal/app/component_version"
	"github.wdf.sap.corp/cc/heureka/internal/app/event"
	"github.wdf.sap.corp/cc/heureka/internal/app/evidence"
	"github.wdf.sap.corp/cc/heureka/internal/app/issue"
	"github.wdf.sap.corp/cc/heureka/internal/app/issue_match"
	"github.wdf.sap.corp/cc/heureka/internal/app/issue_match_change"
	"github.wdf.sap.corp/cc/heureka/internal/app/issue_repository"
	"github.wdf.sap.corp/cc/heureka/internal/app/issue_variant"
	"github.wdf.sap.corp/cc/heureka/internal/app/service"
	"github.wdf.sap.corp/cc/heureka/internal/app/severity"
	"github.wdf.sap.corp/cc/heureka/internal/app/support_group"
	"github.wdf.sap.corp/cc/heureka/internal/app/user"
	"github.wdf.sap.corp/cc/heureka/internal/database"
)

type HeurekaApp struct {
	activity.ActivityHandler
	component.ComponentHandler
	component_instance.ComponentInstanceHandler
	component_version.ComponentVersionHandler
	evidence.EvidenceHandler
	issue.IssueHandler
	issue_match.IssueMatchHandler
	issue_match_change.IssueMatchChangeHandler
	issue_repository.IssueRepositoryHandler
	issue_variant.IssueVariantHandler
	service.ServiceHandler
	severity.SeverityHandler
	support_group.SupportGroupHandler
	user.UserHandler

	eventRegistry event.EventRegistry
	database      database.Database
}

func NewHeurekaApp(db database.Database) *HeurekaApp {
	er := event.NewEventRegistry(db)
	rh := issue_repository.NewIssueRepositoryHandler(db, er)
	ivh := issue_variant.NewIssueVariantHandler(db, er, rh)
	sh := severity.NewSeverityHandler(db, er, ivh)
	er.Run(context.Background())
	return &HeurekaApp{
		ActivityHandler:          activity.NewActivityHandler(db, er),
		ComponentHandler:         component.NewComponentHandler(db, er),
		ComponentInstanceHandler: component_instance.NewComponentInstanceHandler(db, er),
		ComponentVersionHandler:  component_version.NewComponentVersionHandler(db, er),
		EvidenceHandler:          evidence.NewEvidenceHandler(db, er),
		IssueHandler:             issue.NewIssueHandler(db, er),
		IssueMatchHandler:        issue_match.NewIssueMatchHandler(db, er, sh),
		IssueMatchChangeHandler:  issue_match_change.NewIssueMatchChangeHandler(db, er),
		IssueRepositoryHandler:   rh,
		IssueVariantHandler:      ivh,
		ServiceHandler:           service.NewServiceHandler(db, er),
		SeverityHandler:          sh,
		SupportGroupHandler:      support_group.NewSupportGroupHandler(db, er),
		UserHandler:              user.NewUserHandler(db, er),
		eventRegistry:            er,
		database:                 db,
	}
}

func (h *HeurekaApp) SubscribeHandlers() {
	h.eventRegistry.RegisterEventHandler(
		component_instance.CreateComponentInstanceEventName,
		event.EventHandlerFunc(issue_match.OnComponentInstanceCreate),
	)
}

func (h *HeurekaApp) Shutdown() error {
	return h.database.CloseConnection()
}

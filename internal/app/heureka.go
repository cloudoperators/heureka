// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"
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

type HeurekaError struct {
	msg string
}

func (e *HeurekaError) Error() string {
	return fmt.Sprintf("NewHeurekaError: %s", e.msg)
}

type HeurekaApp struct {
	activity.ActivityService
	component.ComponentService
	component_instance.ComponentInstanceService
	component_version.ComponentVersionService
	evidence.EvidenceService
	issue.IssueService
	issue_match.IssueMatchService
	issue_match_change.IssueMatchChangeService
	issue_repository.IssueRepositoryService
	issue_variant.IssueVariantService
	service.ServiceService
	severity.SeverityService
	support_group.SupportGroupService
	user.UserService

	eventRegistry event.EventRegistry
	database      database.Database
}

func NewHeurekaError(msg string) *HeurekaError {
	return &HeurekaError{msg: msg}
}

func NewHeurekaApp(db database.Database) *HeurekaApp {
	er := event.NewEventRegistry()
	rs := issue_repository.NewIssueRepositoryService(db, er)
	ivs := issue_variant.NewIssueVariantService(db, er, rs)
	ss := severity.NewSeverityService(db, er, ivs)
	return &HeurekaApp{
		ActivityService:          activity.NewActivityService(db, er),
		ComponentService:         component.NewComponentService(db, er),
		ComponentInstanceService: component_instance.NewComponentInstanceService(db, er),
		ComponentVersionService:  component_version.NewComponentVersionService(db, er),
		EvidenceService:          evidence.NewEvidenceService(db, er),
		IssueService:             issue.NewIssueService(db, er),
		IssueMatchService:        issue_match.NewIssueMatchService(db, er, ss),
		IssueMatchChangeService:  issue_match_change.NewIssueMatchChangeService(db, er),
		IssueRepositoryService:   rs,
		IssueVariantService:      ivs,
		ServiceService:           service.NewServiceService(db, er),
		SeverityService:          ss,
		SupportGroupService:      support_group.NewSupportGroupService(db, er),
		UserService:              user.NewUserService(db, er),
		database:                 db,
	}
}

func (h *HeurekaApp) Shutdown() error {
	return h.database.CloseConnection()
}

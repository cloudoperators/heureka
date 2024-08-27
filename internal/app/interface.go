// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"github.wdf.sap.corp/cc/heureka/internal/app/activity"
	"github.wdf.sap.corp/cc/heureka/internal/app/component"
	"github.wdf.sap.corp/cc/heureka/internal/app/component_instance"
	"github.wdf.sap.corp/cc/heureka/internal/app/component_version"
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
)

type Heureka interface {
	issue.IssueHandler
	activity.ActivityHandler
	service.ServiceHandler
	user.UserHandler
	component.ComponentHandler
	component_instance.ComponentInstanceHandler
	component_version.ComponentVersionHandler
	support_group.SupportGroupHandler
	issue_variant.IssueVariantHandler
	issue_repository.IssueRepositoryHandler
	issue_match.IssueMatchHandler
	issue_match_change.IssueMatchChangeHandler
	severity.SeverityHandler
	evidence.EvidenceHandler
	issue_match.IssueMatchHandler

	Shutdown() error
}

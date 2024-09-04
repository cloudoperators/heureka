// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"github.com/cloudoperators/heureka/internal/app/activity"
	"github.com/cloudoperators/heureka/internal/app/component"
	"github.com/cloudoperators/heureka/internal/app/component_instance"
	"github.com/cloudoperators/heureka/internal/app/component_version"
	"github.com/cloudoperators/heureka/internal/app/evidence"
	"github.com/cloudoperators/heureka/internal/app/issue"
	"github.com/cloudoperators/heureka/internal/app/issue_match"
	"github.com/cloudoperators/heureka/internal/app/issue_match_change"
	"github.com/cloudoperators/heureka/internal/app/issue_repository"
	"github.com/cloudoperators/heureka/internal/app/issue_variant"
	"github.com/cloudoperators/heureka/internal/app/service"
	"github.com/cloudoperators/heureka/internal/app/severity"
	"github.com/cloudoperators/heureka/internal/app/support_group"
	"github.com/cloudoperators/heureka/internal/app/user"
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

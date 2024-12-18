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
	"github.com/cloudoperators/heureka/internal/app/scanner_run"
	"github.com/cloudoperators/heureka/internal/app/service"
	"github.com/cloudoperators/heureka/internal/app/severity"
	"github.com/cloudoperators/heureka/internal/app/support_group"
	"github.com/cloudoperators/heureka/internal/app/user"
)

type Heureka interface {
	activity.ActivityHandler
	component_instance.ComponentInstanceHandler
	component_version.ComponentVersionHandler
	component.ComponentHandler
	evidence.EvidenceHandler
	issue_match_change.IssueMatchChangeHandler
	issue_match.IssueMatchHandler
	issue_match.IssueMatchHandler
	issue_repository.IssueRepositoryHandler
	issue_variant.IssueVariantHandler
	issue.IssueHandler
	scanner_run.ScannerRunHandler
	service.ServiceHandler
	severity.SeverityHandler
	support_group.SupportGroupHandler
	user.UserHandler

	Shutdown() error
}

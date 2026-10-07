// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package remediation

import (
	"context"

	"github.com/cloudoperators/heureka/internal/entity"
)

// RemediationReference identifies which named reference on a remediation input
// could not be resolved to exactly one domain entity. The transport layer maps
// these to client-facing messages.
type RemediationReference string

const (
	RemediationReferenceService   RemediationReference = "service"
	RemediationReferenceIssue     RemediationReference = "issue"
	RemediationReferenceComponent RemediationReference = "component"
	RemediationReferenceUser      RemediationReference = "user"
)

// RemediationReferenceError is returned when a remediation input references a
// service/issue/component/user that cannot be resolved to exactly one entity.
type RemediationReferenceError struct {
	Reference RemediationReference
	// Ambiguous is true when a component matched more than one entity.
	Ambiguous bool
	// ComponentType carries the matched component's type when Ambiguous, so the
	// transport layer can render a type-specific message.
	ComponentType string
}

func (e *RemediationReferenceError) Error() string {
	return "remediation " + string(e.Reference) + " reference could not be resolved"
}

// RemediationCreateInput carries the already-mapped remediation entity together
// with the raw name references that must be resolved into foreign keys.
type RemediationCreateInput struct {
	Remediation   *entity.Remediation
	Service       *string
	Vulnerability *string
	Image         *string
	RemediatedBy  *string
}

// RemediationUpdateInput carries the remediation entity (with its Id set) and
// the references to re-resolve. A nil reference means "leave that FK untouched".
type RemediationUpdateInput struct {
	Remediation   *entity.Remediation
	Service       *string
	Vulnerability *string
	Image         *string
}

// RemediationDeps is the subset of application operations the remediation logic
// needs. Both *app.HeurekaApp and the generated MockHeureka satisfy it, which
// keeps this logic unit-testable without a database.
type RemediationDeps interface {
	ListServices(context.Context, *entity.ServiceFilter, *entity.ListOptions) (*entity.List[entity.ServiceResult], error)
	ListIssues(context.Context, *entity.IssueFilter, *entity.IssueListOptions) (*entity.IssueList, error)
	ListComponents(context.Context, *entity.ComponentFilter, *entity.ListOptions) (*entity.List[entity.ComponentResult], error)
	ListUniqueUserIDs(context.Context, *entity.UserFilter, *entity.ListOptions) ([]string, error)
	CreateUser(context.Context, *entity.User) (*entity.User, error)
	CreateRemediation(context.Context, *entity.Remediation) (*entity.Remediation, error)
	UpdateRemediation(context.Context, *entity.Remediation) (*entity.Remediation, error)
}

func resolveServiceID(ctx context.Context, deps RemediationDeps, ccrn *string) (int64, error) {
	services, err := deps.ListServices(ctx, &entity.ServiceFilter{CCRN: []*string{ccrn}}, nil)
	if err != nil || len(services.Elements) != 1 {
		return 0, &RemediationReferenceError{Reference: RemediationReferenceService}
	}

	return services.Elements[0].Id, nil
}

func resolveIssueID(ctx context.Context, deps RemediationDeps, primaryName *string) (int64, error) {
	issues, err := deps.ListIssues(ctx, &entity.IssueFilter{PrimaryName: []*string{primaryName}}, nil)
	if err != nil || len(issues.Elements) != 1 {
		return 0, &RemediationReferenceError{Reference: RemediationReferenceIssue}
	}

	return issues.Elements[0].Issue.Id, nil
}

func resolveComponentID(ctx context.Context, deps RemediationDeps, image, serviceCCRN *string) (int64, error) {
	filter := &entity.ComponentFilter{Repository: []*string{image}}
	if serviceCCRN != nil {
		filter.ServiceCCRN = []*string{serviceCCRN}
	}

	components, err := deps.ListComponents(ctx, filter, nil)
	if err != nil {
		return 0, &RemediationReferenceError{Reference: RemediationReferenceComponent}
	}

	if serviceCCRN != nil && len(components.Elements) == 0 {
		components, err = deps.ListComponents(ctx, &entity.ComponentFilter{Repository: []*string{image}}, nil)
		if err != nil {
			return 0, &RemediationReferenceError{Reference: RemediationReferenceComponent}
		}
	}

	if len(components.Elements) != 1 {
		refErr := &RemediationReferenceError{Reference: RemediationReferenceComponent}
		if len(components.Elements) > 1 {
			refErr.Ambiguous = true
			refErr.ComponentType = components.Elements[0].Type
		}

		return 0, refErr
	}

	return components.Elements[0].Id, nil
}

func resolveRemediatedByUser(ctx context.Context, deps RemediationDeps, uniqueUserID *string) (string, error) {
	uids, err := deps.ListUniqueUserIDs(ctx, &entity.UserFilter{UniqueUserID: []*string{uniqueUserID}}, nil)
	if err != nil {
		return "", &RemediationReferenceError{Reference: RemediationReferenceUser}
	}

	if len(uids) == 0 {
		user, err := deps.CreateUser(ctx, &entity.User{UniqueUserID: *uniqueUserID})
		if err != nil {
			return "", &RemediationReferenceError{Reference: RemediationReferenceUser}
		}

		return user.UniqueUserID, nil
	}

	return *uniqueUserID, nil
}

// CreateFromInput resolves the service/issue/component (and optional remediatedBy
// user) references on the input, sets the foreign keys, and persists the remediation.
func CreateFromInput(ctx context.Context, deps RemediationDeps, in RemediationCreateInput) (*entity.Remediation, error) {
	remediation := in.Remediation

	serviceID, err := resolveServiceID(ctx, deps, in.Service)
	if err != nil {
		return nil, err
	}

	remediation.ServiceId = serviceID

	issueID, err := resolveIssueID(ctx, deps, in.Vulnerability)
	if err != nil {
		return nil, err
	}

	remediation.IssueId = issueID

	componentID, err := resolveComponentID(ctx, deps, in.Image, in.Service)
	if err != nil {
		return nil, err
	}

	remediation.ComponentId = componentID

	if in.RemediatedBy != nil {
		remediatedBy, err := resolveRemediatedByUser(ctx, deps, in.RemediatedBy)
		if err != nil {
			return nil, err
		}

		remediation.RemediatedBy = remediatedBy
	}

	if _, err := deps.CreateRemediation(ctx, remediation); err != nil {
		return nil, err
	}

	return remediation, nil
}

// UpdateFromInput re-resolves only the references that were provided (non-nil)
// on the input, sets the corresponding foreign keys, and persists the update.
func UpdateFromInput(ctx context.Context, deps RemediationDeps, in RemediationUpdateInput) (*entity.Remediation, error) {
	remediation := in.Remediation

	if in.Service != nil {
		serviceID, err := resolveServiceID(ctx, deps, in.Service)
		if err != nil {
			return nil, err
		}

		remediation.ServiceId = serviceID
	}

	if in.Image != nil {
		componentID, err := resolveComponentID(ctx, deps, in.Image, nil)
		if err != nil {
			return nil, err
		}

		remediation.ComponentId = componentID
	}

	if in.Vulnerability != nil {
		issueID, err := resolveIssueID(ctx, deps, in.Vulnerability)
		if err != nil {
			return nil, err
		}

		remediation.IssueId = issueID
	}

	return deps.UpdateRemediation(ctx, remediation)
}

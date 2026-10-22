// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package siem_alert

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/cloudoperators/heureka/internal/entity"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/cloudoperators/heureka/internal/util"
)

// SIEMLink is the domain representation of a link attached to a SIEM alert.
// JSON tags mirror model.SIEMAlertLink so that marshaled ExternalUrl values
// stored in IssueVariant are compatible across the two representations.
type SIEMLink struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// SIEMAlertInput carries the fields needed to create a SIEM alert, expressed
// in domain types so this file has no transport dependency.
type SIEMAlertInput struct {
	Service      *string
	SupportGroup *string
	Region       *string
	Cluster      *string
	Namespace    *string
	Pod          *string
	Container    *string
	Name         *string
	Description  *string
	Severity     *string // rating value, e.g. "High"
	Links        []SIEMLink
	Source       *string // SIEM issue-repository name; defaults to "heureka-siem"
}

// SIEMAlertResult carries the domain entities produced by the logic so the
// transport layer can assemble the model response.
type SIEMAlertResult struct {
	Service           *entity.Service
	SupportGroup      *entity.SupportGroup
	ComponentInstance *entity.ComponentInstance
	Issue             *entity.Issue
	IssueVariant      *entity.IssueVariant
}

// SIEMError is returned for both operational failures and data-state validation
// failures inside the SIEM logic. Message holds the text after the
// resolver-name prefix, matching the client-facing error strings the transport
// assembles.
type SIEMError struct {
	Message string
}

func (e *SIEMError) Error() string { return e.Message }

// SIEMAlertDeps is the subset of app.Heureka required by the SIEM alert logic.
// Both *app.HeurekaApp and the generated MockHeureka satisfy it, so the logic
// is unit-testable without a database.
type SIEMAlertDeps interface {
	ListServices(context.Context, *entity.ServiceFilter, *entity.ListOptions) (*entity.List[entity.ServiceResult], error)
	CreateService(context.Context, *entity.Service) (*entity.Service, error)
	ListSupportGroups(context.Context, *entity.SupportGroupFilter, *entity.ListOptions) (*entity.List[entity.SupportGroupResult], error)
	CreateSupportGroup(context.Context, *entity.SupportGroup) (*entity.SupportGroup, error)
	AddServiceToSupportGroup(context.Context, int64, int64) (*entity.SupportGroup, error)
	CreateComponentInstance(context.Context, *entity.ComponentInstance, *string) (*entity.ComponentInstance, error)
	ListComponentInstances(context.Context, *entity.ComponentInstanceFilter, *entity.ListOptions) (*entity.List[entity.ComponentInstanceResult], error)
	ListIssueVariants(context.Context, *entity.IssueVariantFilter, *entity.ListOptions) (*entity.List[entity.IssueVariantResult], error)
	CreateIssue(context.Context, *entity.Issue) (*entity.Issue, error)
	ListIssues(context.Context, *entity.IssueFilter, *entity.IssueListOptions) (*entity.IssueList, error)
	ListIssueRepositories(context.Context, *entity.IssueRepositoryFilter, *entity.ListOptions) (*entity.List[entity.IssueRepositoryResult], error)
	CreateIssueRepository(context.Context, *entity.IssueRepository) (*entity.IssueRepository, error)
	CreateIssueVariant(context.Context, *entity.IssueVariant) (*entity.IssueVariant, error)
	GetIssue(context.Context, int64) (*entity.Issue, error)
	ListUsers(context.Context, *entity.UserFilter, *entity.ListOptions) (*entity.List[entity.UserResult], error)
	CreateIssueMatch(context.Context, *entity.IssueMatch) (*entity.IssueMatch, error)
	AddComponentVersionToIssue(context.Context, int64, int64) (*entity.Issue, error)
}

func getOrCreateService(ctx context.Context, deps SIEMAlertDeps, inputService *string) (*entity.Service, error) {
	if inputService == nil || *inputService == "" {
		return nil, nil
	}

	filter := entity.ServiceFilter{CCRN: []*string{inputService}}

	s, err := deps.ListServices(ctx, &filter, entity.NewListOptions())
	if err != nil {
		return nil, &SIEMError{Message: "Internal Error - when listing services"}
	}

	if len(s.Elements) > 0 {
		return s.Elements[0].Service, nil
	}

	svcEntity := entity.Service{BaseService: entity.BaseService{CCRN: *inputService}}

	newSvc, err := deps.CreateService(ctx, &svcEntity)
	if err != nil {
		s2, err2 := deps.ListServices(ctx, &filter, entity.NewListOptions())
		if err2 != nil || len(s2.Elements) == 0 {
			return nil, &SIEMError{Message: "Internal Error - when creating service"}
		}

		return s2.Elements[0].Service, nil
	}

	return newSvc, nil
}

func getOrCreateSupportGroup(ctx context.Context, deps SIEMAlertDeps, inputSupportGroup *string) (*entity.SupportGroup, error) {
	if inputSupportGroup == nil || *inputSupportGroup == "" {
		return nil, nil
	}

	filter := entity.SupportGroupFilter{CCRN: []*string{inputSupportGroup}}

	sgList, err := deps.ListSupportGroups(ctx, &filter, entity.NewListOptions())
	if err != nil {
		return nil, &SIEMError{Message: "Internal Error - when listing support groups"}
	}

	if len(sgList.Elements) > 0 {
		return sgList.Elements[0].SupportGroup, nil
	}

	sgEntity := entity.SupportGroup{CCRN: *inputSupportGroup}

	newSg, err := deps.CreateSupportGroup(ctx, &sgEntity)
	if err != nil {
		sg2, err2 := deps.ListSupportGroups(ctx, &filter, entity.NewListOptions())
		if err2 != nil || len(sg2.Elements) == 0 {
			return nil, &SIEMError{Message: "Internal Error - when creating support group"}
		}

		return sg2.Elements[0].SupportGroup, nil
	}

	return newSg, nil
}

func buildCCRN(in SIEMAlertInput) string {
	var parts []string

	for _, p := range []*string{in.Service, in.Region, in.Cluster, in.Namespace, in.Pod, in.Container} {
		if p != nil && *p != "" {
			parts = append(parts, *p)
		}
	}

	return strings.Join(parts, "/")
}

func getOrCreateComponentInstance(ctx context.Context, deps SIEMAlertDeps, ccrn string, svc *entity.Service, in SIEMAlertInput) (*entity.ComponentInstance, error) {
	if ccrn == "" || svc == nil {
		return nil, nil
	}

	ciEntity := entity.ComponentInstance{
		CCRN:      ccrn,
		ServiceId: svc.Id,
		Region:    derefStr(in.Region),
		Cluster:   derefStr(in.Cluster),
		Namespace: derefStr(in.Namespace),
		Pod:       derefStr(in.Pod),
		Container: derefStr(in.Container),
		Type:      entity.ComponentInstanceTypeUnknown,
	}

	newCi, err := deps.CreateComponentInstance(ctx, &ciEntity, nil)
	if err != nil {
		filter := entity.ComponentInstanceFilter{CCRN: []*string{&ccrn}}

		cis, err2 := deps.ListComponentInstances(ctx, &filter, &entity.ListOptions{})
		if err2 != nil || len(cis.Elements) == 0 {
			return nil, &SIEMError{Message: "Internal Error - when creating componentInstance"}
		}

		return cis.Elements[0].ComponentInstance, nil
	}

	return newCi, nil
}

func getOrCreateIssueAndVariant(ctx context.Context, deps SIEMAlertDeps, in SIEMAlertInput) (*entity.Issue, *entity.IssueVariant, error) {
	var (
		issue        *entity.Issue
		issueVariant *entity.IssueVariant
	)

	if len(in.Links) > 0 {
		if in.Name != nil && *in.Name != "" {
			linksJSON, _ := json.Marshal(in.Links)

			ivs, err := deps.ListIssueVariants(
				ctx,
				&entity.IssueVariantFilter{SecondaryName: []*string{in.Name}},
				&entity.ListOptions{},
			)
			if err == nil {
				for _, v := range ivs.Elements {
					if v.ExternalUrl == string(linksJSON) {
						issueVariant = v.IssueVariant
						break
					}
				}
			}
		}
	}

	if issueVariant == nil {
		if in.Name == nil || *in.Name == "" {
			return nil, nil, &SIEMError{Message: "Invalid Input - name or url required"}
		}

		description := ""
		if in.Description != nil {
			description = *in.Description
		}

		newIssue, err := deps.CreateIssue(ctx, &entity.Issue{
			PrimaryName: *in.Name,
			Description: description,
			Type:        entity.IssueTypeSecurityEvent,
		})
		if err != nil {
			f := &entity.IssueFilter{PrimaryName: []*string{in.Name}}
			lo := entity.IssueListOptions{ListOptions: *entity.NewListOptions()}

			issues, ierr := deps.ListIssues(ctx, f, &lo)
			if ierr != nil || len(issues.Elements) == 0 {
				return nil, nil, &SIEMError{Message: "Internal Error - when creating issue"}
			}

			issue = issues.Elements[0].Issue
		} else {
			issue = newIssue
		}

		siemRepoName := "heureka-siem"
		if in.Source != nil && *in.Source != "" {
			siemRepoName = *in.Source
		}

		repositories, err := deps.ListIssueRepositories(
			ctx,
			&entity.IssueRepositoryFilter{Name: []*string{&siemRepoName}},
			&entity.ListOptions{},
		)

		var issueRepositoryID int64
		if err == nil && len(repositories.Elements) > 0 {
			issueRepositoryID = repositories.Elements[0].Id
		} else {
			newRepo := entity.IssueRepository{
				BaseIssueRepository: entity.BaseIssueRepository{Name: siemRepoName},
			}

			createdRepo, err := deps.CreateIssueRepository(ctx, &newRepo)
			if err != nil {
				return nil, nil, &SIEMError{Message: "Internal Error - failed to init SIEM repository"}
			}

			issueRepositoryID = createdRepo.Id
		}

		sev := entity.Severity{}
		if in.Severity != nil {
			sev = entity.NewSeverityFromRating(entity.SeverityValues(*in.Severity))
		}

		iv := entity.IssueVariant{
			SecondaryName:     derefStr(in.Name),
			IssueId:           issue.Id,
			IssueRepositoryId: issueRepositoryID,
			Severity:          sev,
			Description:       description,
			ExternalUrl: func() string {
				if len(in.Links) == 0 {
					return ""
				}

				linksJSON, _ := json.Marshal(in.Links)

				return string(linksJSON)
			}(),
		}

		newIv, err := deps.CreateIssueVariant(ctx, &iv)
		if err != nil {
			return nil, nil, &SIEMError{Message: "Internal Error - when creating issueVariant"}
		}

		issueVariant = newIv
	} else {
		iss, err := deps.GetIssue(ctx, issueVariant.IssueId)
		if err != nil {
			return nil, nil, &SIEMError{Message: "Internal Error - when resolving issue"}
		}

		issue = iss
	}

	return issue, issueVariant, nil
}

func createIssueMatchIfCI(ctx context.Context, deps SIEMAlertDeps, ci *entity.ComponentInstance, issue *entity.Issue) error {
	if ci == nil {
		return nil
	}

	userID := util.SystemUserId

	users, err := deps.ListUsers(ctx, &entity.UserFilter{}, &entity.ListOptions{})
	if err == nil && len(users.Elements) > 0 {
		userID = users.Elements[0].Id
	}

	im := entity.IssueMatch{
		IssueId:               issue.Id,
		ComponentInstanceId:   ci.Id,
		UserId:                userID,
		Status:                entity.IssueMatchStatusValuesNew,
		RemediationDate:       time.Now(),
		TargetRemediationDate: time.Now(),
	}

	_, err = deps.CreateIssueMatch(ctx, &im)
	if err != nil {
		return &SIEMError{Message: "Internal Error - when creating issue match"}
	}

	if ci.ComponentVersionId != 0 {
		_, err = deps.AddComponentVersionToIssue(ctx, issue.Id, ci.ComponentVersionId)
		if err != nil && !appErrors.IsAlreadyExists(err) {
			return &SIEMError{Message: "Internal Error - when linking issue to component version"}
		}
	}

	return nil
}

// CreateFromInput orchestrates the get-or-create operations needed to persist
// a SIEM alert: service, support-group, component instance, issue + variant,
// and an issue-match linking the CI to the issue. It returns the domain
// entities so the transport layer can assemble the model response.
func CreateFromInput(ctx context.Context, deps SIEMAlertDeps, in SIEMAlertInput) (SIEMAlertResult, error) {
	svc, err := getOrCreateService(ctx, deps, in.Service)
	if err != nil {
		return SIEMAlertResult{}, err
	}

	sg, err := getOrCreateSupportGroup(ctx, deps, in.SupportGroup)
	if err != nil {
		return SIEMAlertResult{}, err
	}

	if svc != nil && sg != nil {
		if _, err := deps.AddServiceToSupportGroup(ctx, svc.Id, sg.Id); err != nil {
			return SIEMAlertResult{}, &SIEMError{Message: "Internal Error - when adding service to supportGroup"}
		}
	}

	ccrn := buildCCRN(in)

	ci, err := getOrCreateComponentInstance(ctx, deps, ccrn, svc, in)
	if err != nil {
		return SIEMAlertResult{}, err
	}

	issue, issueVariant, err := getOrCreateIssueAndVariant(ctx, deps, in)
	if err != nil {
		return SIEMAlertResult{}, err
	}

	if err := createIssueMatchIfCI(ctx, deps, ci, issue); err != nil {
		return SIEMAlertResult{}, err
	}

	return SIEMAlertResult{
		Service:           svc,
		SupportGroup:      sg,
		ComponentInstance: ci,
		Issue:             issue,
		IssueVariant:      issueVariant,
	}, nil
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}

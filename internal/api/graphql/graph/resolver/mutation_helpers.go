// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/baseResolver"
	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/util"
)

func (r *mutationResolver) getOrCreateService(ctx context.Context, inputService *string) (*entity.Service, error) {
	if inputService == nil || *inputService == "" {
		return nil, nil
	}
	svcFilter := entity.ServiceFilter{CCRN: []*string{inputService}}
	s, err := r.App.ListServices(&svcFilter, entity.NewListOptions())
	if err != nil {
		return nil, baseResolver.NewResolverError("CreateSIEMAlertMutationResolver", "Internal Error - when listing services")
	}
	if len(s.Elements) > 0 {
		return s.Elements[0].Service, nil
	}
	svcInput := model.ServiceInput{Ccrn: inputService}
	svcEntity := model.NewServiceEntity(&svcInput)
	newSvc, err := r.App.CreateService(ctx, &svcEntity)
	if err != nil {
		s2, err2 := r.App.ListServices(&svcFilter, entity.NewListOptions())
		if err2 != nil || len(s2.Elements) == 0 {
			return nil, baseResolver.NewResolverError("CreateSIEMAlertMutationResolver", "Internal Error - when creating service")
		}
		return s2.Elements[0].Service, nil
	}
	return newSvc, nil
}

func (r *mutationResolver) getOrCreateSupportGroup(ctx context.Context, inputSupportGroup *string) (*entity.SupportGroup, error) {
	if inputSupportGroup == nil || *inputSupportGroup == "" {
		return nil, nil
	}
	sgFilter := entity.SupportGroupFilter{CCRN: []*string{inputSupportGroup}}
	sgList, err := r.App.ListSupportGroups(&sgFilter, entity.NewListOptions())
	if err != nil {
		return nil, baseResolver.NewResolverError("CreateSIEMAlertMutationResolver", "Internal Error - when listing support groups")
	}
	if len(sgList.Elements) > 0 {
		return sgList.Elements[0].SupportGroup, nil
	}
	sgEntity := model.NewSupportGroupEntity(&model.SupportGroupInput{Ccrn: inputSupportGroup})
	newSg, err := r.App.CreateSupportGroup(ctx, &sgEntity)
	if err != nil {
		sg2, err2 := r.App.ListSupportGroups(&sgFilter, entity.NewListOptions())
		if err2 != nil || len(sg2.Elements) == 0 {
			return nil, baseResolver.NewResolverError("CreateSIEMAlertMutationResolver", "Internal Error - when creating support group")
		}
		return sg2.Elements[0].SupportGroup, nil
	}
	return newSg, nil
}

func buildCCRN(input model.SIEMAlertInput) string {
	var parts []string
	if input.Service != nil && *input.Service != "" {
		parts = append(parts, *input.Service)
	}
	if input.Region != nil && *input.Region != "" {
		parts = append(parts, *input.Region)
	}
	if input.Cluster != nil && *input.Cluster != "" {
		parts = append(parts, *input.Cluster)
	}
	if input.Namespace != nil && *input.Namespace != "" {
		parts = append(parts, *input.Namespace)
	}
	if input.Pod != nil && *input.Pod != "" {
		parts = append(parts, *input.Pod)
	}
	if input.Container != nil && *input.Container != "" {
		parts = append(parts, *input.Container)
	}
	return strings.Join(parts, "/")
}

func (r *mutationResolver) getOrCreateComponentInstance(ctx context.Context, ccrn string, svc *entity.Service, input model.SIEMAlertInput) (*entity.ComponentInstance, error) {
	if ccrn == "" || svc == nil {
		return nil, nil
	}
	ciInput := model.ComponentInstanceInput{
		Ccrn:      &ccrn,
		ServiceID: func() *string { v := fmt.Sprintf("%d", svc.Id); return &v }(),
		Region:    input.Region,
		Cluster:   input.Cluster,
		Namespace: input.Namespace,
		Pod:       input.Pod,
		Container: input.Container,
	}
	ciEntity := model.NewComponentInstanceEntity(&ciInput)
	newCi, err := r.App.CreateComponentInstance(ctx, &ciEntity, nil)
	if err != nil {
		filter := entity.ComponentInstanceFilter{CCRN: []*string{&ccrn}}
		cis, err2 := r.App.ListComponentInstances(&filter, &entity.ListOptions{})
		if err2 != nil || len(cis.Elements) == 0 {
			return nil, baseResolver.NewResolverError("CreateSIEMAlertMutationResolver", "Internal Error - when creating componentInstance")
		}
		return cis.Elements[0].ComponentInstance, nil
	}
	return newCi, nil
}

func (r *mutationResolver) getOrCreateIssueAndVariant(ctx context.Context, input model.SIEMAlertInput) (*entity.Issue, *entity.IssueVariant, error) {
	var issue *entity.Issue
	var issueVariant *entity.IssueVariant

	if input.URL != nil && *input.URL != "" {
		if input.Name != nil && *input.Name != "" {
			ivs, err := r.App.ListIssueVariants(&entity.IssueVariantFilter{SecondaryName: []*string{input.Name}}, &entity.ListOptions{})
			if err == nil {
				for _, v := range ivs.Elements {
					if v.IssueVariant.ExternalUrl == *input.URL {
						issueVariant = v.IssueVariant
						break
					}
				}
			}
		}
	}

	if issueVariant == nil {
		if input.Name == nil || *input.Name == "" {
			return nil, nil, baseResolver.NewResolverError("CreateSIEMAlertMutationResolver", "Invalid Input - name or url required")
		}
		newIssue, err := r.App.CreateIssue(ctx, &entity.Issue{PrimaryName: *input.Name, Description: func() string {
			if input.Description != nil {
				return *input.Description
			}
			return ""
		}(), Type: entity.IssueTypeSecurityEvent})
		if err != nil {
			f := &entity.IssueFilter{PrimaryName: []*string{input.Name}}
			lo := entity.IssueListOptions{ListOptions: *entity.NewListOptions()}
			issues, ierr := r.App.ListIssues(f, &lo)
			if ierr != nil || len(issues.Elements) == 0 {
				return nil, nil, baseResolver.NewResolverError("CreateSIEMAlertMutationResolver", "Internal Error - when creating issue")
			}
			issue = issues.Elements[0].Issue
		} else {
			issue = newIssue
		}

		siemRepoName := "heureka-siem"
		if input.Source != nil && *input.Source != "" {
			siemRepoName = *input.Source
		}

		repoFilter := entity.IssueRepositoryFilter{
			Name: []*string{&siemRepoName},
		}

		repositories, err := r.App.ListIssueRepositories(&repoFilter, &entity.ListOptions{})

		var issueRepositoryId int64
		if err == nil && len(repositories.Elements) > 0 {
			issueRepositoryId = repositories.Elements[0].IssueRepository.Id
		} else {
			newRepo := entity.IssueRepository{
				BaseIssueRepository: entity.BaseIssueRepository{
					Name: siemRepoName,
				},
			}

			createdRepo, err := r.App.CreateIssueRepository(ctx, &newRepo)
			if err != nil {
				return nil, nil, baseResolver.NewResolverError("CreateSIEMAlertMutationResolver", "Internal Error - failed to init SIEM repository")
			}
			issueRepositoryId = createdRepo.Id
		}

		sev := entity.Severity{}
		if input.Severity != nil {
			sev = entity.NewSeverityFromRating(entity.SeverityValues(input.Severity.String()))
		}
		iv := entity.IssueVariant{
			SecondaryName: func() string {
				if input.Name != nil {
					return *input.Name
				}
				return ""
			}(),
			IssueId:           issue.Id,
			IssueRepositoryId: issueRepositoryId,
			Severity:          sev,
			Description: func() string {
				if input.Description != nil {
					return *input.Description
				}
				return ""
			}(),
			ExternalUrl: func() string {
				if input.URL != nil {
					return *input.URL
				}
				return ""
			}(),
		}
		newIv, err := r.App.CreateIssueVariant(ctx, &iv)
		if err != nil {
			return nil, nil, baseResolver.NewResolverError("CreateSIEMAlertMutationResolver", "Internal Error - when creating issueVariant")
		}
		issueVariant = newIv
	} else {
		iss, err := r.App.GetIssue(issueVariant.IssueId)
		if err != nil {
			return nil, nil, baseResolver.NewResolverError("CreateSIEMAlertMutationResolver", "Internal Error - when resolving issue")
		}
		issue = iss
	}

	return issue, issueVariant, nil
}

func (r *mutationResolver) createIssueMatchIfCI(ctx context.Context, ci *entity.ComponentInstance, issue *entity.Issue) error {
	if ci == nil {
		return nil
	}
	userId := util.SystemUserId
	users, err := r.App.ListUsers(&entity.UserFilter{}, &entity.ListOptions{})
	if err == nil && len(users.Elements) > 0 {
		userId = users.Elements[0].User.Id
	}

	im := entity.IssueMatch{
		IssueId:               issue.Id,
		ComponentInstanceId:   ci.Id,
		UserId:                userId,
		Status:                entity.IssueMatchStatusValuesNew,
		RemediationDate:       time.Now(),
		TargetRemediationDate: time.Now(),
	}
	_, err = r.App.CreateIssueMatch(ctx, &im)
	if err != nil {
		return baseResolver.NewResolverError("CreateSIEMAlertMutationResolver", "Internal Error - when creating issue match")
	}
	return nil
}

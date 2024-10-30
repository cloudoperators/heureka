// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"fmt"
	"strconv"
	"time"

	"github.com/samber/lo"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/pkg/util"
)

// add custom models here

var AllSeverityValuesOrdered = []SeverityValues{
	SeverityValuesCritical,
	SeverityValuesHigh,
	SeverityValuesMedium,
	SeverityValuesLow,
	SeverityValuesNone,
}

var AllIssueTypesOrdered = []IssueTypes{
	IssueTypesPolicyViolation,
	IssueTypesSecurityEvent,
	IssueTypesVulnerability,
}

var AllIssueMatchStatusValuesOrdered = []IssueMatchStatusValues{
	IssueMatchStatusValuesNew,
	IssueMatchStatusValuesRiskAccepted,
	IssueMatchStatusValuesFalsePositive,
	IssueMatchStatusValuesMitigated,
}

func NewPageInfo(p *entity.PageInfo) *PageInfo {
	if p == nil {
		return nil
	}
	return &PageInfo{
		HasNextPage:     p.HasNextPage,
		HasPreviousPage: p.HasPreviousPage,
		IsValidPage:     p.IsValidPage,
		PageNumber:      p.PageNumber,
		NextPageAfter:   p.NextPageAfter,
		Pages: lo.Map(p.Pages, func(page entity.Page, _ int) *Page {
			return NewPage(&page)
		}),
	}
}

func NewPage(p *entity.Page) *Page {
	if p == nil {
		return nil
	}
	return &Page{
		After:      p.After,
		IsCurrent:  util.Ptr(p.IsCurrent),
		PageNumber: p.PageNumber,
		PageCount:  p.PageCount,
	}
}

func NewSeverity(sev entity.Severity) *Severity {
	severity, _ := SeverityValue(sev.Value)

	if severity == "unknown" {
		return &Severity{
			Value: &severity,
			Score: &sev.Score,
			Cvss:  nil,
		}
	}

	baseScore := sev.Cvss.Base.Score()
	av := sev.Cvss.Base.AV.String()
	attackComplexity := sev.Cvss.Base.AC.String()
	privilegesRequired := sev.Cvss.Base.PR.String()
	userInteraction := sev.Cvss.Base.UI.String()
	scope := sev.Cvss.Base.S.String()
	confidentialityImpact := sev.Cvss.Base.C.String()
	integrityImpact := sev.Cvss.Base.I.String()
	availabilityImpact := sev.Cvss.Base.A.String()

	temporalScore := sev.Cvss.Temporal.Score()
	exploitCodeMaturity := sev.Cvss.Temporal.E.String()
	remediationLevel := sev.Cvss.Temporal.RL.String()
	reportConfidence := sev.Cvss.Temporal.RC.String()

	environmentalScore := sev.Cvss.Environmental.Score()
	modifiedAttackVector := sev.Cvss.Environmental.MAV.String()
	modifiedAttackComplexity := sev.Cvss.Environmental.MAC.String()
	modifiedPrivilegesRequired := sev.Cvss.Environmental.MPR.String()
	modifiedUserInteraction := sev.Cvss.Environmental.MUI.String()
	modifiedScope := sev.Cvss.Environmental.MS.String()
	modifiedConfidentialityImpact := sev.Cvss.Environmental.MC.String()
	modifiedIntegrityImpact := sev.Cvss.Environmental.MI.String()
	modifiedAvailabilityImpact := sev.Cvss.Environmental.MA.String()
	availabilityRequirement := sev.Cvss.Environmental.AR.String()
	integrityRequirement := sev.Cvss.Environmental.IR.String()
	confidentialityRequirement := sev.Cvss.Environmental.CR.String()
	s := &Severity{
		Value: &severity,
		Score: &sev.Score,
		Cvss: &Cvss{
			Vector: &sev.Cvss.Vector,
			Base: &CVSSBase{
				Score:                 &baseScore,
				AttackVector:          &av,
				AttackComplexity:      &attackComplexity,
				PrivilegesRequired:    &privilegesRequired,
				UserInteraction:       &userInteraction,
				Scope:                 &scope,
				ConfidentialityImpact: &confidentialityImpact,
				IntegrityImpact:       &integrityImpact,
				AvailabilityImpact:    &availabilityImpact,
			},
			Temporal: &CVSSTemporal{
				Score:               &temporalScore,
				ExploitCodeMaturity: &exploitCodeMaturity,
				RemediationLevel:    &remediationLevel,
				ReportConfidence:    &reportConfidence,
			},
			Environmental: &CVSSEnvironmental{
				Score:                         &environmentalScore,
				ModifiedAttackVector:          &modifiedAttackVector,
				ModifiedAttackComplexity:      &modifiedAttackComplexity,
				ModifiedPrivilegesRequired:    &modifiedPrivilegesRequired,
				ModifiedUserInteraction:       &modifiedUserInteraction,
				ModifiedScope:                 &modifiedScope,
				ModifiedConfidentialityImpact: &modifiedConfidentialityImpact,
				ModifiedIntegrityImpact:       &modifiedIntegrityImpact,
				ModifiedAvailabilityImpact:    &modifiedAvailabilityImpact,
				ConfidentialityRequirement:    &confidentialityRequirement,
				AvailabilityRequirement:       &availabilityRequirement,
				IntegrityRequirement:          &integrityRequirement,
			},
		},
	}
	return s
}

func NewSeverityEntity(severity *SeverityInput) entity.Severity {
	if severity == nil || severity.Vector == nil {
		return entity.Severity{}
	}
	return entity.NewSeverity(*severity.Vector)
}

func NewIssue(issue *entity.Issue) Issue {
	lastModified := issue.UpdatedAt.String()
	issueType := IssueTypes(issue.Type.String())
	return Issue{
		ID:           fmt.Sprintf("%d", issue.Id),
		PrimaryName:  &issue.PrimaryName,
		Type:         &issueType,
		Description:  &issue.Description,
		LastModified: &lastModified,
	}
}

func NewIssueWithAggregations(issue *entity.IssueResult) Issue {
	lastModified := issue.Issue.UpdatedAt.String()
	issueType := IssueTypes(issue.Type.String())

	var metadata IssueMetadata

	if issue.IssueAggregations != nil {
		metadata = IssueMetadata{
			ServiceCount:                  int(issue.IssueAggregations.AffectedServices),
			ActivityCount:                 int(issue.IssueAggregations.Activities),
			IssueMatchCount:               int(issue.IssueAggregations.IssueMatches),
			ComponentInstanceCount:        int(issue.IssueAggregations.AffectedComponentInstances),
			ComponentVersionCount:         int(issue.IssueAggregations.ComponentVersions),
			EarliestDiscoveryDate:         issue.IssueAggregations.EarliestDiscoveryDate.String(),
			EarliestTargetRemediationDate: issue.IssueAggregations.EarliestTargetRemediationDate.String(),
		}
	}

	return Issue{
		ID:           fmt.Sprintf("%d", issue.Issue.Id),
		PrimaryName:  &issue.Issue.PrimaryName,
		Type:         &issueType,
		LastModified: &lastModified,
		Metadata:     &metadata,
	}
}

func NewIssueEntity(issue *IssueInput) entity.Issue {
	issueType := ""
	if issue.Type != nil && issue.Type.IsValid() {
		issueType = issue.Type.String()
	}
	return entity.Issue{
		PrimaryName: lo.FromPtr(issue.PrimaryName),
		Description: lo.FromPtr(issue.Description),
		Type:        entity.NewIssueType(issueType),
	}
}

func NewIssueMatch(im *entity.IssueMatch) IssueMatch {
	status := IssueMatchStatusValue(im.Status.String())
	targetRemediationDate := im.TargetRemediationDate.Format(time.RFC3339)
	discoveryDate := im.CreatedAt.Format(time.RFC3339)
	remediationDate := im.RemediationDate.Format(time.RFC3339)
	severity := NewSeverity(im.Severity)
	return IssueMatch{
		ID:                    fmt.Sprintf("%d", im.Id),
		Status:                &status,
		RemediationDate:       &remediationDate,
		DiscoveryDate:         &discoveryDate,
		TargetRemediationDate: &targetRemediationDate,
		Severity:              severity,
		IssueID:               util.Ptr(fmt.Sprintf("%d", im.IssueId)),
		ComponentInstanceID:   util.Ptr(fmt.Sprintf("%d", im.ComponentInstanceId)),
		UserID:                util.Ptr(fmt.Sprintf("%d", im.UserId)),
	}
}

func NewIssueMatchEntity(im *IssueMatchInput) entity.IssueMatch {
	issueId, _ := strconv.ParseInt(lo.FromPtr(im.IssueID), 10, 64)
	ciId, _ := strconv.ParseInt(lo.FromPtr(im.ComponentInstanceID), 10, 64)
	userId, _ := strconv.ParseInt(lo.FromPtr(im.UserID), 10, 64)
	targetRemediationDate, _ := time.Parse(time.RFC3339, lo.FromPtr(im.TargetRemediationDate))
	remediationDate, _ := time.Parse(time.RFC3339, lo.FromPtr(im.RemediationDate))
	createdAt, _ := time.Parse(time.RFC3339, lo.FromPtr(im.DiscoveryDate))
	status := entity.IssueMatchStatusValuesNone
	if im.Status != nil {
		status = entity.NewIssueMatchStatusValue(im.Status.String())
	}
	return entity.IssueMatch{
		Status:                status,
		TargetRemediationDate: targetRemediationDate,
		RemediationDate:       remediationDate,
		Severity:              entity.Severity{},
		IssueId:               issueId,
		ComponentInstanceId:   ciId,
		UserId:                userId,
		CreatedAt:             createdAt,
	}
}

func NewIssueMatchChange(imc *entity.IssueMatchChange) IssueMatchChange {
	action := IssueMatchChangeAction(imc.Action)
	return IssueMatchChange{
		ID:           fmt.Sprintf("%d", imc.Id),
		Action:       &action,
		IssueMatchID: util.Ptr(fmt.Sprintf("%d", imc.IssueMatchId)),
		IssueMatch:   nil,
		ActivityID:   util.Ptr(fmt.Sprintf("%d", imc.ActivityId)),
		Activity:     nil,
	}
}

func NewIssueMatchChangeEntity(imc *IssueMatchChangeInput) entity.IssueMatchChange {
	action := entity.IssueMatchChangeAction(lo.FromPtr(imc.Action))
	issueMatchId, _ := strconv.ParseInt(lo.FromPtr(imc.IssueMatchID), 10, 64)
	activityId, _ := strconv.ParseInt(lo.FromPtr(imc.ActivityID), 10, 64)
	return entity.IssueMatchChange{
		Action:       action.String(),
		IssueMatchId: issueMatchId,
		ActivityId:   activityId,
	}
}

func NewIssueRepository(repo *entity.IssueRepository) IssueRepository {
	createdAt := repo.BaseIssueRepository.CreatedAt.String()
	updatedAt := repo.BaseIssueRepository.UpdatedAt.String()
	return IssueRepository{
		ID:            fmt.Sprintf("%d", repo.Id),
		Name:          &repo.Name,
		URL:           &repo.Url,
		Services:      nil,
		IssueVariants: nil,
		CreatedAt:     &createdAt,
		UpdatedAt:     &updatedAt,
	}
}

func NewIssueRepositoryEntity(repo *IssueRepositoryInput) entity.IssueRepository {
	return entity.IssueRepository{
		BaseIssueRepository: entity.BaseIssueRepository{
			Name: lo.FromPtr(repo.Name),
			Url:  lo.FromPtr(repo.URL),
		},
	}
}

func NewIssueVariant(issueVariant *entity.IssueVariant) IssueVariant {
	var repo IssueRepository
	if issueVariant.IssueRepository != nil {
		repo = NewIssueRepository(issueVariant.IssueRepository)
	}
	createdAt := issueVariant.CreatedAt.String()
	updatedAt := issueVariant.UpdatedAt.String()
	return IssueVariant{
		ID:                fmt.Sprintf("%d", issueVariant.Id),
		SecondaryName:     &issueVariant.SecondaryName,
		Description:       &issueVariant.Description,
		Severity:          NewSeverity(issueVariant.Severity),
		IssueID:           util.Ptr(fmt.Sprintf("%d", issueVariant.IssueId)),
		IssueRepositoryID: util.Ptr(fmt.Sprintf("%d", issueVariant.IssueRepositoryId)),
		IssueRepository:   &repo,
		CreatedAt:         &createdAt,
		UpdatedAt:         &updatedAt,
	}
}

func NewIssueVariantEdge(issueVariant *entity.IssueVariant) IssueVariantEdge {
	iv := NewIssueVariant(issueVariant)
	edgeCreationDate := issueVariant.CreatedAt.String()
	edgeUpdateDate := issueVariant.UpdatedAt.String()
	issueVariantEdge := IssueVariantEdge{
		Node:      &iv,
		Cursor:    &iv.ID,
		CreatedAt: &edgeCreationDate,
		UpdatedAt: &edgeUpdateDate,
	}
	return issueVariantEdge
}

func NewIssueVariantEntity(issueVariant *IssueVariantInput) entity.IssueVariant {
	issueId, _ := strconv.ParseInt(lo.FromPtr(issueVariant.IssueID), 10, 64)
	irId, _ := strconv.ParseInt(lo.FromPtr(issueVariant.IssueRepositoryID), 10, 64)
	return entity.IssueVariant{
		SecondaryName:     lo.FromPtr(issueVariant.SecondaryName),
		Description:       lo.FromPtr(issueVariant.Description),
		Severity:          NewSeverityEntity(issueVariant.Severity),
		IssueId:           issueId,
		IssueRepositoryId: irId,
	}
}

func NewUser(user *entity.User) User {
	return User{
		ID:           fmt.Sprintf("%d", user.Id),
		UniqueUserID: &user.UniqueUserID,
		Name:         &user.Name,
		Type:         int(user.Type),
	}
}

func NewUserEntity(user *UserInput) entity.User {
	return entity.User{
		Name:         lo.FromPtr(user.Name),
		UniqueUserID: lo.FromPtr(user.UniqueUserID),
		Type:         entity.GetUserTypeFromString(lo.FromPtr(user.Type)),
	}
}

func NewService(s *entity.Service) Service {
	return Service{
		ID:   fmt.Sprintf("%d", s.Id),
		Ccrn: &s.CCRN,
	}
}

func NewServiceWithAggregations(service *entity.ServiceResult) Service {
	var metadata ServiceMetadata

	if service.ServiceAggregations != nil {
		metadata = ServiceMetadata{
			IssueMatchCount:        int(service.ServiceAggregations.IssueMatches),
			ComponentInstanceCount: int(service.ServiceAggregations.ComponentInstances),
		}
	}

	return Service{
		ID:       fmt.Sprintf("%d", service.Id),
		Ccrn:     &service.CCRN,
		Metadata: &metadata,
	}
}

func NewServiceEntity(service *ServiceInput) entity.Service {
	return entity.Service{
		BaseService: entity.BaseService{
			CCRN: lo.FromPtr(service.Ccrn),
		},
	}
}

func NewSupportGroup(supportGroup *entity.SupportGroup) SupportGroup {
	return SupportGroup{
		ID:   fmt.Sprintf("%d", supportGroup.Id),
		Ccrn: &supportGroup.CCRN,
	}
}

func NewSupportGroupEntity(supportGroup *SupportGroupInput) entity.SupportGroup {
	return entity.SupportGroup{
		CCRN: lo.FromPtr(supportGroup.Ccrn),
	}
}

func NewActivity(activity *entity.Activity) Activity {
	status := ActivityStatusValues(activity.Status.String())
	return Activity{
		ID:     fmt.Sprintf("%d", activity.Id),
		Status: &status,
	}
}

func NewActivityEntity(activity *ActivityInput) entity.Activity {
	status := entity.ActivityStatusValuesOpen
	if activity.Status != nil {
		status = entity.NewActivityStatusValue(activity.Status.String())
	}
	return entity.Activity{
		Status: status,
	}
}

func NewEvidence(evidence *entity.Evidence) Evidence {
	authorId := fmt.Sprintf("%d", evidence.UserId)
	activityId := fmt.Sprintf("%d", evidence.ActivityId)
	severity := NewSeverity(evidence.Severity)
	t := evidence.Type.String()
	raaEnd := evidence.RaaEnd.Format(time.RFC3339)
	return Evidence{
		ID:          fmt.Sprintf("%d", evidence.Id),
		Description: &evidence.Description,
		AuthorID:    &authorId,
		ActivityID:  &activityId,
		Vector:      severity.Cvss.Vector,
		Type:        &t,
		RaaEnd:      &raaEnd,
	}
}

func NewEvidenceEntity(evidence *EvidenceInput) entity.Evidence {
	authorId, _ := strconv.ParseInt(lo.FromPtr(evidence.AuthorID), 10, 64)
	activityId, _ := strconv.ParseInt(lo.FromPtr(evidence.ActivityID), 10, 64)
	t := entity.NewEvidenceTypeValue(lo.FromPtr(evidence.Type))
	raaEnd, _ := time.Parse(time.RFC3339, lo.FromPtr(evidence.RaaEnd))
	// raaEnd, _ := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", lo.FromPtr(evidence.RaaEnd))
	return entity.Evidence{
		Description: lo.FromPtr(evidence.Description),
		UserId:      authorId,
		ActivityId:  activityId,
		Severity:    NewSeverityEntity(evidence.Severity),
		Type:        t,
		RaaEnd:      raaEnd,
	}
}

func NewComponent(component *entity.Component) Component {
	componentType, _ := ComponentTypeValue(component.Type)
	return Component{
		ID:   fmt.Sprintf("%d", component.Id),
		Ccrn: &component.CCRN,
		Type: &componentType,
	}
}

func NewComponentEntity(component *ComponentInput) entity.Component {
	componentType := ""
	if component.Type != nil && component.Type.IsValid() {
		componentType = component.Type.String()
	}
	return entity.Component{
		CCRN: lo.FromPtr(component.Ccrn),
		Type: componentType,
	}
}

func NewComponentVersion(componentVersion *entity.ComponentVersion) ComponentVersion {
	return ComponentVersion{
		ID:          fmt.Sprintf("%d", componentVersion.Id),
		Version:     &componentVersion.Version,
		ComponentID: util.Ptr(fmt.Sprintf("%d", componentVersion.ComponentId)),
	}
}

func NewComponentVersionEntity(componentVersion *ComponentVersionInput) entity.ComponentVersion {
	componentId, err := strconv.ParseInt(lo.FromPtr(componentVersion.ComponentID), 10, 64)
	if err != nil {
		componentId = 0
	}
	return entity.ComponentVersion{
		Version:     lo.FromPtr(componentVersion.Version),
		ComponentId: componentId,
	}
}

func NewComponentInstance(componentInstance *entity.ComponentInstance) ComponentInstance {
	count := int(componentInstance.Count)
	return ComponentInstance{
		ID:                 fmt.Sprintf("%d", componentInstance.Id),
		Ccrn:               &componentInstance.CCRN,
		Count:              &count,
		ComponentVersionID: util.Ptr(fmt.Sprintf("%d", componentInstance.ComponentVersionId)),
		ServiceID:          util.Ptr(fmt.Sprintf("%d", componentInstance.ServiceId)),
	}
}

func NewComponentInstanceEntity(componentInstance *ComponentInstanceInput) entity.ComponentInstance {
	componentVersionId, _ := strconv.ParseInt(lo.FromPtr(componentInstance.ComponentVersionID), 10, 64)
	serviceId, _ := strconv.ParseInt(lo.FromPtr(componentInstance.ServiceID), 10, 64)
	return entity.ComponentInstance{
		CCRN:               lo.FromPtr(componentInstance.Ccrn),
		Count:              int16(lo.FromPtr(componentInstance.Count)),
		ComponentVersionId: componentVersionId,
		ServiceId:          serviceId,
	}
}

// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/samber/lo"

	"github.com/cloudoperators/heureka/internal/entity"
	interUtil "github.com/cloudoperators/heureka/internal/util"
	"github.com/cloudoperators/heureka/pkg/util"
)

// add custom models here
func getModelMetadata(em entity.Metadata) *Metadata {
	createdAt := em.CreatedAt.String()
	deletedAt := em.DeletedAt.String()
	updatedAt := em.UpdatedAt.String()
	return &Metadata{
		CreatedAt: util.Ptr(createdAt),
		CreatedBy: util.Ptr(fmt.Sprintf("%d", em.CreatedBy)),
		DeletedAt: util.Ptr(deletedAt),
		UpdatedAt: util.Ptr(updatedAt),
		UpdatedBy: util.Ptr(fmt.Sprintf("%d", em.UpdatedBy)),
	}
}

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

var AllComponentInstanceTypesOrdered = []ComponentInstanceTypes{
	ComponentInstanceTypesUnknown,
	ComponentInstanceTypesProject,
	ComponentInstanceTypesServer,
	ComponentInstanceTypesSecurityGroup,
	ComponentInstanceTypesDNSZone,
	ComponentInstanceTypesFloatingIP,
	ComponentInstanceTypesRbacPolicy,
	ComponentInstanceTypesUser,
	ComponentInstanceTypesContainer,
	ComponentInstanceTypesRecordSet,
	ComponentInstanceTypesSecurityGroupRule,
	ComponentInstanceTypesProjectConfiguration,
}

type HasToEntity interface {
	ToOrderEntity() entity.Order
}

func (od *OrderDirection) ToOrderDirectionEntity() entity.OrderDirection {
	direction := entity.OrderDirectionAsc
	if *od == OrderDirectionDesc {
		direction = entity.OrderDirectionDesc
	}
	return direction
}

func (sg *SupportGroupOrderBy) ToOrderEntity() entity.Order {
	var order entity.Order
	switch *sg.By {
	case SupportGroupOrderByFieldCcrn:
		order.By = entity.SupportGroupCcrn
	}
	order.Direction = sg.Direction.ToOrderDirectionEntity()
	return order
}

func (cv *ComponentVersionOrderBy) ToOrderEntity() entity.Order {
	var order entity.Order
	switch *cv.By {
	case ComponentVersionOrderByFieldRepository:
		order.By = entity.ComponentVersionRepository
	}
	order.Direction = cv.Direction.ToOrderDirectionEntity()
	return order
}

func (io *IssueOrderBy) ToOrderEntity() entity.Order {
	var order entity.Order
	switch *io.By {
	case IssueOrderByFieldPrimaryName:
		order.By = entity.IssuePrimaryName
	case IssueOrderByFieldSeverity:
		order.By = entity.IssueVariantRating
	}
	order.Direction = io.Direction.ToOrderDirectionEntity()
	return order
}

func (cio *ComponentInstanceOrderBy) ToOrderEntity() entity.Order {
	var order entity.Order
	switch *cio.By {
	case ComponentInstanceOrderByFieldCcrn:
		order.By = entity.ComponentInstanceCcrn
	case ComponentInstanceOrderByFieldRegion:
		order.By = entity.ComponentInstanceRegion
	case ComponentInstanceOrderByFieldCluster:
		order.By = entity.ComponentInstanceCluster
	case ComponentInstanceOrderByFieldNamespace:
		order.By = entity.ComponentInstanceNamespace
	case ComponentInstanceOrderByFieldDomain:
		order.By = entity.ComponentInstanceDomain
	case ComponentInstanceOrderByFieldProject:
		order.By = entity.ComponentInstanceProject
	case ComponentInstanceOrderByFieldPod:
		order.By = entity.ComponentInstancePod
	case ComponentInstanceOrderByFieldContainer:
		order.By = entity.ComponentInstanceContainer
	case ComponentInstanceOrderByFieldType:
		order.By = entity.ComponentInstanceTypeOrder
	}
	order.Direction = cio.Direction.ToOrderDirectionEntity()
	return order
}

func (imo *IssueMatchOrderBy) ToOrderEntity() entity.Order {
	var order entity.Order
	switch *imo.By {
	case IssueMatchOrderByFieldPrimaryName:
		order.By = entity.IssuePrimaryName
	case IssueMatchOrderByFieldComponentInstanceCcrn:
		order.By = entity.ComponentInstanceCcrn
	case IssueMatchOrderByFieldTargetRemediationDate:
		order.By = entity.IssueMatchTargetRemediationDate
	case IssueMatchOrderByFieldSeverity:
		order.By = entity.IssueMatchRating
	}
	order.Direction = imo.Direction.ToOrderDirectionEntity()
	return order
}

func (so *ServiceOrderBy) ToOrderEntity() entity.Order {
	var order entity.Order
	switch *so.By {
	case ServiceOrderByFieldCcrn:
		order.By = entity.ServiceCcrn
	}
	order.Direction = so.Direction.ToOrderDirectionEntity()
	return order
}

func (ro *RemediationOrderBy) ToOrderEntity() entity.Order {
	var order entity.Order
	switch *ro.By {
	case RemediationOrderByFieldVulnerability:
		order.By = entity.RemediationIssue
	case RemediationOrderByFieldSeverity:
		order.By = entity.RemediationSeverity
	}
	order.Direction = ro.Direction.ToOrderDirectionEntity()
	return order
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

	if severity == "unknown" || sev.Cvss == (entity.Cvss{}) {
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
			Vector:      &sev.Cvss.Vector,
			ExternalURL: &sev.Cvss.ExternalUrl,
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
	if severity == nil || (severity.Rating == nil && severity.Vector == nil) {
		// no severity information was passed
		return entity.Severity{}
	}
	if (severity.Vector == nil || *severity.Vector == "") && severity.Rating != nil {
		// only rating was passed
		return entity.NewSeverityFromRating(entity.SeverityValues(*severity.Rating))
	}
	// both rating and vector or only vector was passed
	// either way, use the vector as the primary source of information
	return entity.NewSeverity(*severity.Vector)
}

func NewSeverityCounts(counts *entity.IssueSeverityCounts) SeverityCounts {
	return SeverityCounts{
		Critical: int(counts.Critical),
		High:     int(counts.High),
		Medium:   int(counts.Medium),
		Low:      int(counts.Low),
		None:     int(counts.None),
		Total:    int(counts.Total),
	}
}

func NewIssueStatus(status *VulnerabilityStatus) entity.IssueStatus {
	if status == nil {
		return entity.IssueStatusAll
	}
	switch status.String() {
	case VulnerabilityStatusOpen.String():
		return entity.IssueStatusOpen
	case VulnerabilityStatusRemediated.String():
		return entity.IssueStatusRemediated
	case VulnerabilityStatusAll.String():
		return entity.IssueStatusAll
	}

	return entity.IssueStatusAll
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
		Metadata:     getModelMetadata(issue.Metadata),
	}
}

func NewIssueWithAggregations(issue *entity.IssueResult) Issue {
	lastModified := issue.Issue.UpdatedAt.String()
	issueType := IssueTypes(issue.Type.String())

	var objectMetadata IssueMetadata

	if issue.IssueAggregations != nil {
		objectMetadata = IssueMetadata{
			ServiceCount:                  int(issue.IssueAggregations.AffectedServices),
			IssueMatchCount:               int(issue.IssueAggregations.IssueMatches),
			ComponentInstanceCount:        int(issue.IssueAggregations.AffectedComponentInstances),
			ComponentVersionCount:         int(issue.IssueAggregations.ComponentVersions),
			EarliestDiscoveryDate:         issue.IssueAggregations.EarliestDiscoveryDate.String(),
			EarliestTargetRemediationDate: issue.IssueAggregations.EarliestTargetRemediationDate.String(),
		}
	}

	return Issue{
		ID:             fmt.Sprintf("%d", issue.Issue.Id),
		PrimaryName:    &issue.Issue.PrimaryName,
		Type:           &issueType,
		Description:    &issue.Issue.Description,
		LastModified:   &lastModified,
		ObjectMetadata: &objectMetadata,
		Metadata:       getModelMetadata(issue.Issue.Metadata),
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

func NewScannerRunEntity(sr *ScannerRunInput) entity.ScannerRun {
	return entity.ScannerRun{
		RunID:     -1,
		UUID:      lo.FromPtr(sr.UUID),
		Tag:       lo.FromPtr(sr.Tag),
		Completed: false,
		StartRun:  time.Now(),
		EndRun:    time.Now(),
	}
}

func NewScannerRun(sr *entity.ScannerRun) ScannerRun {
	startRun := sr.StartRun.Format(time.RFC3339)
	endRun := sr.EndRun.Format(time.RFC3339)

	return ScannerRun{
		ID:        fmt.Sprintf("%d", sr.RunID),
		UUID:      sr.UUID,
		Tag:       sr.Tag,
		Completed: sr.Completed,
		StartRun:  startRun,
		EndRun:    endRun,
	}
}

func NewImage(component *entity.Component) Image {
	return Image{
		ID:               fmt.Sprintf("%d", component.Id),
		ImageRegistryURL: &component.Url,
		Repository:       &component.Repository,
	}
}

func NewImageVersion(componentVersion *entity.ComponentVersion) ImageVersion {
	return ImageVersion{
		ID:         fmt.Sprintf("%d", componentVersion.Id),
		Tag:        &componentVersion.Tag,
		Version:    &componentVersion.Version,
		Repository: &componentVersion.Repository,
		Metadata:   getModelMetadata(componentVersion.Metadata),
		EndOfLife:  &componentVersion.EndOfLife,
	}
}

func NewVulnerability(issue *entity.Issue) Vulnerability {
	return Vulnerability{
		ID:          fmt.Sprintf("%d", issue.Id),
		Name:        &issue.PrimaryName,
		Description: &issue.Description,
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
		Metadata:              getModelMetadata(im.Metadata),
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
		Metadata:              entity.Metadata{CreatedAt: createdAt},
	}
}

func NewIssueRepository(repo *entity.IssueRepository) IssueRepository {
	return IssueRepository{
		ID:            fmt.Sprintf("%d", repo.Id),
		Name:          &repo.Name,
		URL:           &repo.Url,
		Services:      nil,
		IssueVariants: nil,
		Metadata:      getModelMetadata(repo.BaseIssueRepository.Metadata),
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
	return IssueVariant{
		ID:                fmt.Sprintf("%d", issueVariant.Id),
		SecondaryName:     &issueVariant.SecondaryName,
		Description:       &issueVariant.Description,
		ExternalURL:       &issueVariant.ExternalUrl,
		Severity:          NewSeverity(issueVariant.Severity),
		IssueID:           util.Ptr(fmt.Sprintf("%d", issueVariant.IssueId)),
		IssueRepositoryID: util.Ptr(fmt.Sprintf("%d", issueVariant.IssueRepositoryId)),
		IssueRepository:   &repo,
		Metadata:          getModelMetadata(issueVariant.Metadata),
	}
}

func NewIssueVariantEdge(issueVariant *entity.IssueVariant) IssueVariantEdge {
	iv := NewIssueVariant(issueVariant)
	issueVariantEdge := IssueVariantEdge{
		Node:     &iv,
		Cursor:   &iv.ID,
		Metadata: getModelMetadata(issueVariant.Metadata),
	}
	return issueVariantEdge
}

func NewIssueVariantEntity(issueVariant *IssueVariantInput) entity.IssueVariant {
	issueId, _ := strconv.ParseInt(lo.FromPtr(issueVariant.IssueID), 10, 64)
	irId, _ := strconv.ParseInt(lo.FromPtr(issueVariant.IssueRepositoryID), 10, 64)
	return entity.IssueVariant{
		SecondaryName:     lo.FromPtr(issueVariant.SecondaryName),
		Description:       lo.FromPtr(issueVariant.Description),
		ExternalUrl:       lo.FromPtr(issueVariant.ExternalURL),
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
		Email:        &user.Email,
		Metadata:     getModelMetadata(user.Metadata),
	}
}

func NewUserEntity(user *UserInput) entity.User {
	return entity.User{
		Name:         lo.FromPtr(user.Name),
		UniqueUserID: lo.FromPtr(user.UniqueUserID),
		Type:         entity.GetUserTypeFromString(lo.FromPtr(user.Type)),
		Email:        lo.FromPtr(user.Email),
	}
}

func NewService(s *entity.Service) Service {
	return Service{
		ID:       fmt.Sprintf("%d", s.Id),
		Ccrn:     &s.CCRN,
		Domain:   &s.Domain,
		Region:   &s.Region,
		Metadata: getModelMetadata(s.BaseService.Metadata),
	}
}

func NewServiceWithAggregations(service *entity.ServiceResult) Service {
	var objectMetadata ServiceMetadata

	if service.ServiceAggregations != nil {
		objectMetadata = ServiceMetadata{
			IssueMatchCount:        int(service.ServiceAggregations.IssueMatches),
			ComponentInstanceCount: int(service.ServiceAggregations.ComponentInstances),
		}
	}

	return Service{
		ID:             fmt.Sprintf("%d", service.Id),
		Ccrn:           &service.CCRN,
		Domain:         &service.Domain,
		Region:         &service.Region,
		ObjectMetadata: &objectMetadata,
		Metadata:       getModelMetadata(service.BaseService.Metadata),
	}
}

func NewServiceEntity(service *ServiceInput) entity.Service {
	return entity.Service{
		BaseService: entity.BaseService{
			CCRN:   lo.FromPtr(service.Ccrn),
			Domain: lo.FromPtr(service.Domain),
			Region: lo.FromPtr(service.Region),
		},
	}
}

func NewSupportGroup(supportGroup *entity.SupportGroup) SupportGroup {
	return SupportGroup{
		ID:       fmt.Sprintf("%d", supportGroup.Id),
		Ccrn:     &supportGroup.CCRN,
		Metadata: getModelMetadata(supportGroup.Metadata),
	}
}

func NewSupportGroupEntity(supportGroup *SupportGroupInput) entity.SupportGroup {
	return entity.SupportGroup{
		CCRN: lo.FromPtr(supportGroup.Ccrn),
	}
}

func NewComponent(component *entity.Component) Component {
	componentType, _ := ComponentTypeValue(component.Type)
	return Component{
		ID:           fmt.Sprintf("%d", component.Id),
		Ccrn:         &component.CCRN,
		Repository:   &component.Repository,
		Organization: &component.Organization,
		URL:          &component.Url,
		Type:         &componentType,
		Metadata:     getModelMetadata(component.Metadata),
	}
}

func NewComponentEntity(component *ComponentInput) entity.Component {
	componentType := ""
	if component.Type != nil && component.Type.IsValid() {
		componentType = component.Type.String()
	}
	return entity.Component{
		CCRN:         lo.FromPtr(component.Ccrn),
		Repository:   lo.FromPtr(component.Repository),
		Organization: lo.FromPtr(component.Organization),
		Url:          lo.FromPtr(component.URL),
		Type:         componentType,
	}
}

func NewComponentVersion(componentVersion *entity.ComponentVersion) ComponentVersion {
	return ComponentVersion{
		ID:           fmt.Sprintf("%d", componentVersion.Id),
		Version:      &componentVersion.Version,
		ComponentID:  util.Ptr(fmt.Sprintf("%d", componentVersion.ComponentId)),
		Repository:   &componentVersion.Repository,
		Organization: &componentVersion.Organization,
		Tag:          &componentVersion.Tag,
		Metadata:     getModelMetadata(componentVersion.Metadata),
		EndOfLife:    &componentVersion.EndOfLife,
	}
}

func NewComponentVersionEntity(componentVersion *ComponentVersionInput) entity.ComponentVersion {
	componentId, err := strconv.ParseInt(lo.FromPtr(componentVersion.ComponentID), 10, 64)
	if err != nil {
		componentId = 0
	}

	return entity.ComponentVersion{
		Version:      lo.FromPtr(componentVersion.Version),
		ComponentId:  componentId,
		Repository:   lo.FromPtr(componentVersion.Repository),
		Organization: lo.FromPtr(componentVersion.Organization),
		Tag:          lo.FromPtr(componentVersion.Tag),
		EndOfLife:    lo.FromPtr(componentVersion.EndOfLife),
	}
}

func NewComponentInstance(componentInstance *entity.ComponentInstance) ComponentInstance {
	count := int(componentInstance.Count)
	componentInstanceType := ComponentInstanceTypes(componentInstance.Type.String())

	var parentID *string
	if componentInstance.ParentId == -1 {
		parentID = nil
	} else {
		parentID = util.Ptr(fmt.Sprintf("%d", componentInstance.ParentId))
	}
	return ComponentInstance{
		ID:                 fmt.Sprintf("%d", componentInstance.Id),
		Ccrn:               &componentInstance.CCRN,
		Region:             &componentInstance.Region,
		Cluster:            &componentInstance.Cluster,
		Namespace:          &componentInstance.Namespace,
		Domain:             &componentInstance.Domain,
		Project:            &componentInstance.Project,
		Pod:                &componentInstance.Pod,
		Container:          &componentInstance.Container,
		Type:               &componentInstanceType,
		ParentID:           parentID,
		Context:            interUtil.ConvertJsonPointerToValue((*map[string]interface{})(componentInstance.Context)),
		Count:              &count,
		ComponentVersionID: util.Ptr(fmt.Sprintf("%d", componentInstance.ComponentVersionId)),
		ServiceID:          util.Ptr(fmt.Sprintf("%d", componentInstance.ServiceId)),
		Metadata:           getModelMetadata(componentInstance.Metadata),
	}
}

type Ccrn struct {
	Region    string
	Cluster   string
	Namespace string
	Domain    string
	Project   string
}

func getCcrnVal(rawCcrn string, k string) string {
	pattern := k + `=([^,]+)`
	rgx := regexp.MustCompile(pattern)
	matches := rgx.FindAllStringSubmatch(rawCcrn, -1)
	if len(matches) > 0 {
		return matches[0][1]
	}
	return ""
}

func ParseCcrn(rawCcrn string) Ccrn {
	var ccrn Ccrn
	ccrn.Region = getCcrnVal(rawCcrn, "region")
	ccrn.Cluster = getCcrnVal(rawCcrn, "cluster")
	ccrn.Namespace = getCcrnVal(rawCcrn, "namespace")
	ccrn.Domain = getCcrnVal(rawCcrn, "domain")
	ccrn.Project = getCcrnVal(rawCcrn, "project")
	return ccrn
}

func NewComponentInstanceEntity(componentInstance *ComponentInstanceInput) entity.ComponentInstance {
	componentVersionId, _ := strconv.ParseInt(lo.FromPtr(componentInstance.ComponentVersionID), 10, 64)
	serviceId, _ := strconv.ParseInt(lo.FromPtr(componentInstance.ServiceID), 10, 64)

	var parentId int64
	if componentInstance.ParentID != nil && *componentInstance.ParentID != "" {
		parentId, _ = strconv.ParseInt(*componentInstance.ParentID, 10, 64)
	}

	rawCcrn := lo.FromPtr(componentInstance.Ccrn)
	ciType := ""
	if componentInstance.Type != nil && componentInstance.Type.IsValid() {
		ciType = componentInstance.Type.String()
	}
	return entity.ComponentInstance{
		CCRN:               rawCcrn,
		Region:             lo.FromPtr(componentInstance.Region),
		Cluster:            lo.FromPtr(componentInstance.Cluster),
		Namespace:          lo.FromPtr(componentInstance.Namespace),
		Domain:             lo.FromPtr(componentInstance.Domain),
		Project:            lo.FromPtr(componentInstance.Project),
		Pod:                lo.FromPtr(componentInstance.Pod),
		Container:          lo.FromPtr(componentInstance.Container),
		Type:               entity.NewComponentInstanceType(ciType),
		Context:            (*entity.Json)(&componentInstance.Context),
		Count:              int16(lo.FromPtr(componentInstance.Count)),
		ComponentVersionId: componentVersionId,
		ServiceId:          serviceId,
		ParentId:           parentId,
	}
}

func GetStateFilterType(sf []StateFilter) []entity.StateFilterType {
	if len(sf) > 0 {
		s := make([]entity.StateFilterType, len(sf))
		for i := range sf {
			if sf[i] == StateFilterDeleted {
				s[i] = entity.Deleted
			} else if sf[i] == StateFilterActive {
				s[i] = entity.Active
			}
		}
		return s
	}
	return []entity.StateFilterType{entity.Active}
}

func NewRemediationEntity(r *RemediationInput) entity.Remediation {
	remediationDate, _ := time.Parse(time.RFC3339, lo.FromPtr(r.RemediationDate))
	expirationDate, _ := time.Parse(time.RFC3339, lo.FromPtr(r.ExpirationDate))
	rType := entity.NewRemediationType(lo.FromPtr(r.Type).String())
	rSeverity := entity.NewSeverityValues(lo.FromPtr(r.Severity).String())
	return entity.Remediation{
		Description:     lo.FromPtr(r.Description),
		Severity:        rSeverity,
		Service:         lo.FromPtr(r.Service),
		Component:       lo.FromPtr(r.Image),
		Issue:           lo.FromPtr(r.Vulnerability),
		Type:            rType,
		RemediatedBy:    lo.FromPtr(r.RemediatedBy),
		RemediationDate: remediationDate,
		ExpirationDate:  expirationDate,
	}
}

func NewRemediation(r *entity.Remediation) Remediation {
	remediationDate := r.RemediationDate.Format(time.RFC3339)
	expirationDate := r.ExpirationDate.Format(time.RFC3339)
	remediationType := RemediationTypeValues(r.Type)
	remediationSeverity := SeverityValues(r.Severity.String())
	return Remediation{
		ID:              fmt.Sprintf("%d", r.Id),
		Description:     &r.Description,
		Severity:        &remediationSeverity,
		Type:            &remediationType,
		Service:         &r.Service,
		ServiceID:       lo.ToPtr(fmt.Sprintf("%d", r.ServiceId)),
		Image:           &r.Component,
		ImageID:         lo.ToPtr(fmt.Sprintf("%d", r.ComponentId)),
		Vulnerability:   &r.Issue,
		VulnerabilityID: lo.ToPtr(fmt.Sprintf("%d", r.IssueId)),
		RemediationDate: &remediationDate,
		ExpirationDate:  &expirationDate,
		RemediatedBy:    &r.RemediatedBy,
		Metadata:        getModelMetadata(r.Metadata),
	}
}

func NewPatch(p *entity.Patch) Patch {
	return Patch{
		ID:                   fmt.Sprintf("%d", p.Id),
		ServiceID:            lo.ToPtr(fmt.Sprintf("%d", p.ServiceId)),
		ServiceName:          &p.ServiceName,
		ComponentVersionID:   lo.ToPtr(fmt.Sprintf("%d", p.ComponentVersionId)),
		ComponentVersionName: &p.ComponentVersionName,
		Metadata:             getModelMetadata(p.Metadata),
	}
}

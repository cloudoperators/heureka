// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"database/sql"
	"time"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
)

func GetInt64Value(v sql.NullInt64) int64 {
	if v.Valid {
		return v.Int64
	} else {
		return -1
	}
}

func GetInt16Value(v sql.NullInt16) int16 {
	if v.Valid {
		return v.Int16
	} else {
		return -1
	}
}

func GetStringValue(v sql.NullString) string {
	if v.Valid {
		return v.String
	} else {
		return ""
	}
}

func GetTimeValue(v sql.NullTime) time.Time {
	if v.Valid {
		return v.Time
	} else {
		return time.Unix(0, 0)
	}
}

func GetUserTypeValue(v sql.NullInt64) entity.UserType {
	if v.Valid {
		return entity.UserType(v.Int64)
	} else {
		return entity.InvalidUserType
	}
}

type DatabaseRow interface {
	IssueRow |
		IssueCountRow |
		GetIssuesByRow |
		IssueMatchRow |
		IssueAggregationsRow |
		IssueVariantRow |
		BaseIssueRepositoryRow |
		IssueRepositoryRow |
		IssueVariantWithRepository |
		ComponentRow |
		ComponentInstanceRow |
		ComponentVersionRow |
		BaseServiceRow |
		ServiceRow |
		GetServicesByRow |
		ServiceAggregationsRow |
		ActivityRow |
		UserRow |
		EvidenceRow |
		OwnerRow |
		SupportGroupRow |
		SupportGroupServiceRow |
		ActivityHasIssueRow |
		ActivityHasServiceRow |
		IssueRepositoryServiceRow |
		IssueMatchChangeRow |
		ServiceIssueVariantRow
}

type IssueRow struct {
	Id          sql.NullInt64  `db:"issue_id" json:"id"`
	Type        sql.NullString `db:"issue_type" json:"type"`
	PrimaryName sql.NullString `db:"issue_primary_name" json:"primary_name"`
	Description sql.NullString `db:"issue_description" json:"description"`
	CreatedAt   sql.NullTime   `db:"issue_created_at" json:"created_at"`
	DeletedAt   sql.NullTime   `db:"issue_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt   sql.NullTime   `db:"issue_updated_at" json:"updated_at"`
}

func (ir *IssueRow) AsIssue() entity.Issue {
	return entity.Issue{
		Id:            GetInt64Value(ir.Id),
		PrimaryName:   GetStringValue(ir.PrimaryName),
		Type:          entity.NewIssueType(GetStringValue(ir.Type)),
		Description:   GetStringValue(ir.Description),
		IssueVariants: []entity.IssueVariant{},
		IssueMatches:  []entity.IssueMatch{},
		Activity:      []entity.Activity{},
		CreatedAt:     GetTimeValue(ir.CreatedAt),
		DeletedAt:     GetTimeValue(ir.DeletedAt),
		UpdatedAt:     GetTimeValue(ir.UpdatedAt),
	}
}

type GetIssuesByRow struct {
	IssueAggregationsRow
	IssueRow
}

type IssueAggregationsRow struct {
	Activites                     sql.NullInt64 `db:"agg_activities"`
	IssueMatches                  sql.NullInt64 `db:"agg_issue_matches"`
	AffectedServices              sql.NullInt64 `db:"agg_affected_services"`
	ComponentVersions             sql.NullInt64 `db:"agg_component_versions"`
	AffectedComponentInstances    sql.NullInt64 `db:"agg_affected_component_instances"`
	EarliestTargetRemediationDate sql.NullTime  `db:"agg_earliest_target_remediation_date"`
	EarliestDiscoveryDate         sql.NullTime  `db:"agg_earliest_discovery_date"`
}

func (ibr *GetIssuesByRow) AsIssueWithAggregations() entity.IssueWithAggregations {
	return entity.IssueWithAggregations{
		IssueAggregations: entity.IssueAggregations{
			Activites:                     GetInt64Value(ibr.IssueAggregationsRow.Activites),
			IssueMatches:                  GetInt64Value(ibr.IssueAggregationsRow.IssueMatches),
			AffectedServices:              GetInt64Value(ibr.IssueAggregationsRow.AffectedServices),
			ComponentVersions:             GetInt64Value(ibr.IssueAggregationsRow.ComponentVersions),
			AffectedComponentInstances:    GetInt64Value(ibr.IssueAggregationsRow.AffectedComponentInstances),
			EarliestTargetRemediationDate: GetTimeValue(ibr.IssueAggregationsRow.EarliestTargetRemediationDate),
			EarliestDiscoveryDate:         GetTimeValue(ibr.IssueAggregationsRow.EarliestDiscoveryDate),
		},
		Issue: entity.Issue{
			Id:            GetInt64Value(ibr.IssueRow.Id),
			PrimaryName:   GetStringValue(ibr.IssueRow.PrimaryName),
			Type:          entity.NewIssueType(GetStringValue(ibr.Type)),
			Description:   GetStringValue(ibr.IssueRow.Description),
			IssueVariants: []entity.IssueVariant{},
			IssueMatches:  []entity.IssueMatch{},
			Activity:      []entity.Activity{},
			CreatedAt:     GetTimeValue(ibr.IssueRow.CreatedAt),
			DeletedAt:     GetTimeValue(ibr.IssueRow.DeletedAt),
			UpdatedAt:     GetTimeValue(ibr.IssueRow.UpdatedAt),
		},
	}
}

func (ibr *GetIssuesByRow) AsIssue() entity.Issue {
	return entity.Issue{
		Id:            GetInt64Value(ibr.IssueRow.Id),
		PrimaryName:   GetStringValue(ibr.IssueRow.PrimaryName),
		Type:          entity.NewIssueType(GetStringValue(ibr.Type)),
		Description:   GetStringValue(ibr.IssueRow.Description),
		IssueVariants: []entity.IssueVariant{},
		IssueMatches:  []entity.IssueMatch{},
		Activity:      []entity.Activity{},
		CreatedAt:     GetTimeValue(ibr.IssueRow.CreatedAt),
		DeletedAt:     GetTimeValue(ibr.IssueRow.DeletedAt),
		UpdatedAt:     GetTimeValue(ibr.IssueRow.UpdatedAt),
	}
}

type IssueCountRow struct {
	Count sql.NullInt64  `db:"issue_count"`
	Type  sql.NullString `db:"issue_type"`
}

func (icr *IssueCountRow) AsIssueCount() entity.IssueCount {
	return entity.IssueCount{
		Count: GetInt64Value(icr.Count),
		Type:  entity.NewIssueType(GetStringValue(icr.Type)),
	}
}

func (ir *IssueRow) FromIssue(i *entity.Issue) {
	ir.Id = sql.NullInt64{Int64: i.Id, Valid: true}
	ir.PrimaryName = sql.NullString{String: i.PrimaryName, Valid: true}
	ir.Type = sql.NullString{String: i.Type.String(), Valid: true}
	ir.Description = sql.NullString{String: i.Description, Valid: true}
	ir.CreatedAt = sql.NullTime{Time: i.CreatedAt, Valid: true}
	ir.DeletedAt = sql.NullTime{Time: i.DeletedAt, Valid: true}
	ir.UpdatedAt = sql.NullTime{Time: i.UpdatedAt, Valid: true}
}

type IssueMatchRow struct {
	Id                    sql.NullInt64  `db:"issuematch_id" json:"id"`
	Status                sql.NullString `db:"issuematch_status" json:"status"`
	Vector                sql.NullString `db:"issuematch_vector" json:"vector"`
	Rating                sql.NullString `db:"issuematch_rating" json:"rating"`
	UserId                sql.NullInt64  `db:"issuematch_user_id" json:"user_id"`
	ComponentInstanceId   sql.NullInt64  `db:"issuematch_component_instance_id" json:"component_instance_id"`
	IssueId               sql.NullInt64  `db:"issuematch_issue_id" json:"issue_id"`
	RemediationDate       sql.NullTime   `db:"issuematch_remediation_date" json:"remediation_date"`
	TargetRemediationDate sql.NullTime   `db:"issuematch_target_remediation_date" json:"target_remediation_date"`
	CreatedAt             sql.NullTime   `db:"issuematch_created_at" json:"created_at"`
	DeletedAt             sql.NullTime   `db:"issuematch_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt             sql.NullTime   `db:"issuematch_updated_at" json:"updated_at"`
}

func (imr IssueMatchRow) AsIssueMatch() entity.IssueMatch {
	return entity.IssueMatch{
		Id:                    GetInt64Value(imr.Id),
		Status:                entity.NewIssueMatchStatusValue(GetStringValue(imr.Status)),
		User:                  nil,
		UserId:                GetInt64Value(imr.UserId),
		ComponentInstance:     nil,
		ComponentInstanceId:   GetInt64Value(imr.ComponentInstanceId),
		Issue:                 nil,
		IssueId:               GetInt64Value(imr.IssueId),
		RemediationDate:       GetTimeValue(imr.RemediationDate),
		TargetRemediationDate: GetTimeValue(imr.TargetRemediationDate),
		Severity:              entity.NewSeverity(GetStringValue(imr.Vector)),
		CreatedAt:             GetTimeValue(imr.CreatedAt),
		DeletedAt:             GetTimeValue(imr.DeletedAt),
		UpdatedAt:             GetTimeValue(imr.UpdatedAt),
	}
}

func (imr *IssueMatchRow) FromIssueMatch(im *entity.IssueMatch) {
	imr.Id = sql.NullInt64{Int64: im.Id, Valid: true}
	imr.Status = sql.NullString{String: im.Status.String(), Valid: true}
	imr.Vector = sql.NullString{String: im.Severity.Cvss.Vector, Valid: true}
	imr.Rating = sql.NullString{String: im.Severity.Value, Valid: true}
	imr.UserId = sql.NullInt64{Int64: im.UserId, Valid: true}
	imr.ComponentInstanceId = sql.NullInt64{Int64: im.ComponentInstanceId, Valid: true}
	imr.IssueId = sql.NullInt64{Int64: im.IssueId, Valid: true}
	imr.RemediationDate = sql.NullTime{Time: im.RemediationDate, Valid: true}
	imr.TargetRemediationDate = sql.NullTime{Time: im.TargetRemediationDate, Valid: true}
	imr.CreatedAt = sql.NullTime{Time: im.CreatedAt, Valid: true}
	imr.DeletedAt = sql.NullTime{Time: im.DeletedAt, Valid: true}
	imr.UpdatedAt = sql.NullTime{Time: im.UpdatedAt, Valid: true}
}

type IssueRepositoryRow struct {
	BaseIssueRepositoryRow
	IssueRepositoryServiceRow
}

type BaseIssueRepositoryRow struct {
	Id        sql.NullInt64  `db:"issuerepository_id"`
	Name      sql.NullString `db:"issuerepository_name"`
	Url       sql.NullString `db:"issuerepository_url"`
	CreatedAt sql.NullTime   `db:"issuerepository_created_at" json:"created_at"`
	DeletedAt sql.NullTime   `db:"issuerepository_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt sql.NullTime   `db:"issuerepository_updated_at" json:"updated_at"`
}

func (irr *IssueRepositoryRow) FromIssueRepository(ir *entity.IssueRepository) {
	irr.Id = sql.NullInt64{Int64: ir.Id, Valid: true}
	irr.Name = sql.NullString{String: ir.Name, Valid: true}
	irr.Url = sql.NullString{String: ir.Url, Valid: true}
	irr.Priority = sql.NullInt64{Int64: ir.Priority, Valid: true}
	irr.ServiceId = sql.NullInt64{Int64: ir.ServiceId, Valid: true}
	irr.IssueRepositoryId = sql.NullInt64{Int64: ir.IssueRepositoryId, Valid: true}
}

func (birr *BaseIssueRepositoryRow) AsBaseIssueRepository() entity.BaseIssueRepository {
	return entity.BaseIssueRepository{
		Id:            GetInt64Value(birr.Id),
		Name:          GetStringValue(birr.Name),
		Url:           GetStringValue(birr.Url),
		IssueVariants: nil,
		Services:      nil,
		CreatedAt:     GetTimeValue(birr.CreatedAt),
		DeletedAt:     GetTimeValue(birr.DeletedAt),
		UpdatedAt:     GetTimeValue(birr.UpdatedAt),
	}
}

func (barr *BaseIssueRepositoryRow) AsIssueRepository() entity.IssueRepository {
	return entity.IssueRepository{
		BaseIssueRepository: entity.BaseIssueRepository{
			Id:            GetInt64Value(barr.Id),
			Name:          GetStringValue(barr.Name),
			Url:           GetStringValue(barr.Url),
			IssueVariants: nil,
			Services:      nil,
			CreatedAt:     GetTimeValue(barr.CreatedAt),
			DeletedAt:     GetTimeValue(barr.DeletedAt),
			UpdatedAt:     GetTimeValue(barr.UpdatedAt),
		},
	}
}

func (irr *IssueRepositoryRow) AsIssueRepository() entity.IssueRepository {
	return entity.IssueRepository{
		BaseIssueRepository: entity.BaseIssueRepository{
			Id:            GetInt64Value(irr.Id),
			Name:          GetStringValue(irr.Name),
			Url:           GetStringValue(irr.Url),
			IssueVariants: nil,
			Services:      nil,
			CreatedAt:     GetTimeValue(irr.BaseIssueRepositoryRow.CreatedAt),
			DeletedAt:     GetTimeValue(irr.BaseIssueRepositoryRow.DeletedAt),
			UpdatedAt:     GetTimeValue(irr.BaseIssueRepositoryRow.UpdatedAt),
		},
		IssueRepositoryService: entity.IssueRepositoryService{
			ServiceId:         GetInt64Value(irr.ServiceId),
			IssueRepositoryId: GetInt64Value(irr.IssueRepositoryId),
			Priority:          GetInt64Value(irr.Priority),
		}}
}

type IssueVariantRow struct {
	Id                sql.NullInt64  `db:"issuevariant_id" json:"id"`
	IssueId           sql.NullInt64  `db:"issuevariant_issue_id" json:"issue_id"`
	IssueRepositoryId sql.NullInt64  `db:"issuevariant_repository_id" json:"issue_repository_id"`
	SecondaryName     sql.NullString `db:"issuevariant_secondary_name" json:"secondary_name"`
	Vector            sql.NullString `db:"issuevariant_vector" json:"vector"`
	Rating            sql.NullString `db:"issuevariant_rating" json:"rating"`
	Description       sql.NullString `db:"issuevariant_description" json:"description"`
	CreatedAt         sql.NullTime   `db:"issuevariant_created_at" json:"created_at"`
	DeletedAt         sql.NullTime   `db:"issuevariant_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt         sql.NullTime   `db:"issuevariant_updated_at" json:"updated_at"`
}

func (ivr *IssueVariantRow) AsIssueVariant(repository *entity.IssueRepository) entity.IssueVariant {
	return entity.IssueVariant{
		Id:                GetInt64Value(ivr.Id),
		IssueRepositoryId: GetInt64Value(ivr.IssueRepositoryId),
		IssueRepository:   repository,
		SecondaryName:     GetStringValue(ivr.SecondaryName),
		IssueId:           GetInt64Value(ivr.IssueId),
		Issue:             nil,
		Severity:          entity.NewSeverity(GetStringValue(ivr.Vector)),
		Description:       GetStringValue(ivr.Description),
		CreatedAt:         GetTimeValue(ivr.CreatedAt),
		DeletedAt:         GetTimeValue(ivr.DeletedAt),
		UpdatedAt:         GetTimeValue(ivr.UpdatedAt),
	}
}

func (ivr *IssueVariantRow) FromIssueVariant(iv *entity.IssueVariant) {
	ivr.Id = sql.NullInt64{Int64: iv.Id, Valid: true}
	ivr.IssueRepositoryId = sql.NullInt64{Int64: iv.IssueRepositoryId, Valid: true}
	ivr.SecondaryName = sql.NullString{String: iv.SecondaryName, Valid: true}
	ivr.IssueId = sql.NullInt64{Int64: iv.IssueId, Valid: true}
	ivr.Vector = sql.NullString{String: iv.Severity.Cvss.Vector, Valid: true}
	ivr.Rating = sql.NullString{String: iv.Severity.Value, Valid: true}
	ivr.Description = sql.NullString{String: iv.Description, Valid: true}
	ivr.CreatedAt = sql.NullTime{Time: iv.CreatedAt, Valid: true}
	ivr.DeletedAt = sql.NullTime{Time: iv.DeletedAt, Valid: true}
	ivr.UpdatedAt = sql.NullTime{Time: iv.UpdatedAt, Valid: true}
}

type IssueVariantWithRepository struct {
	IssueRepositoryRow
	IssueVariantRow
}

func (ivwr *IssueVariantWithRepository) AsIssueVariantEntry() entity.IssueVariant {
	rep := ivwr.IssueRepositoryRow.AsIssueRepository()
	return entity.IssueVariant{
		Id:                GetInt64Value(ivwr.IssueVariantRow.Id),
		IssueRepositoryId: GetInt64Value(ivwr.IssueRepositoryId),
		IssueRepository:   &rep,
		SecondaryName:     GetStringValue(ivwr.IssueVariantRow.SecondaryName),
		IssueId:           GetInt64Value(ivwr.IssueId),
		Issue:             nil,
		Severity:          entity.NewSeverity(GetStringValue(ivwr.Vector)),
		Description:       GetStringValue(ivwr.Description),
		CreatedAt:         GetTimeValue(ivwr.IssueVariantRow.CreatedAt),
		DeletedAt:         GetTimeValue(ivwr.IssueVariantRow.DeletedAt),
		UpdatedAt:         GetTimeValue(ivwr.IssueVariantRow.UpdatedAt),
	}
}

type ServiceIssueVariantRow struct {
	IssueRepositoryRow
	IssueVariantRow
	IssueRepositoryServiceRow
	ComponentInstanceRow
}

func (siv *ServiceIssueVariantRow) AsServiceIssueVariantEntry() entity.ServiceIssueVariant {
	rep := siv.IssueRepositoryRow.AsIssueRepository()
	return entity.ServiceIssueVariant{
		IssueVariant: entity.IssueVariant{
			Id:                GetInt64Value(siv.IssueVariantRow.Id),
			IssueRepositoryId: GetInt64Value(siv.IssueRepositoryRow.IssueRepositoryId),
			IssueRepository:   &rep,
			SecondaryName:     GetStringValue(siv.IssueVariantRow.SecondaryName),
			IssueId:           GetInt64Value(siv.IssueId),
			Issue:             nil,
			Severity:          entity.NewSeverity(GetStringValue(siv.Vector)),
			Description:       GetStringValue(siv.Description),
			CreatedAt:         GetTimeValue(siv.IssueVariantRow.CreatedAt),
			DeletedAt:         GetTimeValue(siv.IssueVariantRow.DeletedAt),
			UpdatedAt:         GetTimeValue(siv.IssueVariantRow.UpdatedAt),
		},
		ServiceId: GetInt64Value(siv.IssueRepositoryServiceRow.ServiceId),
		Priority:  GetInt64Value(siv.Priority),
	}
}

type ComponentRow struct {
	Id        sql.NullInt64  `db:"component_id" json:"id"`
	Name      sql.NullString `db:"component_name" json:"name"`
	Type      sql.NullString `db:"component_type" json:"type"`
	CreatedAt sql.NullTime   `db:"component_created_at" json:"created_at"`
	DeletedAt sql.NullTime   `db:"component_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt sql.NullTime   `db:"component_updated_at" json:"updated_at"`
}

func (cr *ComponentRow) AsComponent() entity.Component {
	return entity.Component{
		Id:        GetInt64Value(cr.Id),
		Name:      GetStringValue(cr.Name),
		Type:      GetStringValue(cr.Type),
		CreatedAt: GetTimeValue(cr.CreatedAt),
		DeletedAt: GetTimeValue(cr.DeletedAt),
		UpdatedAt: GetTimeValue(cr.UpdatedAt),
	}
}

func (cr *ComponentRow) FromComponent(c *entity.Component) {
	cr.Id = sql.NullInt64{Int64: c.Id, Valid: true}
	cr.Name = sql.NullString{String: c.Name, Valid: true}
	cr.Type = sql.NullString{String: c.Type, Valid: true}
	cr.CreatedAt = sql.NullTime{Time: c.CreatedAt, Valid: true}
	cr.DeletedAt = sql.NullTime{Time: c.DeletedAt, Valid: true}
	cr.UpdatedAt = sql.NullTime{Time: c.UpdatedAt, Valid: true}
}

type ComponentVersionRow struct {
	Id          sql.NullInt64  `db:"componentversion_id" json:"id"`
	Version     sql.NullString `db:"componentversion_version" json:"version"`
	ComponentId sql.NullInt64  `db:"componentversion_component_id"`
	CreatedAt   sql.NullTime   `db:"componentversion_created_at" json:"created_at"`
	DeletedAt   sql.NullTime   `db:"componentversion_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt   sql.NullTime   `db:"componentversion_updated_at" json:"updated_at"`
}

func (cvr *ComponentVersionRow) AsComponentVersion() entity.ComponentVersion {
	return entity.ComponentVersion{
		Id:          GetInt64Value(cvr.Id),
		Version:     GetStringValue(cvr.Version),
		ComponentId: GetInt64Value(cvr.ComponentId),
		CreatedAt:   GetTimeValue(cvr.CreatedAt),
		DeletedAt:   GetTimeValue(cvr.DeletedAt),
		UpdatedAt:   GetTimeValue(cvr.UpdatedAt),
	}
}

func (cvr *ComponentVersionRow) FromComponentVersion(cv *entity.ComponentVersion) {
	cvr.Id = sql.NullInt64{Int64: cv.Id, Valid: true}
	cvr.Version = sql.NullString{String: cv.Version, Valid: true}
	cvr.ComponentId = sql.NullInt64{Int64: cv.ComponentId, Valid: true}
	cvr.CreatedAt = sql.NullTime{Time: cv.CreatedAt, Valid: true}
	cvr.DeletedAt = sql.NullTime{Time: cv.DeletedAt, Valid: true}
	cvr.UpdatedAt = sql.NullTime{Time: cv.UpdatedAt, Valid: true}
}

type SupportGroupRow struct {
	Id        sql.NullInt64  `db:"supportgroup_id" json:"id"`
	Name      sql.NullString `db:"supportgroup_name" json:"name"`
	CreatedAt sql.NullTime   `db:"supportgroup_created_at" json:"created_at"`
	DeletedAt sql.NullTime   `db:"supportgroup_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt sql.NullTime   `db:"supportgroup_updated_at" json:"updated_at"`
}

func (sgr *SupportGroupRow) AsSupportGroup() entity.SupportGroup {
	return entity.SupportGroup{
		Id:        GetInt64Value(sgr.Id),
		Name:      GetStringValue(sgr.Name),
		CreatedAt: GetTimeValue(sgr.CreatedAt),
		DeletedAt: GetTimeValue(sgr.DeletedAt),
		UpdatedAt: GetTimeValue(sgr.UpdatedAt),
	}
}

func (sgr *SupportGroupRow) FromSupportGroup(sg *entity.SupportGroup) {
	sgr.Id = sql.NullInt64{Int64: sg.Id, Valid: true}
	sgr.Name = sql.NullString{String: sg.Name, Valid: true}
	sgr.CreatedAt = sql.NullTime{Time: sg.CreatedAt, Valid: true}
	sgr.DeletedAt = sql.NullTime{Time: sg.DeletedAt, Valid: true}
	sgr.UpdatedAt = sql.NullTime{Time: sg.UpdatedAt, Valid: true}
}

type ServiceRow struct {
	BaseServiceRow
	IssueRepositoryServiceRow
}

type BaseServiceRow struct {
	Id        sql.NullInt64  `db:"service_id" json:"id"`
	Name      sql.NullString `db:"service_name" json:"name"`
	CreatedAt sql.NullTime   `db:"service_created_at" json:"created_at"`
	DeletedAt sql.NullTime   `db:"service_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt sql.NullTime   `db:"service_updated_at" json:"updated_at"`
}

func (bsr *BaseServiceRow) AsBaseService() entity.BaseService {
	return entity.BaseService{
		Id:         GetInt64Value(bsr.Id),
		Name:       GetStringValue(bsr.Name),
		Owners:     []entity.User{},
		Activities: []entity.Activity{},
		CreatedAt:  GetTimeValue(bsr.CreatedAt),
		DeletedAt:  GetTimeValue(bsr.DeletedAt),
		UpdatedAt:  GetTimeValue(bsr.UpdatedAt),
	}
}

func (bsr *BaseServiceRow) AsService() entity.Service {
	bs := bsr.AsBaseService()
	return entity.Service{
		BaseService: bs,
	}
}

func (sr *ServiceRow) AsService() entity.Service {
	return entity.Service{
		BaseService: entity.BaseService{
			Id:         GetInt64Value(sr.Id),
			Name:       GetStringValue(sr.Name),
			Owners:     []entity.User{},
			Activities: []entity.Activity{},
			CreatedAt:  GetTimeValue(sr.BaseServiceRow.CreatedAt),
			DeletedAt:  GetTimeValue(sr.BaseServiceRow.DeletedAt),
			UpdatedAt:  GetTimeValue(sr.BaseServiceRow.UpdatedAt),
		},
		IssueRepositoryService: entity.IssueRepositoryService{
			ServiceId:         GetInt64Value(sr.ServiceId),
			IssueRepositoryId: GetInt64Value(sr.IssueRepositoryId),
			Priority:          GetInt64Value(sr.Priority),
		}}
}

func (sr *ServiceRow) FromService(s *entity.Service) {
	sr.Id = sql.NullInt64{Int64: s.Id, Valid: true}
	sr.Name = sql.NullString{String: s.Name, Valid: true}
	sr.BaseServiceRow.CreatedAt = sql.NullTime{Time: s.BaseService.CreatedAt, Valid: true}
	sr.BaseServiceRow.DeletedAt = sql.NullTime{Time: s.BaseService.DeletedAt, Valid: true}
	sr.BaseServiceRow.UpdatedAt = sql.NullTime{Time: s.BaseService.UpdatedAt, Valid: true}
}

type GetServicesByRow struct {
	ServiceAggregationsRow
	ServiceRow
}

type ServiceAggregationsRow struct {
	ComponentInstances sql.NullInt64 `db:"agg_component_instances"`
	IssueMatches       sql.NullInt64 `db:"agg_issue_matches"`
}

func (sbr *GetServicesByRow) AsServiceWithAggregations() entity.ServiceWithAggregations {
	return entity.ServiceWithAggregations{
		ServiceAggregations: entity.ServiceAggregations{
			ComponentInstances: lo.Max([]int64{0, GetInt64Value(sbr.ServiceAggregationsRow.ComponentInstances)}),
			IssueMatches:       lo.Max([]int64{0, GetInt64Value(sbr.ServiceAggregationsRow.IssueMatches)}),
		},
		Service: entity.Service{
			BaseService: entity.BaseService{
				Id:         GetInt64Value(sbr.BaseServiceRow.Id),
				Name:       GetStringValue(sbr.BaseServiceRow.Name),
				Owners:     []entity.User{},
				Activities: []entity.Activity{},
				CreatedAt:  GetTimeValue(sbr.BaseServiceRow.CreatedAt),
				DeletedAt:  GetTimeValue(sbr.BaseServiceRow.DeletedAt),
				UpdatedAt:  GetTimeValue(sbr.BaseServiceRow.UpdatedAt),
			},
			IssueRepositoryService: entity.IssueRepositoryService{
				ServiceId:         GetInt64Value(sbr.IssueRepositoryServiceRow.ServiceId),
				IssueRepositoryId: GetInt64Value(sbr.IssueRepositoryServiceRow.IssueRepositoryId),
				Priority:          GetInt64Value(sbr.IssueRepositoryServiceRow.Priority),
			},
		},
	}
}

type ActivityRow struct {
	Id        sql.NullInt64  `db:"activity_id" json:"id"`
	Status    sql.NullString `db:"activity_status" json:"status"`
	CreatedAt sql.NullTime   `db:"activity_created_at" json:"created_at"`
	DeletedAt sql.NullTime   `db:"activity_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt sql.NullTime   `db:"activity_updated_at" json:"updated_at"`
}

func (ar *ActivityRow) AsActivity() entity.Activity {
	return entity.Activity{
		Id:        GetInt64Value(ar.Id),
		Status:    entity.ActivityStatusValue(GetStringValue(ar.Status)),
		Issues:    []entity.Issue{},
		Evidences: []entity.Evidence{},
		CreatedAt: GetTimeValue(ar.CreatedAt),
		DeletedAt: GetTimeValue(ar.DeletedAt),
		UpdatedAt: GetTimeValue(ar.UpdatedAt),
	}
}

func (ar *ActivityRow) FromActivity(a *entity.Activity) {
	ar.Id = sql.NullInt64{Int64: a.Id, Valid: true}
	ar.Status = sql.NullString{String: a.Status.String(), Valid: true}
	ar.CreatedAt = sql.NullTime{Time: a.CreatedAt, Valid: true}
	ar.DeletedAt = sql.NullTime{Time: a.DeletedAt, Valid: true}
	ar.UpdatedAt = sql.NullTime{Time: a.UpdatedAt, Valid: true}
}

type ComponentInstanceRow struct {
	Id                 sql.NullInt64  `db:"componentinstance_id" json:"id"`
	CCRN               sql.NullString `db:"componentinstance_ccrn" json:"ccrn"`
	Count              sql.NullInt16  `db:"componentinstance_count" json:"count"`
	ComponentVersionId sql.NullInt64  `db:"componentinstance_component_version_id"`
	ServiceId          sql.NullInt64  `db:"componentinstance_service_id"`
	CreatedAt          sql.NullTime   `db:"componentinstance_created_at" json:"created_at"`
	DeletedAt          sql.NullTime   `db:"componentinstance_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt          sql.NullTime   `db:"componentinstance_updated_at" json:"updated_at"`
}

func (cir *ComponentInstanceRow) AsComponentInstance() entity.ComponentInstance {
	return entity.ComponentInstance{
		Id:                 GetInt64Value(cir.Id),
		CCRN:               GetStringValue(cir.CCRN),
		Count:              GetInt16Value(cir.Count),
		ComponentVersion:   nil,
		ComponentVersionId: GetInt64Value(cir.ComponentVersionId),
		Service:            nil,
		ServiceId:          GetInt64Value(cir.ServiceId),
		CreatedAt:          GetTimeValue(cir.CreatedAt),
		DeletedAt:          GetTimeValue(cir.DeletedAt),
		UpdatedAt:          GetTimeValue(cir.UpdatedAt),
	}
}

func (cir *ComponentInstanceRow) FromComponentInstance(ci *entity.ComponentInstance) {
	cir.Id = sql.NullInt64{Int64: ci.Id, Valid: true}
	cir.CCRN = sql.NullString{String: ci.CCRN, Valid: true}
	cir.Count = sql.NullInt16{Int16: ci.Count, Valid: true}
	cir.ComponentVersionId = sql.NullInt64{Int64: ci.ComponentVersionId, Valid: true}
	cir.ServiceId = sql.NullInt64{Int64: ci.ServiceId, Valid: true}
	cir.CreatedAt = sql.NullTime{Time: ci.CreatedAt, Valid: true}
	cir.DeletedAt = sql.NullTime{Time: ci.DeletedAt, Valid: true}
	cir.UpdatedAt = sql.NullTime{Time: ci.UpdatedAt, Valid: true}
}

type UserRow struct {
	Id           sql.NullInt64  `db:"user_id" json:"id"`
	Name         sql.NullString `db:"user_name" json:"ccrn"`
	UniqueUserID sql.NullString `db:"user_unique_user_id" json:"unique_user_id"`
	Type         sql.NullInt64  `db:"user_type" json:"type"`
	CreatedAt    sql.NullTime   `db:"user_created_at" json:"created_at"`
	DeletedAt    sql.NullTime   `db:"user_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt    sql.NullTime   `db:"user_updated_at" json:"updated_at"`
}

func (ur *UserRow) AsUser() entity.User {
	return entity.User{
		Id:           GetInt64Value(ur.Id),
		Name:         GetStringValue(ur.Name),
		UniqueUserID: GetStringValue(ur.UniqueUserID),
		Type:         GetUserTypeValue(ur.Type),
		CreatedAt:    GetTimeValue(ur.CreatedAt),
		DeletedAt:    GetTimeValue(ur.DeletedAt),
		UpdatedAt:    GetTimeValue(ur.UpdatedAt),
	}
}

func (ur *UserRow) FromUser(u *entity.User) {
	ur.Id = sql.NullInt64{Int64: u.Id, Valid: true}
	ur.Name = sql.NullString{String: u.Name, Valid: true}
	ur.UniqueUserID = sql.NullString{String: u.UniqueUserID, Valid: true}
	ur.Type = sql.NullInt64{Int64: int64(u.Type), Valid: true}
	ur.CreatedAt = sql.NullTime{Time: u.CreatedAt, Valid: true}
	ur.DeletedAt = sql.NullTime{Time: u.DeletedAt, Valid: true}
	ur.UpdatedAt = sql.NullTime{Time: u.UpdatedAt, Valid: true}
}

type EvidenceRow struct {
	Id          sql.NullInt64  `db:"evidence_id" json:"id"`
	Description sql.NullString `db:"evidence_description" json:"description"`
	Type        sql.NullString `db:"evidence_type" json:"type"`
	Vector      sql.NullString `db:"evidence_vector" json:"vector"`
	Rating      sql.NullString `db:"evidence_rating" json:"rating"`
	RAAEnd      sql.NullTime   `db:"evidence_raa_end" json:"raa_end"`
	User        *UserRow       `json:"user,omitempty"`
	UserId      sql.NullInt64  `db:"evidence_author_id"`
	Activity    *ActivityRow   `json:"activity,omitempty"`
	ActivityId  sql.NullInt64  `db:"evidence_activity_id"`
	CreatedAt   sql.NullTime   `db:"evidence_created_at" json:"created_at"`
	DeletedAt   sql.NullTime   `db:"evidence_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt   sql.NullTime   `db:"evidence_updated_at" json:"updated_at"`
}

func (er *EvidenceRow) AsEvidence() entity.Evidence {
	return entity.Evidence{
		Id:          GetInt64Value(er.Id),
		Description: GetStringValue(er.Description),
		Type:        entity.NewEvidenceTypeValue(GetStringValue(er.Type)),
		Severity:    entity.NewSeverity(GetStringValue(er.Vector)),
		RaaEnd:      GetTimeValue(er.RAAEnd),
		User:        nil,
		UserId:      GetInt64Value(er.UserId),
		Activity:    nil,
		ActivityId:  GetInt64Value(er.ActivityId),
		CreatedAt:   GetTimeValue(er.CreatedAt),
		DeletedAt:   GetTimeValue(er.DeletedAt),
		UpdatedAt:   GetTimeValue(er.UpdatedAt),
	}
}

func (er *EvidenceRow) FromEvidence(e *entity.Evidence) {
	er.Id = sql.NullInt64{Int64: e.Id, Valid: true}
	er.Description = sql.NullString{String: e.Description, Valid: true}
	er.Type = sql.NullString{String: e.Type.String(), Valid: true}
	er.Vector = sql.NullString{String: e.Severity.Cvss.Vector, Valid: true}
	er.Rating = sql.NullString{String: e.Severity.Value, Valid: true}
	er.RAAEnd = sql.NullTime{Time: e.RaaEnd, Valid: true}
	er.UserId = sql.NullInt64{Int64: e.UserId, Valid: true}
	er.ActivityId = sql.NullInt64{Int64: e.ActivityId, Valid: true}
	er.CreatedAt = sql.NullTime{Time: e.CreatedAt, Valid: true}
	er.DeletedAt = sql.NullTime{Time: e.DeletedAt, Valid: true}
	er.UpdatedAt = sql.NullTime{Time: e.UpdatedAt, Valid: true}
}

type IssueMatchChangeRow struct {
	Id           sql.NullInt64  `db:"issuematchchange_id" json:"id"`
	IssueMatchId sql.NullInt64  `db:"issuematchchange_issue_match_id" json:"issue_match_id"`
	ActivityId   sql.NullInt64  `db:"issuematchchange_activity_id" json:"activity_id"`
	Action       sql.NullString `db:"issuematchchange_action" json:"action"`
	CreatedAt    sql.NullTime   `db:"issuematchchange_created_at" json:"created_at"`
	DeletedAt    sql.NullTime   `db:"issuematchchange_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt    sql.NullTime   `db:"issuematchchange_updated_at" json:"updated_at"`
}

func (imcr *IssueMatchChangeRow) AsIssueMatchChange() entity.IssueMatchChange {
	return entity.IssueMatchChange{
		Id:           GetInt64Value(imcr.Id),
		IssueMatchId: GetInt64Value(imcr.IssueMatchId),
		ActivityId:   GetInt64Value(imcr.ActivityId),
		Action:       GetStringValue(imcr.Action),
		CreatedAt:    GetTimeValue(imcr.CreatedAt),
		DeletedAt:    GetTimeValue(imcr.DeletedAt),
		UpdatedAt:    GetTimeValue(imcr.UpdatedAt),
	}
}

func (imcr *IssueMatchChangeRow) FromIssueMatchChange(imc *entity.IssueMatchChange) {
	imcr.Id = sql.NullInt64{Int64: imc.Id, Valid: true}
	imcr.IssueMatchId = sql.NullInt64{Int64: imc.IssueMatchId, Valid: true}
	imcr.ActivityId = sql.NullInt64{Int64: imc.ActivityId, Valid: true}
	imcr.Action = sql.NullString{String: imc.Action, Valid: true}
	imcr.CreatedAt = sql.NullTime{Time: imc.CreatedAt, Valid: true}
	imcr.DeletedAt = sql.NullTime{Time: imc.DeletedAt, Valid: true}
	imcr.UpdatedAt = sql.NullTime{Time: imc.UpdatedAt, Valid: true}
}

type OwnerRow struct {
	ServiceId sql.NullInt64 `db:"owner_service_id" json:"service_id"`
	UserId    sql.NullInt64 `db:"owner_user_id" json:"user_id"`
	CreatedAt sql.NullTime  `db:"owner_created_at" json:"created_at"`
	DeletedAt sql.NullTime  `db:"owner_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt sql.NullTime  `db:"owner_updated_at" json:"updated_at"`
}

type SupportGroupUserRow struct {
	SupportGroupId sql.NullInt64 `db:"supportgroupuser_support_group_id" json:"support_group_id"`
	UserId         sql.NullInt64 `db:"supportgroupuser_user_id" json:"user_id"`
	CreatedAt      sql.NullTime  `db:"supportgroupuser_created_at" json:"created_at"`
	DeletedAt      sql.NullTime  `db:"supportgroupuser_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt      sql.NullTime  `db:"supportgroupuser_updated_at" json:"updated_at"`
}

type SupportGroupServiceRow struct {
	SupportGroupId sql.NullInt64 `db:"supportgroupservice_support_group_id" json:"support_group_id"`
	ServiceId      sql.NullInt64 `db:"supportgroupservice_service_id" json:"service_id"`
	CreatedAt      sql.NullTime  `db:"supportgroupservice_created_at" json:"created_at"`
	DeletedAt      sql.NullTime  `db:"supportgroupservice_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt      sql.NullTime  `db:"supportgroupservice_updated_at" json:"updated_at"`
}

type ActivityHasIssueRow struct {
	ActivityId sql.NullInt64 `db:"activityhasissue_activity_id" json:"activity_id"`
	IssueId    sql.NullInt64 `db:"activityhasissue_issue_id" json:"issue_id"`
	CreatedAt  sql.NullTime  `db:"activityhasissue_created_at" json:"created_at"`
	DeletedAt  sql.NullTime  `db:"activityhasissue_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt  sql.NullTime  `db:"activityhasissue_updated_at" json:"updated_at"`
}

type ActivityHasServiceRow struct {
	ActivityId sql.NullInt64 `db:"activityhasservice_activity_id" json:"activity_id"`
	ServiceId  sql.NullInt64 `db:"activityhasservice_service_id" json:"service_id"`
	CreatedAt  sql.NullTime  `db:"activityhasservice_created_at" json:"created_at"`
	DeletedAt  sql.NullTime  `db:"activityhasservice_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt  sql.NullTime  `db:"activityhasservice_updated_at" json:"updated_at"`
}

type ComponentVersionIssueRow struct {
	ComponentVersionId sql.NullInt64 `db:"componentversionissue_component_version_id" json:"component_version_id"`
	IssueId            sql.NullInt64 `db:"componentversionissue_issue_id" json:"issue_id"`
	CreatedAt          sql.NullTime  `db:"componentversionissue_created_at" json:"created_at"`
	DeletedAt          sql.NullTime  `db:"componentversionissue_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt          sql.NullTime  `db:"componentversionissue_updated_at" json:"updated_at"`
}

type IssueMatchEvidenceRow struct {
	EvidenceId   sql.NullInt64 `db:"issuematchevidence_evidence_id" json:"evidence_id"`
	IssueMatchId sql.NullInt64 `db:"issuematchevidence_issue_match_id" json:"issue_match_id"`
	CreatedAt    sql.NullTime  `db:"issuematchevidence_created_at" json:"created_at"`
	DeletedAt    sql.NullTime  `db:"issuematchevidence_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt    sql.NullTime  `db:"issuematchevidence_updated_at" json:"updated_at"`
}

type IssueRepositoryServiceRow struct {
	IssueRepositoryId sql.NullInt64 `db:"issuerepositoryservice_issue_repository_id" json:"issue_repository_id"`
	ServiceId         sql.NullInt64 `db:"issuerepositoryservice_service_id" json:"service_id"`
	Priority          sql.NullInt64 `db:"issuerepositoryservice_priority" json:"priority"`
	CreatedAt         sql.NullTime  `db:"issuerepositoryservice_created_at" json:"created_at"`
	DeletedAt         sql.NullTime  `db:"issuerepositoryservice_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt         sql.NullTime  `db:"issuerepositoryservice_updated_at" json:"updated_at"`
}

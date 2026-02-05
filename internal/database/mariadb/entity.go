// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"database/sql"
	"time"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/util"
	"github.com/samber/lo"
)

func GetInt64Value(v sql.NullInt64) int64 {
	if v.Valid {
		return v.Int64
	} else {
		return -1
	}
}

func GetBoolValue(v sql.NullBool) bool {
	if v.Valid {
		return v.Bool
	} else {
		return false
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

func IsValidId(id int64) bool {
	return id > 0
}

// RowComposite is a composite type that contains all the row types for the database
// This is used to unmarshal the database rows into the corresponding entity types in a dynamical manner
type RowComposite struct {
	*IssueRow
	*IssueCountRow
	*GetIssuesByRow
	*IssueMatchRow
	*IssueAggregationsRow
	*IssueVariantRow
	*BaseIssueRepositoryRow
	*IssueRepositoryRow
	*IssueVariantWithRepository
	*ComponentRow
	*ComponentInstanceRow
	*ComponentVersionRow
	*BaseServiceRow
	*ServiceRow
	*GetServicesByRow
	*ServiceAggregationsRow
	*ActivityRow
	*UserRow
	*EvidenceRow
	*OwnerRow
	*SupportGroupRow
	*SupportGroupServiceRow
	*ActivityHasIssueRow
	*ActivityHasServiceRow
	*IssueRepositoryServiceRow
	*ServiceIssueVariantRow
	*RatingCount
	*RemediationRow
	*PatchRow
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
		ServiceIssueVariantRow |
		RatingCount |
		RowComposite |
		PatchRow
}

type RatingCount struct {
	Critical sql.NullInt64 `db:"critical_count"`
	High     sql.NullInt64 `db:"high_count"`
	Medium   sql.NullInt64 `db:"medium_count"`
	Low      sql.NullInt64 `db:"low_count"`
	None     sql.NullInt64 `db:"none_count"`
}

func (rc *RatingCount) AsIssueSeverityCounts() entity.IssueSeverityCounts {
	isc := entity.IssueSeverityCounts{
		Critical: GetInt64Value(rc.Critical),
		High:     GetInt64Value(rc.High),
		Medium:   GetInt64Value(rc.Medium),
		Low:      GetInt64Value(rc.Low),
		None:     GetInt64Value(rc.None),
	}
	isc.Total = isc.Critical + isc.High + isc.Medium + isc.Low + isc.None
	return isc
}

type IssueRow struct {
	Id          sql.NullInt64  `db:"issue_id" json:"id"`
	Type        sql.NullString `db:"issue_type" json:"type"`
	PrimaryName sql.NullString `db:"issue_primary_name" json:"primary_name"`
	Description sql.NullString `db:"issue_description" json:"description"`
	CreatedAt   sql.NullTime   `db:"issue_created_at" json:"created_at"`
	CreatedBy   sql.NullInt64  `db:"issue_created_by" json:"created_by"`
	DeletedAt   sql.NullTime   `db:"issue_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt   sql.NullTime   `db:"issue_updated_at" json:"updated_at"`
	UpdatedBy   sql.NullInt64  `db:"issue_updated_by" json:"updated_by"`
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
		Metadata: entity.Metadata{
			CreatedAt: GetTimeValue(ir.CreatedAt),
			CreatedBy: GetInt64Value(ir.CreatedBy),
			DeletedAt: GetTimeValue(ir.DeletedAt),
			UpdatedAt: GetTimeValue(ir.UpdatedAt),
			UpdatedBy: GetInt64Value(ir.UpdatedBy),
		},
	}
}

type GetIssuesByRow struct {
	IssueAggregationsRow
	IssueRow
}

type IssueAggregationsRow struct {
	Activities                    sql.NullInt64 `db:"agg_activities"`
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
			Activities:                    lo.Max([]int64{0, GetInt64Value(ibr.IssueAggregationsRow.Activities)}),
			IssueMatches:                  lo.Max([]int64{0, GetInt64Value(ibr.IssueAggregationsRow.IssueMatches)}),
			AffectedServices:              lo.Max([]int64{0, GetInt64Value(ibr.IssueAggregationsRow.AffectedServices)}),
			ComponentVersions:             lo.Max([]int64{0, GetInt64Value(ibr.IssueAggregationsRow.ComponentVersions)}),
			AffectedComponentInstances:    lo.Max([]int64{0, GetInt64Value(ibr.IssueAggregationsRow.AffectedComponentInstances)}),
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
			Metadata: entity.Metadata{
				CreatedAt: GetTimeValue(ibr.IssueRow.CreatedAt),
				CreatedBy: GetInt64Value(ibr.IssueRow.CreatedBy),
				DeletedAt: GetTimeValue(ibr.IssueRow.DeletedAt),
				UpdatedAt: GetTimeValue(ibr.IssueRow.UpdatedAt),
				UpdatedBy: GetInt64Value(ibr.IssueRow.UpdatedBy),
			},
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
		Metadata: entity.Metadata{
			CreatedAt: GetTimeValue(ibr.IssueRow.CreatedAt),
			CreatedBy: GetInt64Value(ibr.IssueRow.CreatedBy),
			DeletedAt: GetTimeValue(ibr.IssueRow.DeletedAt),
			UpdatedAt: GetTimeValue(ibr.IssueRow.UpdatedAt),
			UpdatedBy: GetInt64Value(ibr.IssueRow.UpdatedBy),
		},
	}
}

type IssueCountRow struct {
	Count sql.NullInt64  `db:"issue_count"`
	Value sql.NullString `db:"issue_value"`
}

func (icr *IssueCountRow) AsIssueCount() entity.IssueCount {
	return entity.IssueCount{
		Count: GetInt64Value(icr.Count),
		Value: GetStringValue(icr.Value),
	}
}

func (ir *IssueRow) FromIssue(i *entity.Issue) {
	ir.Id = sql.NullInt64{Int64: i.Id, Valid: true}
	ir.PrimaryName = sql.NullString{String: i.PrimaryName, Valid: true}
	ir.Type = sql.NullString{String: i.Type.String(), Valid: true}
	ir.Description = sql.NullString{String: i.Description, Valid: true}
	ir.CreatedAt = sql.NullTime{Time: i.CreatedAt, Valid: true}
	ir.CreatedBy = sql.NullInt64{Int64: i.CreatedBy, Valid: true}
	ir.DeletedAt = sql.NullTime{Time: i.DeletedAt, Valid: true}
	ir.UpdatedAt = sql.NullTime{Time: i.UpdatedAt, Valid: true}
	ir.UpdatedBy = sql.NullInt64{Int64: i.UpdatedBy, Valid: true}
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
	CreatedBy             sql.NullInt64  `db:"issuematch_created_by" json:"created_by"`
	DeletedAt             sql.NullTime   `db:"issuematch_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt             sql.NullTime   `db:"issuematch_updated_at" json:"updated_at"`
	UpdatedBy             sql.NullInt64  `db:"issuematch_updated_by" json:"updated_by"`
}

func (imr IssueMatchRow) AsIssueMatch() entity.IssueMatch {
	var severity entity.Severity
	if imr.Vector.String == "" {
		severity = entity.NewSeverityFromRating(entity.SeverityValues(imr.Rating.String))
	} else {
		severity = entity.NewSeverity(GetStringValue(imr.Vector))
	}
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
		Severity:              severity,
		Metadata: entity.Metadata{
			CreatedAt: GetTimeValue(imr.CreatedAt),
			CreatedBy: GetInt64Value(imr.CreatedBy),
			DeletedAt: GetTimeValue(imr.DeletedAt),
			UpdatedAt: GetTimeValue(imr.UpdatedAt),
			UpdatedBy: GetInt64Value(imr.UpdatedBy),
		},
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
	imr.CreatedBy = sql.NullInt64{Int64: im.CreatedBy, Valid: true}
	imr.DeletedAt = sql.NullTime{Time: im.DeletedAt, Valid: true}
	imr.UpdatedAt = sql.NullTime{Time: im.UpdatedAt, Valid: true}
	imr.UpdatedBy = sql.NullInt64{Int64: im.UpdatedBy, Valid: true}
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
	CreatedBy sql.NullInt64  `db:"issuerepository_created_by" json:"created_by"`
	DeletedAt sql.NullTime   `db:"issuerepository_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt sql.NullTime   `db:"issuerepository_updated_at" json:"updated_at"`
	UpdatedBy sql.NullInt64  `db:"issuerepository_updated_by" json:"updated_by"`
}

func (irr *IssueRepositoryRow) FromIssueRepository(ir *entity.IssueRepository) {
	irr.Id = sql.NullInt64{Int64: ir.Id, Valid: true}
	irr.Name = sql.NullString{String: ir.Name, Valid: true}
	irr.Url = sql.NullString{String: ir.Url, Valid: true}
	irr.Priority = sql.NullInt64{Int64: ir.Priority, Valid: true}
	irr.ServiceId = sql.NullInt64{Int64: ir.ServiceId, Valid: true}
	irr.IssueRepositoryId = sql.NullInt64{Int64: ir.IssueRepositoryId, Valid: true}
	irr.BaseIssueRepositoryRow.CreatedAt = sql.NullTime{Time: ir.BaseIssueRepository.CreatedAt, Valid: true}
	irr.BaseIssueRepositoryRow.CreatedBy = sql.NullInt64{Int64: ir.BaseIssueRepository.CreatedBy, Valid: true}
	irr.BaseIssueRepositoryRow.DeletedAt = sql.NullTime{Time: ir.BaseIssueRepository.DeletedAt, Valid: true}
	irr.BaseIssueRepositoryRow.UpdatedAt = sql.NullTime{Time: ir.BaseIssueRepository.UpdatedAt, Valid: true}
	irr.BaseIssueRepositoryRow.UpdatedBy = sql.NullInt64{Int64: ir.BaseIssueRepository.UpdatedBy, Valid: true}
}

func (birr *BaseIssueRepositoryRow) AsBaseIssueRepository() entity.BaseIssueRepository {
	return entity.BaseIssueRepository{
		Id:            GetInt64Value(birr.Id),
		Name:          GetStringValue(birr.Name),
		Url:           GetStringValue(birr.Url),
		IssueVariants: nil,
		Services:      nil,
		Metadata: entity.Metadata{
			CreatedAt: GetTimeValue(birr.CreatedAt),
			CreatedBy: GetInt64Value(birr.CreatedBy),
			DeletedAt: GetTimeValue(birr.DeletedAt),
			UpdatedAt: GetTimeValue(birr.UpdatedAt),
			UpdatedBy: GetInt64Value(birr.UpdatedBy),
		},
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
			Metadata: entity.Metadata{
				CreatedAt: GetTimeValue(barr.CreatedAt),
				CreatedBy: GetInt64Value(barr.CreatedBy),
				DeletedAt: GetTimeValue(barr.DeletedAt),
				UpdatedAt: GetTimeValue(barr.UpdatedAt),
				UpdatedBy: GetInt64Value(barr.UpdatedBy),
			},
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
			Metadata: entity.Metadata{
				CreatedAt: GetTimeValue(irr.BaseIssueRepositoryRow.CreatedAt),
				CreatedBy: GetInt64Value(irr.BaseIssueRepositoryRow.CreatedBy),
				DeletedAt: GetTimeValue(irr.BaseIssueRepositoryRow.DeletedAt),
				UpdatedAt: GetTimeValue(irr.BaseIssueRepositoryRow.UpdatedAt),
				UpdatedBy: GetInt64Value(irr.BaseIssueRepositoryRow.UpdatedBy),
			},
		},
		IssueRepositoryService: entity.IssueRepositoryService{
			ServiceId:         GetInt64Value(irr.ServiceId),
			IssueRepositoryId: GetInt64Value(irr.IssueRepositoryId),
			Priority:          GetInt64Value(irr.Priority),
		},
	}
}

type IssueVariantRow struct {
	Id                sql.NullInt64  `db:"issuevariant_id" json:"id"`
	IssueId           sql.NullInt64  `db:"issuevariant_issue_id" json:"issue_id"`
	IssueRepositoryId sql.NullInt64  `db:"issuevariant_repository_id" json:"issue_repository_id"`
	SecondaryName     sql.NullString `db:"issuevariant_secondary_name" json:"secondary_name"`
	Vector            sql.NullString `db:"issuevariant_vector" json:"vector"`
	Rating            sql.NullString `db:"issuevariant_rating" json:"rating"`
	RatingNumerical   sql.NullInt64  `db:"issuevariant_rating_num" json:"rating_numerical"`
	Description       sql.NullString `db:"issuevariant_description" json:"description"`
	ExternalUrl       sql.NullString `db:"issuevariant_external_url" json:"external_url"`
	CreatedAt         sql.NullTime   `db:"issuevariant_created_at" json:"created_at"`
	CreatedBy         sql.NullInt64  `db:"issuevariant_created_by" json:"created_by"`
	DeletedAt         sql.NullTime   `db:"issuevariant_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt         sql.NullTime   `db:"issuevariant_updated_at" json:"updated_at"`
	UpdatedBy         sql.NullInt64  `db:"issuevariant_updated_by" json:"updated_by"`
}

func (ivr *IssueVariantRow) AsIssueVariant(repository *entity.IssueRepository) entity.IssueVariant {
	var severity entity.Severity
	if ivr.Vector.String == "" {
		severity = entity.NewSeverityFromRating(entity.SeverityValues(ivr.Rating.String))
	} else {
		severity = entity.NewSeverity(GetStringValue(ivr.Vector))
	}

	return entity.IssueVariant{
		Id:                GetInt64Value(ivr.Id),
		IssueRepositoryId: GetInt64Value(ivr.IssueRepositoryId),
		IssueRepository:   repository,
		SecondaryName:     GetStringValue(ivr.SecondaryName),
		IssueId:           GetInt64Value(ivr.IssueId),
		Issue:             nil,
		Severity:          severity,
		Description:       GetStringValue(ivr.Description),
		ExternalUrl:       GetStringValue(ivr.ExternalUrl),
		Metadata: entity.Metadata{
			CreatedAt: GetTimeValue(ivr.CreatedAt),
			CreatedBy: GetInt64Value(ivr.CreatedBy),
			DeletedAt: GetTimeValue(ivr.DeletedAt),
			UpdatedAt: GetTimeValue(ivr.UpdatedAt),
			UpdatedBy: GetInt64Value(ivr.UpdatedBy),
		},
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
	ivr.ExternalUrl = sql.NullString{String: iv.ExternalUrl, Valid: true}
	ivr.CreatedAt = sql.NullTime{Time: iv.CreatedAt, Valid: true}
	ivr.CreatedBy = sql.NullInt64{Int64: iv.CreatedBy, Valid: true}
	ivr.DeletedAt = sql.NullTime{Time: iv.DeletedAt, Valid: true}
	ivr.UpdatedAt = sql.NullTime{Time: iv.UpdatedAt, Valid: true}
	ivr.UpdatedBy = sql.NullInt64{Int64: iv.UpdatedBy, Valid: true}
}

type IssueVariantWithRepository struct {
	IssueRepositoryRow
	IssueVariantRow
}

func (ivwr *IssueVariantWithRepository) AsIssueVariantEntry() entity.IssueVariant {
	rep := ivwr.IssueRepositoryRow.AsIssueRepository()

	var severity entity.Severity
	if ivwr.Vector.String == "" {
		severity = entity.NewSeverityFromRating(entity.SeverityValues(ivwr.Rating.String))
	} else {
		severity = entity.NewSeverity(GetStringValue(ivwr.Vector))
	}

	return entity.IssueVariant{
		Id:                GetInt64Value(ivwr.IssueVariantRow.Id),
		IssueRepositoryId: GetInt64Value(ivwr.IssueRepositoryId),
		IssueRepository:   &rep,
		SecondaryName:     GetStringValue(ivwr.IssueVariantRow.SecondaryName),
		IssueId:           GetInt64Value(ivwr.IssueId),
		Issue:             nil,
		Severity:          severity,
		Description:       GetStringValue(ivwr.Description),
		ExternalUrl:       GetStringValue(ivwr.ExternalUrl),
		Metadata: entity.Metadata{
			CreatedAt: GetTimeValue(ivwr.IssueVariantRow.CreatedAt),
			CreatedBy: GetInt64Value(ivwr.CreatedBy),
			DeletedAt: GetTimeValue(ivwr.IssueVariantRow.DeletedAt),
			UpdatedAt: GetTimeValue(ivwr.IssueVariantRow.UpdatedAt),
			UpdatedBy: GetInt64Value(ivwr.UpdatedBy),
		},
	}
}

type ServiceIssueVariantRow struct {
	IssueRepositoryRow
	IssueVariantRow
}

func (siv *ServiceIssueVariantRow) AsServiceIssueVariantEntry() entity.ServiceIssueVariant {
	rep := siv.IssueRepositoryRow.AsIssueRepository()

	var severity entity.Severity
	if siv.Vector.String == "" {
		severity = entity.NewSeverityFromRating(entity.SeverityValues(siv.Rating.String))
	} else {
		severity = entity.NewSeverity(GetStringValue(siv.Vector))
	}

	return entity.ServiceIssueVariant{
		IssueVariant: entity.IssueVariant{
			Id:                GetInt64Value(siv.IssueVariantRow.Id),
			IssueRepositoryId: GetInt64Value(siv.IssueRepositoryRow.IssueRepositoryId),
			IssueRepository:   &rep,
			SecondaryName:     GetStringValue(siv.IssueVariantRow.SecondaryName),
			IssueId:           GetInt64Value(siv.IssueId),
			Issue:             nil,
			Severity:          severity,
			Description:       GetStringValue(siv.Description),
			Metadata: entity.Metadata{
				CreatedAt: GetTimeValue(siv.IssueVariantRow.CreatedAt),
				CreatedBy: GetInt64Value(siv.CreatedBy),
				DeletedAt: GetTimeValue(siv.IssueVariantRow.DeletedAt),
				UpdatedAt: GetTimeValue(siv.IssueVariantRow.UpdatedAt),
				UpdatedBy: GetInt64Value(siv.UpdatedBy),
			},
		},
		ServiceId: GetInt64Value(siv.IssueRepositoryServiceRow.ServiceId),
		Priority:  GetInt64Value(siv.Priority),
	}
}

type ComponentRow struct {
	Id           sql.NullInt64  `db:"component_id" json:"id"`
	CCRN         sql.NullString `db:"component_ccrn" json:"ccrn"`
	Repository   sql.NullString `db:"component_repository" json:"repository"`
	Organization sql.NullString `db:"component_organization" json:"organization"`
	Url          sql.NullString `db:"component_url" json:"url"`
	Type         sql.NullString `db:"component_type" json:"type"`
	CreatedAt    sql.NullTime   `db:"component_created_at" json:"created_at"`
	CreatedBy    sql.NullInt64  `db:"component_created_by" json:"created_by"`
	DeletedAt    sql.NullTime   `db:"component_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt    sql.NullTime   `db:"component_updated_at" json:"updated_at"`
	UpdatedBy    sql.NullInt64  `db:"component_updated_by" json:"updated_by"`
}

func (cr *ComponentRow) AsComponent() entity.Component {
	return entity.Component{
		Id:           GetInt64Value(cr.Id),
		CCRN:         GetStringValue(cr.CCRN),
		Repository:   GetStringValue(cr.Repository),
		Organization: GetStringValue(cr.Organization),
		Url:          GetStringValue(cr.Url),
		Type:         GetStringValue(cr.Type),
		Metadata: entity.Metadata{
			CreatedAt: GetTimeValue(cr.CreatedAt),
			CreatedBy: GetInt64Value(cr.CreatedBy),
			DeletedAt: GetTimeValue(cr.DeletedAt),
			UpdatedAt: GetTimeValue(cr.UpdatedAt),
			UpdatedBy: GetInt64Value(cr.UpdatedBy),
		},
	}
}

func (cr *ComponentRow) FromComponent(c *entity.Component) {
	cr.Id = sql.NullInt64{Int64: c.Id, Valid: true}
	cr.CCRN = sql.NullString{String: c.CCRN, Valid: true}
	cr.Repository = sql.NullString{String: c.Repository, Valid: true}
	cr.Organization = sql.NullString{String: c.Organization, Valid: true}
	cr.Url = sql.NullString{String: c.Url, Valid: true}
	cr.Type = sql.NullString{String: c.Type, Valid: true}
	cr.CreatedAt = sql.NullTime{Time: c.CreatedAt, Valid: true}
	cr.CreatedBy = sql.NullInt64{Int64: c.CreatedBy, Valid: true}
	cr.DeletedAt = sql.NullTime{Time: c.DeletedAt, Valid: true}
	cr.UpdatedAt = sql.NullTime{Time: c.UpdatedAt, Valid: true}
	cr.UpdatedBy = sql.NullInt64{Int64: c.UpdatedBy, Valid: true}
}

type ComponentVersionRow struct {
	Id           sql.NullInt64  `db:"componentversion_id" json:"id"`
	Version      sql.NullString `db:"componentversion_version" json:"version"`
	Tag          sql.NullString `db:"componentversion_tag" json:"tag"`
	Repository   sql.NullString `db:"componentversion_repository" json:"repository"`
	Organization sql.NullString `db:"componentversion_organization" json:"organization"`
	ComponentId  sql.NullInt64  `db:"componentversion_component_id"`
	CreatedAt    sql.NullTime   `db:"componentversion_created_at" json:"created_at"`
	CreatedBy    sql.NullInt64  `db:"componentversion_created_by" json:"created_by"`
	DeletedAt    sql.NullTime   `db:"componentversion_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt    sql.NullTime   `db:"componentversion_updated_at" json:"updated_at"`
	UpdatedBy    sql.NullInt64  `db:"componentversion_updated_by" json:"updated_by"`
}

func (cvr *ComponentVersionRow) AsComponentVersion() entity.ComponentVersion {
	return entity.ComponentVersion{
		Id:           GetInt64Value(cvr.Id),
		Version:      GetStringValue(cvr.Version),
		Tag:          GetStringValue(cvr.Tag),
		Repository:   GetStringValue(cvr.Repository),
		Organization: GetStringValue(cvr.Organization),
		ComponentId:  GetInt64Value(cvr.ComponentId),
		Metadata: entity.Metadata{
			CreatedAt: GetTimeValue(cvr.CreatedAt),
			CreatedBy: GetInt64Value(cvr.CreatedBy),
			DeletedAt: GetTimeValue(cvr.DeletedAt),
			UpdatedAt: GetTimeValue(cvr.UpdatedAt),
			UpdatedBy: GetInt64Value(cvr.UpdatedBy),
		},
	}
}

func (cvr *ComponentVersionRow) FromComponentVersion(cv *entity.ComponentVersion) {
	cvr.Id = sql.NullInt64{Int64: cv.Id, Valid: true}
	cvr.Version = sql.NullString{String: cv.Version, Valid: true}
	cvr.Tag = sql.NullString{String: cv.Tag, Valid: true}
	cvr.Repository = sql.NullString{String: cv.Repository, Valid: true}
	cvr.Organization = sql.NullString{String: cv.Organization, Valid: true}
	cvr.ComponentId = sql.NullInt64{Int64: cv.ComponentId, Valid: true}
	cvr.CreatedAt = sql.NullTime{Time: cv.CreatedAt, Valid: true}
	cvr.CreatedBy = sql.NullInt64{Int64: cv.CreatedBy, Valid: true}
	cvr.DeletedAt = sql.NullTime{Time: cv.DeletedAt, Valid: true}
	cvr.UpdatedAt = sql.NullTime{Time: cv.UpdatedAt, Valid: true}
	cvr.UpdatedBy = sql.NullInt64{Int64: cv.UpdatedBy, Valid: true}
}

type SupportGroupRow struct {
	Id        sql.NullInt64  `db:"supportgroup_id" json:"id"`
	CCRN      sql.NullString `db:"supportgroup_ccrn" json:"ccrn"`
	CreatedAt sql.NullTime   `db:"supportgroup_created_at" json:"created_at"`
	CreatedBy sql.NullInt64  `db:"supportgroup_created_by" json:"created_by"`
	DeletedAt sql.NullTime   `db:"supportgroup_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt sql.NullTime   `db:"supportgroup_updated_at" json:"updated_at"`
	UpdatedBy sql.NullInt64  `db:"supportgroup_updated_by" json:"updated_by"`
}

func (sgr *SupportGroupRow) AsSupportGroup() entity.SupportGroup {
	return entity.SupportGroup{
		Id:   GetInt64Value(sgr.Id),
		CCRN: GetStringValue(sgr.CCRN),
		Metadata: entity.Metadata{
			CreatedAt: GetTimeValue(sgr.CreatedAt),
			CreatedBy: GetInt64Value(sgr.CreatedBy),
			DeletedAt: GetTimeValue(sgr.DeletedAt),
			UpdatedAt: GetTimeValue(sgr.UpdatedAt),
			UpdatedBy: GetInt64Value(sgr.UpdatedBy),
		},
	}
}

func (sgr *SupportGroupRow) FromSupportGroup(sg *entity.SupportGroup) {
	sgr.Id = sql.NullInt64{Int64: sg.Id, Valid: true}
	sgr.CCRN = sql.NullString{String: sg.CCRN, Valid: true}
	sgr.CreatedAt = sql.NullTime{Time: sg.CreatedAt, Valid: true}
	sgr.CreatedBy = sql.NullInt64{Int64: sg.CreatedBy, Valid: true}
	sgr.DeletedAt = sql.NullTime{Time: sg.DeletedAt, Valid: true}
	sgr.UpdatedAt = sql.NullTime{Time: sg.UpdatedAt, Valid: true}
	sgr.UpdatedBy = sql.NullInt64{Int64: sg.UpdatedBy, Valid: true}
}

type ServiceRow struct {
	BaseServiceRow
	IssueRepositoryServiceRow
}

type BaseServiceRow struct {
	Id        sql.NullInt64  `db:"service_id" json:"id"`
	CCRN      sql.NullString `db:"service_ccrn" json:"ccrn"`
	Domain    sql.NullString `db:"service_domain" json:"domain"`
	Region    sql.NullString `db:"service_region" json:"region"`
	CreatedAt sql.NullTime   `db:"service_created_at" json:"created_at"`
	CreatedBy sql.NullInt64  `db:"service_created_by" json:"created_by"`
	DeletedAt sql.NullTime   `db:"service_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt sql.NullTime   `db:"service_updated_at" json:"updated_at"`
	UpdatedBy sql.NullInt64  `db:"service_updated_by" json:"updated_by"`
}

func (bsr *BaseServiceRow) AsBaseService() entity.BaseService {
	return entity.BaseService{
		Id:         GetInt64Value(bsr.Id),
		CCRN:       GetStringValue(bsr.CCRN),
		Domain:     GetStringValue(bsr.Domain),
		Region:     GetStringValue(bsr.Region),
		Owners:     []entity.User{},
		Activities: []entity.Activity{},
		Metadata: entity.Metadata{
			CreatedAt: GetTimeValue(bsr.CreatedAt),
			CreatedBy: GetInt64Value(bsr.CreatedBy),
			DeletedAt: GetTimeValue(bsr.DeletedAt),
			UpdatedAt: GetTimeValue(bsr.UpdatedAt),
			UpdatedBy: GetInt64Value(bsr.UpdatedBy),
		},
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
			CCRN:       GetStringValue(sr.CCRN),
			Domain:     GetStringValue(sr.Domain),
			Region:     GetStringValue(sr.Region),
			Owners:     []entity.User{},
			Activities: []entity.Activity{},
			Metadata: entity.Metadata{
				CreatedAt: GetTimeValue(sr.BaseServiceRow.CreatedAt),
				CreatedBy: GetInt64Value(sr.BaseServiceRow.CreatedBy),
				DeletedAt: GetTimeValue(sr.BaseServiceRow.DeletedAt),
				UpdatedAt: GetTimeValue(sr.BaseServiceRow.UpdatedAt),
				UpdatedBy: GetInt64Value(sr.BaseServiceRow.UpdatedBy),
			},
		},
		IssueRepositoryService: entity.IssueRepositoryService{
			ServiceId:         GetInt64Value(sr.ServiceId),
			IssueRepositoryId: GetInt64Value(sr.IssueRepositoryId),
			Priority:          GetInt64Value(sr.Priority),
		},
	}
}

func (sr *ServiceRow) FromService(s *entity.Service) {
	sr.Id = sql.NullInt64{Int64: s.Id, Valid: true}
	sr.CCRN = sql.NullString{String: s.CCRN, Valid: true}
	sr.Domain = sql.NullString{String: s.Domain, Valid: true}
	sr.Region = sql.NullString{String: s.Region, Valid: true}
	sr.BaseServiceRow.CreatedAt = sql.NullTime{Time: s.BaseService.CreatedAt, Valid: true}
	sr.BaseServiceRow.CreatedBy = sql.NullInt64{Int64: s.BaseService.CreatedBy, Valid: true}
	sr.BaseServiceRow.DeletedAt = sql.NullTime{Time: s.BaseService.DeletedAt, Valid: true}
	sr.BaseServiceRow.UpdatedAt = sql.NullTime{Time: s.BaseService.UpdatedAt, Valid: true}
	sr.BaseServiceRow.UpdatedBy = sql.NullInt64{Int64: s.BaseService.UpdatedBy, Valid: true}
}

type GetServicesByRow struct {
	ServiceAggregationsRow
	ServiceRow
}

type ServiceAggregationsRow struct {
	ComponentInstances sql.NullInt64 `db:"service_agg_component_instances"`
	IssueMatches       sql.NullInt64 `db:"service_agg_issue_matches"`
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
				CCRN:       GetStringValue(sbr.BaseServiceRow.CCRN),
				Domain:     GetStringValue(sbr.BaseServiceRow.Domain),
				Region:     GetStringValue(sbr.BaseServiceRow.Region),
				Owners:     []entity.User{},
				Activities: []entity.Activity{},
				Metadata: entity.Metadata{
					CreatedAt: GetTimeValue(sbr.BaseServiceRow.CreatedAt),
					CreatedBy: GetInt64Value(sbr.BaseServiceRow.CreatedBy),
					DeletedAt: GetTimeValue(sbr.BaseServiceRow.DeletedAt),
					UpdatedAt: GetTimeValue(sbr.BaseServiceRow.UpdatedAt),
					UpdatedBy: GetInt64Value(sbr.BaseServiceRow.UpdatedBy),
				},
			},
			IssueRepositoryService: entity.IssueRepositoryService{
				ServiceId:         GetInt64Value(sbr.IssueRepositoryServiceRow.ServiceId),
				IssueRepositoryId: GetInt64Value(sbr.IssueRepositoryServiceRow.IssueRepositoryId),
				Priority:          GetInt64Value(sbr.IssueRepositoryServiceRow.Priority),
			},
		},
	}
}

func (sar *ServiceAggregationsRow) AsServiceAggregations() entity.ServiceAggregations {
	return entity.ServiceAggregations{
		ComponentInstances: lo.Max([]int64{0, GetInt64Value(sar.ComponentInstances)}),
		IssueMatches:       lo.Max([]int64{0, GetInt64Value(sar.IssueMatches)}),
	}
}

type ActivityRow struct {
	Id        sql.NullInt64  `db:"activity_id" json:"id"`
	Status    sql.NullString `db:"activity_status" json:"status"`
	CreatedAt sql.NullTime   `db:"activity_created_at" json:"created_at"`
	CreatedBy sql.NullInt64  `db:"activity_created_by" json:"created_by"`
	DeletedAt sql.NullTime   `db:"activity_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt sql.NullTime   `db:"activity_updated_at" json:"updated_at"`
	UpdatedBy sql.NullInt64  `db:"activity_updated_by" json:"updated_by"`
}

func (ar *ActivityRow) AsActivity() entity.Activity {
	return entity.Activity{
		Id:        GetInt64Value(ar.Id),
		Status:    entity.ActivityStatusValue(GetStringValue(ar.Status)),
		Issues:    []entity.Issue{},
		Evidences: []entity.Evidence{},
		Metadata: entity.Metadata{
			CreatedAt: GetTimeValue(ar.CreatedAt),
			CreatedBy: GetInt64Value(ar.CreatedBy),
			DeletedAt: GetTimeValue(ar.DeletedAt),
			UpdatedAt: GetTimeValue(ar.UpdatedAt),
			UpdatedBy: GetInt64Value(ar.UpdatedBy),
		},
	}
}

func (ar *ActivityRow) FromActivity(a *entity.Activity) {
	ar.Id = sql.NullInt64{Int64: a.Id, Valid: true}
	ar.Status = sql.NullString{String: a.Status.String(), Valid: true}
	ar.CreatedAt = sql.NullTime{Time: a.CreatedAt, Valid: true}
	ar.CreatedBy = sql.NullInt64{Int64: a.CreatedBy, Valid: true}
	ar.DeletedAt = sql.NullTime{Time: a.DeletedAt, Valid: true}
	ar.UpdatedAt = sql.NullTime{Time: a.UpdatedAt, Valid: true}
	ar.UpdatedBy = sql.NullInt64{Int64: a.UpdatedBy, Valid: true}
}

type ComponentInstanceRow struct {
	Id                 sql.NullInt64  `db:"componentinstance_id" json:"id"`
	CCRN               sql.NullString `db:"componentinstance_ccrn" json:"ccrn"`
	Region             sql.NullString `db:"componentinstance_region" json:"region"`
	Cluster            sql.NullString `db:"componentinstance_cluster" json:"cluster"`
	Namespace          sql.NullString `db:"componentinstance_namespace" json:"namespace"`
	Domain             sql.NullString `db:"componentinstance_domain" json:"domain"`
	Project            sql.NullString `db:"componentinstance_project" json:"project"`
	Pod                sql.NullString `db:"componentinstance_pod" json:"pod"`
	Container          sql.NullString `db:"componentinstance_container" json:"container"`
	Type               sql.NullString `db:"componentinstance_type" json:"type"`
	ParentId           sql.NullInt64  `db:"componentinstance_parent_id"`
	Context            sql.NullString `db:"componentinstance_context" json:"context"`
	Count              sql.NullInt16  `db:"componentinstance_count" json:"count"`
	ComponentVersionId sql.NullInt64  `db:"componentinstance_component_version_id"`
	ServiceId          sql.NullInt64  `db:"componentinstance_service_id"`
	CreatedAt          sql.NullTime   `db:"componentinstance_created_at" json:"created_at"`
	CreatedBy          sql.NullInt64  `db:"componentinstance_created_by" json:"created_by"`
	DeletedAt          sql.NullTime   `db:"componentinstance_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt          sql.NullTime   `db:"componentinstance_updated_at" json:"updated_at"`
	UpdatedBy          sql.NullInt64  `db:"componentinstance_updated_by" json:"udpated_by"`
}

func (cir *ComponentInstanceRow) AsComponentInstance() entity.ComponentInstance {
	return entity.ComponentInstance{
		Id:                 GetInt64Value(cir.Id),
		CCRN:               GetStringValue(cir.CCRN),
		Region:             GetStringValue(cir.Region),
		Cluster:            GetStringValue(cir.Cluster),
		Namespace:          GetStringValue(cir.Namespace),
		Domain:             GetStringValue(cir.Domain),
		Project:            GetStringValue(cir.Project),
		Pod:                GetStringValue(cir.Pod),
		Container:          GetStringValue(cir.Container),
		Type:               entity.NewComponentInstanceType(GetStringValue(cir.Type)),
		ParentId:           GetInt64Value(cir.ParentId),
		Context:            (*entity.Json)(util.ConvertStrToJsonNoError(&cir.Context.String)),
		Count:              GetInt16Value(cir.Count),
		ComponentVersion:   nil,
		ComponentVersionId: GetInt64Value(cir.ComponentVersionId),
		Service:            nil,
		ServiceId:          GetInt64Value(cir.ServiceId),
		Metadata: entity.Metadata{
			CreatedAt: GetTimeValue(cir.CreatedAt),
			CreatedBy: GetInt64Value(cir.CreatedBy),
			DeletedAt: GetTimeValue(cir.DeletedAt),
			UpdatedAt: GetTimeValue(cir.UpdatedAt),
			UpdatedBy: GetInt64Value(cir.UpdatedBy),
		},
	}
}

func (cir *ComponentInstanceRow) FromComponentInstance(ci *entity.ComponentInstance) {
	if ci.ParentId > 0 {
		cir.ParentId = sql.NullInt64{Int64: ci.ParentId, Valid: true}
	}

	if ci.ComponentVersionId > 0 {
		cir.ComponentVersionId = sql.NullInt64{Int64: ci.ComponentVersionId, Valid: true}
	} else {
		cir.ComponentVersionId = sql.NullInt64{Valid: false}
	}

	cir.Id = sql.NullInt64{Int64: ci.Id, Valid: true}
	cir.CCRN = sql.NullString{String: ci.CCRN, Valid: true}
	cir.Region = sql.NullString{String: ci.Region, Valid: true}
	cir.Cluster = sql.NullString{String: ci.Cluster, Valid: true}
	cir.Namespace = sql.NullString{String: ci.Namespace, Valid: true}
	cir.Domain = sql.NullString{String: ci.Domain, Valid: true}
	cir.Project = sql.NullString{String: ci.Project, Valid: true}
	cir.Pod = sql.NullString{String: ci.Pod, Valid: true}
	cir.Container = sql.NullString{String: ci.Container, Valid: true}
	cir.Type = sql.NullString{String: ci.Type.String(), Valid: true}
	cir.Context = sql.NullString{String: ci.Context.String(), Valid: true}
	cir.Count = sql.NullInt16{Int16: ci.Count, Valid: true}
	cir.ServiceId = sql.NullInt64{Int64: ci.ServiceId, Valid: true}
	cir.CreatedAt = sql.NullTime{Time: ci.CreatedAt, Valid: true}
	cir.CreatedBy = sql.NullInt64{Int64: ci.CreatedBy, Valid: true}
	cir.DeletedAt = sql.NullTime{Time: ci.DeletedAt, Valid: true}
	cir.UpdatedAt = sql.NullTime{Time: ci.UpdatedAt, Valid: true}
	cir.UpdatedBy = sql.NullInt64{Int64: ci.UpdatedBy, Valid: true}
}

type UserRow struct {
	Id           sql.NullInt64  `db:"user_id" json:"id"`
	Name         sql.NullString `db:"user_name" json:"ccrn"`
	UniqueUserID sql.NullString `db:"user_unique_user_id" json:"unique_user_id"`
	Type         sql.NullInt64  `db:"user_type" json:"type"`
	CreatedAt    sql.NullTime   `db:"user_created_at" json:"created_at"`
	CreatedBy    sql.NullInt64  `db:"user_created_by" json:"created_by"`
	DeletedAt    sql.NullTime   `db:"user_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt    sql.NullTime   `db:"user_updated_at" json:"updated_at"`
	UpdatedBy    sql.NullInt64  `db:"user_updated_by" json:"Updated_by"`
	Email        sql.NullString `db:"user_email" json:"email"`
}

func (ur *UserRow) AsUser() entity.User {
	return entity.User{
		Id:           GetInt64Value(ur.Id),
		Name:         GetStringValue(ur.Name),
		UniqueUserID: GetStringValue(ur.UniqueUserID),
		Type:         GetUserTypeValue(ur.Type),
		Email:        GetStringValue(ur.Email),
		Metadata: entity.Metadata{
			CreatedAt: GetTimeValue(ur.CreatedAt),
			CreatedBy: GetInt64Value(ur.CreatedBy),
			DeletedAt: GetTimeValue(ur.DeletedAt),
			UpdatedAt: GetTimeValue(ur.UpdatedAt),
			UpdatedBy: GetInt64Value(ur.UpdatedBy),
		},
	}
}

func (ur *UserRow) FromUser(u *entity.User) {
	ur.Id = sql.NullInt64{Int64: u.Id, Valid: true}
	ur.Name = sql.NullString{String: u.Name, Valid: true}
	ur.UniqueUserID = sql.NullString{String: u.UniqueUserID, Valid: true}
	ur.Type = sql.NullInt64{Int64: int64(u.Type), Valid: true}
	ur.CreatedAt = sql.NullTime{Time: u.CreatedAt, Valid: true}
	ur.CreatedBy = sql.NullInt64{Int64: u.CreatedBy, Valid: true}
	ur.DeletedAt = sql.NullTime{Time: u.DeletedAt, Valid: true}
	ur.UpdatedAt = sql.NullTime{Time: u.UpdatedAt, Valid: true}
	ur.UpdatedBy = sql.NullInt64{Int64: u.UpdatedBy, Valid: true}
	ur.Email = sql.NullString{String: u.Email, Valid: true}
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
	CreatedBy   sql.NullInt64  `db:"evidence_created_by" json:"created_by"`
	DeletedAt   sql.NullTime   `db:"evidence_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt   sql.NullTime   `db:"evidence_updated_at" json:"updated_at"`
	UpdatedBy   sql.NullInt64  `db:"evidence_updated_by" json:"updated_by"`
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
		Metadata: entity.Metadata{
			CreatedAt: GetTimeValue(er.CreatedAt),
			CreatedBy: GetInt64Value(er.CreatedBy),
			DeletedAt: GetTimeValue(er.DeletedAt),
			UpdatedAt: GetTimeValue(er.UpdatedAt),
			UpdatedBy: GetInt64Value(er.UpdatedBy),
		},
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
	er.CreatedBy = sql.NullInt64{Int64: e.CreatedBy, Valid: true}
	er.DeletedAt = sql.NullTime{Time: e.DeletedAt, Valid: true}
	er.UpdatedAt = sql.NullTime{Time: e.UpdatedAt, Valid: true}
	er.UpdatedBy = sql.NullInt64{Int64: e.UpdatedBy, Valid: true}
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

type ScannerRunRow struct {
	RunID       sql.NullInt64  `db:"scannerrun_run_id"`
	UUID        sql.NullString `db:"scannerrun_uuid"`
	Tag         sql.NullString `db:"scannerrun_tag"`
	StartRun    sql.NullTime   `db:"scannerrun_start_run"`
	EndRun      sql.NullTime   `db:"scannerrun_end_run"`
	IsCompleted sql.NullBool   `db:"scannerrun_is_completed"`
	CreatedAt   sql.NullTime   `db:"scannerrun_created_at" json:"created_at"`
	CreatedBy   sql.NullInt64  `db:"scannerrun_created_by" json:"created_by"`
	DeletedAt   sql.NullTime   `db:"scannerrun_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt   sql.NullTime   `db:"scannerrun_updated_at" json:"updated_at"`
	UpdatedBy   sql.NullInt64  `db:"scannerrun_updated_by" json:"updated_by"`
}

func (srr *ScannerRunRow) AsScannerRun() entity.ScannerRun {
	return entity.ScannerRun{
		RunID:     GetInt64Value(srr.RunID),
		UUID:      GetStringValue(srr.UUID),
		Tag:       GetStringValue(srr.Tag),
		StartRun:  GetTimeValue(srr.StartRun),
		EndRun:    GetTimeValue(srr.EndRun),
		Completed: GetBoolValue(srr.IsCompleted),
		Metadata: entity.Metadata{
			CreatedAt: GetTimeValue(srr.CreatedAt),
			CreatedBy: GetInt64Value(srr.CreatedBy),
			DeletedAt: GetTimeValue(srr.DeletedAt),
			UpdatedAt: GetTimeValue(srr.UpdatedAt),
			UpdatedBy: GetInt64Value(srr.UpdatedBy),
		},
	}
}

func (srr *ScannerRunRow) FromScannerRun(sr *entity.ScannerRun) {
	srr.RunID = sql.NullInt64{Int64: sr.RunID, Valid: true}
	srr.UUID = sql.NullString{String: sr.UUID, Valid: true}
	srr.Tag = sql.NullString{String: sr.Tag, Valid: true}
	srr.StartRun = sql.NullTime{Time: sr.StartRun, Valid: true}
	srr.EndRun = sql.NullTime{Time: sr.EndRun, Valid: true}
	srr.IsCompleted = sql.NullBool{Bool: sr.Completed, Valid: true}
	srr.CreatedAt = sql.NullTime{Time: sr.CreatedAt, Valid: true}
	srr.CreatedBy = sql.NullInt64{Int64: sr.CreatedBy, Valid: true}
	srr.DeletedAt = sql.NullTime{Time: sr.DeletedAt, Valid: true}
	srr.UpdatedAt = sql.NullTime{Time: sr.UpdatedAt, Valid: true}
	srr.UpdatedBy = sql.NullInt64{Int64: sr.UpdatedBy, Valid: true}
}

type RemediationRow struct {
	Id              sql.NullInt64  `db:"remediation_id" json:"id"`
	Type            sql.NullString `db:"remediation_type" json:"type"`
	Description     sql.NullString `db:"remediation_description" json:"description"`
	RemediationDate sql.NullTime   `db:"remediation_remediation_date" json:"remediation_date"`
	ExpirationDate  sql.NullTime   `db:"remediation_expiration_date" json:"expiry_date"`
	Severity        sql.NullString `db:"remediation_severity" json:"severity"`
	RemediatedBy    sql.NullString `db:"remediation_remediated_by" json:"remediated_by"`
	RemediatedById  sql.NullInt64  `db:"remediation_remediated_by_id" json:"remediated_by_id"`
	Service         sql.NullString `db:"remediation_service" json:"service"`
	ServiceId       sql.NullInt64  `db:"remediation_service_id" json:"service_id"`
	Component       sql.NullString `db:"remediation_component" json:"component"`
	ComponentId     sql.NullInt64  `db:"remediation_component_id" json:"component_id"`
	Issue           sql.NullString `db:"remediation_issue" json:"issue"`
	IssueId         sql.NullInt64  `db:"remediation_issue_id" json:"issue_id"`
	CreatedAt       sql.NullTime   `db:"remediation_created_at" json:"created_at"`
	CreatedBy       sql.NullInt64  `db:"remediation_created_by" json:"created_by"`
	DeletedAt       sql.NullTime   `db:"remediation_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt       sql.NullTime   `db:"remediation_updated_at" json:"updated_at"`
	UpdatedBy       sql.NullInt64  `db:"remediation_updated_by" json:"updated_by"`
}

func (rr *RemediationRow) AsRemediation() entity.Remediation {
	return entity.Remediation{
		Id:              GetInt64Value(rr.Id),
		Description:     GetStringValue(rr.Description),
		Type:            entity.NewRemediationType(GetStringValue(rr.Type)),
		Severity:        entity.NewSeverityValues(GetStringValue(rr.Severity)),
		Component:       GetStringValue(rr.Component),
		ComponentId:     GetInt64Value(rr.ComponentId),
		Service:         GetStringValue(rr.Service),
		ServiceId:       GetInt64Value(rr.ServiceId),
		Issue:           GetStringValue(rr.Issue),
		IssueId:         GetInt64Value(rr.IssueId),
		RemediationDate: GetTimeValue(rr.RemediationDate),
		ExpirationDate:  GetTimeValue(rr.ExpirationDate),
		RemediatedBy:    GetStringValue(rr.RemediatedBy),
		RemediatedById:  GetInt64Value(rr.RemediatedById),
		Metadata: entity.Metadata{
			CreatedAt: GetTimeValue(rr.CreatedAt),
			CreatedBy: GetInt64Value(rr.CreatedBy),
			DeletedAt: GetTimeValue(rr.DeletedAt),
			UpdatedAt: GetTimeValue(rr.UpdatedAt),
			UpdatedBy: GetInt64Value(rr.UpdatedBy),
		},
	}
}

func (rr *RemediationRow) FromRemediation(r *entity.Remediation) {
	rr.Id = sql.NullInt64{Int64: r.Id, Valid: true}
	rr.Description = sql.NullString{String: r.Description, Valid: true}
	rr.Type = sql.NullString{String: r.Type.String(), Valid: true}
	rr.Component = sql.NullString{String: r.Component, Valid: true}
	rr.ComponentId = sql.NullInt64{Int64: r.ComponentId, Valid: IsValidId(r.ComponentId)}
	rr.Service = sql.NullString{String: r.Service, Valid: true}
	rr.ServiceId = sql.NullInt64{Int64: r.ServiceId, Valid: true}
	rr.Issue = sql.NullString{String: r.Issue, Valid: true}
	rr.IssueId = sql.NullInt64{Int64: r.IssueId, Valid: true}
	rr.RemediationDate = sql.NullTime{Time: r.RemediationDate, Valid: true}
	rr.ExpirationDate = sql.NullTime{Time: r.ExpirationDate, Valid: true}
	rr.Severity = sql.NullString{String: r.Severity.String(), Valid: true}
	rr.RemediatedBy = sql.NullString{String: r.RemediatedBy, Valid: true}
	rr.RemediatedById = sql.NullInt64{Int64: r.RemediatedById, Valid: IsValidId(r.RemediatedById)}
	rr.CreatedAt = sql.NullTime{Time: r.CreatedAt, Valid: true}
	rr.CreatedBy = sql.NullInt64{Int64: r.CreatedBy, Valid: true}
	rr.DeletedAt = sql.NullTime{Time: r.DeletedAt, Valid: true}
	rr.UpdatedAt = sql.NullTime{Time: r.UpdatedAt, Valid: true}
	rr.UpdatedBy = sql.NullInt64{Int64: r.UpdatedBy, Valid: true}
}

type PatchRow struct {
	Id                   sql.NullInt64  `db:"patch_id" json:"id"`
	ServiceId            sql.NullInt64  `db:"patch_service_id" json:"service_id"`
	ServiceName          sql.NullString `db:"patch_service_name" json:"service_name"`
	ComponentVersionId   sql.NullInt64  `db:"patch_component_version_id" json:"compoidnent_version_id"`
	ComponentVersionName sql.NullString `db:"patch_component_version_name" json:"compoidnent_version_name"`
	CreatedAt            sql.NullTime   `db:"patch_created_at" json:"created_at"`
	CreatedBy            sql.NullInt64  `db:"patch_created_by" json:"created_by"`
	DeletedAt            sql.NullTime   `db:"patch_deleted_at" json:"deleted_at,omitempty"`
	UpdatedAt            sql.NullTime   `db:"patch_updated_at" json:"updated_at"`
	UpdatedBy            sql.NullInt64  `db:"patch_updated_by" json:"Updated_by"`
}

func (pr *PatchRow) AsPatch() entity.Patch {
	return entity.Patch{
		Id:                   GetInt64Value(pr.Id),
		ServiceId:            GetInt64Value(pr.ServiceId),
		ServiceName:          GetStringValue(pr.ServiceName),
		ComponentVersionId:   GetInt64Value(pr.ComponentVersionId),
		ComponentVersionName: GetStringValue(pr.ComponentVersionName),
		Metadata: entity.Metadata{
			CreatedAt: GetTimeValue(pr.CreatedAt),
			CreatedBy: GetInt64Value(pr.CreatedBy),
			DeletedAt: GetTimeValue(pr.DeletedAt),
			UpdatedAt: GetTimeValue(pr.UpdatedAt),
			UpdatedBy: GetInt64Value(pr.UpdatedBy),
		},
	}
}

func (pr *PatchRow) FromPatch(p *entity.Patch) {
	pr.Id = sql.NullInt64{Int64: p.Id, Valid: true}
	pr.ServiceId = sql.NullInt64{Int64: p.ServiceId, Valid: true}
	pr.ServiceName = sql.NullString{String: p.ServiceName, Valid: true}
	pr.ComponentVersionId = sql.NullInt64{Int64: p.ComponentVersionId, Valid: true}
	pr.ComponentVersionName = sql.NullString{String: p.ComponentVersionName, Valid: true}
	pr.CreatedAt = sql.NullTime{Time: p.CreatedAt, Valid: true}
	pr.CreatedBy = sql.NullInt64{Int64: p.CreatedBy, Valid: true}
	pr.DeletedAt = sql.NullTime{Time: p.DeletedAt, Valid: true}
	pr.UpdatedAt = sql.NullTime{Time: p.UpdatedAt, Valid: true}
	pr.UpdatedBy = sql.NullInt64{Int64: p.UpdatedBy, Valid: true}
}

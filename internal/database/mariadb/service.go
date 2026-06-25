// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"

	"github.com/cloudoperators/heureka/internal/entity"
)

var serviceObject = DbObject[*entity.Service, *entity.ServiceFilter, entity.ServiceResult, *entity.ServiceAggregations]{
	Prefix:       "service",
	TableName:    "Service",
	TableKey:     "S",
	DefaultOrder: entity.Order{By: entity.ServiceId, Direction: entity.OrderDirectionAsc},
	Aggregated:   true,
	Properties: []*Property[*entity.Service]{
		NewProperty("service_ccrn", func(s *entity.Service) (any, bool) { return s.CCRN, s.CCRN != "" }),
		NewProperty("service_domain", func(s *entity.Service) (any, bool) { return s.Domain, s.Domain != "" }),
		NewProperty("service_region", func(s *entity.Service) (any, bool) { return s.Region, s.Region != "" }),
		NewProperty("service_created_by", func(s *entity.Service) (any, bool) { return s.BaseService.CreatedBy, NoUpdate }),
		NewProperty("service_updated_by", func(s *entity.Service) (any, bool) { return s.BaseService.UpdatedBy, s.BaseService.UpdatedBy != 0 }),
	},
	FilterProperties: []*FilterProperty[*entity.ServiceFilter]{
		NewFilterProperty("S.service_ccrn = ?", func(filter *entity.ServiceFilter) any { return filter.CCRN }),
		NewFilterProperty("S.service_domain = ?", func(filter *entity.ServiceFilter) any { return filter.Domain }),
		NewFilterProperty("S.service_region = ?", func(filter *entity.ServiceFilter) any { return filter.Region }),
		NewFilterProperty("S.service_id = ?", func(filter *entity.ServiceFilter) any { return filter.Id }),
		NewFilterProperty("SG.supportgroup_ccrn = ?", func(filter *entity.ServiceFilter) any { return filter.SupportGroupCCRN }),
		NewFilterProperty("U.user_name = ?", func(filter *entity.ServiceFilter) any { return filter.OwnerName }),
		NewFilterProperty("IM.issuematch_issue_id = ?", func(filter *entity.ServiceFilter) any { return filter.IssueId }),
		NewFilterProperty("CI.componentinstance_id = ?", func(filter *entity.ServiceFilter) any { return filter.ComponentInstanceId }),
		NewFilterProperty("IRS.issuerepositoryservice_issue_repository_id = ?", func(filter *entity.ServiceFilter) any { return filter.IssueRepositoryId }),
		NewFilterProperty("SGS.supportgroupservice_support_group_id = ?", func(filter *entity.ServiceFilter) any { return filter.SupportGroupId }),
		NewFilterProperty("O.owner_user_id = ?", func(filter *entity.ServiceFilter) any { return filter.OwnerId }),
		NewFilterProperty("S.service_ccrn LIKE Concat('%',?,'%')", func(filter *entity.ServiceFilter) any { return filter.Search }),
		NewStateFilterProperty("S.service", func(filter *entity.ServiceFilter) any { return filter.State }),
	},
	JoinDefs: []*JoinDef[*entity.ServiceFilter]{
		{
			Name:      "O",
			Type:      LeftJoin,
			Table:     "Owner O",
			On:        "S.service_id = O.owner_service_id",
			Condition: func(f *entity.ServiceFilter, _ *Order) bool { return len(f.OwnerId) > 0 },
		},
		{
			Name:      "U",
			Type:      LeftJoin,
			Table:     "User U",
			On:        "O.owner_user_id = U.user_id",
			DependsOn: []string{"O"},
			Condition: func(f *entity.ServiceFilter, _ *Order) bool { return len(f.OwnerName) > 0 },
		},
		{
			Name:      "SGS",
			Type:      LeftJoin,
			Table:     "SupportGroupService SGS",
			On:        "S.service_id = SGS.supportgroupservice_service_id",
			Condition: func(f *entity.ServiceFilter, _ *Order) bool { return len(f.SupportGroupId) > 0 },
		},
		{
			Name:      "SG",
			Type:      LeftJoin,
			Table:     "SupportGroup SG",
			On:        "SGS.supportgroupservice_support_group_id = SG.supportgroup_id",
			DependsOn: []string{"SGS"},
			Condition: func(f *entity.ServiceFilter, _ *Order) bool { return len(f.SupportGroupCCRN) > 0 },
		},
		{
			Name:      "CI",
			Type:      LeftJoin,
			Table:     "ComponentInstance CI",
			On:        "S.service_id = CI.componentinstance_service_id",
			Condition: func(f *entity.ServiceFilter, _ *Order) bool { return len(f.ComponentInstanceId) > 0 },
		},
		{
			Name:      "IM",
			Type:      LeftJoin,
			Table:     "IssueMatch IM",
			On:        "CI.componentinstance_id = IM.issuematch_component_instance_id",
			DependsOn: []string{"CI"},
			Condition: func(f *entity.ServiceFilter, _ *Order) bool { return len(f.IssueId) > 0 },
		},
		{
			Name:      "IRS",
			Type:      LeftJoin,
			Table:     "IssueRepositoryService IRS",
			On:        "S.service_id = IRS.issuerepositoryservice_service_id",
			Condition: func(f *entity.ServiceFilter, _ *Order) bool { return len(f.IssueRepositoryId) > 0 },
		},
		{
			Name:      "SIC",
			Type:      LeftJoin,
			Table:     "mvServiceIssueCounts SIC",
			On:        "S.service_id = SIC.service_id",
			Condition: func(f *entity.ServiceFilter, order *Order) bool { return order.ByCount() },
		},
	},
	Attributes: []Attr{
		{Name: "ccrn", Order: entity.Order{By: entity.ServiceCcrn, Direction: entity.OrderDirectionAsc}},
		{Name: "domain", Order: entity.Order{By: entity.ServiceCcrn, Direction: entity.OrderDirectionAsc}},
		{Name: "region", Order: entity.Order{By: entity.ServiceCcrn, Direction: entity.OrderDirectionAsc}},
	},
	ExtraColumnsSelector: func(filter *entity.ServiceFilter, order *Order) []string {
		s := []string{}
		if filter != nil && len(filter.IssueRepositoryId) > 0 {
			s = append(s, "IRS.*")
		}

		for _, o := range order.Sequence() {
			switch o.By {
			case entity.CriticalCount:
				s = append(s, "SIC.critical_count")
			case entity.HighCount:
				s = append(s, "SIC.high_count")
			case entity.MediumCount:
				s = append(s, "SIC.medium_count")
			case entity.LowCount:
				s = append(s, "SIC.low_count")
			case entity.NoneCount:
				s = append(s, "SIC.none_count")
			}
		}

		return s
	},
	RowToData: func(e RowComposite, order []entity.Order) (*entity.Service, string) {
		s := entity.Service{BaseService: e.AsBaseService()}

		var isc entity.IssueSeverityCounts
		if e.RatingCount != nil {
			isc = e.AsIssueSeverityCounts()
		}

		cursor, _ := EncodeCursor(WithService(order, s, isc))

		return &s, cursor
	},
	RowToAggregatedData: func(e RowComposite, order []entity.Order) (*entity.Service, *entity.ServiceAggregations, string) {
		s := entity.Service{
			BaseService: e.AsBaseService(),
		}
		sa := e.AsServiceAggregations()

		var isc entity.IssueSeverityCounts
		if e.RatingCount != nil {
			isc = e.AsIssueSeverityCounts()
		}

		cursor, _ := EncodeCursor(WithService(order, s, isc))

		return &s, &sa, cursor
	},
	NewResult: func(s *entity.Service, sa *entity.ServiceAggregations, cursor string) entity.ServiceResult {
		return entity.ServiceResult{
			WithCursor:          entity.WithCursor{Value: cursor},
			Service:             s,
			ServiceAggregations: sa,
		}
	},
}

func (s *SqlDatabase) CountServices(ctx context.Context, filter *entity.ServiceFilter) (int64, error) {
	return serviceObject.Count(ctx, s.db, filter)
}

func (s *SqlDatabase) GetServices(ctx context.Context, filter *entity.ServiceFilter, order []entity.Order) ([]entity.ServiceResult, error) {
	return serviceObject.Get(ctx, s.db, filter, order)
}

func (s *SqlDatabase) GetServicesWithAggregations(ctx context.Context, filter *entity.ServiceFilter, order []entity.Order) ([]entity.ServiceResult, error) {
	aggDef := AggregationDef{
		Aggregations: []Aggregation{
			{
				Table:       "IssueMatchCounts",
				TableKey:    "IMC",
				Columns:     []string{"COUNT(IM.issuematch_id) AS service_agg_issue_matches"},
				ForcedJoins: []string{"IM"},
			},
			{
				Table:       "ComponentInstanceCounts",
				TableKey:    "CIC",
				Columns:     []string{"SUM(CI.componentinstance_count) AS service_agg_component_instances"},
				ForcedJoins: []string{"CI"},
			},
		},
		From:          "ComponentInstanceCounts CIC",
		Joins:         []string{"JOIN IssueMatchCounts IMC on CIC.service_id = IMC.service_id"},
		OrderByPrefix: "IMC",
	}

	return serviceObject.GetWithAggregations(ctx, s.db, aggDef, filter, order)
}

func (s *SqlDatabase) GetAllServiceCursors(ctx context.Context, filter *entity.ServiceFilter, order []entity.Order) ([]string, error) {
	return serviceObject.GetAllCursors(ctx, s.db, filter, order)
}

func (s *SqlDatabase) CreateService(service *entity.Service) (*entity.Service, error) {
	return serviceObject.Create(s.db, service)
}

func (s *SqlDatabase) UpdateService(service *entity.Service) error {
	return serviceObject.Update(s.db, service)
}

func (s *SqlDatabase) DeleteService(id int64, userId int64) error {
	return serviceObject.Delete(s.db, id, userId)
}

func (s *SqlDatabase) AddOwnerToService(serviceId int64, userId int64) error {
	return AssociateId(s.db, "Owner", "owner", "service", serviceId, "user", userId)
}

func (s *SqlDatabase) RemoveOwnerFromService(serviceId int64, userId int64) error {
	return DissociateId(s.db, "Owner", "owner", "service", serviceId, "user", userId)
}

func (s *SqlDatabase) AddIssueRepositoryToService(serviceId int64, issueRepositoryId int64, priority int64) error {
	return AssociateIdWithVal(s.db, "IssueRepositoryService", "issuerepositoryservice", "service", serviceId, "issue_repository", issueRepositoryId, "priority", priority)
}

func (s *SqlDatabase) RemoveIssueRepositoryFromService(serviceId int64, issueRepositoryId int64) error {
	return DissociateId(s.db, "IssueRepositoryService", "issuerepositoryservice", "service", serviceId, "issue_repository", issueRepositoryId)
}

func (s *SqlDatabase) GetServiceCcrns(ctx context.Context, filter *entity.ServiceFilter) ([]string, error) {
	return serviceObject.GetAttr(ctx, s.db, "ccrn", filter)
}

func (s *SqlDatabase) GetServiceDomains(ctx context.Context, filter *entity.ServiceFilter) ([]string, error) {
	return serviceObject.GetAttr(ctx, s.db, "domain", filter)
}

func (s *SqlDatabase) GetServiceRegions(ctx context.Context, filter *entity.ServiceFilter) ([]string, error) {
	return serviceObject.GetAttr(ctx, s.db, "region", filter)
}

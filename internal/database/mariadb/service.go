// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

var serviceObject = DbObject[*entity.Service, *entity.ServiceFilter, entity.ServiceResult]{
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
	GetItemAppender: func(l []entity.ServiceResult, e RowComposite, order []entity.Order) []entity.ServiceResult {
		s := entity.Service{
			BaseService: e.AsBaseService(),
		}

		var isc entity.IssueSeverityCounts
		if e.RatingCount != nil {
			isc = e.AsIssueSeverityCounts()
		}

		cursor, _ := EncodeCursor(WithService(order, s, isc))

		sr := entity.ServiceResult{
			WithCursor: entity.WithCursor{
				Value: cursor,
			},
			Service: &s,
		}

		return append(l, sr)
	},

	GetAllCursorItemAppender: func(l []string, e RowComposite, order []entity.Order) []string {
		s := entity.Service{
			BaseService: e.AsBaseService(),
		}

		var isc entity.IssueSeverityCounts
		if e.RatingCount != nil {
			isc = e.AsIssueSeverityCounts()
		}

		cursor, _ := EncodeCursor(WithService(order, s, isc))

		return append(l, cursor)
	},
}

func appendServiceColumns(s []string, filter *entity.ServiceFilter, order []entity.Order) []string {
	if filter != nil && len(filter.IssueRepositoryId) > 0 {
		s = append(s, "IRS.*")
	}

	for _, o := range order {
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
}

func (s *SqlDatabase) CountServices(ctx context.Context, filter *entity.ServiceFilter) (int64, error) {
	return serviceObject.Count(ctx, s.db, filter)
}

func (s *SqlDatabase) GetServices(ctx context.Context, filter *entity.ServiceFilter, order []entity.Order) ([]entity.ServiceResult, error) {
	return serviceObject.Get(ctx, s.db, filter, order)
}

// TODO use DbObject
func (s *SqlDatabase) GetServicesWithAggregations(
	ctx context.Context,
	filter *entity.ServiceFilter,
	order []entity.Order,
) ([]entity.ServiceResult, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetServicesWithAggregations",
	})

	baseImQuery := `
        SELECT %s, COUNT(IM.issuematch_id) AS service_agg_issue_matches FROM Service S
        %s
        LEFT JOIN IssueMatch IM on CI.componentinstance_id = IM.issuematch_component_instance_id
        %s
        GROUP BY S.service_id %s ORDER BY %s LIMIT ?
    `

	baseCiQuery := `
        SELECT %s, SUM(CI.componentinstance_count) AS service_agg_component_instances FROM Service S
        %s
        %s
        GROUP BY S.service_id %s ORDER BY %s LIMIT ?
    `

	baseQuery := `
        WITH IssueMatchCounts AS (
            %s
        ),
        ComponentInstanceCounts AS (
            %s
        )
        SELECT IMC.*, CIC.*
        FROM ComponentInstanceCounts CIC
        JOIN IssueMatchCounts IMC ON CIC.service_id = IMC.service_id
        ORDER BY %s;
    `
	filter = EnsureFilter(filter)
	ord := NewOrder(order, entity.Order{By: entity.ServiceId, Direction: entity.OrderDirectionAsc})
	joins := serviceObject.GetJoins_tmp(filter, ord)

	// Ensure ComponentInstance is joined for aggregations
	if !strings.Contains(joins, "ComponentInstance CI") {
		joins = fmt.Sprintf("%s LEFT JOIN ComponentInstance CI on S.service_id = CI.componentinstance_service_id", joins)
	}

	columns := strings.Join(appendServiceColumns([]string{"S.*"}, filter, ord.Sequence()), ",")

	cursorFields, err := DecodeCursor(filter.After)
	if err != nil {
		return nil, err
	}

	cursorQuery := CreateCursorQuery("", cursorFields)

	filterStr := serviceObject.GetFilterQuery_tmp(filter)

	whereClause := ""
	if filterStr != "" || cursorQuery != "" {
		whereClause = fmt.Sprintf("WHERE %s", filterStr)
	}

	if filterStr != "" && cursorQuery != "" {
		cursorQuery = fmt.Sprintf(" HAVING (%s)", cursorQuery)
	}

	imQuery := fmt.Sprintf(baseImQuery, columns, joins, whereClause, cursorQuery, ord.ToSql())
	ciQuery := fmt.Sprintf(baseCiQuery, columns, joins, whereClause, cursorQuery, ord.ToSql())
	query := fmt.Sprintf(baseQuery, imQuery, ciQuery, ord.ToSqlWithPrefixAll("IMC"))

	stmt, err := s.db.PreparexContext(ctx, query)
	if err != nil {
		msg := ERROR_MSG_PREPARED_STMT
		l.WithFields(
			logrus.Fields{
				"error": err,
				"query": query,
				"stmt":  stmt,
			},
		).Error(msg)

		return nil, fmt.Errorf("%s", msg)
	}

	// parameters for issue match query
	filterParameters := serviceObject.GetFilterParameters_tmp(filter, true, cursorFields)
	// parameters for component instance query
	filterParameters = append(
		filterParameters,
		serviceObject.GetFilterParameters_tmp(filter, true, cursorFields)...,
	)

	defer func() {
		if err := stmt.Close(); err != nil {
			logrus.Warnf("error during close stmt: %s", err)
		}
	}()

	return performListScan(
		ctx,
		stmt,
		filterParameters,
		l,
		func(l []entity.ServiceResult, e RowComposite) []entity.ServiceResult {
			service := entity.Service{
				BaseService: e.AsBaseService(),
			}
			aggregations := e.AsServiceAggregations()

			var isc entity.IssueSeverityCounts
			if e.RatingCount != nil {
				isc = e.AsIssueSeverityCounts()
			}

			cursor, _ := EncodeCursor(WithService(order, service, isc))

			sr := entity.ServiceResult{
				WithCursor: entity.WithCursor{
					Value: cursor,
				},
				Service:             &service,
				ServiceAggregations: &aggregations,
			}

			return append(l, sr)
		},
	)
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

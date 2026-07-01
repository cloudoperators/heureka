// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"

	"github.com/cloudoperators/heureka/internal/entity"
)

var supportGroupObject = DbObject[*entity.SupportGroup, *entity.SupportGroupFilter, entity.SupportGroupResult, *any]{
	Prefix:       "supportgroup",
	TableName:    "SupportGroup",
	TableKey:     "SG",
	DefaultOrder: entity.Order{By: entity.SupportGroupId, Direction: entity.OrderDirectionAsc},
	Properties: []*Property[*entity.SupportGroup]{
		NewProperty("supportgroup_ccrn", func(sg *entity.SupportGroup) (any, bool) { return sg.CCRN, sg.CCRN != "" }),
		NewProperty("supportgroup_created_by", func(sg *entity.SupportGroup) (any, bool) { return sg.CreatedBy, NoUpdate }),
		NewProperty("supportgroup_updated_by", func(sg *entity.SupportGroup) (any, bool) { return sg.UpdatedBy, sg.UpdatedBy != 0 }),
	},
	FilterProperties: []*FilterProperty[*entity.SupportGroupFilter]{
		NewFilterProperty("SG.supportgroup_id = ?", func(filter *entity.SupportGroupFilter) any { return filter.Id }),
		NewFilterProperty("SGS.supportgroupservice_service_id = ?", func(filter *entity.SupportGroupFilter) any { return filter.ServiceId }),
		NewFilterProperty("SG.supportgroup_ccrn = ?", func(filter *entity.SupportGroupFilter) any { return filter.CCRN }),
		NewFilterProperty("SGU.supportgroupuser_user_id = ?", func(filter *entity.SupportGroupFilter) any { return filter.UserId }),
		NewFilterProperty("IM.issuematch_issue_id = ?", func(filter *entity.SupportGroupFilter) any { return filter.IssueId }),
		NewStateFilterProperty("SG.supportgroup", func(filter *entity.SupportGroupFilter) any { return filter.State }),
	},
	JoinDefs: []*JoinDef[*entity.SupportGroupFilter]{
		{
			Name:      "SGS",
			Type:      InnerJoin,
			Table:     "SupportGroupService SGS",
			On:        "SG.supportgroup_id = SGS.supportgroupservice_support_group_id",
			Condition: func(f *entity.SupportGroupFilter, _ *Order) bool { return len(f.ServiceId) > 0 },
		},
		{
			Name:      "CI",
			Type:      InnerJoin,
			Table:     "ComponentInstance CI",
			On:        "SGS.supportgroupservice_service_id = CI.componentinstance_service_id",
			DependsOn: []string{"SGS"},
			Condition: DependentJoin[*entity.SupportGroupFilter],
		},
		{
			Name:      "IM",
			Type:      InnerJoin,
			Table:     "IssueMatch IM",
			On:        "CI.componentinstance_id = IM.issuematch_component_instance_id",
			DependsOn: []string{"CI"},
			Condition: func(f *entity.SupportGroupFilter, _ *Order) bool { return len(f.IssueId) > 0 },
		},
		{
			Name:      "SGU",
			Type:      InnerJoin,
			Table:     "SupportGroupUser SGU",
			On:        "SG.supportgroup_id = SGU.supportgroupuser_support_group_id",
			Condition: func(f *entity.SupportGroupFilter, _ *Order) bool { return len(f.UserId) > 0 },
		},
	},
	Attributes: []Attr{{Name: "ccrn", Order: entity.Order{By: entity.SupportGroupCcrn, Direction: entity.OrderDirectionAsc}}},
	RowToData: func(e RowComposite, order []entity.Order) (*entity.SupportGroup, string) {
		sg := e.AsSupportGroup()

		cursor, _ := EncodeCursor(WithSupportGroup(order, sg))

		return &sg, cursor
	},
	NewResult: func(sg *entity.SupportGroup, _ *any, cursor string) entity.SupportGroupResult {
		return entity.SupportGroupResult{
			WithCursor:   entity.WithCursor{Value: cursor},
			SupportGroup: sg,
		}
	},
}

func (s *SqlDatabase) GetAllSupportGroupCursors(ctx context.Context, filter *entity.SupportGroupFilter, order []entity.Order) ([]string, error) {
	return supportGroupObject.GetAllCursors(ctx, s.db, filter, order)
}

func (s *SqlDatabase) GetSupportGroups(ctx context.Context, filter *entity.SupportGroupFilter, order []entity.Order) ([]entity.SupportGroupResult, error) {
	return supportGroupObject.Get(ctx, s.db, filter, order)
}

func (s *SqlDatabase) CountSupportGroups(ctx context.Context, filter *entity.SupportGroupFilter) (int64, error) {
	return supportGroupObject.Count(ctx, s.db, filter)
}

func (s *SqlDatabase) CreateSupportGroup(supportGroup *entity.SupportGroup) (*entity.SupportGroup, error) {
	return supportGroupObject.Create(s.db, supportGroup)
}

func (s *SqlDatabase) UpdateSupportGroup(supportGroup *entity.SupportGroup) error {
	return supportGroupObject.Update(s.db, supportGroup)
}

func (s *SqlDatabase) DeleteSupportGroup(id int64, userId int64) error {
	return supportGroupObject.Delete(s.db, id, userId)
}

func (s *SqlDatabase) AddServiceToSupportGroup(supportGroupId int64, serviceId int64) error {
	return AssociateId(s.db, "SupportGroupService", "supportgroupservice", "service", serviceId, "support_group", supportGroupId)
}

func (s *SqlDatabase) RemoveServiceFromSupportGroup(supportGroupId int64, serviceId int64) error {
	return DissociateId(s.db, "SupportGroupService", "supportgroupservice", "service", serviceId, "support_group", supportGroupId)
}

func (s *SqlDatabase) AddUserToSupportGroup(supportGroupId int64, userId int64) error {
	return AssociateId(s.db, "SupportGroupUser", "supportgroupuser", "user", userId, "support_group", supportGroupId)
}

func (s *SqlDatabase) RemoveUserFromSupportGroup(supportGroupId int64, userId int64) error {
	return DissociateId(s.db, "SupportGroupUser", "supportgroupuser", "user", userId, "support_group", supportGroupId)
}

func (s *SqlDatabase) GetSupportGroupCcrns(ctx context.Context, filter *entity.SupportGroupFilter) ([]string, error) {
	return supportGroupObject.GetAttr(ctx, s.db, "ccrn", filter)
}

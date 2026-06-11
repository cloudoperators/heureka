// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

var supportGroupObject = DbObject[*entity.SupportGroup, *entity.SupportGroupFilter, entity.SupportGroupResult]{
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
	GetItemAppender: func(l []entity.SupportGroupResult, e RowComposite, order []entity.Order) []entity.SupportGroupResult {
		sg := e.AsSupportGroup()
		cursor, _ := EncodeCursor(WithSupportGroup(order, sg))

		sgr := entity.SupportGroupResult{
			WithCursor: entity.WithCursor{
				Value: cursor,
			},
			SupportGroup: &sg,
		}

		return append(l, sgr)
	},
	GetAllCursorItemAppender: func(l []string, e RowComposite, order []entity.Order) []string {
		sg := e.AsSupportGroup()

		cursor, _ := EncodeCursor(WithSupportGroup(order, sg))

		return append(l, cursor)
	},
}

func (s *SqlDatabase) buildSupportGroupStatement(
	ctx context.Context,
	baseQuery sq.SelectBuilder,
	filter *entity.SupportGroupFilter,
	withCursor bool,
	order []entity.Order,
	l *logrus.Entry,
) (Stmt, []any, error) {
	statement := Statement[*entity.SupportGroupFilter]{
		Db:         s.db,
		L:          l,
		Obj:        &supportGroupObject,
		BaseQuery:  baseQuery,
		Order:      NewOrder(order, supportGroupObject.DefaultOrder),
		WithCursor: withCursor,
	}

	return BuildStatement(ctx, statement, filter)
}

func (s *SqlDatabase) GetAllSupportGroupCursors(
	ctx context.Context,
	filter *entity.SupportGroupFilter,
	order []entity.Order,
) ([]string, error) {
	return supportGroupObject.GetAllCursors(ctx, s.db, filter, order)
}

func (s *SqlDatabase) GetSupportGroups(
	ctx context.Context,
	filter *entity.SupportGroupFilter,
	order []entity.Order,
) ([]entity.SupportGroupResult, error) {
	return supportGroupObject.Get(ctx, s.db, filter, order)
}

func (s *SqlDatabase) CountSupportGroups(ctx context.Context, filter *entity.SupportGroupFilter) (int64, error) {
	return supportGroupObject.Count(ctx, s.db, filter)
}

func (s *SqlDatabase) CreateSupportGroup(
	supportGroup *entity.SupportGroup,
) (*entity.SupportGroup, error) {
	return supportGroupObject.Create(s.db, supportGroup)
}

func (s *SqlDatabase) UpdateSupportGroup(supportGroup *entity.SupportGroup) error {
	return supportGroupObject.Update(s.db, supportGroup)
}

func (s *SqlDatabase) DeleteSupportGroup(id int64, userId int64) error {
	return supportGroupObject.Delete(s.db, id, userId)
}

func (s *SqlDatabase) AddServiceToSupportGroup(supportGroupId int64, serviceId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"serviceId":      serviceId,
		"supportGroupId": supportGroupId,
		"event":          "database.AddServiceToSupportGroup",
	})

	query := `
		INSERT INTO SupportGroupService (
			supportgroupservice_service_id,
			supportgroupservice_support_group_id
		) VALUES (
			:service_id,
			:support_group_id
		)
	`

	args := map[string]any{
		"service_id":       serviceId,
		"support_group_id": supportGroupId,
	}

	var mysqlErr *mysql.MySQLError

	_, err := performExec(s, query, args, l)
	if err != nil {
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == database.ErrCodeDuplicateEntry {
				return nil
			}
		}

		return err
	}

	return nil
}

func (s *SqlDatabase) RemoveServiceFromSupportGroup(supportGroupId int64, serviceId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"serviceId":      serviceId,
		"supportGroupId": supportGroupId,
		"event":          "database.RemoveServiceFromSupportGroup",
	})

	query := `
		DELETE FROM SupportGroupService
		WHERE supportgroupservice_service_id = :service_id
		AND supportgroupservice_support_group_id = :support_group_id
	`

	args := map[string]any{
		"service_id":       serviceId,
		"support_group_id": supportGroupId,
	}

	_, err := performExec(s, query, args, l)

	return err
}

func (s *SqlDatabase) AddUserToSupportGroup(supportGroupId int64, userId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"userId":         userId,
		"supportGroupId": supportGroupId,
		"event":          "database.AddUserToSupportGroup",
	})

	query := `
		INSERT INTO SupportGroupUser (
			supportgroupuser_user_id,
			supportgroupuser_support_group_id
		) VALUES (
			:user_id,
			:support_group_id
		)
	`

	args := map[string]any{
		"user_id":          userId,
		"support_group_id": supportGroupId,
	}

	var mysqlErr *mysql.MySQLError

	_, err := performExec(s, query, args, l)
	if err != nil {
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == database.ErrCodeDuplicateEntry {
				return nil
			}
		}

		return err
	}

	return nil
}

func (s *SqlDatabase) RemoveUserFromSupportGroup(supportGroupId int64, userId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"userId":         userId,
		"supportGroupId": supportGroupId,
		"event":          "database.RemoveUserFromSupportGroup",
	})

	query := `
		DELETE FROM SupportGroupUser
		WHERE supportgroupuser_user_id = :user_id
		AND supportgroupuser_support_group_id = :support_group_id
	`

	args := map[string]any{
		"user_id":          userId,
		"support_group_id": supportGroupId,
	}

	_, err := performExec(s, query, args, l)

	return err
}

func (s *SqlDatabase) GetSupportGroupCcrns(ctx context.Context, filter *entity.SupportGroupFilter) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetSupportGroupCcrns",
	})

	baseQuery := sq.Select("SG.supportgroup_ccrn").From("SupportGroup SG")

	order := []entity.Order{
		{
			By:        entity.SupportGroupCcrn,
			Direction: entity.OrderDirectionAsc,
		},
	}

	// Builds full statement with possible joins and filters
	stmt, filterParameters, err := s.buildSupportGroupStatement(ctx, baseQuery, filter, false, order, l)
	if err != nil {
		l.Error("Error preparing statement: ", err)
		return nil, err
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			logrus.Warnf("error during close stmt: %s", err)
		}
	}()

	// Execute the query
	rows, err := stmt.QueryxContext(ctx, filterParameters...)
	if err != nil {
		l.Error("Error executing query: ", err)
		return nil, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			logrus.Warnf("error during close rows: %s", err)
		}
	}()

	// Collect the results
	supportGroupCcrns := []string{}

	var ccrn string
	for rows.Next() {
		if err := rows.Scan(&ccrn); err != nil {
			l.Error("Error scanning row: ", err)
			continue
		}

		supportGroupCcrns = append(supportGroupCcrns, ccrn)
	}

	if err = rows.Err(); err != nil {
		l.Error("Row iteration error: ", err)
		return nil, err
	}

	return supportGroupCcrns, nil
}

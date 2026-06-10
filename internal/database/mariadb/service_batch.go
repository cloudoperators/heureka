// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/cloudoperators/heureka/internal/database/querycounter"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

// ownerWithServiceRow is a scan target for the batch owners query.
type ownerWithServiceRow struct {
	UserID       sql.NullInt64  `db:"user_id"`
	UserName     sql.NullString `db:"user_name"`
	UniqueUserID sql.NullString `db:"user_unique_user_id"`
	UserType     sql.NullInt64  `db:"user_type"`
	UserEmail    sql.NullString `db:"user_email"`
	CreatedAt    sql.NullTime   `db:"user_created_at"`
	CreatedBy    sql.NullInt64  `db:"user_created_by"`
	DeletedAt    sql.NullTime   `db:"user_deleted_at"`
	UpdatedAt    sql.NullTime   `db:"user_updated_at"`
	UpdatedBy    sql.NullInt64  `db:"user_updated_by"`
	ServiceID    sql.NullInt64  `db:"owner_service_id"`
}

func (r *ownerWithServiceRow) asUser() entity.User {
	return entity.User{
		Id:           GetInt64Value(r.UserID),
		Name:         GetStringValue(r.UserName),
		UniqueUserID: GetStringValue(r.UniqueUserID),
		Type:         GetUserTypeValue(r.UserType),
		Email:        GetStringValue(r.UserEmail),
		Metadata: entity.Metadata{
			CreatedAt: GetTimeValue(r.CreatedAt),
			CreatedBy: GetInt64Value(r.CreatedBy),
			DeletedAt: GetTimeValue(r.DeletedAt),
			UpdatedAt: GetTimeValue(r.UpdatedAt),
			UpdatedBy: GetInt64Value(r.UpdatedBy),
		},
	}
}

// supportGroupWithServiceRow is a scan target for the batch support groups query.
type supportGroupWithServiceRow struct {
	SGID      sql.NullInt64  `db:"supportgroup_id"`
	SGCCRN    sql.NullString `db:"supportgroup_ccrn"`
	CreatedAt sql.NullTime   `db:"supportgroup_created_at"`
	CreatedBy sql.NullInt64  `db:"supportgroup_created_by"`
	DeletedAt sql.NullTime   `db:"supportgroup_deleted_at"`
	UpdatedAt sql.NullTime   `db:"supportgroup_updated_at"`
	UpdatedBy sql.NullInt64  `db:"supportgroup_updated_by"`
	ServiceID sql.NullInt64  `db:"supportgroupservice_service_id"`
}

func (r *supportGroupWithServiceRow) asSupportGroup() entity.SupportGroup {
	return entity.SupportGroup{
		Id:   GetInt64Value(r.SGID),
		CCRN: GetStringValue(r.SGCCRN),
		Metadata: entity.Metadata{
			CreatedAt: GetTimeValue(r.CreatedAt),
			CreatedBy: GetInt64Value(r.CreatedBy),
			DeletedAt: GetTimeValue(r.DeletedAt),
			UpdatedAt: GetTimeValue(r.UpdatedAt),
			UpdatedBy: GetInt64Value(r.UpdatedBy),
		},
	}
}

// issueCountWithServiceRow is a scan target for the batch issue counts query.
type issueCountWithServiceRow struct {
	ServiceID int64         `db:"service_id"`
	Critical  sql.NullInt64 `db:"critical_count"`
	High      sql.NullInt64 `db:"high_count"`
	Medium    sql.NullInt64 `db:"medium_count"`
	Low       sql.NullInt64 `db:"low_count"`
	None      sql.NullInt64 `db:"none_count"`
}

func (r *issueCountWithServiceRow) asIssueSeverityCounts() entity.IssueSeverityCounts {
	isc := entity.IssueSeverityCounts{
		Critical: GetInt64Value(r.Critical),
		High:     GetInt64Value(r.High),
		Medium:   GetInt64Value(r.Medium),
		Low:      GetInt64Value(r.Low),
		None:     GetInt64Value(r.None),
	}
	isc.Total = isc.Critical + isc.High + isc.Medium + isc.Low + isc.None

	return isc
}

// GetOwnersByServiceIDs returns owners (users) grouped by service ID in a single query.
// This eliminates N+1 queries when loading owners for multiple services.
func (s *SqlDatabase) GetOwnersByServiceIDs(ctx context.Context, serviceIDs []int64) (map[int64][]entity.User, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":      "database.GetOwnersByServiceIDs",
		"serviceIDs": serviceIDs,
	})

	if len(serviceIDs) == 0 {
		return map[int64][]entity.User{}, nil
	}

	placeholders := make([]string, len(serviceIDs))
	args := make([]any, len(serviceIDs))

	for i, id := range serviceIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT U.user_id, U.user_name, U.user_unique_user_id, U.user_type, U.user_email,
		       U.user_created_at, U.user_created_by, U.user_deleted_at, U.user_updated_at, U.user_updated_by,
		       O.owner_service_id
		FROM Owner O
		JOIN User U ON O.owner_user_id = U.user_id
		WHERE O.owner_service_id IN (%s)
		  AND O.owner_deleted_at IS NULL
		  AND U.user_deleted_at IS NULL
		ORDER BY O.owner_service_id, U.user_id
	`, strings.Join(placeholders, ","))

	stmt, err := s.db.PreparexContext(ctx, query)
	if err != nil {
		l.WithField("error", err).Error("Error preparing statement")
		return nil, fmt.Errorf("GetOwnersByServiceIDs: error preparing statement: %w", err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			l.Warnf("error closing statement: %s", err)
		}
	}()

	querycounter.Increment(ctx)

	rows, err := stmt.QueryxContext(ctx, args...)
	if err != nil {
		l.WithField("error", err).Error("Error executing query")
		return nil, fmt.Errorf("GetOwnersByServiceIDs: error executing query: %w", err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			l.Warnf("error closing rows: %s", err)
		}
	}()

	result := make(map[int64][]entity.User)

	for rows.Next() {
		var row ownerWithServiceRow
		if err := rows.StructScan(&row); err != nil {
			l.WithField("error", err).Error("Error scanning row")
			return nil, fmt.Errorf("GetOwnersByServiceIDs: error scanning row: %w", err)
		}

		serviceID := GetInt64Value(row.ServiceID)
		user := row.asUser()
		result[serviceID] = append(result[serviceID], user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetOwnersByServiceIDs: row iteration error: %w", err)
	}

	return result, nil
}

// GetSupportGroupsByServiceIDs returns support groups grouped by service ID in a single query.
// This eliminates N+1 queries when loading support groups for multiple services.
func (s *SqlDatabase) GetSupportGroupsByServiceIDs(ctx context.Context, serviceIDs []int64) (map[int64][]entity.SupportGroup, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":      "database.GetSupportGroupsByServiceIDs",
		"serviceIDs": serviceIDs,
	})

	if len(serviceIDs) == 0 {
		return map[int64][]entity.SupportGroup{}, nil
	}

	placeholders := make([]string, len(serviceIDs))
	args := make([]any, len(serviceIDs))

	for i, id := range serviceIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT SG.supportgroup_id, SG.supportgroup_ccrn,
		       SG.supportgroup_created_at, SG.supportgroup_created_by,
		       SG.supportgroup_deleted_at, SG.supportgroup_updated_at, SG.supportgroup_updated_by,
		       SGS.supportgroupservice_service_id
		FROM SupportGroupService SGS
		JOIN SupportGroup SG ON SGS.supportgroupservice_support_group_id = SG.supportgroup_id
		WHERE SGS.supportgroupservice_service_id IN (%s)
		  AND SGS.supportgroupservice_deleted_at IS NULL
		  AND SG.supportgroup_deleted_at IS NULL
		ORDER BY SGS.supportgroupservice_service_id, SG.supportgroup_id
	`, strings.Join(placeholders, ","))

	stmt, err := s.db.PreparexContext(ctx, query)
	if err != nil {
		l.WithField("error", err).Error("Error preparing statement")
		return nil, fmt.Errorf("GetSupportGroupsByServiceIDs: error preparing statement: %w", err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			l.Warnf("error closing statement: %s", err)
		}
	}()

	querycounter.Increment(ctx)

	rows, err := stmt.QueryxContext(ctx, args...)
	if err != nil {
		l.WithField("error", err).Error("Error executing query")
		return nil, fmt.Errorf("GetSupportGroupsByServiceIDs: error executing query: %w", err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			l.Warnf("error closing rows: %s", err)
		}
	}()

	result := make(map[int64][]entity.SupportGroup)

	for rows.Next() {
		var row supportGroupWithServiceRow
		if err := rows.StructScan(&row); err != nil {
			l.WithField("error", err).Error("Error scanning row")
			return nil, fmt.Errorf("GetSupportGroupsByServiceIDs: error scanning row: %w", err)
		}

		serviceID := GetInt64Value(row.ServiceID)
		sg := row.asSupportGroup()
		result[serviceID] = append(result[serviceID], sg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetSupportGroupsByServiceIDs: row iteration error: %w", err)
	}

	return result, nil
}

// GetIssueCountsByServiceIDs returns issue severity counts grouped by service ID in a single query.
// This eliminates N+1 queries when loading issue counts for multiple services.
func (s *SqlDatabase) GetIssueCountsByServiceIDs(ctx context.Context, serviceIDs []int64) (map[int64]entity.IssueSeverityCounts, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":      "database.GetIssueCountsByServiceIDs",
		"serviceIDs": serviceIDs,
	})

	if len(serviceIDs) == 0 {
		return map[int64]entity.IssueSeverityCounts{}, nil
	}

	placeholders := make([]string, len(serviceIDs))
	args := make([]any, len(serviceIDs))

	for i, id := range serviceIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT service_id, critical_count, high_count, medium_count, low_count, none_count
		FROM mvServiceIssueCounts
		WHERE service_id IN (%s)
	`, strings.Join(placeholders, ","))

	stmt, err := s.db.PreparexContext(ctx, query)
	if err != nil {
		l.WithField("error", err).Error("Error preparing statement")
		return nil, fmt.Errorf("GetIssueCountsByServiceIDs: error preparing statement: %w", err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			l.Warnf("error closing statement: %s", err)
		}
	}()

	querycounter.Increment(ctx)

	rows, err := stmt.QueryxContext(ctx, args...)
	if err != nil {
		l.WithField("error", err).Error("Error executing query")
		return nil, fmt.Errorf("GetIssueCountsByServiceIDs: error executing query: %w", err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			l.Warnf("error closing rows: %s", err)
		}
	}()

	result := make(map[int64]entity.IssueSeverityCounts)

	for rows.Next() {
		var row issueCountWithServiceRow
		if err := rows.StructScan(&row); err != nil {
			l.WithField("error", err).Error("Error scanning row")
			return nil, fmt.Errorf("GetIssueCountsByServiceIDs: error scanning row: %w", err)
		}

		isc := row.asIssueSeverityCounts()
		result[row.ServiceID] = isc
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetIssueCountsByServiceIDs: row iteration error: %w", err)
	}

	return result, nil
}

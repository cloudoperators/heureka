// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/cloudoperators/heureka/internal/database/querycounter"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

// GetVersionsByComponentIDs returns component versions grouped by component ID in a single query.
// This eliminates N+1 queries when loading versions for multiple images.
// The optional serviceCCRN filter restricts results to versions that have component instances
// associated with the specified services.
func (s *SqlDatabase) GetVersionsByComponentIDs(ctx context.Context, componentIDs []int64, serviceCCRN []*string) (map[int64][]entity.ComponentVersionResult, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":        "database.GetVersionsByComponentIDs",
		"componentIDs": componentIDs,
	})

	if len(componentIDs) == 0 {
		return map[int64][]entity.ComponentVersionResult{}, nil
	}

	query := sq.Select(
		"CV.componentversion_id",
		"CV.componentversion_component_id",
		"CV.componentversion_version",
		"CV.componentversion_tag",
		"CV.componentversion_repository",
		"CV.componentversion_organization",
		"CV.componentversion_created_at",
		"CV.componentversion_created_by",
		"CV.componentversion_deleted_at",
		"CV.componentversion_updated_at",
		"CV.componentversion_updated_by",
		"CV.componentversion_end_of_life",
	).
		From("ComponentVersion CV").
		Where(sq.Eq{"CV.componentversion_component_id": componentIDs}).
		Where("CV.componentversion_deleted_at IS NULL")

	if len(serviceCCRN) > 0 {
		nonNilCCRNs := make([]string, 0, len(serviceCCRN))
		for _, c := range serviceCCRN {
			if c != nil {
				nonNilCCRNs = append(nonNilCCRNs, *c)
			}
		}

		if len(nonNilCCRNs) > 0 {
			query = query.
				Join("ComponentInstance CI ON CV.componentversion_id = CI.componentinstance_component_version_id").
				Join("Service S ON CI.componentinstance_service_id = S.service_id").
				Where(sq.Eq{"S.service_ccrn": nonNilCCRNs}).
				Where("CI.componentinstance_deleted_at IS NULL").
				Where("S.service_deleted_at IS NULL")
		}
	}

	query = query.
		GroupBy("CV.componentversion_id").
		OrderBy("CV.componentversion_component_id", "CV.componentversion_version DESC")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		l.WithError(err).Error("Error building GetVersionsByComponentIDs query")
		return nil, fmt.Errorf("GetVersionsByComponentIDs: failed to build query: %w", err)
	}

	l.WithField("query", sqlStr).Debug("Executing GetVersionsByComponentIDs")

	querycounter.Increment(ctx)

	rows, err := s.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		l.WithError(err).Error("Error executing GetVersionsByComponentIDs query")
		return nil, fmt.Errorf("GetVersionsByComponentIDs: error executing query: %w", err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			l.Warnf("error closing rows: %s", err)
		}
	}()

	result := make(map[int64][]entity.ComponentVersionResult)

	for rows.Next() {
		var row componentVersionWithComponentRow
		if err := rows.Scan(
			&row.ID,
			&row.ComponentID,
			&row.Version,
			&row.Tag,
			&row.Repository,
			&row.Organization,
			&row.CreatedAt,
			&row.CreatedBy,
			&row.DeletedAt,
			&row.UpdatedAt,
			&row.UpdatedBy,
			&row.EndOfLife,
		); err != nil {
			l.WithError(err).Error("Error scanning GetVersionsByComponentIDs row")
			return nil, fmt.Errorf("GetVersionsByComponentIDs: error scanning row: %w", err)
		}

		componentID := GetInt64Value(row.ComponentID)
		cv := row.asComponentVersion()
		cvr := entity.ComponentVersionResult{
			ComponentVersion: &cv,
		}
		result[componentID] = append(result[componentID], cvr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetVersionsByComponentIDs: row iteration error: %w", err)
	}

	return result, nil
}

// GetIssueCountsByComponentIDs returns vulnerability severity counts grouped by component ID.
// Counts are derived from the same mvVulnerabilityList source used by GetVulnerabilitiesByComponentIDs,
// so the badge counts always match the number of items shown in the vulnerability list.
// The optional serviceCCRN filter restricts counts to vulnerabilities active in the specified services
// (via mvVulnerabilityService).
func (s *SqlDatabase) GetIssueCountsByComponentIDs(ctx context.Context, componentIDs []int64, serviceCCRN []*string) (map[int64]entity.IssueSeverityCounts, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":        "database.GetIssueCountsByComponentIDs",
		"componentIDs": componentIDs,
	})

	if len(componentIDs) == 0 {
		return map[int64]entity.IssueSeverityCounts{}, nil
	}

	query := sq.Select(
		"CV.componentversion_component_id",
		"COUNT(DISTINCT CASE WHEN MVL.max_severity = 'Critical' THEN I.issue_id END) as critical_count",
		"COUNT(DISTINCT CASE WHEN MVL.max_severity = 'High' THEN I.issue_id END) as high_count",
		"COUNT(DISTINCT CASE WHEN MVL.max_severity = 'Medium' THEN I.issue_id END) as medium_count",
		"COUNT(DISTINCT CASE WHEN MVL.max_severity = 'Low' THEN I.issue_id END) as low_count",
		"COUNT(DISTINCT CASE WHEN MVL.max_severity = 'None' THEN I.issue_id END) as none_count",
	).
		From("ComponentVersionIssue CVI").
		Join("ComponentVersion CV ON CVI.componentversionissue_component_version_id = CV.componentversion_id").
		Join("Issue I ON CVI.componentversionissue_issue_id = I.issue_id").
		Join("mvVulnerabilityList MVL ON I.issue_id = MVL.issue_id").
		Where(sq.Eq{"CV.componentversion_component_id": componentIDs}).
		Where("CVI.componentversionissue_deleted_at IS NULL").
		Where("CV.componentversion_deleted_at IS NULL").
		Where("I.issue_deleted_at IS NULL")

	if len(serviceCCRN) > 0 {
		nonNilCCRNs := make([]string, 0, len(serviceCCRN))
		for _, c := range serviceCCRN {
			if c != nil {
				nonNilCCRNs = append(nonNilCCRNs, *c)
			}
		}

		if len(nonNilCCRNs) > 0 {
			query = query.
				Join("mvVulnerabilityService MVS ON I.issue_id = MVS.issue_id").
				Join("Service S ON MVS.service_id = S.service_id").
				Where(sq.Eq{"S.service_ccrn": nonNilCCRNs})
		}
	}

	query = query.GroupBy("CV.componentversion_component_id")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		l.WithError(err).Error("Error building GetIssueCountsByComponentIDs query")
		return nil, fmt.Errorf("GetIssueCountsByComponentIDs: failed to build query: %w", err)
	}

	l.WithField("query", sqlStr).Debug("Executing GetIssueCountsByComponentIDs")

	querycounter.Increment(ctx)

	rows, err := s.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		l.WithError(err).Error("Error executing GetIssueCountsByComponentIDs query")
		return nil, fmt.Errorf("GetIssueCountsByComponentIDs: error executing query: %w", err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			l.Warnf("error closing rows: %s", err)
		}
	}()

	result := make(map[int64]entity.IssueSeverityCounts)

	for rows.Next() {
		var row issueCountWithComponentRow
		if err := rows.Scan(
			&row.ComponentID,
			&row.Critical,
			&row.High,
			&row.Medium,
			&row.Low,
			&row.None,
		); err != nil {
			l.WithError(err).Error("Error scanning GetIssueCountsByComponentIDs row")
			return nil, fmt.Errorf("GetIssueCountsByComponentIDs: error scanning row: %w", err)
		}

		isc := row.asIssueSeverityCounts()
		result[row.ComponentID] = isc
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetIssueCountsByComponentIDs: row iteration error: %w", err)
	}

	return result, nil
}

// componentVersionWithComponentRow is a scan target for the batch versions query.
type componentVersionWithComponentRow struct {
	ID           sql.NullInt64  `db:"componentversion_id"`
	ComponentID  sql.NullInt64  `db:"componentversion_component_id"`
	Version      sql.NullString `db:"componentversion_version"`
	Tag          sql.NullString `db:"componentversion_tag"`
	Repository   sql.NullString `db:"componentversion_repository"`
	Organization sql.NullString `db:"componentversion_organization"`
	CreatedAt    sql.NullTime   `db:"componentversion_created_at"`
	CreatedBy    sql.NullInt64  `db:"componentversion_created_by"`
	DeletedAt    sql.NullTime   `db:"componentversion_deleted_at"`
	UpdatedAt    sql.NullTime   `db:"componentversion_updated_at"`
	UpdatedBy    sql.NullInt64  `db:"componentversion_updated_by"`
	EndOfLife    sql.NullBool   `db:"componentversion_end_of_life"`
}

func (r *componentVersionWithComponentRow) asComponentVersion() entity.ComponentVersion {
	var eol *bool
	if r.EndOfLife.Valid {
		eol = &r.EndOfLife.Bool
	}

	return entity.ComponentVersion{
		Metadata: entity.Metadata{
			CreatedAt: GetTimeValue(r.CreatedAt),
			CreatedBy: GetInt64Value(r.CreatedBy),
			DeletedAt: GetTimeValue(r.DeletedAt),
			UpdatedAt: GetTimeValue(r.UpdatedAt),
			UpdatedBy: GetInt64Value(r.UpdatedBy),
		},
		Id:           GetInt64Value(r.ID),
		ComponentId:  GetInt64Value(r.ComponentID),
		Version:      GetStringValue(r.Version),
		Tag:          GetStringValue(r.Tag),
		Repository:   GetStringValue(r.Repository),
		Organization: GetStringValue(r.Organization),
		EndOfLife:    eol,
	}
}

// issueCountWithComponentRow is a scan target for the batch issue counts query.
type issueCountWithComponentRow struct {
	ComponentID int64         `db:"component_id"`
	Critical    sql.NullInt64 `db:"critical_count"`
	High        sql.NullInt64 `db:"high_count"`
	Medium      sql.NullInt64 `db:"medium_count"`
	Low         sql.NullInt64 `db:"low_count"`
	None        sql.NullInt64 `db:"none_count"`
}

func (r *issueCountWithComponentRow) asIssueSeverityCounts() entity.IssueSeverityCounts {
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

// GetVulnerabilitiesByComponentIDs returns active vulnerabilities grouped by component ID
// using the mvVulnerabilityList materialized view. This eliminates N+1 queries when loading
// nested vulnerabilities for multiple images.
// Join path: ComponentVersionIssue → ComponentVersion → Issue → mvVulnerabilityList
func (s *SqlDatabase) GetVulnerabilitiesByComponentIDs(ctx context.Context, componentIDs []int64) (map[int64][]entity.VulnerabilityResult, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":        "database.GetVulnerabilitiesByComponentIDs",
		"componentIDs": componentIDs,
	})

	if len(componentIDs) == 0 {
		return map[int64][]entity.VulnerabilityResult{}, nil
	}

	query := sq.Select(
		"CV.componentversion_component_id",
		"I.issue_id",
		"I.issue_primary_name",
		"I.issue_description",
		"MVL.max_severity",
		"MVL.earliest_remediation_date",
		"MVL.source_url",
	).
		From("ComponentVersionIssue CVI").
		Join("ComponentVersion CV ON CVI.componentversionissue_component_version_id = CV.componentversion_id").
		Join("Issue I ON CVI.componentversionissue_issue_id = I.issue_id").
		Join("mvVulnerabilityList MVL ON I.issue_id = MVL.issue_id").
		Where(sq.Eq{"CV.componentversion_component_id": componentIDs}).
		Where("CVI.componentversionissue_deleted_at IS NULL").
		Where("CV.componentversion_deleted_at IS NULL").
		Where("I.issue_deleted_at IS NULL").
		GroupBy("CV.componentversion_component_id", "I.issue_id").
		OrderBy(
			"CV.componentversion_component_id",
			"FIELD(MVL.max_severity, 'Critical','High','Medium','Low','None') ASC",
			"I.issue_primary_name ASC",
		)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		l.WithError(err).Error("Error building GetVulnerabilitiesByComponentIDs query")
		return nil, fmt.Errorf("GetVulnerabilitiesByComponentIDs: failed to build query: %w", err)
	}

	l.WithField("query", sqlStr).Debug("Executing GetVulnerabilitiesByComponentIDs")

	querycounter.Increment(ctx)

	rows, err := s.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		l.WithError(err).Error("Error executing GetVulnerabilitiesByComponentIDs query")
		return nil, fmt.Errorf("GetVulnerabilitiesByComponentIDs: error executing query: %w", err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			l.Warnf("error closing rows: %s", err)
		}
	}()

	result := make(map[int64][]entity.VulnerabilityResult)

	for rows.Next() {
		var componentID sql.NullInt64

		var issueID sql.NullInt64

		var primaryName sql.NullString

		var description sql.NullString

		var severity sql.NullString

		var remediationDate sql.NullTime

		var sourceURL sql.NullString

		if err := rows.Scan(
			&componentID,
			&issueID,
			&primaryName,
			&description,
			&severity,
			&remediationDate,
			&sourceURL,
		); err != nil {
			l.WithError(err).Error("Error scanning GetVulnerabilitiesByComponentIDs row")
			return nil, fmt.Errorf("GetVulnerabilitiesByComponentIDs: error scanning row: %w", err)
		}

		if !componentID.Valid || !issueID.Valid {
			continue
		}

		vr := entity.VulnerabilityResult{
			IssueID:     issueID.Int64,
			PrimaryName: GetStringValue(primaryName),
			Description: GetStringValue(description),
			MaxSeverity: GetStringValue(severity),
		}
		if remediationDate.Valid {
			vr.EarliestRemediationDate = &remediationDate.Time
		}

		if sourceURL.Valid {
			vr.SourceURL = sourceURL.String
		}

		result[componentID.Int64] = append(result[componentID.Int64], vr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetVulnerabilitiesByComponentIDs: row iteration error: %w", err)
	}

	return result, nil
}

// GetVulnerabilityCountsByComponentIDs returns the total count of active vulnerabilities
// per component ID using the mvVulnerabilityList materialized view.
func (s *SqlDatabase) GetVulnerabilityCountsByComponentIDs(ctx context.Context, componentIDs []int64) (map[int64]int, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":        "database.GetVulnerabilityCountsByComponentIDs",
		"componentIDs": componentIDs,
	})

	if len(componentIDs) == 0 {
		return map[int64]int{}, nil
	}

	query := sq.Select(
		"CV.componentversion_component_id",
		"COUNT(DISTINCT CVI.componentversionissue_issue_id) as vuln_count",
	).
		From("ComponentVersionIssue CVI").
		Join("ComponentVersion CV ON CVI.componentversionissue_component_version_id = CV.componentversion_id").
		Join("mvVulnerabilityList MVL ON CVI.componentversionissue_issue_id = MVL.issue_id").
		Where(sq.Eq{"CV.componentversion_component_id": componentIDs}).
		Where("CVI.componentversionissue_deleted_at IS NULL").
		Where("CV.componentversion_deleted_at IS NULL").
		GroupBy("CV.componentversion_component_id")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		l.WithError(err).Error("Error building GetVulnerabilityCountsByComponentIDs query")
		return nil, fmt.Errorf("GetVulnerabilityCountsByComponentIDs: failed to build query: %w", err)
	}

	l.WithField("query", sqlStr).Debug("Executing GetVulnerabilityCountsByComponentIDs")

	querycounter.Increment(ctx)

	rows, err := s.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		l.WithError(err).Error("Error executing GetVulnerabilityCountsByComponentIDs query")
		return nil, fmt.Errorf("GetVulnerabilityCountsByComponentIDs: error executing query: %w", err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			l.Warnf("error closing rows: %s", err)
		}
	}()

	result := make(map[int64]int)

	for rows.Next() {
		var componentID sql.NullInt64

		var count sql.NullInt64

		if err := rows.Scan(&componentID, &count); err != nil {
			l.WithError(err).Error("Error scanning GetVulnerabilityCountsByComponentIDs row")
			return nil, fmt.Errorf("GetVulnerabilityCountsByComponentIDs: error scanning row: %w", err)
		}

		if componentID.Valid {
			result[componentID.Int64] = int(GetInt64Value(count))
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetVulnerabilityCountsByComponentIDs: row iteration error: %w", err)
	}

	return result, nil
}

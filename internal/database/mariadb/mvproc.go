// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

type DBTX interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

type MVProcedure func(ctx context.Context, db DBTX) error

var MVProcedures [][]MVProcedure = [][]MVProcedure{
	// Keep this list in sync with the Refresh* functions in this file.
	//   vim helper:
	//      r! grep -E '^func\s' internal/database/mariadb/mvproc.go | sed -e 's@func\s\(.*\)[(].*@\1@' | sed -e 's/$/,/'
	{RefreshMVServiceIssueCounts},
	{RefreshMVCountIssueRatingsServiceId},
	{RefreshMVCountIssueRatingsUniqueService},
	{RefreshMVCountIssueRatingsOther}, // 4 - this one is not tested in database nor in e2e
	{RefreshMVCountIssueRatingsService},
	{RefreshMVCountIssueRatingsServiceWithoutSupportGroup},
	{RefreshMVCountIssueRatingsSupportGroup},
	{RefreshMVCountIssueRatingsComponentVersion},
	// The following two have to be called in sequence
	{
		RefreshMVVulnerabilityList,
		RefreshMVVulnerabilityService, // 10 - this one is not tested in database nor in e2e
	},
	{RefreshMVComponentService},
	{RefreshMVSingleComponentByServiceVulnerabilityCounts},
	{RefreshMVAllComponentsByServiceVulnerabilityCounts},
}

func RefreshMVServiceIssueCounts(ctx context.Context, db DBTX) error {
	if err := PrepareTmpTables(ctx, db, "mvServiceIssueCounts"); err != nil {
		return err
	}

	selectBuilder := sq.
		Select(
			"S.service_id",
			`COUNT(DISTINCT CASE
				WHEN IM.issuematch_rating = 'Critical'
				THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IM.issuematch_rating = 'High'
				THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IM.issuematch_rating = 'Medium'
				THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IM.issuematch_rating = 'Low'
				THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IM.issuematch_rating = 'None'
				THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id)
			END)`,
		).
		From("Service S").
		LeftJoin(`ComponentInstance CI
			ON S.service_id = CI.componentinstance_service_id
			AND CI.componentinstance_deleted_at IS NULL`).
		LeftJoin(`ComponentVersion CV
			ON CI.componentinstance_component_version_id = CV.componentversion_id
			AND CV.componentversion_deleted_at IS NULL`).
		LeftJoin(`IssueMatch IM
			ON CI.componentinstance_id = IM.issuematch_component_instance_id
			AND IM.issuematch_deleted_at IS NULL`).
		LeftJoin(`Issue I
			ON IM.issuematch_issue_id = I.issue_id
			AND I.issue_deleted_at IS NULL`).
		Where("S.service_deleted_at IS NULL").
		Where(`
			NOT EXISTS (
				SELECT 1
				FROM Remediation R
				WHERE R.remediation_service_id = S.service_id
				  AND R.remediation_issue_id = I.issue_id
				  AND R.remediation_deleted_at IS NULL
				  AND (
					  R.remediation_expiration_date IS NULL
					  OR R.remediation_expiration_date >= CURDATE()
				  )
			)`).
		GroupBy("S.service_id")

	insertBuilder := sq.
		Insert("mvServiceIssueCounts_tmp").
		Columns(
			"service_id",
			"critical_count",
			"high_count",
			"medium_count",
			"low_count",
			"none_count",
		).
		Select(selectBuilder)

	insertSQL, args, err := insertBuilder.ToSql()
	if err != nil {
		return err
	}

	if _, err = db.ExecContext(ctx, insertSQL, args...); err != nil {
		return err
	}

	return SwapTmpTables(ctx, db, "mvServiceIssueCounts")
}

func RefreshMVCountIssueRatingsServiceId(ctx context.Context, db DBTX) error {
	// Mark all current rows as inactive.
	if _, err := db.ExecContext(ctx, `
		UPDATE mvCountIssueRatingsServiceId
		SET is_active = 0
		WHERE is_active = 1`); err != nil {
		return err
	}

	selectBuilder := sq.
		Select(
			"S.service_id",
			"S.service_ccrn",
			`COUNT(DISTINCT CASE
				WHEN IM.issuematch_rating = 'Critical'
				THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IM.issuematch_rating = 'High'
				THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IM.issuematch_rating = 'Medium'
				THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IM.issuematch_rating = 'Low'
				THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IM.issuematch_rating = 'None'
				THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id)
			END)`,
			"1",
		).
		From("Service S").
		LeftJoin(`ComponentInstance CI
			ON S.service_id = CI.componentinstance_service_id
			AND CI.componentinstance_deleted_at IS NULL`).
		LeftJoin(`ComponentVersion CV
			ON CI.componentinstance_component_version_id = CV.componentversion_id
			AND CV.componentversion_deleted_at IS NULL`).
		LeftJoin(`IssueMatch IM
			ON CI.componentinstance_id = IM.issuematch_component_instance_id
			AND IM.issuematch_deleted_at IS NULL`).
		LeftJoin(`Issue I
			ON IM.issuematch_issue_id = I.issue_id
			AND I.issue_deleted_at IS NULL`).
		Where("S.service_deleted_at IS NULL").
		Where("IM.issuematch_id IS NOT NULL").
		Where(sq.Expr(`
			NOT EXISTS (
				SELECT 1
				FROM Remediation R
				WHERE R.remediation_service_id = S.service_id
				  AND R.remediation_issue_id = I.issue_id
				  AND R.remediation_deleted_at IS NULL
				  AND (
					  R.remediation_expiration_date IS NULL
					  OR R.remediation_expiration_date >= CURDATE()
				  )
			)`)).
		GroupBy("S.service_id")

	insertBuilder := sq.
		Insert("mvCountIssueRatingsServiceId").
		Columns(
			"service_id",
			"service_ccrn",
			"critical_count",
			"high_count",
			"medium_count",
			"low_count",
			"none_count",
			"is_active",
		).
		Select(selectBuilder)

	insertSQL, args, err := insertBuilder.ToSql()
	if err != nil {
		return err
	}

	insertSQL += `
		ON DUPLICATE KEY UPDATE
			service_ccrn   = VALUES(service_ccrn),
			critical_count = VALUES(critical_count),
			high_count     = VALUES(high_count),
			medium_count   = VALUES(medium_count),
			low_count      = VALUES(low_count),
			none_count     = VALUES(none_count),
			is_active      = 1`

	if _, err = db.ExecContext(ctx, insertSQL, args...); err != nil {
		return err
	}

	// Remove rows that were not refreshed.
	if _, err = db.ExecContext(ctx, `
		DELETE FROM mvCountIssueRatingsServiceId
		WHERE is_active = 0`); err != nil {
		return err
	}

	return nil
}

func RefreshMVCountIssueRatingsUniqueService(ctx context.Context, db DBTX) error {
	// Mark all current rows as inactive.
	if _, err := db.ExecContext(ctx, `
		UPDATE mvCountIssueRatingsUniqueService
		SET is_active = 0
		WHERE is_active = 1`); err != nil {
		return err
	}

	selectBuilder := sq.
		Select(
			"1",
			`COUNT(DISTINCT CASE
				WHEN IV.issuevariant_rating = 'Critical'
				THEN IV.issuevariant_issue_id
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IV.issuevariant_rating = 'High'
				THEN IV.issuevariant_issue_id
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IV.issuevariant_rating = 'Medium'
				THEN IV.issuevariant_issue_id
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IV.issuevariant_rating = 'Low'
				THEN IV.issuevariant_issue_id
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IV.issuevariant_rating = 'None'
				THEN IV.issuevariant_issue_id
			END)`,
			"1",
		).
		From("Issue I").
		LeftJoin("IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id").
		Where("I.issue_deleted_at IS NULL")

	insertBuilder := sq.
		Insert("mvCountIssueRatingsUniqueService").
		Columns(
			"id",
			"critical_count",
			"high_count",
			"medium_count",
			"low_count",
			"none_count",
			"is_active",
		).
		Select(selectBuilder)

	insertSQL, args, err := insertBuilder.ToSql()
	if err != nil {
		return err
	}

	insertSQL += `
		ON DUPLICATE KEY UPDATE
			critical_count = VALUES(critical_count),
			high_count     = VALUES(high_count),
			medium_count   = VALUES(medium_count),
			low_count      = VALUES(low_count),
			none_count     = VALUES(none_count),
			is_active      = 1`

	if _, err = db.ExecContext(ctx, insertSQL, args...); err != nil {
		return err
	}

	// Remove rows that were not refreshed.
	if _, err = db.ExecContext(ctx, `
		DELETE FROM mvCountIssueRatingsUniqueService
		WHERE is_active = 0`); err != nil {
		return err
	}

	return nil
}

func RefreshMVCountIssueRatingsOther(ctx context.Context, db DBTX) error {
	// Mark all current rows as inactive.
	if _, err := db.ExecContext(ctx, `
			UPDATE mvCountIssueRatingsOther
			SET is_active = 0
			WHERE is_active = 1`); err != nil {
		return err
	}

	selectBuilder := sq.
		Select(
			"1",
			`COUNT(DISTINCT CASE
					WHEN IV.issuevariant_rating = 'Critical'
					THEN IV.issuevariant_issue_id
				END)`,
			`COUNT(DISTINCT CASE
					WHEN IV.issuevariant_rating = 'High'
					THEN IV.issuevariant_issue_id
				END)`,
			`COUNT(DISTINCT CASE
					WHEN IV.issuevariant_rating = 'Medium'
					THEN IV.issuevariant_issue_id
				END)`,
			`COUNT(DISTINCT CASE
					WHEN IV.issuevariant_rating = 'Low'
					THEN IV.issuevariant_issue_id
				END)`,
			`COUNT(DISTINCT CASE
					WHEN IV.issuevariant_rating = 'None'
					THEN IV.issuevariant_issue_id
				END)`,
			"1",
		).
		From("Issue I").
		LeftJoin("IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id").
		Where("I.issue_deleted_at IS NULL")

	insertBuilder := sq.
		Insert("mvCountIssueRatingsOther").
		Columns(
			"id",
			"critical_count",
			"high_count",
			"medium_count",
			"low_count",
			"none_count",
			"is_active",
		).
		Select(selectBuilder)

	insertSQL, args, err := insertBuilder.ToSql()
	if err != nil {
		return err
	}

	insertSQL += `
			ON DUPLICATE KEY UPDATE
				critical_count = VALUES(critical_count),
				high_count     = VALUES(high_count),
				medium_count   = VALUES(medium_count),
				low_count      = VALUES(low_count),
				none_count     = VALUES(none_count),
				is_active      = 1`

	if _, err = db.ExecContext(ctx, insertSQL, args...); err != nil {
		return err
	}

	// Remove rows that were not refreshed.
	if _, err = db.ExecContext(ctx, `
			DELETE FROM mvCountIssueRatingsOther
			WHERE is_active = 0`); err != nil {
		return err
	}

	return nil
}

func RefreshMVCountIssueRatingsService(ctx context.Context, db DBTX) error {
	// Mark all current rows as inactive.
	if _, err := db.ExecContext(ctx, `
		UPDATE mvCountIssueRatingsService
		SET is_active = 0
		WHERE is_active = 1`); err != nil {
		return err
	}

	selectBuilder := sq.
		Select(
			`COALESCE(SG.supportgroup_ccrn, 'UNKNOWN')`,
			`COUNT(DISTINCT CASE
				WHEN IM.issuematch_rating = 'Critical'
				THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id, ',', S.service_id)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IM.issuematch_rating = 'High'
				THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id, ',', S.service_id)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IM.issuematch_rating = 'Medium'
				THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id, ',', S.service_id)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IM.issuematch_rating = 'Low'
				THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id, ',', S.service_id)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IM.issuematch_rating = 'None'
				THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id, ',', S.service_id)
			END)`,
			"1",
		).
		From("Service S").
		LeftJoin("ComponentInstance CI ON S.service_id = CI.componentinstance_service_id AND CI.componentinstance_deleted_at IS NULL").
		LeftJoin("ComponentVersion CV ON CI.componentinstance_component_version_id = CV.componentversion_id AND CV.componentversion_deleted_at IS NULL").
		LeftJoin("IssueMatch IM ON CI.componentinstance_id = IM.issuematch_component_instance_id AND IM.issuematch_deleted_at IS NULL").
		LeftJoin("Issue I ON IM.issuematch_issue_id = I.issue_id AND I.issue_deleted_at IS NULL").
		LeftJoin("SupportGroupService SGS ON SGS.supportgroupservice_service_id = S.service_id AND SGS.supportgroupservice_deleted_at IS NULL").
		LeftJoin("SupportGroup SG ON SGS.supportgroupservice_support_group_id = SG.supportgroup_id AND SG.supportgroup_deleted_at IS NULL").
		Where("S.service_deleted_at IS NULL").
		Where(sq.Expr(`
			NOT EXISTS (
				SELECT 1
				FROM Remediation R
				WHERE R.remediation_service_id = S.service_id
				  AND R.remediation_issue_id = I.issue_id
				  AND R.remediation_deleted_at IS NULL
				  AND (
					  R.remediation_expiration_date IS NULL
					  OR R.remediation_expiration_date >= CURDATE()
				  )
			)`)).
		GroupBy("SG.supportgroup_ccrn")

	insertBuilder := sq.
		Insert("mvCountIssueRatingsService").
		Columns(
			"supportgroup_ccrn",
			"critical_count",
			"high_count",
			"medium_count",
			"low_count",
			"none_count",
			"is_active",
		).
		Select(selectBuilder)

	insertSQL, args, err := insertBuilder.ToSql()
	if err != nil {
		return err
	}

	insertSQL += `
		ON DUPLICATE KEY UPDATE
			critical_count = VALUES(critical_count),
			high_count     = VALUES(high_count),
			medium_count   = VALUES(medium_count),
			low_count      = VALUES(low_count),
			none_count     = VALUES(none_count),
			is_active      = 1`

	if _, err = db.ExecContext(ctx, insertSQL, args...); err != nil {
		return err
	}

	// Remove rows that were not refreshed.
	if _, err = db.ExecContext(ctx, `
		DELETE FROM mvCountIssueRatingsService
		WHERE is_active = 0`); err != nil {
		return err
	}

	return nil
}

func RefreshMVCountIssueRatingsServiceWithoutSupportGroup(ctx context.Context, db DBTX) error {
	// Mark all current rows as inactive.
	if _, err := db.ExecContext(ctx, `
		UPDATE mvCountIssueRatingsServiceWithoutSupportGroup
		SET is_active = 0
		WHERE is_active = 1`); err != nil {
		return err
	}

	selectBuilder := sq.
		Select(
			"1",
			`COUNT(DISTINCT CASE
				WHEN IM.issuematch_rating = 'Critical'
				THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id, ',', S.service_id)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IM.issuematch_rating = 'High'
				THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id, ',', S.service_id)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IM.issuematch_rating = 'Medium'
				THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id, ',', S.service_id)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IM.issuematch_rating = 'Low'
				THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id, ',', S.service_id)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IM.issuematch_rating = 'None'
				THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id, ',', S.service_id)
			END)`,
			"1",
		).
		From("Service S").
		LeftJoin("ComponentInstance CI ON S.service_id = CI.componentinstance_service_id AND CI.componentinstance_deleted_at IS NULL").
		LeftJoin("ComponentVersion CV ON CI.componentinstance_component_version_id = CV.componentversion_id AND CV.componentversion_deleted_at IS NULL").
		LeftJoin("IssueMatch IM ON CI.componentinstance_id = IM.issuematch_component_instance_id AND IM.issuematch_deleted_at IS NULL").
		LeftJoin("Issue I ON IM.issuematch_issue_id = I.issue_id AND I.issue_deleted_at IS NULL").
		Where("S.service_deleted_at IS NULL").
		Where(sq.Expr(`
			NOT EXISTS (
				SELECT 1
				FROM Remediation R
				WHERE R.remediation_service_id = S.service_id
				  AND R.remediation_issue_id = I.issue_id
				  AND R.remediation_deleted_at IS NULL
				  AND (
					  R.remediation_expiration_date IS NULL
					  OR R.remediation_expiration_date >= CURDATE()
				  )
			)`))

	insertBuilder := sq.
		Insert("mvCountIssueRatingsServiceWithoutSupportGroup").
		Columns(
			"id",
			"critical_count",
			"high_count",
			"medium_count",
			"low_count",
			"none_count",
			"is_active",
		).
		Select(selectBuilder)

	insertSQL, args, err := insertBuilder.ToSql()
	if err != nil {
		return err
	}

	insertSQL += `
		ON DUPLICATE KEY UPDATE
			critical_count = VALUES(critical_count),
			high_count     = VALUES(high_count),
			medium_count   = VALUES(medium_count),
			low_count      = VALUES(low_count),
			none_count     = VALUES(none_count),
			is_active      = 1`

	if _, err = db.ExecContext(ctx, insertSQL, args...); err != nil {
		return err
	}

	// Remove rows that were not refreshed.
	if _, err = db.ExecContext(ctx, `
		DELETE FROM mvCountIssueRatingsServiceWithoutSupportGroup
		WHERE is_active = 0`); err != nil {
		return err
	}

	return nil
}

func RefreshMVCountIssueRatingsSupportGroup(ctx context.Context, db DBTX) error {
	// Mark all current rows as inactive.
	if _, err := db.ExecContext(ctx, `
		UPDATE mvCountIssueRatingsSupportGroup
		SET is_active = 0
		WHERE is_active = 1`); err != nil {
		return err
	}

	selectBuilder := sq.
		Select(
			`COALESCE(SG.supportgroup_ccrn, 'UNKNOWN')`,
			`COUNT(DISTINCT CASE
				WHEN IM.issuematch_rating = 'Critical'
				THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id, ',', S.service_id)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IM.issuematch_rating = 'High'
				THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id, ',', S.service_id)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IM.issuematch_rating = 'Medium'
				THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id, ',', S.service_id)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IM.issuematch_rating = 'Low'
				THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id, ',', S.service_id)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IM.issuematch_rating = 'None'
				THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id, ',', S.service_id)
			END)`,
			"1",
		).
		From("Service S").
		LeftJoin("ComponentInstance CI ON S.service_id = CI.componentinstance_service_id AND CI.componentinstance_deleted_at IS NULL").
		LeftJoin("ComponentVersion CV ON CI.componentinstance_component_version_id = CV.componentversion_id AND CV.componentversion_deleted_at IS NULL").
		LeftJoin("IssueMatch IM ON CI.componentinstance_id = IM.issuematch_component_instance_id AND IM.issuematch_deleted_at IS NULL").
		LeftJoin("Issue I ON IM.issuematch_issue_id = I.issue_id AND I.issue_deleted_at IS NULL").
		LeftJoin("SupportGroupService SGS ON SGS.supportgroupservice_service_id = S.service_id AND SGS.supportgroupservice_deleted_at IS NULL").
		LeftJoin("SupportGroup SG ON SGS.supportgroupservice_support_group_id = SG.supportgroup_id AND SG.supportgroup_deleted_at IS NULL").
		Where("S.service_deleted_at IS NULL").
		Where("IM.issuematch_id IS NOT NULL").
		Where(sq.Expr(`
			NOT EXISTS (
				SELECT 1
				FROM Remediation R
				WHERE R.remediation_service_id = S.service_id
				  AND R.remediation_issue_id = I.issue_id
				  AND R.remediation_deleted_at IS NULL
				  AND (
					  R.remediation_expiration_date IS NULL
					  OR R.remediation_expiration_date >= CURDATE()
				  )
			)`)).
		GroupBy("SG.supportgroup_ccrn")

	insertBuilder := sq.
		Insert("mvCountIssueRatingsSupportGroup").
		Columns(
			"supportgroup_ccrn",
			"critical_count",
			"high_count",
			"medium_count",
			"low_count",
			"none_count",
			"is_active",
		).
		Select(selectBuilder)

	insertSQL, args, err := insertBuilder.ToSql()
	if err != nil {
		return err
	}

	insertSQL += `
		ON DUPLICATE KEY UPDATE
			critical_count = VALUES(critical_count),
			high_count     = VALUES(high_count),
			medium_count   = VALUES(medium_count),
			low_count      = VALUES(low_count),
			none_count     = VALUES(none_count),
			is_active      = 1`

	if _, err = db.ExecContext(ctx, insertSQL, args...); err != nil {
		return err
	}

	// Remove rows that were not refreshed.
	if _, err = db.ExecContext(ctx, `
		DELETE FROM mvCountIssueRatingsSupportGroup
		WHERE is_active = 0`); err != nil {
		return err
	}

	return nil
}

func RefreshMVCountIssueRatingsComponentVersion(ctx context.Context, db DBTX) error {
	// Mark all current rows as inactive.
	if _, err := db.ExecContext(ctx, `
		UPDATE mvCountIssueRatingsComponentVersion
		SET is_active = 0
		WHERE is_active = 1`); err != nil {
		return err
	}

	selectBuilder := sq.
		Select(
			"CVI.componentversionissue_component_version_id",
			"CI.componentinstance_service_id",
			"S.service_ccrn",
			`COUNT(DISTINCT CASE
				WHEN IV.issuevariant_rating = 'Critical'
				THEN CONCAT(
					CVI.componentversionissue_component_version_id,
					',',
					CVI.componentversionissue_issue_id
				)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IV.issuevariant_rating = 'High'
				THEN CONCAT(
					CVI.componentversionissue_component_version_id,
					',',
					CVI.componentversionissue_issue_id
				)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IV.issuevariant_rating = 'Medium'
				THEN CONCAT(
					CVI.componentversionissue_component_version_id,
					',',
					CVI.componentversionissue_issue_id
				)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IV.issuevariant_rating = 'Low'
				THEN CONCAT(
					CVI.componentversionissue_component_version_id,
					',',
					CVI.componentversionissue_issue_id
				)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IV.issuevariant_rating = 'None'
				THEN CONCAT(
					CVI.componentversionissue_component_version_id,
					',',
					CVI.componentversionissue_issue_id
				)
			END)`,
			"1",
		).
		From("ComponentVersionIssue CVI").
		LeftJoin("IssueVariant IV ON IV.issuevariant_issue_id = CVI.componentversionissue_issue_id").
		Join("ComponentInstance CI ON CVI.componentversionissue_component_version_id = CI.componentinstance_component_version_id").
		Join("Service S ON CI.componentinstance_service_id = S.service_id").
		Where(sq.Expr(`
			NOT EXISTS (
				SELECT 1
				FROM Remediation R
				WHERE R.remediation_service_id = CI.componentinstance_service_id
				  AND R.remediation_issue_id = CVI.componentversionissue_issue_id
				  AND R.remediation_deleted_at IS NULL
				  AND (
					  R.remediation_expiration_date IS NULL
					  OR R.remediation_expiration_date >= CURDATE()
				  )
			)`)).
		GroupBy("CVI.componentversionissue_component_version_id")

	insertBuilder := sq.
		Insert("mvCountIssueRatingsComponentVersion").
		Columns(
			"component_version_id",
			"service_id",
			"service_ccrn",
			"critical_count",
			"high_count",
			"medium_count",
			"low_count",
			"none_count",
			"is_active",
		).
		Select(selectBuilder)

	insertSQL, args, err := insertBuilder.ToSql()
	if err != nil {
		return err
	}

	insertSQL += `
		ON DUPLICATE KEY UPDATE
			service_id     = VALUES(service_id),
			service_ccrn   = VALUES(service_ccrn),
			critical_count = VALUES(critical_count),
			high_count     = VALUES(high_count),
			medium_count   = VALUES(medium_count),
			low_count      = VALUES(low_count),
			none_count     = VALUES(none_count),
			is_active      = 1`

	if _, err = db.ExecContext(ctx, insertSQL, args...); err != nil {
		return err
	}

	// Remove rows that were not refreshed.
	if _, err = db.ExecContext(ctx, `
		DELETE FROM mvCountIssueRatingsComponentVersion
		WHERE is_active = 0`); err != nil {
		return err
	}

	return nil
}

func RefreshMVVulnerabilityList(ctx context.Context, db DBTX) error {
	if err := PrepareTmpTables(ctx, db, "mvVulnerabilityList"); err != nil {
		return err
	}

	selectBuilder := sq.
		Select(
			"I.issue_id",
			"MAX(IM.issuematch_rating)",
			"MIN(IM.issuematch_target_remediation_date)",
			`(
				SELECT MIN(IV.issuevariant_external_url)
				FROM IssueVariant IV
				WHERE IV.issuevariant_issue_id = I.issue_id
				  AND IV.issuevariant_deleted_at IS NULL
				  AND IV.issuevariant_external_url != ''
			)`,
		).
		From("Issue I").
		RightJoin("IssueMatch IM ON I.issue_id = IM.issuematch_issue_id").
		Where("IM.issuematch_status = 'new'").
		Where("IM.issuematch_deleted_at IS NULL").
		Where("I.issue_type = 'Vulnerability'").
		Where("I.issue_deleted_at IS NULL").
		GroupBy("I.issue_id")

	insertBuilder := sq.
		Insert("mvVulnerabilityList_tmp").
		Columns(
			"issue_id",
			"max_severity",
			"earliest_remediation_date",
			"source_url",
		).
		Select(selectBuilder)

	insertSQL, args, err := insertBuilder.ToSql()
	if err != nil {
		return err
	}

	if _, err = db.ExecContext(ctx, insertSQL, args...); err != nil {
		return err
	}

	return SwapTmpTables(ctx, db, "mvVulnerabilityList")
}

func RefreshMVVulnerabilityService(ctx context.Context, db DBTX) error {
	if err := PrepareTmpTables(ctx, db, "mvVulnerabilityService"); err != nil {
		return err
	}

	selectBuilder := sq.
		Select("DISTINCT MVL.issue_id", "CI.componentinstance_service_id").
		From("mvVulnerabilityList MVL").
		Join("IssueMatch IM ON MVL.issue_id = IM.issuematch_issue_id").
		Join("ComponentInstance CI ON IM.issuematch_component_instance_id = CI.componentinstance_id").
		Where("IM.issuematch_deleted_at IS NULL").
		Where("CI.componentinstance_deleted_at IS NULL")

	insertBuilder := sq.
		Insert("mvVulnerabilityService_tmp").
		Columns(
			"issue_id",
			"service_id",
		).
		Select(selectBuilder)

	insertSQL, args, err := insertBuilder.ToSql()
	if err != nil {
		return err
	}

	if _, err = db.ExecContext(ctx, insertSQL, args...); err != nil {
		return err
	}

	return SwapTmpTables(ctx, db, "mvVulnerabilityService")
}

func RefreshMVComponentService(ctx context.Context, db DBTX) error {
	if err := PrepareTmpTables(ctx, db, "mvComponentService"); err != nil {
		return err
	}

	selectBuilder := sq.
		Select(
			"CI.componentinstance_service_id",
			"CV.componentversion_component_id",
		).
		Distinct().
		From("ComponentInstance CI").
		Join("ComponentVersion CV ON CI.componentinstance_component_version_id = CV.componentversion_id").
		Where("CI.componentinstance_deleted_at IS NULL")

	insertBuilder := sq.
		Insert("mvComponentService_tmp").
		Columns(
			"service_id",
			"component_id",
		).
		Select(selectBuilder)

	insertSQL, args, err := insertBuilder.ToSql()
	if err != nil {
		return err
	}

	if _, err = db.ExecContext(ctx, insertSQL, args...); err != nil {
		return err
	}

	return SwapTmpTables(ctx, db, "mvComponentService")
}

func RefreshMVSingleComponentByServiceVulnerabilityCounts(ctx context.Context, db DBTX) error {
	// Mark all current rows as inactive.
	if _, err := db.ExecContext(ctx, `
		UPDATE mvSingleComponentByServiceVulnerabilityCounts
		SET is_active = 0`); err != nil {
		return err
	}

	selectBuilder := sq.
		Select(
			"CI.componentinstance_service_id",
			"CV.componentversion_component_id",
			`COUNT(DISTINCT CASE
				WHEN IM.issuematch_rating = 'Critical'
				THEN CONCAT(CV.componentversion_component_id, ',', IM.issuematch_issue_id)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IM.issuematch_rating = 'High'
				THEN CONCAT(CV.componentversion_component_id, ',', IM.issuematch_issue_id)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IM.issuematch_rating = 'Medium'
				THEN CONCAT(CV.componentversion_component_id, ',', IM.issuematch_issue_id)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IM.issuematch_rating = 'Low'
				THEN CONCAT(CV.componentversion_component_id, ',', IM.issuematch_issue_id)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IM.issuematch_rating = 'None'
				THEN CONCAT(CV.componentversion_component_id, ',', IM.issuematch_issue_id)
			END)`,
			"1",
		).
		From("IssueMatch IM").
		Join("ComponentInstance CI ON CI.componentinstance_id = IM.issuematch_component_instance_id").
		Join("ComponentVersion CV ON CV.componentversion_id = CI.componentinstance_component_version_id").
		Join("Issue I ON I.issue_id = IM.issuematch_issue_id").
		Where("IM.issuematch_status = 'new'").
		Where("I.issue_type = 'Vulnerability'").
		Where("IM.issuematch_deleted_at IS NULL").
		Where("I.issue_deleted_at IS NULL").
		Where("CI.componentinstance_deleted_at IS NULL").
		Where("CV.componentversion_deleted_at IS NULL").
		Where(sq.Expr(`
			NOT EXISTS (
				SELECT 1
				FROM Remediation R
				WHERE R.remediation_service_id = CI.componentinstance_service_id
				  AND R.remediation_issue_id = I.issue_id
				  AND R.remediation_deleted_at IS NULL
				  AND (
					  R.remediation_expiration_date IS NULL
					  OR R.remediation_expiration_date >= CURDATE()
				  )
			)`)).
		GroupBy(
			"CI.componentinstance_service_id",
			"CV.componentversion_component_id",
		)

	insertBuilder := sq.
		Insert("mvSingleComponentByServiceVulnerabilityCounts").
		Columns(
			"service_id",
			"component_id",
			"critical_count",
			"high_count",
			"medium_count",
			"low_count",
			"none_count",
			"is_active",
		).
		Select(selectBuilder)

	insertSQL, args, err := insertBuilder.ToSql()
	if err != nil {
		return err
	}

	insertSQL += `
		ON DUPLICATE KEY UPDATE
			critical_count = VALUES(critical_count),
			high_count     = VALUES(high_count),
			medium_count   = VALUES(medium_count),
			low_count      = VALUES(low_count),
			none_count     = VALUES(none_count),
			is_active      = 1`

	if _, err = db.ExecContext(ctx, insertSQL, args...); err != nil {
		return err
	}

	return nil
}

func RefreshMVAllComponentsByServiceVulnerabilityCounts(ctx context.Context, db DBTX) error {
	// Mark all current rows as inactive.
	if _, err := db.ExecContext(ctx, `
		UPDATE mvAllComponentsByServiceVulnerabilityCounts
		SET is_active = 0`); err != nil {
		return err
	}

	selectBuilder := sq.
		Select(
			"CI.componentinstance_service_id",
			`COUNT(DISTINCT CASE
				WHEN IV.issuevariant_rating = 'Critical'
				THEN CONCAT(CV.componentversion_component_id, ',', IV.issuevariant_issue_id)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IV.issuevariant_rating = 'High'
				THEN CONCAT(CV.componentversion_component_id, ',', IV.issuevariant_issue_id)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IV.issuevariant_rating = 'Medium'
				THEN CONCAT(CV.componentversion_component_id, ',', IV.issuevariant_issue_id)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IV.issuevariant_rating = 'Low'
				THEN CONCAT(CV.componentversion_component_id, ',', IV.issuevariant_issue_id)
			END)`,
			`COUNT(DISTINCT CASE
				WHEN IV.issuevariant_rating = 'None'
				THEN CONCAT(CV.componentversion_component_id, ',', IV.issuevariant_issue_id)
			END)`,
			"1",
		).
		From("IssueMatch IM").
		Join("ComponentInstance CI ON CI.componentinstance_id = IM.issuematch_component_instance_id").
		Join("ComponentVersion CV ON CV.componentversion_id = CI.componentinstance_component_version_id").
		Join("IssueVariant IV ON IV.issuevariant_issue_id = IM.issuematch_issue_id").
		Join("Issue I ON I.issue_id = IV.issuevariant_issue_id").
		Where("IM.issuematch_status = 'new'").
		Where("I.issue_type = 'Vulnerability'").
		Where("IM.issuematch_deleted_at IS NULL").
		Where("I.issue_deleted_at IS NULL").
		Where("CI.componentinstance_deleted_at IS NULL").
		Where("CV.componentversion_deleted_at IS NULL").
		Where(sq.Expr(`
			NOT EXISTS (
				SELECT 1
				FROM Remediation R
				WHERE R.remediation_service_id = CI.componentinstance_service_id
				  AND R.remediation_issue_id = I.issue_id
				  AND R.remediation_deleted_at IS NULL
				  AND (
					  R.remediation_expiration_date IS NULL
					  OR R.remediation_expiration_date >= CURDATE()
				  )
			)`)).
		GroupBy("CI.componentinstance_service_id")

	insertBuilder := sq.
		Insert("mvAllComponentsByServiceVulnerabilityCounts").
		Columns(
			"service_id",
			"critical_count",
			"high_count",
			"medium_count",
			"low_count",
			"none_count",
			"is_active",
		).
		Select(selectBuilder)

	insertSQL, args, err := insertBuilder.ToSql()
	if err != nil {
		return err
	}

	insertSQL += `
		ON DUPLICATE KEY UPDATE
			critical_count = VALUES(critical_count),
			high_count     = VALUES(high_count),
			medium_count   = VALUES(medium_count),
			low_count      = VALUES(low_count),
			none_count     = VALUES(none_count),
			is_active      = 1`

	if _, err = db.ExecContext(ctx, insertSQL, args...); err != nil {
		return err
	}

	return nil
}

func DropTableIfExists(ctx context.Context, db DBTX, tables ...string) error {
	for _, table := range tables {
		if _, err := db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", table)); err != nil {
			return err
		}
	}

	return nil
}

func PrepareTmpTables(ctx context.Context, db DBTX, tableName string) error {
	// Ensure clean state for atomic swap.
	tmpTable := tableName + "_tmp"
	oldTable := tableName + "_old"

	if err := DropTableIfExists(ctx, db, tmpTable, oldTable); err != nil {
		return err
	}

	// Create temporary table.
	_, err := db.ExecContext(ctx, fmt.Sprintf("CREATE TABLE %s LIKE %s", tmpTable, tableName))

	return err
}

func SwapTmpTables(ctx context.Context, db DBTX, tableName string) error {
	tmpTable := tableName + "_tmp"
	oldTable := tableName + "_old"

	// Atomically swap the tables.
	if _, err := db.ExecContext(ctx, fmt.Sprintf("RENAME TABLE %s TO %s, %s TO %s", tableName, oldTable, tmpTable, tableName)); err != nil {
		return err
	}

	// Remove the previous table.
	return DropTableIfExists(ctx, db, oldTable)
}

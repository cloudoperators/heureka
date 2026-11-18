// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

// Targeted MV refresh for a single remediation change.
//
// A full MVProcedures run rebuilds every MV table, which is expensive and makes
// a rescore's new severity take seconds to appear. The vulnerability severity the
// UI reads only lives in mvVulnerabilityList (per issue) and mvVulnerabilityService
// (per service+issue), so a rescore for one (service, issue) only needs those two
// rows recomputed. These functions do exactly that with DELETE+INSERT on the live
// tables, mirroring the SELECT logic of the full procedures with a key filter.
//
// The count tables are left to the periodic/async full refresh.

// RefreshMVVulnerabilityListForIssue recomputes the single mvVulnerabilityList row
// for one issue (its max_severity is the max across all the issue's services, so the
// whole issue row is recomputed).
func RefreshMVVulnerabilityListForIssue(ctx context.Context, db DBTX, issueID int64) error {
	if _, err := db.ExecContext(ctx, "DELETE FROM mvVulnerabilityList WHERE issue_id = ?", issueID); err != nil {
		return err
	}

	effectiveRating := effectiveRatingCol("IM.issuematch_rating")
	rescoreJoinExpr, rescoreJoinArgs := latestRescoreJoin("CI.componentinstance_service_id", "IM.issuematch_issue_id")

	selectBuilder := sq.
		Select(
			"I.issue_id",
			fmt.Sprintf(
				`ELT(
					MAX(FIELD(%s, 'None', 'Low', 'Medium', 'High', 'Critical')),
					'None', 'Low', 'Medium', 'High', 'Critical'
				)`,
				effectiveRating,
			),
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
		LeftJoin("ComponentInstance CI ON CI.componentinstance_id = IM.issuematch_component_instance_id AND CI.componentinstance_deleted_at IS NULL").
		LeftJoin(rescoreJoinExpr, rescoreJoinArgs...).
		Where("IM.issuematch_status = 'new'").
		Where("IM.issuematch_deleted_at IS NULL").
		Where("I.issue_type = 'Vulnerability'").
		Where("I.issue_deleted_at IS NULL").
		Where(sq.Eq{"I.issue_id": issueID}).
		GroupBy("I.issue_id")

	insertSQL, args, err := sq.
		Insert("mvVulnerabilityList").
		Columns("issue_id", "max_severity", "earliest_remediation_date", "source_url").
		Select(selectBuilder).
		ToSql()
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, insertSQL, args...)

	return err
}

// RefreshMVVulnerabilityServiceForServiceIssue recomputes the single
// mvVulnerabilityService row for one (service, issue). Reads mvVulnerabilityList, so
// call RefreshMVVulnerabilityListForIssue first in the same transaction.
func RefreshMVVulnerabilityServiceForServiceIssue(ctx context.Context, db DBTX, serviceID, issueID int64) error {
	if _, err := db.ExecContext(ctx,
		"DELETE FROM mvVulnerabilityService WHERE service_id = ? AND issue_id = ?", serviceID, issueID,
	); err != nil {
		return err
	}

	effectiveRating := effectiveRatingCol("IM.issuematch_rating")
	maxSeverity := fmt.Sprintf(`ELT(
			COALESCE(
				MAX(CASE WHEN IM.issuematch_status = 'new' THEN FIELD(%s, 'None', 'Low', 'Medium', 'High', 'Critical') END),
				MAX(FIELD(%s, 'None', 'Low', 'Medium', 'High', 'Critical'))
			),
			'None', 'Low', 'Medium', 'High', 'Critical'
		)`, effectiveRating, effectiveRating)

	rescoreJoinExpr, rescoreJoinArgs := latestRescoreJoin("CI.componentinstance_service_id", "IM.issuematch_issue_id")

	selectBuilder := sq.
		Select(
			"MVL.issue_id",
			"CI.componentinstance_service_id",
			maxSeverity,
		).
		From("mvVulnerabilityList MVL").
		Join("IssueMatch IM ON MVL.issue_id = IM.issuematch_issue_id").
		Join("ComponentInstance CI ON IM.issuematch_component_instance_id = CI.componentinstance_id").
		LeftJoin(rescoreJoinExpr, rescoreJoinArgs...).
		Where("IM.issuematch_deleted_at IS NULL").
		Where("CI.componentinstance_deleted_at IS NULL").
		Where(sq.Eq{"MVL.issue_id": issueID}).
		Where(sq.Eq{"CI.componentinstance_service_id": serviceID}).
		GroupBy("MVL.issue_id", "CI.componentinstance_service_id")

	insertSQL, args, err := sq.
		Insert("mvVulnerabilityService").
		Columns("issue_id", "service_id", "max_severity").
		Select(selectBuilder).
		ToSql()
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, insertSQL, args...)

	return err
}

// RefreshVulnerabilityForRemediation recomputes the severity MV rows affected by a
// remediation on (serviceID, issueID), in one transaction. Cheap enough to run
// synchronously on a remediation change so the new severity is visible immediately,
// while the count tables are refreshed by the periodic/async full run.
func RefreshVulnerabilityForRemediation(ctx context.Context, db Db, serviceID, issueID int64) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}

	if err := RefreshMVVulnerabilityListForIssue(ctx, tx, issueID); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := RefreshMVVulnerabilityServiceForServiceIssue(ctx, tx, serviceID, issueID); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

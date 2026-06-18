-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

-- Fix: The mvSingleComponentByServiceVulnerabilityCounts MV was using
-- IV.issuevariant_rating to bucket severity counts, but the detail view
-- (VulnerabilityBaseResolver with HasIssueMatches) filters on IM.issuematch_rating.
-- These can diverge when a scanner re-rates a vulnerability without updating
-- existing IssueMatches. Switch to IM.issuematch_rating so badge counts always
-- match the detail view.
--
-- Additionally removes the IssueVariant JOIN from this procedure — it's no longer
-- needed since we read severity directly from IssueMatch.

DROP PROCEDURE IF EXISTS refresh_mvSingleComponentByServiceVulnerabilityCounts_proc;

CREATE PROCEDURE refresh_mvSingleComponentByServiceVulnerabilityCounts_proc()
BEGIN
    SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED;

    UPDATE mvSingleComponentByServiceVulnerabilityCounts
    SET is_active = 0;

    INSERT INTO mvSingleComponentByServiceVulnerabilityCounts (
        service_id,
        component_id,
        critical_count,
        high_count,
        medium_count,
        low_count,
        none_count,
        is_active
    )
    SELECT
        CI.componentinstance_service_id,
        CV.componentversion_component_id,
        COUNT(DISTINCT CASE WHEN IM.issuematch_rating = 'Critical'
            THEN CONCAT(CV.componentversion_component_id, ',', IM.issuematch_issue_id) END),
        COUNT(DISTINCT CASE WHEN IM.issuematch_rating = 'High'
            THEN CONCAT(CV.componentversion_component_id, ',', IM.issuematch_issue_id) END),
        COUNT(DISTINCT CASE WHEN IM.issuematch_rating = 'Medium'
            THEN CONCAT(CV.componentversion_component_id, ',', IM.issuematch_issue_id) END),
        COUNT(DISTINCT CASE WHEN IM.issuematch_rating = 'Low'
            THEN CONCAT(CV.componentversion_component_id, ',', IM.issuematch_issue_id) END),
        COUNT(DISTINCT CASE WHEN IM.issuematch_rating = 'None'
            THEN CONCAT(CV.componentversion_component_id, ',', IM.issuematch_issue_id) END),
        1
    FROM IssueMatch IM
    JOIN ComponentInstance CI ON CI.componentinstance_id = IM.issuematch_component_instance_id
    JOIN ComponentVersion CV ON CV.componentversion_id = CI.componentinstance_component_version_id
    JOIN Issue I ON I.issue_id = IM.issuematch_issue_id
    WHERE IM.issuematch_status = 'new'
      AND I.issue_type = 'Vulnerability'
      AND IM.issuematch_deleted_at IS NULL
      AND I.issue_deleted_at IS NULL
      AND CI.componentinstance_deleted_at IS NULL
      AND CV.componentversion_deleted_at IS NULL
      AND NOT EXISTS (
          SELECT 1 FROM Remediation R
          WHERE R.remediation_service_id = CI.componentinstance_service_id
            AND R.remediation_issue_id = I.issue_id
            AND R.remediation_deleted_at IS NULL
            AND (R.remediation_expiration_date IS NULL OR R.remediation_expiration_date >= CURDATE())
      )
    GROUP BY CI.componentinstance_service_id, CV.componentversion_component_id
    ON DUPLICATE KEY UPDATE
        critical_count = VALUES(critical_count),
        high_count     = VALUES(high_count),
        medium_count   = VALUES(medium_count),
        low_count      = VALUES(low_count),
        none_count     = VALUES(none_count),
        is_active      = 1;
END;

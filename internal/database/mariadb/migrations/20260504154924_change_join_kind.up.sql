-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsServiceId_proc;

CREATE PROCEDURE refresh_mvCountIssueRatingsServiceId_proc()
BEGIN
    UPDATE mvCountIssueRatingsServiceId
    SET is_active = 0
    WHERE is_active = 1;

    INSERT INTO mvCountIssueRatingsServiceId (
        service_id,
        service_ccrn,
        critical_count,
        high_count,
        medium_count,
        low_count,
        none_count,
        is_active
    )
    SELECT
        CI.componentinstance_service_id,
        S.service_ccrn,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id) END),
        1
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    LEFT JOIN IssueMatch IM ON I.issue_id = IM.issuematch_issue_id
    LEFT JOIN ComponentInstance CI
        ON CI.componentinstance_id = IM.issuematch_component_instance_id
       AND CI.componentinstance_deleted_at IS NULL
    JOIN Service S ON S.service_id = CI.componentinstance_service_id
    LEFT JOIN Remediation R
        ON CI.componentinstance_service_id = R.remediation_service_id
       AND I.issue_id = R.remediation_issue_id
       AND R.remediation_deleted_at IS NULL
    WHERE I.issue_deleted_at IS NULL
      AND IM.issuematch_id IS NOT NULL
      AND (R.remediation_id IS NULL OR R.remediation_expiration_date < CURDATE())
    GROUP BY CI.componentinstance_service_id
    ON DUPLICATE KEY UPDATE
        service_ccrn   = VALUES(service_ccrn),
        critical_count = VALUES(critical_count),
        high_count     = VALUES(high_count),
        medium_count   = VALUES(medium_count),
        low_count      = VALUES(low_count),
        none_count     = VALUES(none_count),
        is_active      = 1;

    DELETE FROM mvCountIssueRatingsServiceId
    WHERE is_active = 0;
END;
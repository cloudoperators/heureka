-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0



DROP PROCEDURE IF EXISTS refresh_mvServiceIssueCounts_proc;
CREATE PROCEDURE refresh_mvServiceIssueCounts_proc()
BEGIN
    -- Ensure clean state for atomic swap
    DROP TABLE IF EXISTS mvServiceIssueCounts_tmp;
    DROP TABLE IF EXISTS mvServiceIssueCounts_old;

    CREATE TABLE mvServiceIssueCounts_tmp LIKE mvServiceIssueCounts;

    INSERT INTO mvServiceIssueCounts_tmp (
        service_id,
        critical_count,
        high_count,
        medium_count,
        low_count,
        none_count
    )
    SELECT
        S.service_id,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id) END) AS critical_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id) END) AS high_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id) END) AS medium_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id) END) AS low_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id) END) AS none_count
    FROM Service S
    LEFT JOIN ComponentInstance CI ON S.service_id = CI.componentinstance_service_id AND CI.componentinstance_deleted_at IS NULL
    LEFT JOIN IssueMatch IM ON CI.componentinstance_id = IM.issuematch_component_instance_id AND IM.issuematch_deleted_at IS NULL
    LEFT JOIN Issue I ON IM.issuematch_issue_id = I.issue_id AND I.issue_deleted_at IS NULL
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    LEFT JOIN Remediation R
        ON S.service_id = R.remediation_service_id
       AND I.issue_id = R.remediation_issue_id
       AND R.remediation_deleted_at IS NULL
    WHERE S.service_deleted_at IS NULL
      AND (R.remediation_id IS NULL OR R.remediation_expiration_date < CURDATE())
    GROUP BY S.service_id;

    RENAME TABLE
        mvServiceIssueCounts TO mvServiceIssueCounts_old,
        mvServiceIssueCounts_tmp TO mvServiceIssueCounts;

    DROP TABLE IF EXISTS mvServiceIssueCounts_old;
END;


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
        S.service_id,
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
    FROM Service S
    LEFT JOIN ComponentInstance CI ON S.service_id = CI.componentinstance_service_id AND CI.componentinstance_deleted_at IS NULL
    LEFT JOIN IssueMatch IM ON CI.componentinstance_id = IM.issuematch_component_instance_id AND IM.issuematch_deleted_at IS NULL
    LEFT JOIN Issue I ON IM.issuematch_issue_id = I.issue_id AND I.issue_deleted_at IS NULL
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    LEFT JOIN Remediation R
        ON S.service_id = R.remediation_service_id
       AND I.issue_id = R.remediation_issue_id
       AND R.remediation_deleted_at IS NULL
    WHERE S.service_deleted_at IS NULL
      AND (R.remediation_id IS NULL OR R.remediation_expiration_date < CURDATE())
    GROUP BY S.service_id
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

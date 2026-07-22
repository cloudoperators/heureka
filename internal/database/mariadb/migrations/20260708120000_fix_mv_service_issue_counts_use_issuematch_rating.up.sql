-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

-- Fix: multiple MVs were overcounting vulnerabilities when a service has multiple deployed
-- versions of the same image. The old dedup key was
-- CONCAT(componentinstance_component_version_id, issue_id), so the same CVE appearing in
-- N versions of an image was counted N times. The per-image badges
-- (mvSingleComponentByServiceVulnerabilityCounts) deduplicate at the component level,
-- so the service-level and support-group totals diverged proportionally to the number of
-- deployed versions.
-- Fix: deduplicate using CONCAT(componentversion_component_id, issue_id) to count each
-- unique (image, vulnerability) pair exactly once, matching the per-image badge semantics.
-- Also switches from IV.issuevariant_rating to IM.issuematch_rating to match the source
-- used by the per-image MV and the vulnerability detail view, and removes the now-unused
-- IssueVariant JOINs.
--
-- Affected procedures:
--   refresh_mvCountIssueRatingsServiceId_proc              (service detail page top total)
--   refresh_mvServiceIssueCounts_proc                      (service list per-row counts)
--   refresh_mvCountIssueRatingsSupportGroup_proc           (support group page top total)
--   refresh_mvCountIssueRatingsService_proc                (service list + SG filter total)
--   refresh_mvCountIssueRatingsServiceWithoutSupportGroup_proc (service list no-filter total)

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsServiceId_proc;

CREATE PROCEDURE refresh_mvCountIssueRatingsServiceId_proc()
BEGIN
    SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED;

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
        COUNT(DISTINCT CASE WHEN IM.issuematch_rating = 'Critical'
            THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id) END),
        COUNT(DISTINCT CASE WHEN IM.issuematch_rating = 'High'
            THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id) END),
        COUNT(DISTINCT CASE WHEN IM.issuematch_rating = 'Medium'
            THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id) END),
        COUNT(DISTINCT CASE WHEN IM.issuematch_rating = 'Low'
            THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id) END),
        COUNT(DISTINCT CASE WHEN IM.issuematch_rating = 'None'
            THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id) END),
        1
    FROM Service S
    LEFT JOIN ComponentInstance CI ON S.service_id = CI.componentinstance_service_id AND CI.componentinstance_deleted_at IS NULL
    LEFT JOIN ComponentVersion CV ON CI.componentinstance_component_version_id = CV.componentversion_id AND CV.componentversion_deleted_at IS NULL
    LEFT JOIN IssueMatch IM ON CI.componentinstance_id = IM.issuematch_component_instance_id AND IM.issuematch_deleted_at IS NULL
    LEFT JOIN Issue I ON IM.issuematch_issue_id = I.issue_id AND I.issue_deleted_at IS NULL
    WHERE S.service_deleted_at IS NULL
      AND IM.issuematch_id IS NOT NULL
      AND NOT EXISTS (
          SELECT 1 FROM Remediation R
          WHERE R.remediation_service_id = S.service_id
            AND R.remediation_issue_id = I.issue_id
            AND R.remediation_deleted_at IS NULL
            AND (R.remediation_expiration_date IS NULL OR R.remediation_expiration_date >= CURDATE())
      )
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

DROP PROCEDURE IF EXISTS refresh_mvServiceIssueCounts_proc;

CREATE PROCEDURE refresh_mvServiceIssueCounts_proc()
BEGIN
    SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED;

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
        COUNT(DISTINCT CASE WHEN IM.issuematch_rating = 'Critical'
            THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id) END) AS critical_count,
        COUNT(DISTINCT CASE WHEN IM.issuematch_rating = 'High'
            THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id) END) AS high_count,
        COUNT(DISTINCT CASE WHEN IM.issuematch_rating = 'Medium'
            THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id) END) AS medium_count,
        COUNT(DISTINCT CASE WHEN IM.issuematch_rating = 'Low'
            THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id) END) AS low_count,
        COUNT(DISTINCT CASE WHEN IM.issuematch_rating = 'None'
            THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id) END) AS none_count
    FROM Service S
    LEFT JOIN ComponentInstance CI ON S.service_id = CI.componentinstance_service_id AND CI.componentinstance_deleted_at IS NULL
    LEFT JOIN ComponentVersion CV ON CI.componentinstance_component_version_id = CV.componentversion_id AND CV.componentversion_deleted_at IS NULL
    LEFT JOIN IssueMatch IM ON CI.componentinstance_id = IM.issuematch_component_instance_id AND IM.issuematch_deleted_at IS NULL
    LEFT JOIN Issue I ON IM.issuematch_issue_id = I.issue_id AND I.issue_deleted_at IS NULL
    WHERE S.service_deleted_at IS NULL
      AND NOT EXISTS (
          SELECT 1 FROM Remediation R
          WHERE R.remediation_service_id = S.service_id
            AND R.remediation_issue_id = I.issue_id
            AND R.remediation_deleted_at IS NULL
            AND (R.remediation_expiration_date IS NULL OR R.remediation_expiration_date >= CURDATE())
      )
    GROUP BY S.service_id;

    RENAME TABLE
        mvServiceIssueCounts TO mvServiceIssueCounts_old,
        mvServiceIssueCounts_tmp TO mvServiceIssueCounts;

    DROP TABLE IF EXISTS mvServiceIssueCounts_old;
END;

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsSupportGroup_proc;

CREATE PROCEDURE refresh_mvCountIssueRatingsSupportGroup_proc()
BEGIN
    SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED;

    UPDATE mvCountIssueRatingsSupportGroup
    SET is_active = 0
    WHERE is_active = 1;

    INSERT INTO mvCountIssueRatingsSupportGroup (
        supportgroup_ccrn,
        critical_count,
        high_count,
        medium_count,
        low_count,
        none_count,
        is_active
    )
    SELECT
        COALESCE(SG.supportgroup_ccrn, 'UNKNOWN'),
        COUNT(DISTINCT CASE WHEN IM.issuematch_rating = 'Critical'
            THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id, ',', S.service_id)
        END),
        COUNT(DISTINCT CASE WHEN IM.issuematch_rating = 'High'
            THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id, ',', S.service_id)
        END),
        COUNT(DISTINCT CASE WHEN IM.issuematch_rating = 'Medium'
            THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id, ',', S.service_id)
        END),
        COUNT(DISTINCT CASE WHEN IM.issuematch_rating = 'Low'
            THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id, ',', S.service_id)
        END),
        COUNT(DISTINCT CASE WHEN IM.issuematch_rating = 'None'
            THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id, ',', S.service_id)
        END),
        1
    FROM Service S
    LEFT JOIN ComponentInstance CI ON S.service_id = CI.componentinstance_service_id AND CI.componentinstance_deleted_at IS NULL
    LEFT JOIN ComponentVersion CV ON CI.componentinstance_component_version_id = CV.componentversion_id AND CV.componentversion_deleted_at IS NULL
    LEFT JOIN IssueMatch IM ON CI.componentinstance_id = IM.issuematch_component_instance_id AND IM.issuematch_deleted_at IS NULL
    LEFT JOIN Issue I ON IM.issuematch_issue_id = I.issue_id AND I.issue_deleted_at IS NULL
    LEFT JOIN SupportGroupService SGS ON SGS.supportgroupservice_service_id = S.service_id AND SGS.supportgroupservice_deleted_at IS NULL
    LEFT JOIN SupportGroup SG ON SGS.supportgroupservice_support_group_id = SG.supportgroup_id AND SG.supportgroup_deleted_at IS NULL
    WHERE S.service_deleted_at IS NULL
      AND IM.issuematch_id IS NOT NULL
      AND NOT EXISTS (
          SELECT 1 FROM Remediation R
          WHERE R.remediation_service_id = S.service_id
            AND R.remediation_issue_id = I.issue_id
            AND R.remediation_deleted_at IS NULL
            AND (R.remediation_expiration_date IS NULL OR R.remediation_expiration_date >= CURDATE())
      )
    GROUP BY SG.supportgroup_ccrn
    ON DUPLICATE KEY UPDATE
        critical_count = VALUES(critical_count),
        high_count     = VALUES(high_count),
        medium_count   = VALUES(medium_count),
        low_count      = VALUES(low_count),
        none_count     = VALUES(none_count),
        is_active      = 1;

    DELETE FROM mvCountIssueRatingsSupportGroup
    WHERE is_active = 0;
END;

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsService_proc;

CREATE PROCEDURE refresh_mvCountIssueRatingsService_proc()
BEGIN
    SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED;

    UPDATE mvCountIssueRatingsService
    SET is_active = 0
    WHERE is_active = 1;

    INSERT INTO mvCountIssueRatingsService (
        supportgroup_ccrn,
        critical_count,
        high_count,
        medium_count,
        low_count,
        none_count,
        is_active
    )
    SELECT
        COALESCE(SG.supportgroup_ccrn, 'UNKNOWN'),
        COUNT(DISTINCT CASE WHEN IM.issuematch_rating = 'Critical'
            THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IM.issuematch_rating = 'High'
            THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IM.issuematch_rating = 'Medium'
            THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IM.issuematch_rating = 'Low'
            THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IM.issuematch_rating = 'None'
            THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id, ',', S.service_id) END),
        1
    FROM Service S
    LEFT JOIN ComponentInstance CI ON S.service_id = CI.componentinstance_service_id AND CI.componentinstance_deleted_at IS NULL
    LEFT JOIN ComponentVersion CV ON CI.componentinstance_component_version_id = CV.componentversion_id AND CV.componentversion_deleted_at IS NULL
    LEFT JOIN IssueMatch IM ON CI.componentinstance_id = IM.issuematch_component_instance_id AND IM.issuematch_deleted_at IS NULL
    LEFT JOIN Issue I ON IM.issuematch_issue_id = I.issue_id AND I.issue_deleted_at IS NULL
    LEFT JOIN SupportGroupService SGS ON SGS.supportgroupservice_service_id = S.service_id AND SGS.supportgroupservice_deleted_at IS NULL
    LEFT JOIN SupportGroup SG ON SGS.supportgroupservice_support_group_id = SG.supportgroup_id AND SG.supportgroup_deleted_at IS NULL
    WHERE S.service_deleted_at IS NULL
      AND NOT EXISTS (
          SELECT 1 FROM Remediation R
          WHERE R.remediation_service_id = S.service_id
            AND R.remediation_issue_id = I.issue_id
            AND R.remediation_deleted_at IS NULL
            AND (R.remediation_expiration_date IS NULL OR R.remediation_expiration_date >= CURDATE())
      )
    GROUP BY SG.supportgroup_ccrn
    ON DUPLICATE KEY UPDATE
        critical_count = VALUES(critical_count),
        high_count     = VALUES(high_count),
        medium_count   = VALUES(medium_count),
        low_count      = VALUES(low_count),
        none_count     = VALUES(none_count),
        is_active      = 1;

    DELETE FROM mvCountIssueRatingsService
    WHERE is_active = 0;
END;

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsServiceWithoutSupportGroup_proc;

CREATE PROCEDURE refresh_mvCountIssueRatingsServiceWithoutSupportGroup_proc()
BEGIN
    SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED;

    UPDATE mvCountIssueRatingsServiceWithoutSupportGroup
    SET is_active = 0
    WHERE is_active = 1;

    INSERT INTO mvCountIssueRatingsServiceWithoutSupportGroup (
        id,
        critical_count,
        high_count,
        medium_count,
        low_count,
        none_count,
        is_active
    )
    SELECT
        1,
        COUNT(DISTINCT CASE WHEN IM.issuematch_rating = 'Critical'
            THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IM.issuematch_rating = 'High'
            THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IM.issuematch_rating = 'Medium'
            THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IM.issuematch_rating = 'Low'
            THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IM.issuematch_rating = 'None'
            THEN CONCAT(CV.componentversion_component_id, ',', I.issue_id, ',', S.service_id) END),
        1
    FROM Service S
    LEFT JOIN ComponentInstance CI ON S.service_id = CI.componentinstance_service_id AND CI.componentinstance_deleted_at IS NULL
    LEFT JOIN ComponentVersion CV ON CI.componentinstance_component_version_id = CV.componentversion_id AND CV.componentversion_deleted_at IS NULL
    LEFT JOIN IssueMatch IM ON CI.componentinstance_id = IM.issuematch_component_instance_id AND IM.issuematch_deleted_at IS NULL
    LEFT JOIN Issue I ON IM.issuematch_issue_id = I.issue_id AND I.issue_deleted_at IS NULL
    WHERE S.service_deleted_at IS NULL
      AND NOT EXISTS (
          SELECT 1 FROM Remediation R
          WHERE R.remediation_service_id = S.service_id
            AND R.remediation_issue_id = I.issue_id
            AND R.remediation_deleted_at IS NULL
            AND (R.remediation_expiration_date IS NULL OR R.remediation_expiration_date >= CURDATE())
      )
    ON DUPLICATE KEY UPDATE
        critical_count = VALUES(critical_count),
        high_count     = VALUES(high_count),
        medium_count   = VALUES(medium_count),
        low_count      = VALUES(low_count),
        none_count     = VALUES(none_count),
        is_active      = 1;

    DELETE FROM mvCountIssueRatingsServiceWithoutSupportGroup
    WHERE is_active = 0;
END;

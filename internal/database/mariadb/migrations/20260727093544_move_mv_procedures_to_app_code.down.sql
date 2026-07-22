-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS post_migration_procedure_registry (name VARCHAR(128) PRIMARY KEY);

DROP PROCEDURE IF EXISTS call_registered_post_migration_procedures;

CREATE PROCEDURE call_registered_post_migration_procedures()
BEGIN
    DECLARE done INT DEFAULT 0;
    DECLARE pname VARCHAR(128);
    DECLARE cur CURSOR FOR SELECT name FROM post_migration_procedure_registry;
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = 1;

    OPEN cur;

    read_loop: LOOP
        FETCH cur INTO pname;
        IF done THEN
            LEAVE read_loop;
        END IF;

        SET @sql = CONCAT('CALL ', pname, '();');
        PREPARE stmt FROM @sql;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
    END LOOP;

    CLOSE cur;
END;

DROP PROCEDURE IF EXISTS add_post_migration_procedure;

CREATE PROCEDURE add_post_migration_procedure(IN p_name VARCHAR(128))
BEGIN
    INSERT IGNORE INTO post_migration_procedure_registry (name)
    VALUES (p_name);
END
;

DROP PROCEDURE IF EXISTS remove_post_migration_procedure;

CREATE PROCEDURE remove_post_migration_procedure(IN p_name VARCHAR(128))
BEGIN
    DELETE FROM post_migration_procedure_registry
    WHERE name = p_name;
END;

-- MV Procedures

DROP PROCEDURE IF EXISTS refresh_mvServiceIssueCounts_proc;
CREATE PROCEDURE refresh_mvServiceIssueCounts_proc()
BEGIN
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

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsUniqueService_proc;
CREATE PROCEDURE refresh_mvCountIssueRatingsUniqueService_proc()
BEGIN
    SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED;

    UPDATE mvCountIssueRatingsUniqueService
    SET is_active = 0
    WHERE is_active = 1;

    INSERT INTO mvCountIssueRatingsUniqueService (
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
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical' THEN IV.issuevariant_issue_id END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High'     THEN IV.issuevariant_issue_id END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium'   THEN IV.issuevariant_issue_id END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low'      THEN IV.issuevariant_issue_id END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None'     THEN IV.issuevariant_issue_id END),
        1
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    WHERE I.issue_deleted_at IS NULL
    ON DUPLICATE KEY UPDATE
        critical_count = VALUES(critical_count),
        high_count     = VALUES(high_count),
        medium_count   = VALUES(medium_count),
        low_count      = VALUES(low_count),
        none_count     = VALUES(none_count),
        is_active      = 1;

    DELETE FROM mvCountIssueRatingsUniqueService
    WHERE is_active = 0;
END;

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsOther_proc;
CREATE PROCEDURE refresh_mvCountIssueRatingsOther_proc()
BEGIN
    SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED;

    UPDATE mvCountIssueRatingsOther
    SET is_active = 0
    WHERE is_active = 1;

    INSERT INTO mvCountIssueRatingsOther (
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
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical' THEN IV.issuevariant_issue_id END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High'     THEN IV.issuevariant_issue_id END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium'   THEN IV.issuevariant_issue_id END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low'      THEN IV.issuevariant_issue_id END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None'     THEN IV.issuevariant_issue_id END),
        1
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    WHERE I.issue_deleted_at IS NULL
    ON DUPLICATE KEY UPDATE
        critical_count = VALUES(critical_count),
        high_count     = VALUES(high_count),
        medium_count   = VALUES(medium_count),
        low_count      = VALUES(low_count),
        none_count     = VALUES(none_count),
        is_active      = 1;

    DELETE FROM mvCountIssueRatingsOther
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
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END),
        1
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    RIGHT JOIN IssueMatch IM ON I.issue_id = IM.issuematch_issue_id
    LEFT JOIN ComponentInstance CI ON CI.componentinstance_id = IM.issuematch_component_instance_id
    LEFT JOIN Service S ON S.service_id = CI.componentinstance_service_id
    LEFT JOIN SupportGroupService SGS ON SGS.supportgroupservice_service_id = CI.componentinstance_service_id
    LEFT JOIN SupportGroup SG ON SGS.supportgroupservice_support_group_id = SG.supportgroup_id
    WHERE I.issue_deleted_at IS NULL
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
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END),
        1
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    RIGHT JOIN IssueMatch IM ON I.issue_id = IM.issuematch_issue_id
    LEFT JOIN ComponentInstance CI ON CI.componentinstance_id = IM.issuematch_component_instance_id
    LEFT JOIN Service S ON S.service_id = CI.componentinstance_service_id
    WHERE I.issue_deleted_at IS NULL
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

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsSupportGroup_proc;
CREATE PROCEDURE refresh_mvCountIssueRatingsSupportGroup_proc()
BEGIN
    SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED;

    UPDATE mvCountIssueRatingsSupportGroup
    SET is_active = 0
    WHERE is_active = 1;

    INSERT INTO mvCountIssueRatingsSupportGroup (                                                                                                                                                                           supportgroup_ccrn,                                                                                                                                                                                                  critical_count,                                                                                                                                                                                                     high_count,                                                                                                                                                                                                         medium_count,                                                                                                                                                                                                       low_count,                                                                                                                                                                                                          none_count,                                                                                                                                                                                                         is_active                                                                                                                                                                                                       )                                                                                                                                                                                                                   SELECT                                                                                                                                                                                                                  COALESCE(SG.supportgroup_ccrn, 'UNKNOWN'),                                                                                                                                                                          COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical'                                                                                                                                                            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', SGS.supportgroupservice_service_id, ',', SG.supportgroup_id)                                                                       END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', SGS.supportgroupservice_service_id, ',', SG.supportgroup_id)
        END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', SGS.supportgroupservice_service_id, ',', SG.supportgroup_id)
        END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', SGS.supportgroupservice_service_id, ',', SG.supportgroup_id)
        END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', SGS.supportgroupservice_service_id, ',', SG.supportgroup_id)
        END),
        1
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    LEFT JOIN IssueMatch IM ON I.issue_id = IM.issuematch_issue_id
    LEFT JOIN ComponentInstance CI ON CI.componentinstance_id = IM.issuematch_component_instance_id
    LEFT JOIN SupportGroupService SGS ON SGS.supportgroupservice_service_id = CI.componentinstance_service_id
    LEFT JOIN SupportGroup SG ON SGS.supportgroupservice_support_group_id = SG.supportgroup_id
    WHERE I.issue_deleted_at IS NULL
      AND CI.componentinstance_deleted_at IS NULL
      AND NOT EXISTS (
          SELECT 1 FROM Remediation R
          WHERE R.remediation_service_id = SGS.supportgroupservice_service_id
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

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsComponentVersion_proc;
CREATE PROCEDURE refresh_mvCountIssueRatingsComponentVersion_proc()
BEGIN
    SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED;
                                                                                                                                                                                                                        UPDATE mvCountIssueRatingsComponentVersion                                                                                                                                                                          SET is_active = 0                                                                                                                                                                                                   WHERE is_active = 1;
    INSERT INTO mvCountIssueRatingsComponentVersion (
        component_version_id,
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
        CVI.componentversionissue_component_version_id,
        CI.componentinstance_service_id,
        S.service_ccrn,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical'
            THEN CONCAT(CVI.componentversionissue_component_version_id, ',', CVI.componentversionissue_issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High'
            THEN CONCAT(CVI.componentversionissue_component_version_id, ',', CVI.componentversionissue_issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium'
            THEN CONCAT(CVI.componentversionissue_component_version_id, ',', CVI.componentversionissue_issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low'
            THEN CONCAT(CVI.componentversionissue_component_version_id, ',', CVI.componentversionissue_issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None'
            THEN CONCAT(CVI.componentversionissue_component_version_id, ',', CVI.componentversionissue_issue_id) END),
        1
    FROM ComponentVersionIssue CVI
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = CVI.componentversionissue_issue_id
    INNER JOIN ComponentInstance CI ON CVI.componentversionissue_component_version_id = CI.componentinstance_component_version_id
    INNER JOIN Service S ON CI.componentinstance_service_id = S.service_id
    WHERE NOT EXISTS (
          SELECT 1 FROM Remediation R
          WHERE R.remediation_service_id = CI.componentinstance_service_id
            AND R.remediation_issue_id = CVI.componentversionissue_issue_id
            AND R.remediation_deleted_at IS NULL
            AND (R.remediation_expiration_date IS NULL OR R.remediation_expiration_date >= CURDATE())
      )
    GROUP BY CVI.componentversionissue_component_version_id
    ON DUPLICATE KEY UPDATE
        service_id     = VALUES(service_id),
        service_ccrn   = VALUES(service_ccrn),
        critical_count = VALUES(critical_count),
        high_count     = VALUES(high_count),
        medium_count   = VALUES(medium_count),
        low_count      = VALUES(low_count),
        none_count     = VALUES(none_count),
        is_active      = 1;

    DELETE FROM mvCountIssueRatingsComponentVersion
    WHERE is_active = 0;
END;

DROP PROCEDURE IF EXISTS refresh_mvVulnerabilityList_proc;
CREATE PROCEDURE refresh_mvVulnerabilityList_proc()
BEGIN
    -- Ensure clean state for atomic swap
    DROP TABLE IF EXISTS mvVulnerabilityList_tmp;
    DROP TABLE IF EXISTS mvVulnerabilityList_old;

    CREATE TABLE mvVulnerabilityList_tmp LIKE mvVulnerabilityList;

    INSERT INTO mvVulnerabilityList_tmp (issue_id, max_severity, earliest_remediation_date, source_url)
    SELECT
        I.issue_id,
        MAX(IM.issuematch_rating) as max_severity,
        MIN(IM.issuematch_target_remediation_date) as earliest_remediation_date,
        (SELECT MIN(IV.issuevariant_external_url)
         FROM IssueVariant IV
         WHERE IV.issuevariant_issue_id = I.issue_id
           AND IV.issuevariant_deleted_at IS NULL
           AND IV.issuevariant_external_url != '') as source_url
    FROM Issue I
    RIGHT JOIN IssueMatch IM ON I.issue_id = IM.issuematch_issue_id
    WHERE IM.issuematch_status = 'new'
      AND IM.issuematch_deleted_at IS NULL
      AND I.issue_type = 'Vulnerability'
      AND I.issue_deleted_at IS NULL
    GROUP BY I.issue_id;

    RENAME TABLE
        mvVulnerabilityList TO mvVulnerabilityList_old,
        mvVulnerabilityList_tmp TO mvVulnerabilityList;

    DROP TABLE IF EXISTS mvVulnerabilityList_old;
END;

DROP PROCEDURE IF EXISTS refresh_mvVulnerabilityService_proc;
CREATE PROCEDURE refresh_mvVulnerabilityService_proc()
BEGIN
    DROP TABLE IF EXISTS mvVulnerabilityService_tmp;
    DROP TABLE IF EXISTS mvVulnerabilityService_old;

    CREATE TABLE mvVulnerabilityService_tmp LIKE mvVulnerabilityService;

    INSERT INTO mvVulnerabilityService_tmp (issue_id, service_id)
    SELECT DISTINCT MVL.issue_id, CI.componentinstance_service_id
    FROM mvVulnerabilityList MVL
    JOIN IssueMatch IM ON MVL.issue_id = IM.issuematch_issue_id
    JOIN ComponentInstance CI ON IM.issuematch_component_instance_id = CI.componentinstance_id
    WHERE IM.issuematch_deleted_at IS NULL
      AND CI.componentinstance_deleted_at IS NULL;

    RENAME TABLE
        mvVulnerabilityService TO mvVulnerabilityService_old,
        mvVulnerabilityService_tmp TO mvVulnerabilityService;

    DROP TABLE IF EXISTS mvVulnerabilityService_old;
END;

DROP PROCEDURE IF EXISTS refresh_mvComponentService_proc;
CREATE PROCEDURE refresh_mvComponentService_proc()
BEGIN
    -- Ensure clean state for atomic swap
    DROP TABLE IF EXISTS mvComponentService_tmp;
    DROP TABLE IF EXISTS mvComponentService_old;

    CREATE TABLE mvComponentService_tmp LIKE mvComponentService;

    INSERT INTO mvComponentService_tmp (service_id, component_id)
    SELECT DISTINCT CI.componentinstance_service_id, CV.componentversion_component_id
    FROM ComponentInstance CI
    JOIN ComponentVersion CV ON CI.componentinstance_component_version_id = CV.componentversion_id
    WHERE CI.componentinstance_deleted_at IS NULL;

    RENAME TABLE
        mvComponentService TO mvComponentService_old,
        mvComponentService_tmp TO mvComponentService;

    DROP TABLE IF EXISTS mvComponentService_old;
END;

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

DROP PROCEDURE IF EXISTS refresh_mvAllComponentsByServiceVulnerabilityCounts_proc;
CREATE PROCEDURE refreshAllComponentsByServiceVulnerabilityCounts_proc()
BEGIN
    SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED;

    UPDATE mvAllComponentsByServiceVulnerabilityCounts
    SET is_active = 0;

    INSERT INTO mvAllComponentsByServiceVulnerabilityCounts (
        service_id,
        critical_count,
        high_count,
        medium_count,
        low_count,
        none_count,
        is_active
    )
    SELECT
        CI.componentinstance_service_id,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical'
            THEN CONCAT(CV.componentversion_component_id, ',', IV.issuevariant_issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High'
            THEN CONCAT(CV.componentversion_component_id, ',', IV.issuevariant_issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium'
            THEN CONCAT(CV.componentversion_component_id, ',', IV.issuevariant_issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low'
            THEN CONCAT(CV.componentversion_component_id, ',', IV.issuevariant_issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None'
            THEN CONCAT(CV.componentversion_component_id, ',', IV.issuevariant_issue_id) END),
        1
    FROM IssueMatch IM
    JOIN ComponentInstance CI ON CI.componentinstance_id = IM.issuematch_component_instance_id
    JOIN ComponentVersion CV ON CV.componentversion_id = CI.componentinstance_component_version_id
    JOIN IssueVariant IV ON IV.issuevariant_issue_id = IM.issuematch_issue_id
    JOIN Issue I ON I.issue_id = IV.issuevariant_issue_id
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
    GROUP BY CI.componentinstance_service_id
    ON DUPLICATE KEY UPDATE
        critical_count = VALUES(critical_count),
        high_count     = VALUES(high_count),
        medium_count   = VALUES(medium_count),
        low_count      = VALUES(low_count),
        none_count     = VALUES(none_count),
        is_active      = 1;
END;

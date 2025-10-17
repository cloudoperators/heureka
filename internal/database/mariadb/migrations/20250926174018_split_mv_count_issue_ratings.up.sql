-- SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

-- Drop obsolette event calling single procedure update of 7 tables
DROP EVENT IF EXISTS refresh_mvCountIssueRatings;

-- Remove post migration procedure updating 7 tables
CALL remove_post_migration_procedure('refresh_mvCountIssueRatings_proc');

-- Drop single procedure for update of 7 tables
DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatings_proc;


-- Create 7 procedures to update particular counters
CREATE PROCEDURE refresh_mvCountIssueRatingsUniqueService_proc()
BEGIN
    CREATE TABLE IF NOT EXISTS mvCountIssueRatingsUniqueService_tmp LIKE mvCountIssueRatingsUniqueService;
    DELETE FROM mvCountIssueRatingsUniqueService_tmp;

    INSERT INTO mvCountIssueRatingsUniqueService_tmp (issue_value, issue_count)
    SELECT
        IV.issuevariant_rating AS issue_value,
        COUNT(DISTINCT IV.issuevariant_issue_id) AS issue_count
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    WHERE I.issue_deleted_at IS NULL
    GROUP BY IV.issuevariant_rating
    ORDER BY issue_id ASC;

    RENAME TABLE
        mvCountIssueRatingsUniqueService TO mvCountIssueRatingsUniqueService_old,
        mvCountIssueRatingsUniqueService_tmp TO mvCountIssueRatingsUniqueService;

    DROP TABLE mvCountIssueRatingsUniqueService_old;
END;

CREATE PROCEDURE refresh_mvCountIssueRatingsService_proc()
BEGIN
    CREATE TABLE IF NOT EXISTS mvCountIssueRatingsService_tmp LIKE mvCountIssueRatingsService;
    DELETE FROM mvCountIssueRatingsService_tmp;

    INSERT INTO mvCountIssueRatingsService_tmp (supportgroup_ccrn, issue_value, issue_count)
    SELECT
        COALESCE(SG.supportgroup_ccrn, 'UNKNOWN') AS supportgroup_ccrn,
        IV.issuevariant_rating AS issue_value,
        COUNT(DISTINCT CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id)) AS issue_count
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    RIGHT JOIN IssueMatch IM ON I.issue_id = IM.issuematch_issue_id
    LEFT JOIN ComponentInstance CI ON CI.componentinstance_id = IM.issuematch_component_instance_id
    LEFT JOIN ComponentVersion CV ON CI.componentinstance_component_version_id = CV.componentversion_id
    LEFT JOIN Service S ON S.service_id = CI.componentinstance_service_id
    LEFT JOIN SupportGroupService SGS ON SGS.supportgroupservice_service_id = CI.componentinstance_service_id
    LEFT JOIN SupportGroup SG ON SGS.supportgroupservice_support_group_id = SG.supportgroup_id
    WHERE I.issue_deleted_at IS NULL
    GROUP BY SG.supportgroup_ccrn, IV.issuevariant_rating
    ORDER BY issue_id ASC;

    RENAME TABLE
        mvCountIssueRatingsService TO mvCountIssueRatingsService_old,
        mvCountIssueRatingsService_tmp TO mvCountIssueRatingsService;

    DROP TABLE mvCountIssueRatingsService_old;
END;

CREATE PROCEDURE refresh_mvCountIssueRatingsServiceWithoutSupportGroup_proc()
BEGIN
    CREATE TABLE IF NOT EXISTS mvCountIssueRatingsServiceWithoutSupportGroup_tmp LIKE mvCountIssueRatingsServiceWithoutSupportGroup;
    DELETE FROM mvCountIssueRatingsServiceWithoutSupportGroup_tmp;

    INSERT INTO mvCountIssueRatingsServiceWithoutSupportGroup_tmp (issue_value, issue_count)
    SELECT
        IV.issuevariant_rating AS issue_value,
        COUNT(DISTINCT CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id)) AS issue_count
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    RIGHT JOIN IssueMatch IM ON I.issue_id = IM.issuematch_issue_id
    LEFT JOIN ComponentInstance CI ON CI.componentinstance_id = IM.issuematch_component_instance_id
    LEFT JOIN ComponentVersion CV ON CI.componentinstance_component_version_id = CV.componentversion_id
    LEFT JOIN Service S ON S.service_id = CI.componentinstance_service_id
    WHERE I.issue_deleted_at IS NULL
    GROUP BY IV.issuevariant_rating
    ORDER BY issue_id ASC;

    RENAME TABLE
        mvCountIssueRatingsServiceWithoutSupportGroup TO mvCountIssueRatingsServiceWithoutSupportGroup_old,
        mvCountIssueRatingsServiceWithoutSupportGroup_tmp TO mvCountIssueRatingsServiceWithoutSupportGroup;

    DROP TABLE mvCountIssueRatingsServiceWithoutSupportGroup_old;
END;

CREATE PROCEDURE refresh_mvCountIssueRatingsSupportGroup_proc()
BEGIN
    CREATE TABLE IF NOT EXISTS mvCountIssueRatingsSupportGroup_tmp LIKE mvCountIssueRatingsSupportGroup;
    DELETE FROM mvCountIssueRatingsSupportGroup_tmp;

    INSERT INTO mvCountIssueRatingsSupportGroup_tmp (supportgroup_ccrn, issue_value, issue_count)
    SELECT
        COALESCE(SG.supportgroup_ccrn, 'UNKNOWN') AS supportgroup_ccrn,
        IV.issuevariant_rating AS issue_value,
        COUNT(DISTINCT CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', SGS.supportgroupservice_service_id, ',', SG.supportgroup_id)) AS issue_count
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    LEFT JOIN IssueMatch IM ON I.issue_id = IM.issuematch_issue_id
    LEFT JOIN ComponentInstance CI ON CI.componentinstance_id = IM.issuematch_component_instance_id
    LEFT JOIN SupportGroupService SGS ON SGS.supportgroupservice_service_id = CI.componentinstance_service_id
    LEFT JOIN SupportGroup SG ON SGS.supportgroupservice_support_group_id = SG.supportgroup_id
    WHERE I.issue_deleted_at IS NULL
    GROUP BY SG.supportgroup_ccrn, IV.issuevariant_rating
    ORDER BY issue_id ASC;

    RENAME TABLE
        mvCountIssueRatingsSupportGroup TO mvCountIssueRatingsSupportGroup_old,
        mvCountIssueRatingsSupportGroup_tmp TO mvCountIssueRatingsSupportGroup;

    DROP TABLE mvCountIssueRatingsSupportGroup_old;
END;

CREATE PROCEDURE refresh_mvCountIssueRatingsComponentVersion_proc()
BEGIN
    CREATE TABLE IF NOT EXISTS mvCountIssueRatingsComponentVersion_tmp LIKE mvCountIssueRatingsComponentVersion;
    DELETE FROM mvCountIssueRatingsComponentVersion_tmp;

    INSERT INTO mvCountIssueRatingsComponentVersion_tmp (component_version_id, issue_value, issue_count)
    SELECT
        CVI.componentversionissue_component_version_id AS component_version_id,
        IV.issuevariant_rating AS issue_value,
        COUNT(DISTINCT CONCAT(CVI.componentversionissue_component_version_id, ',', CVI.componentversionissue_issue_id)) AS issue_count
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    LEFT JOIN ComponentVersionIssue CVI ON I.issue_id = CVI.componentversionissue_issue_id
    WHERE I.issue_deleted_at IS NULL
    GROUP BY CVI.componentversionissue_component_version_id, IV.issuevariant_rating
    ORDER BY issue_id ASC;

    RENAME TABLE
        mvCountIssueRatingsComponentVersion TO mvCountIssueRatingsComponentVersion_old,
        mvCountIssueRatingsComponentVersion_tmp TO mvCountIssueRatingsComponentVersion;

    DROP TABLE mvCountIssueRatingsComponentVersion_old;
END;

CREATE PROCEDURE refresh_mvCountIssueRatingsServiceId_proc()
BEGIN
    CREATE TABLE IF NOT EXISTS mvCountIssueRatingsServiceId_tmp LIKE mvCountIssueRatingsServiceId;
    DELETE FROM mvCountIssueRatingsServiceId_tmp;

    INSERT INTO mvCountIssueRatingsServiceId_tmp (service_id, issue_value, issue_count)
    SELECT
        CI.componentinstance_service_id AS service_id,
        IV.issuevariant_rating AS issue_value,
        COUNT(DISTINCT CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id)) AS issue_count
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    LEFT JOIN IssueMatch IM ON I.issue_id = IM.issuematch_issue_id
    LEFT JOIN ComponentInstance CI ON CI.componentinstance_id = IM.issuematch_component_instance_id
    WHERE I.issue_deleted_at IS NULL
    GROUP BY CI.componentinstance_service_id, IV.issuevariant_rating
    ORDER BY issue_id ASC;

    RENAME TABLE
        mvCountIssueRatingsServiceId TO mvCountIssueRatingsServiceId_old,
        mvCountIssueRatingsServiceId_tmp TO mvCountIssueRatingsServiceId;

    DROP TABLE mvCountIssueRatingsServiceId_old;
END;

CREATE PROCEDURE refresh_mvCountIssueRatingsOther_proc()
BEGIN
    CREATE TABLE IF NOT EXISTS mvCountIssueRatingsOther_tmp LIKE mvCountIssueRatingsOther;
    DELETE FROM mvCountIssueRatingsOther_tmp;

    INSERT INTO mvCountIssueRatingsOther_tmp (issue_value, issue_count)
    SELECT
        IV.issuevariant_rating AS issue_value,
        COUNT(DISTINCT IV.issuevariant_issue_id) AS issue_count
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    WHERE I.issue_deleted_at IS NULL
    GROUP BY IV.issuevariant_rating
    ORDER BY issue_id ASC;

    RENAME TABLE
        mvCountIssueRatingsOther TO mvCountIssueRatingsOther_old,
        mvCountIssueRatingsOther_tmp TO mvCountIssueRatingsOther;

    DROP TABLE mvCountIssueRatingsOther_old;
END;

-- Create new event to update counters using split procedures
CREATE EVENT refresh_mvCountIssueRatingsUniqueService
ON SCHEDULE EVERY 1 HOUR
DO
  CALL refresh_mvCountIssueRatingsUniqueService_proc();

CREATE EVENT refresh_mvCountIssueRatingsService
ON SCHEDULE EVERY 1 HOUR
DO
  CALL refresh_mvCountIssueRatingsService_proc();

CREATE EVENT refresh_mvCountIssueRatingsServiceWithoutSupportGroup
ON SCHEDULE EVERY 1 HOUR
DO
  CALL refresh_mvCountIssueRatingsServiceWithoutSupportGroup_proc();

CREATE EVENT refresh_mvCountIssueRatingsSupportGroup
ON SCHEDULE EVERY 1 HOUR
DO
  CALL refresh_mvCountIssueRatingsSupportGroup_proc();

CREATE EVENT refresh_mvCountIssueRatingsComponentVersion
ON SCHEDULE EVERY 1 HOUR
DO
  CALL refresh_mvCountIssueRatingsComponentVersion_proc();

CREATE EVENT refresh_mvCountIssueRatingsServiceId
ON SCHEDULE EVERY 1 HOUR
DO
  CALL refresh_mvCountIssueRatingsServiceId_proc();

CREATE EVENT refresh_mvCountIssueRatingsOther
ON SCHEDULE EVERY 1 HOUR
DO
  CALL refresh_mvCountIssueRatingsOther_proc();

-- Add post migration procedures for all 7 mv tables updating particular counters
CALL add_post_migration_procedure('refresh_mvCountIssueRatingsUniqueService_proc');
CALL add_post_migration_procedure('refresh_mvCountIssueRatingsService_proc');
CALL add_post_migration_procedure('refresh_mvCountIssueRatingsServiceWithoutSupportGroup_proc');
CALL add_post_migration_procedure('refresh_mvCountIssueRatingsSupportGroup_proc');
CALL add_post_migration_procedure('refresh_mvCountIssueRatingsComponentVersion_proc');
CALL add_post_migration_procedure('refresh_mvCountIssueRatingsServiceId_proc');
CALL add_post_migration_procedure('refresh_mvCountIssueRatingsOther_proc');

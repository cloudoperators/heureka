-- SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

--
-- SET GLOBAL event_scheduler = ON;
--

-- 1. Create the table (only if not exists to avoid errors if re-run)
CREATE TABLE IF NOT EXISTS mvCountIssueRatingsUniqueService (
    issue_value enum ('None','Low','Medium', 'High', 'Critical') NULL,
    issue_count INT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS mvCountIssueRatingsService (
    supportgroup_ccrn varchar(255) NULL,
    issue_value enum ('None','Low','Medium', 'High', 'Critical') NULL,
    issue_count INT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS mvCountIssueRatingsServiceWithoutSupportGroup (
    issue_value enum ('None','Low','Medium', 'High', 'Critical') NULL,
    issue_count INT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS mvCountIssueRatingsSupportGroup (
    supportgroup_ccrn varchar(255) NULL,
    issue_value enum ('None','Low','Medium', 'High', 'Critical') NULL,
    issue_count INT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS mvCountIssueRatingsComponentVersion (
    component_version_id INT UNSIGNED NULL,
    issue_value enum ('None','Low','Medium', 'High', 'Critical') NULL,
    issue_count INT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS mvCountIssueRatingsServiceId (
    service_id INT NULL,
    issue_value enum ('None','Low','Medium', 'High', 'Critical') NULL,
    issue_count INT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS mvCountIssueRatingsOther (
    issue_value enum ('None','Low','Medium', 'High', 'Critical') NULL,
    issue_count INT DEFAULT 0
);

-- 2. Create or replace the procedure that refreshes the table
CREATE PROCEDURE refresh_mvCountIssueRatings_proc()
BEGIN
    TRUNCATE TABLE mvCountIssueRatingsUniqueService;
    INSERT INTO mvCountIssueRatingsUniqueService (issue_value, issue_count)
    SELECT
        IV.issuevariant_rating AS issue_value,
        COUNT(DISTINCT IV.issuevariant_issue_id) AS issue_count
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    WHERE I.issue_deleted_at IS NULL
    GROUP BY IV.issuevariant_rating ORDER BY issue_id ASC;

    TRUNCATE TABLE mvCountIssueRatingsService;
    INSERT INTO mvCountIssueRatingsService (supportgroup_ccrn, issue_value, issue_count)
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
    GROUP BY SG.supportgroup_ccrn, IV.issuevariant_rating ORDER BY issue_id ASC;

    TRUNCATE TABLE mvCountIssueRatingsServiceWithoutSupportGroup;
    INSERT INTO mvCountIssueRatingsServiceWithoutSupportGroup (issue_value, issue_count)
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
    GROUP BY IV.issuevariant_rating ORDER BY issue_id ASC;

    TRUNCATE TABLE mvCountIssueRatingsSupportGroup;
    INSERT INTO mvCountIssueRatingsSupportGroup (supportgroup_ccrn, issue_value, issue_count)
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
    GROUP BY SG.supportgroup_ccrn, IV.issuevariant_rating ORDER BY issue_id ASC;

    TRUNCATE TABLE mvCountIssueRatingsComponentVersion;
    INSERT INTO mvCountIssueRatingsComponentVersion (component_version_id, issue_value, issue_count)
    SELECT
        CVI.componentversionissue_component_version_id as component_version_id,
        IV.issuevariant_rating AS issue_value,
        COUNT(DISTINCT CONCAT(CVI.componentversionissue_component_version_id, ',', CVI.componentversionissue_issue_id)) AS issue_count
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    LEFT JOIN ComponentVersionIssue CVI ON I.issue_id = CVI.componentversionissue_issue_id
    WHERE I.issue_deleted_at IS NULL
    GROUP BY CVI.componentversionissue_component_version_id, IV.issuevariant_rating ORDER BY issue_id ASC;

    TRUNCATE TABLE mvCountIssueRatingsServiceId;
    INSERT INTO mvCountIssueRatingsServiceId (service_id, issue_value, issue_count)
    SELECT
        CI.componentinstance_service_id AS service_id,
        IV.issuevariant_rating AS issue_value,
        COUNT(DISTINCT CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id)) AS issue_count
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    LEFT JOIN IssueMatch IM ON I.issue_id = IM.issuematch_issue_id
    LEFT JOIN ComponentInstance CI ON CI.componentinstance_id = IM.issuematch_component_instance_id
    WHERE I.issue_deleted_at IS NULL
    GROUP BY CI.componentinstance_service_id, IV.issuevariant_rating ORDER BY issue_id ASC;

    TRUNCATE TABLE mvCountIssueRatingsOther;
    INSERT INTO mvCountIssueRatingsOther (issue_value, issue_count)
    SELECT
        IV.issuevariant_rating AS issue_value,
        COUNT(DISTINCT IV.issuevariant_issue_id) AS issue_count
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    WHERE I.issue_deleted_at IS NULL
    GROUP BY IV.issuevariant_rating ORDER BY issue_id ASC;
END;

-- 3. Run the procedure once to populate the table initially
-- CALL refresh_mvCountIssueRatings_proc();

-- 4. Create the event to run the procedure every hour
CREATE EVENT IF NOT EXISTS refresh_mvCountIssueRatings
ON SCHEDULE EVERY 1 HOUR
DO
  CALL refresh_mvCountIssueRatings_proc();


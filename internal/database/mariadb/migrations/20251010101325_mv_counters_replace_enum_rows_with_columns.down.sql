-- SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

DROP TABLE IF EXISTS mvCountIssueRatingsUniqueService_new;
CREATE TABLE mvCountIssueRatingsUniqueService_new (
    issue_value ENUM('None','Low','Medium','High','Critical') NULL,
    issue_count INT DEFAULT 0
);

INSERT INTO mvCountIssueRatingsUniqueService_new (issue_value, issue_count)
SELECT 'Critical', critical_count FROM mvCountIssueRatingsUniqueService
UNION ALL
SELECT 'High', high_count FROM mvCountIssueRatingsUniqueService
UNION ALL
SELECT 'Medium', medium_count FROM mvCountIssueRatingsUniqueService
UNION ALL
SELECT 'Low', low_count FROM mvCountIssueRatingsUniqueService
UNION ALL
SELECT 'None', none_count FROM mvCountIssueRatingsUniqueService;

DROP TABLE mvCountIssueRatingsUniqueService;
RENAME TABLE mvCountIssueRatingsUniqueService_new TO mvCountIssueRatingsUniqueService;

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsUniqueService_proc;
CREATE PROCEDURE refresh_mvCountIssueRatingsUniqueService_proc()
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
END;

--

DROP TABLE IF EXISTS mvCountIssueRatingsService_new;
CREATE TABLE mvCountIssueRatingsService_new (
    issue_value ENUM('None','Low','Medium','High','Critical') NULL,
    issue_count INT DEFAULT 0
);

INSERT INTO mvCountIssueRatingsService_new (issue_value, issue_count)
SELECT 'Critical', critical_count FROM mvCountIssueRatingsService
UNION ALL
SELECT 'High', high_count FROM mvCountIssueRatingsService
UNION ALL
SELECT 'Medium', medium_count FROM mvCountIssueRatingsService
UNION ALL
SELECT 'Low', low_count FROM mvCountIssueRatingsService
UNION ALL
SELECT 'None', none_count FROM mvCountIssueRatingsService;

DROP TABLE mvCountIssueRatingsService;
RENAME TABLE mvCountIssueRatingsService_new TO mvCountIssueRatingsService;

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsService_proc;
CREATE PROCEDURE refresh_mvCountIssueRatingsService_proc()
BEGIN
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
END;

--

DROP TABLE IF EXISTS mvCountIssueRatingsServiceWithoutSupportGroup_new;
CREATE TABLE mvCountIssueRatingsServiceWithoutSupportGroup_new (
    issue_value ENUM('None','Low','Medium','High','Critical') NULL,
    issue_count INT DEFAULT 0
);

INSERT INTO mvCountIssueRatingsServiceWithoutSupportGroup_new (issue_value, issue_count)
SELECT 'Critical', critical_count FROM mvCountIssueRatingsServiceWithoutSupportGroup
UNION ALL
SELECT 'High', high_count FROM mvCountIssueRatingsServiceWithoutSupportGroup
UNION ALL
SELECT 'Medium', medium_count FROM mvCountIssueRatingsServiceWithoutSupportGroup
UNION ALL
SELECT 'Low', low_count FROM mvCountIssueRatingsServiceWithoutSupportGroup
UNION ALL
SELECT 'None', none_count FROM mvCountIssueRatingsServiceWithoutSupportGroup;

DROP TABLE mvCountIssueRatingsServiceWithoutSupportGroup;
RENAME TABLE mvCountIssueRatingsServiceWithoutSupportGroup_new TO mvCountIssueRatingsServiceWithoutSupportGroup;

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsServiceWithoutSupportGroup_proc;
CREATE PROCEDURE refresh_mvCountIssueRatingsServiceWithoutSupportGroup_proc()
BEGIN
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
END;

--

DROP TABLE IF EXISTS mvCountIssueRatingsSupportGroup_new;
CREATE TABLE mvCountIssueRatingsSupportGroup_new (
    issue_value ENUM('None','Low','Medium','High','Critical') NULL,
    issue_count INT DEFAULT 0
);

INSERT INTO mvCountIssueRatingsSupportGroup_new (issue_value, issue_count)
SELECT 'Critical', critical_count FROM mvCountIssueRatingsSupportGroup
UNION ALL
SELECT 'High', high_count FROM mvCountIssueRatingsSupportGroup
UNION ALL
SELECT 'Medium', medium_count FROM mvCountIssueRatingsSupportGroup
UNION ALL
SELECT 'Low', low_count FROM mvCountIssueRatingsSupportGroup
UNION ALL
SELECT 'None', none_count FROM mvCountIssueRatingsSupportGroup;

DROP TABLE mvCountIssueRatingsSupportGroup;
RENAME TABLE mvCountIssueRatingsSupportGroup_new TO mvCountIssueRatingsSupportGroup;

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsSupportGroup_proc;
CREATE PROCEDURE refresh_mvCountIssueRatingsSupportGroup_proc()
BEGIN
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
END;

--

DROP TABLE IF EXISTS mvCountIssueRatingsComponentVersion_new;
CREATE TABLE mvCountIssueRatingsComponentVersion_new (
    issue_value ENUM('None','Low','Medium','High','Critical') NULL,
    issue_count INT DEFAULT 0
);

INSERT INTO mvCountIssueRatingsComponentVersion_new (issue_value, issue_count)
SELECT 'Critical', critical_count FROM mvCountIssueRatingsComponentVersion
UNION ALL
SELECT 'High', high_count FROM mvCountIssueRatingsComponentVersion
UNION ALL
SELECT 'Medium', medium_count FROM mvCountIssueRatingsComponentVersion
UNION ALL
SELECT 'Low', low_count FROM mvCountIssueRatingsComponentVersion
UNION ALL
SELECT 'None', none_count FROM mvCountIssueRatingsComponentVersion;

DROP TABLE mvCountIssueRatingsComponentVersion;
RENAME TABLE mvCountIssueRatingsComponentVersion_new TO mvCountIssueRatingsComponentVersion;

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsComponentVersion_proc;
CREATE PROCEDURE refresh_mvCountIssueRatingsComponentVersion_proc()
BEGIN
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
END;

--

DROP TABLE IF EXISTS mvCountIssueRatingsServiceId_new;
CREATE TABLE mvCountIssueRatingsServiceId_new (
    issue_value ENUM('None','Low','Medium','High','Critical') NULL,
    issue_count INT DEFAULT 0
);

INSERT INTO mvCountIssueRatingsServiceId_new (issue_value, issue_count)
SELECT 'Critical', critical_count FROM mvCountIssueRatingsServiceId
UNION ALL
SELECT 'High', high_count FROM mvCountIssueRatingsServiceId
UNION ALL
SELECT 'Medium', medium_count FROM mvCountIssueRatingsServiceId
UNION ALL
SELECT 'Low', low_count FROM mvCountIssueRatingsServiceId
UNION ALL
SELECT 'None', none_count FROM mvCountIssueRatingsServiceId;

DROP TABLE mvCountIssueRatingsServiceId;
RENAME TABLE mvCountIssueRatingsServiceId_new TO mvCountIssueRatingsServiceId;

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsServiceId_proc;
CREATE PROCEDURE refresh_mvCountIssueRatingsServiceId_proc()
BEGIN
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
END;

--

DROP TABLE IF EXISTS mvCountIssueRatingsOther_new;
CREATE TABLE mvCountIssueRatingsOther_new (
    issue_value ENUM('None','Low','Medium','High','Critical') NULL,
    issue_count INT DEFAULT 0
);

INSERT INTO mvCountIssueRatingsOther_new (issue_value, issue_count)
SELECT 'Critical', critical_count FROM mvCountIssueRatingsOther
UNION ALL
SELECT 'High', high_count FROM mvCountIssueRatingsOther
UNION ALL
SELECT 'Medium', medium_count FROM mvCountIssueRatingsOther
UNION ALL
SELECT 'Low', low_count FROM mvCountIssueRatingsOther
UNION ALL
SELECT 'None', none_count FROM mvCountIssueRatingsOther;

DROP TABLE mvCountIssueRatingsOther;
RENAME TABLE mvCountIssueRatingsOther_new TO mvCountIssueRatingsOther;

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsOther_proc;
CREATE PROCEDURE refresh_mvCountIssueRatingsOther_proc()
BEGIN
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

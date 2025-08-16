-- SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

--
-- SET GLOBAL event_scheduler = ON;
--

-- 1. Create the table (only if not exists to avoid errors if re-run)
CREATE TABLE IF NOT EXISTS mvCountIssueRatings1 (
    issue_value enum ('None','Low','Medium', 'High', 'Critical') NOT NULL,
    issue_count INT DEFAULT 0,
    PRIMARY KEY (issue_value)
);

CREATE TABLE IF NOT EXISTS mvCountIssueRatings2 (
    supportgroup_ccrn varchar(255) NOT NULL,
    issue_value enum ('None','Low','Medium', 'High', 'Critical') NOT NULL,
    issue_count INT DEFAULT 0,
    PRIMARY KEY (supportgroup_ccrn, issue_value)
);

CREATE TABLE IF NOT EXISTS mvCountIssueRatings2a (
    issue_value enum ('None','Low','Medium', 'High', 'Critical') NOT NULL,
    issue_count INT DEFAULT 0,
    PRIMARY KEY (issue_value)
);

CREATE TABLE IF NOT EXISTS mvCountIssueRatings3 (
    supportgroup_ccrn varchar(255) NOT NULL,
    issue_value enum ('None','Low','Medium', 'High', 'Critical') NOT NULL,
    issue_count INT DEFAULT 0,
    PRIMARY KEY (supportgroup_ccrn, issue_value)
);

CREATE TABLE IF NOT EXISTS mvCountIssueRatings4 (
    component_version_id INT UNSIGNED NOT NULL,
    issue_value enum ('None','Low','Medium', 'High', 'Critical') NOT NULL,
    issue_count INT DEFAULT 0,
    PRIMARY KEY (component_version_id, issue_value)
);

CREATE TABLE IF NOT EXISTS mvCountIssueRatings5 (
    service_id INT NOT NULL,
    issue_value enum ('None','Low','Medium', 'High', 'Critical') NOT NULL,
    issue_count INT DEFAULT 0,
    PRIMARY KEY (service_id, issue_value)
);

CREATE TABLE IF NOT EXISTS mvCountIssueRatings6 (
    issue_value enum ('None','Low','Medium', 'High', 'Critical') NOT NULL,
    issue_count INT DEFAULT 0,
    PRIMARY KEY (issue_value)
);

-- 2. Create or replace the procedure that refreshes the table
CREATE PROCEDURE refresh_mvCountIssueRatings_proc()
BEGIN
    TRUNCATE TABLE mvCountIssueRatings1;
    INSERT INTO mvCountIssueRatings1 (issue_value, issue_count)
    SELECT
        IV.issuevariant_rating AS issue_value,
        COUNT(DISTINCT IV.issuevariant_issue_id) AS issue_count
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    WHERE I.issue_deleted_at IS NULL
    GROUP BY IV.issuevariant_rating ORDER BY issue_id ASC;

    TRUNCATE TABLE mvCountIssueRatings2;
    INSERT INTO mvCountIssueRatings2 (supportgroup_ccrn, issue_value, issue_count)
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

    TRUNCATE TABLE mvCountIssueRatings2a;
    INSERT INTO mvCountIssueRatings2a (issue_value, issue_count)
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

    TRUNCATE TABLE mvCountIssueRatings3;
    INSERT INTO mvCountIssueRatings3 (supportgroup_ccrn, issue_value, issue_count)
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

    TRUNCATE TABLE mvCountIssueRatings4;
    INSERT INTO mvCountIssueRatings4 (component_version_id, issue_value, issue_count)
    SELECT
        CVI.componentversionissue_component_version_id as supportgroup_ccrn,
        IV.issuevariant_rating AS issue_value,
        COUNT(DISTINCT CONCAT(CVI.componentversionissue_component_version_id, ',', CVI.componentversionissue_issue_id)) AS issue_count
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    LEFT JOIN ComponentVersionIssue CVI ON I.issue_id = CVI.componentversionissue_issue_id
    WHERE I.issue_deleted_at IS NULL
    GROUP BY CVI.componentversionissue_component_version_id, IV.issuevariant_rating ORDER BY issue_id ASC;

    TRUNCATE TABLE mvCountIssueRatings5;
    INSERT INTO mvCountIssueRatings5 (service_id, issue_value, issue_count)
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

    TRUNCATE TABLE mvCountIssueRatings6;
    INSERT INTO mvCountIssueRatings6 (issue_value, issue_count)
    SELECT
        IV.issuevariant_rating AS issue_value,
        COUNT(DISTINCT IV.issuevariant_issue_id) AS issue_count
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    WHERE I.issue_deleted_at IS NULL
    GROUP BY IV.issuevariant_rating ORDER BY issue_id ASC;
END;

-- 3. Run the procedure once to populate the table initially
CALL refresh_mvCountIssueRatings_proc();

-- 4. Create the event to run the procedure every hour
CREATE EVENT IF NOT EXISTS refresh_mvCountIssueRatings
ON SCHEDULE EVERY 1 HOUR
DO
  CALL refresh_mvCountIssueRatings_proc();


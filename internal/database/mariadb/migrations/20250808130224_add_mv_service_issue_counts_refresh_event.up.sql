-- SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

--
-- SET GLOBAL event_scheduler = ON;
--

-- 1. Create the table (only if not exists to avoid errors if re-run)
CREATE TABLE IF NOT EXISTS mvServiceIssueCounts (
    service_id INT NOT NULL PRIMARY KEY,
    critical_count INT DEFAULT 0,
    high_count INT DEFAULT 0,
    medium_count INT DEFAULT 0,
    low_count INT DEFAULT 0,
    none_count INT DEFAULT 0
);

-- 2. Create or replace the procedure that refreshes the table
CREATE PROCEDURE refresh_mvServiceIssueCounts_proc()
BEGIN
    TRUNCATE TABLE mvServiceIssueCounts;

    INSERT INTO mvServiceIssueCounts
    SELECT
        S.service_id,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical'
            THEN CONCAT(CV.componentversion_id, ',', IV.issuevariant_issue_id) END) AS critical_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High'
            THEN CONCAT(CV.componentversion_id, ',', IV.issuevariant_issue_id) END) AS high_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium'
            THEN CONCAT(CV.componentversion_id, ',', IV.issuevariant_issue_id) END) AS medium_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low'
            THEN CONCAT(CV.componentversion_id, ',', IV.issuevariant_issue_id) END) AS low_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None'
            THEN CONCAT(CV.componentversion_id, ',', IV.issuevariant_issue_id) END) AS none_count
    FROM Service S
    LEFT JOIN ComponentInstance CI ON S.service_id = CI.componentinstance_service_id
    LEFT JOIN ComponentVersion CV ON CV.componentversion_id = CI.componentinstance_component_version_id
    LEFT JOIN ComponentVersionIssue CVI ON CV.componentversion_id = CVI.componentversionissue_component_version_id
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = CVI.componentversionissue_issue_id
    WHERE S.service_deleted_at IS NULL
    GROUP BY S.service_id;
END;

-- 3. Run the procedure once to populate the table initially
-- CALL refresh_mvServiceIssueCounts_proc();

-- 4. Create the event to run the procedure every hour
CREATE EVENT IF NOT EXISTS refresh_mvServiceIssueCounts
ON SCHEDULE EVERY 1 HOUR
DO
  CALL refresh_mvServiceIssueCounts_proc();


-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS mvCountIssueRatingsRegion (
    region varchar(1024) NULL,
    critical_count INT DEFAULT 0,
    high_count INT DEFAULT 0,
    medium_count INT DEFAULT 0,
    low_count INT DEFAULT 0,
    none_count INT DEFAULT 0,
    is_active TINYINT(1) NOT NULL DEFAULT 1
);

CREATE PROCEDURE IF NOT EXISTS refresh_mvCountIssueRatingsRegion_proc()
BEGIN
    UPDATE mvCountIssueRatingsRegion
    SET is_active = 0
    WHERE is_active = 1;

    INSERT INTO mvCountIssueRatingsRegion
        (region, critical_count, high_count, medium_count, low_count, none_count, is_active)
    SELECT
        DISTINCT CI.componentinstance_region,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical' THEN IV.issuevariant_issue_id END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High'     THEN IV.issuevariant_issue_id END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium'   THEN IV.issuevariant_issue_id END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low'      THEN IV.issuevariant_issue_id END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None'     THEN IV.issuevariant_issue_id END),
        1
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    RIGHT JOIN IssueMatch IM ON I.issue_id = IM.issuematch_issue_id
    LEFT JOIN ComponentInstance CI ON CI.componentinstance_id = IM.issuematch_component_instance_id
    WHERE I.issue_deleted_at IS NULL
    GROUP BY CI.componentinstance_region
    ON DUPLICATE KEY UPDATE
        critical_count = VALUES(critical_count),
        high_count     = VALUES(high_count),
        medium_count   = VALUES(medium_count),
        low_count      = VALUES(low_count),
        none_count     = VALUES(none_count),
        is_active      = 1;

    DELETE FROM mvCountIssueRatingsRegion
    WHERE is_active = 0;
END;

CREATE EVENT IF NOT EXISTS refresh_mvCountIssueRatingsRegion
ON SCHEDULE EVERY 2 HOUR
DO
  CALL refresh_mvCountIssueRatingsRegion();
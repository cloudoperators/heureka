-- SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

DROP TABLE IF EXISTS post_migration_procedure_registry;
DROP PROCEDURE IF EXISTS call_registered_post_migration_procedures;
DROP PROCEDURE IF EXISTS add_post_migration_procedure;
DROP PROCEDURE IF EXISTS remove_post_migration_procedure;

CREATE TABLE IF NOT EXISTS post_migration_procedure_registry (
    caller  VARCHAR(128) NOT NULL PRIMARY KEY,
    checker VARCHAR(128) NOT NULL
);

CREATE PROCEDURE add_post_migration_procedure(IN p_caller VARCHAR(128), IN p_checker VARCHAR(128))
BEGIN
    INSERT IGNORE INTO post_migration_procedure_registry (caller, checker)
    VALUES (p_caller, p_checker);
END;

CREATE PROCEDURE remove_post_migration_procedure(IN p_caller VARCHAR(128))
BEGIN
    DELETE FROM post_migration_procedure_registry
    WHERE caller = p_caller;
END;


CREATE PROCEDURE has_mvServiceIssueCounts_proc(OUT has_records BOOLEAN)
BEGIN
    DECLARE cnt INT;

    SELECT COUNT(*) INTO cnt
    FROM mvServiceIssueCounts;

    SET has_records = (cnt > 0);
END;

CREATE PROCEDURE has_mvCountIssueRatings_proc(OUT has_records BOOLEAN)
BEGIN
    DECLARE cntMvCountIssueRatingsUniqueService,
            cntMvCountIssueRatingsService,
            cntMvCountIssueRatingsServiceWithoutSupportGroup,
            cntMvCountIssueRatingsSupportGroup,
            cntMvCountIssueRatingsComponentVersion,
            cntMvCountIssueRatingsServiceId,
            cntMvCountIssueRatingsOther
    INT;

    -- Check each table row count
    SELECT COUNT(*) INTO cntMvCountIssueRatingsUniqueService FROM mvCountIssueRatingsUniqueService;
    SELECT COUNT(*) INTO cntMvCountIssueRatingsService FROM mvCountIssueRatingsService;
    SELECT COUNT(*) INTO cntMvCountIssueRatingsServiceWithoutSupportGroup FROM mvCountIssueRatingsServiceWithoutSupportGroup;
    SELECT COUNT(*) INTO cntMvCountIssueRatingsSupportGroup FROM mvCountIssueRatingsSupportGroup;
    SELECT COUNT(*) INTO cntMvCountIssueRatingsComponentVersion FROM mvCountIssueRatingsComponentVersion;
    SELECT COUNT(*) INTO cntMvCountIssueRatingsServiceId FROM mvCountIssueRatingsServiceId;
    SELECT COUNT(*) INTO cntMvCountIssueRatingsOther FROM mvCountIssueRatingsOther;

    -- Set has_records = TRUE only if all > 0
    SET has_records = (cntMvCountIssueRatingsUniqueService > 0 AND
                       cntMvCountIssueRatingsService > 0 AND
                       cntMvCountIssueRatingsServiceWithoutSupportGroup > 0 AND
                       cntMvCountIssueRatingsSupportGroup > 0 AND
                       cntMvCountIssueRatingsComponentVersion > 0 AND
                       cntMvCountIssueRatingsServiceId > 0 AND
                       cntMvCountIssueRatingsOther > 0);
END;

CREATE PROCEDURE has_mvSingleComponentByServiceVulnerabilityCounts_proc(OUT has_records BOOLEAN)
BEGIN
    DECLARE cnt INT;

    SELECT COUNT(*) INTO cnt
    FROM mvSingleComponentByServiceVulnerabilityCounts;

    SET has_records = (cnt > 0);
END;

CREATE PROCEDURE has_mvAllComponentsByServiceVulnerabilityCounts_proc(OUT has_records BOOLEAN)
BEGIN
    DECLARE cnt INT;

    SELECT COUNT(*) INTO cnt
    FROM mvSingleComponentByServiceVulnerabilityCounts;

    SET has_records = (cnt > 0);
END;

CALL add_post_migration_procedure('refresh_mvServiceIssueCounts_proc', 'has_mvServiceIssueCounts_proc');
CALL add_post_migration_procedure('refresh_mvCountIssueRatings_proc', 'has_mvCountIssueRatings_proc');
CALL add_post_migration_procedure('refresh_mvSingleComponentByServiceVulnerabilityCounts_proc', 'has_mvSingleComponentByServiceVulnerabilityCounts_proc');
CALL add_post_migration_procedure('refresh_mvAllComponentsByServiceVulnerabilityCounts_proc', 'has_mvAllComponentsByServiceVulnerabilityCounts_proc');


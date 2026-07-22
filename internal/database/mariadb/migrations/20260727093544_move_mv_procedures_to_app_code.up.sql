-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

DROP TABLE IF EXISTS post_migration_procedure_registry;
DROP PROCEDURE IF EXISTS call_registered_post_migration_procedures;
DROP PROCEDURE IF EXISTS add_post_migration_procedure;
DROP PROCEDURE IF EXISTS remove_post_migration_procedure;

DROP PROCEDURE IF EXISTS refresh_mvServiceIssueCounts_proc;
DROP PROCEDURE IF EXISTS refresh_mvServiceIssueCounts_proc;
DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsServiceId_proc;
DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsUniqueService_proc;
DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsOther_proc;
DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsService_proc;
DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsServiceWithoutSupportGroup_proc;
DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsSupportGroup_proc;
DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsComponentVersion_proc;
DROP PROCEDURE IF EXISTS refresh_mvVulnerabilityList_proc;
DROP PROCEDURE IF EXISTS refresh_mvVulnerabilityService_proc;
DROP PROCEDURE IF EXISTS refresh_mvComponentService_proc;
DROP PROCEDURE IF EXISTS refresh_mvSingleComponentByServiceVulnerabilityCounts_proc;
DROP PROCEDURE IF EXISTS refresh_mvAllComponentsByServiceVulnerabilityCounts_proc;

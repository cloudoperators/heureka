DROP EVENT IF EXISTS refresh_mvCountIssueRatingsOther;

DROP EVENT IF EXISTS refresh_mvCountIssueRatingsServiceId;

DROP EVENT IF EXISTS refresh_mvCountIssueRatingsComponentVersion;

DROP EVENT IF EXISTS refresh_mvCountIssueRatingsSupportGroup;

DROP EVENT IF EXISTS refresh_mvCountIssueRatingsServiceWithoutSupportGroup;

DROP EVENT IF EXISTS refresh_mvCountIssueRatingsService;

DROP EVENT IF EXISTS refresh_mvCountIssueRatingsUniqueService;

DROP EVENT IF EXISTS refresh_mvAllComponentsByServiceVulnerabilityCounts;

DROP EVENT IF EXISTS refresh_mvSingleComponentByServiceVulnerabilityCounts;

DROP EVENT IF EXISTS refresh_mvServiceIssueCounts;

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsOther_proc;

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsServiceId_proc;

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsComponentVersion_proc;

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsSupportGroup_proc;

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsServiceWithoutSupportGroup_proc;

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsService_proc;

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsUniqueService_proc;

DROP PROCEDURE IF EXISTS refresh_mvAllComponentsByServiceVulnerabilityCounts_proc;

DROP PROCEDURE IF EXISTS refresh_mvSingleComponentByServiceVulnerabilityCounts_proc;

DROP PROCEDURE IF EXISTS refresh_mvServiceIssueCounts_proc;

DROP PROCEDURE IF EXISTS remove_post_migration_procedure;

DROP PROCEDURE IF EXISTS add_post_migration_procedure;

DROP PROCEDURE IF EXISTS call_registered_post_migration_procedures;

DROP TABLE IF EXISTS mvServiceIssueCounts_tmp;

DROP TABLE IF EXISTS ComponentVersionIssue;

DROP TABLE IF EXISTS IssueMatch;

DROP TABLE IF EXISTS IssueRepositoryService;

DROP TABLE IF EXISTS IssueVariant;

DROP TABLE IF EXISTS Owner;

DROP TABLE IF EXISTS Patch;

DROP TABLE IF EXISTS Remediation;

DROP TABLE IF EXISTS ScannerRunComponentInstanceTracker;

DROP TABLE IF EXISTS ScannerRunError;

DROP TABLE IF EXISTS SupportGroupService;

DROP TABLE IF EXISTS SupportGroupUser;

DROP TABLE IF EXISTS mvAllComponentsByServiceVulnerabilityCounts;

DROP TABLE IF EXISTS mvCountIssueRatingsComponentVersion;

DROP TABLE IF EXISTS mvCountIssueRatingsOther;

DROP TABLE IF EXISTS mvCountIssueRatingsService;

DROP TABLE IF EXISTS mvCountIssueRatingsServiceId;

DROP TABLE IF EXISTS mvCountIssueRatingsServiceWithoutSupportGroup;

DROP TABLE IF EXISTS mvCountIssueRatingsSupportGroup;

DROP TABLE IF EXISTS mvCountIssueRatingsUniqueService;

DROP TABLE IF EXISTS mvServiceIssueCounts;

DROP TABLE IF EXISTS mvSingleComponentByServiceVulnerabilityCounts;

DROP TABLE IF EXISTS post_migration_procedure_registry;

DROP TABLE IF EXISTS ComponentInstance;

DROP TABLE IF EXISTS Issue;

DROP TABLE IF EXISTS IssueRepository;

DROP TABLE IF EXISTS ScannerRun;

DROP TABLE IF EXISTS SupportGroup;

DROP TABLE IF EXISTS ComponentVersion;

DROP TABLE IF EXISTS Service;

DROP TABLE IF EXISTS Component;

DROP TABLE IF EXISTS User;
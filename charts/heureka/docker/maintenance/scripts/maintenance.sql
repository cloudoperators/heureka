-- Maintenance script for database truncation
-- Enable logging
SET @start_time = NOW();
SELECT CONCAT('Starting maintenance at ', @start_time) AS log_message;

-- Start transaction
START TRANSACTION;

-- Disable foreign key checks temporarily
SET FOREIGN_KEY_CHECKS = 0;

-- Truncate junction join tables first
TRUNCATE TABLE IssueMatchEvidence;
TRUNCATE TABLE ComponentVersionIssue;
TRUNCATE TABLE IssueRepositoryService;
TRUNCATE TABLE ActivityHasIssue;
TRUNCATE TABLE ActivityHasService;
TRUNCATE TABLE IssueMatchChange;
TRUNCATE TABLE SupportGroupService;
TRUNCATE TABLE SupportGroupUser;
TRUNCATE TABLE Owner;

-- Truncate dependent entity tables
TRUNCATE TABLE IssueMatch;
TRUNCATE TABLE Evidence;
TRUNCATE TABLE ComponentInstance;
TRUNCATE TABLE IssueVariant;
TRUNCATE TABLE ComponentVersion;
TRUNCATE TABLE Activity;

-- Truncate main entity tables
TRUNCATE TABLE Component;
TRUNCATE TABLE Service;
TRUNCATE TABLE SupportGroup;
TRUNCATE TABLE Issue;
TRUNCATE TABLE IssueRepository;

-- Re-enable foreign key checks
SET FOREIGN_KEY_CHECKS = 1;

-- Log completion
SELECT CONCAT('Maintenance completed at ', NOW(), '. Duration: ', 
       TIMESTAMPDIFF(SECOND, @start_time, NOW()), ' seconds') AS log_message;

-- Commit transaction
COMMIT;

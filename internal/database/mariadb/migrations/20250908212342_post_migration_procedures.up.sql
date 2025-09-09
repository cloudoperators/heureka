-- SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS post_migration_procedure_registry (
    name VARCHAR(128) PRIMARY KEY
);

CREATE PROCEDURE call_registered_post_migration_procedures()
BEGIN
    DECLARE done INT DEFAULT 0;
    DECLARE pname VARCHAR(128);
    DECLARE cur CURSOR FOR SELECT name FROM post_migration_procedure_registry;
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = 1;

    OPEN cur;
    read_loop: LOOP
        FETCH cur INTO pname;
        IF done THEN
            LEAVE read_loop;
        END IF;

        SET @sql = CONCAT('CALL ', pname, '();');
        PREPARE stmt FROM @sql;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
    END LOOP;
    CLOSE cur;
END;

CREATE PROCEDURE add_post_migration_procedure(IN p_name VARCHAR(128))
BEGIN
    INSERT IGNORE INTO post_migration_procedure_registry (name)
    VALUES (p_name);
END;

CREATE PROCEDURE remove_post_migration_procedure(IN p_name VARCHAR(128))
BEGIN
    DELETE FROM post_migration_procedure_registry
    WHERE name = p_name;
END;

CALL add_post_migration_procedure('refresh_mvServiceIssueCounts_proc');
CALL add_post_migration_procedure('refresh_mvCountIssueRatings_proc');
CALL add_post_migration_procedure('refresh_mvSingleComponentByServiceVulnerabilityCounts_proc');
CALL add_post_migration_procedure('refresh_mvAllComponentsByServiceVulnerabilityCounts_proc');

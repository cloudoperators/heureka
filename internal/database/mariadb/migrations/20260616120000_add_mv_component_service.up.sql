-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

-- Materialized view for the component-to-service relationship.
-- Pre-computes which components are deployed to which services by
-- resolving the ComponentVersion→ComponentInstance→Service chain.
-- This eliminates the expensive LEFT JOIN through ComponentInstance (millions
-- of rows) when the Images query is filtered by service CCRN.

CREATE TABLE IF NOT EXISTS mvComponentService (
    service_id INT UNSIGNED NOT NULL,
    component_id INT UNSIGNED NOT NULL,
    PRIMARY KEY (service_id, component_id),
    INDEX idx_mvcs_component (component_id)
);

DROP PROCEDURE IF EXISTS refresh_mvComponentService_proc;
CREATE PROCEDURE refresh_mvComponentService_proc()
BEGIN
    -- Ensure clean state for atomic swap
    DROP TABLE IF EXISTS mvComponentService_tmp;
    DROP TABLE IF EXISTS mvComponentService_old;

    CREATE TABLE mvComponentService_tmp LIKE mvComponentService;

    INSERT INTO mvComponentService_tmp (service_id, component_id)
    SELECT DISTINCT CI.componentinstance_service_id, CV.componentversion_component_id
    FROM ComponentInstance CI
    JOIN ComponentVersion CV ON CI.componentinstance_component_version_id = CV.componentversion_id
    WHERE CI.componentinstance_deleted_at IS NULL;

    RENAME TABLE
        mvComponentService TO mvComponentService_old,
        mvComponentService_tmp TO mvComponentService;

    DROP TABLE IF EXISTS mvComponentService_old;
END;

-- Initial population
CALL refresh_mvComponentService_proc();

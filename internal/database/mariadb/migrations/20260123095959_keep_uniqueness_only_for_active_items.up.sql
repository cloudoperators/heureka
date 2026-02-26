--  SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
--  SPDX-License-Identifier: Apache-2.0


-- User

ALTER TABLE User
ADD COLUMN user_active_unique_user_id VARCHAR(64)
    GENERATED ALWAYS AS (
        IF(user_deleted_at IS NULL, user_unique_user_id, NULL)
    ) STORED INVISIBLE;

ALTER TABLE User
DROP INDEX unique_user_id_UNIQUE;

ALTER TABLE User
ADD UNIQUE KEY user_unique_active_unique_user_id (user_active_unique_user_id);


-- Component

ALTER TABLE Component
ADD COLUMN component_active_ccrn VARCHAR(255)
    GENERATED ALWAYS AS (
        IF(component_deleted_at IS NULL, component_ccrn, NULL)
    ) STORED INVISIBLE;

ALTER TABLE Component
DROP INDEX component_ccrn_UNIQUE;

ALTER TABLE Component
ADD UNIQUE KEY component_unique_active_ccrn (component_active_ccrn);


-- Component Version

ALTER TABLE ComponentVersion
ADD COLUMN componentversion_active_version VARCHAR(255)
    GENERATED ALWAYS AS (
        IF(componentversion_deleted_at IS NULL, componentversion_version, NULL)
    ) STORED INVISIBLE;

ALTER TABLE ComponentVersion
ADD COLUMN componentversion_active_component_id INT UNSIGNED
    GENERATED ALWAYS AS (
        IF(componentversion_deleted_at IS NULL, componentversion_component_id, NULL)
    ) VIRTUAL INVISIBLE;

ALTER TABLE ComponentVersion
DROP INDEX version_component_UNIQUE;

ALTER TABLE ComponentVersion
ADD UNIQUE KEY componentversion_unique_active_version (componentversion_active_version, componentversion_active_component_id);


-- Service

ALTER TABLE Service
ADD COLUMN service_active_ccrn VARCHAR(255)
    GENERATED ALWAYS AS (
        IF(service_deleted_at IS NULL, service_ccrn, NULL)
    ) STORED INVISIBLE;

ALTER TABLE Service
DROP INDEX service_ccrn_UNIQUE;

ALTER TABLE Service
ADD UNIQUE KEY service_unique_active_ccrn (service_active_ccrn);


-- Component Instance

ALTER TABLE ComponentInstance
ADD COLUMN componentinstance_active_ccrn VARCHAR(2048)
    GENERATED ALWAYS AS (
        IF(componentinstance_deleted_at IS NULL, componentinstance_ccrn, NULL)
    ) STORED INVISIBLE;

ALTER TABLE ComponentInstance
ADD COLUMN componentinstance_active_service_id INT UNSIGNED
    GENERATED ALWAYS AS (
        IF(componentinstance_deleted_at IS NULL, componentinstance_service_id, NULL)
    ) VIRTUAL INVISIBLE;

ALTER TABLE ComponentInstance
DROP INDEX componentinstance_ccrn_service_id_unique;

ALTER TABLE ComponentInstance
ADD UNIQUE KEY componentinstance_unique_active_ccrn (componentinstance_active_ccrn, componentinstance_active_service_id);


-- Issue Repository

ALTER TABLE IssueRepository
ADD COLUMN issuerepository_active_name VARCHAR(2048)
    GENERATED ALWAYS AS (
        IF(issuerepository_deleted_at IS NULL, issuerepository_name, NULL)
    ) STORED INVISIBLE;

ALTER TABLE IssueRepository
DROP INDEX issuerepository_name_UNIQUE;

ALTER TABLE IssueRepository
ADD UNIQUE KEY issuerepository_unique_active_name (issuerepository_active_name);


-- Issue

ALTER TABLE Issue
ADD COLUMN issue_active_primary_name VARCHAR(255)
    GENERATED ALWAYS AS (
        IF(issue_deleted_at IS NULL, issue_primary_name, NULL)
    ) STORED INVISIBLE;

ALTER TABLE Issue
DROP INDEX issue_primary_name_UNIQUE;

ALTER TABLE Issue
ADD UNIQUE KEY issue_unique_active_primary_name (issue_active_primary_name);


-- Issue Variant

ALTER TABLE IssueVariant
ADD COLUMN issuevariant_active_secondary_name VARCHAR(255)
    GENERATED ALWAYS AS (
        IF(issuevariant_deleted_at IS NULL, issuevariant_secondary_name, NULL)
    ) STORED INVISIBLE;

ALTER TABLE IssueVariant
DROP INDEX issuevariant_secondary_name_UNIQUE;

ALTER TABLE IssueVariant
ADD UNIQUE KEY issuevariant_unique_active_secondary_name (issuevariant_active_secondary_name);

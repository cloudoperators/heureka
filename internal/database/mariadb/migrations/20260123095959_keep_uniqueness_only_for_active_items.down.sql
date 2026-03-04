--  SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
--  SPDX-License-Identifier: Apache-2.0


-- User

ALTER TABLE User
DROP INDEX user_unique_active_unique_user_id;

ALTER TABLE User
DROP COLUMN user_active_unique_user_id;

ALTER TABLE User
ADD UNIQUE KEY unique_user_id_UNIQUE (user_unique_user_id);


-- Component

ALTER TABLE Component
DROP INDEX component_unique_active_ccrn;

ALTER TABLE Component
DROP COLUMN component_active_ccrn;

ALTER TABLE Component
ADD UNIQUE KEY component_ccrn_UNIQUE (component_ccrn);


-- Component Version

ALTER TABLE ComponentVersion
DROP INDEX componentversion_unique_active_version;

ALTER TABLE ComponentVersion
DROP COLUMN componentversion_active_version;

ALTER TABLE ComponentVersion
DROP COLUMN componentversion_active_component_id;

ALTER TABLE ComponentVersion
ADD UNIQUE KEY version_component_UNIQUE (componentversion_version, componentversion_component_id);


-- Service

ALTER TABLE Service
DROP INDEX service_unique_active_ccrn;

ALTER TABLE Service
DROP COLUMN service_active_ccrn;

ALTER TABLE Service
ADD UNIQUE KEY service_ccrn_UNIQUE (service_ccrn);


-- Component Instance

ALTER TABLE ComponentInstance
DROP INDEX componentinstance_unique_active_ccrn;

ALTER TABLE ComponentInstance
DROP COLUMN componentinstance_active_ccrn;

ALTER TABLE ComponentInstance
DROP COLUMN componentinstance_active_service_id;

ALTER TABLE ComponentInstance
ADD UNIQUE KEY componentinstance_ccrn_service_id_unique (componentinstance_ccrn, componentinstance_service_id);


-- Issue Repository

ALTER TABLE IssueRepository
DROP INDEX issuerepository_unique_active_name;

ALTER TABLE IssueRepository
DROP COLUMN issuerepository_active_name;

ALTER TABLE IssueRepository
ADD UNIQUE KEY issuerepository_name_UNIQUE (issuerepository_name);


-- Issue

ALTER TABLE Issue
DROP INDEX issue_unique_active_primary_name;

ALTER TABLE Issue
DROP COLUMN issue_active_primary_name;

ALTER TABLE Issue
ADD UNIQUE KEY issue_primary_name_UNIQUE (issue_primary_name);


-- Issue Variant

ALTER TABLE IssueVariant
DROP INDEX issuevariant_unique_active_secondary_name;

ALTER TABLE IssueVariant
DROP COLUMN issuevariant_active_secondary_name;

ALTER TABLE IssueVariant
ADD UNIQUE KEY issuevariant_secondary_name_UNIQUE (issuevariant_secondary_name);

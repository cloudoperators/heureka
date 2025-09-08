-- SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

DROP TABLE IF EXISTS post_migration_procedure_registry;
DROP PROCEDURE IF EXISTS call_registered_post_migration_procedures;
DROP PROCEDURE IF EXISTS add_post_migration_procedure;
DROP PROCEDURE IF EXISTS remove_post_migration_procedure;

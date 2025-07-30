-- SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

-- Drop the foreign key constraint
ALTER TABLE ComponentInstance
DROP FOREIGN KEY fk_componentinstance_parent_id;

-- Remove the column from the ComponentInstance table
ALTER TABLE ComponentInstance
DROP COLUMN componentinstance_parent_id;
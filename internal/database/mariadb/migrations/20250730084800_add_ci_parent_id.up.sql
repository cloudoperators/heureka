-- SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

-- Add the new column to the ComponentInstance table
ALTER TABLE ComponentInstance
ADD COLUMN componentinstance_parent_id INT UNSIGNED NULL;

-- Add the foreign key constraint
ALTER TABLE ComponentInstance
ADD CONSTRAINT fk_componentinstance_parent_id
FOREIGN KEY (componentinstance_parent_id)
REFERENCES ComponentInstance (componentinstance_id)
ON UPDATE CASCADE
ON DELETE CASCADE;
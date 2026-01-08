-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE Component
ADD COLUMN component_url VARCHAR(255) NULL,
ADD COLUMN component_repository VARCHAR(255) NULL,
ADD COLUMN component_organization VARCHAR(255) NULL;
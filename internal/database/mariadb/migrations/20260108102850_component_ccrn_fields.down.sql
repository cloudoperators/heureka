-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE Component
DROP COLUMN IF EXISTS component_url,
DROP COLUMN IF EXISTS component_repository,
DROP COLUMN IF EXISTS component_organization;
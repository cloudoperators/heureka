-- SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE Service
DROP COLUMN IF EXISTS service_domain,
DROP COLUMN IF EXISTS service_region;

-- SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE Service
ADD COLUMN service_domain VARCHAR(255) NULL,
ADD COLUMN service_region VARCHAR(255) NULL;
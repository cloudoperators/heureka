-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE ComponentVersion
    ADD COLUMN IF NOT EXISTS componentversion_is_outdated BOOLEAN NULL;

-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE Issue
    ADD COLUMN issue_known_exploited BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN issue_known_exploited_added_date TIMESTAMP NULL,
    ADD COLUMN issue_known_exploited_due_date TIMESTAMP NULL;

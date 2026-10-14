-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE Issue
    ADD COLUMN issue_epss_score DOUBLE NULL,
    ADD COLUMN issue_epss_percentile DOUBLE NULL;

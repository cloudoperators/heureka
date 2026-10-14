-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE Issue
    DROP COLUMN issue_epss_score,
    DROP COLUMN issue_epss_percentile;

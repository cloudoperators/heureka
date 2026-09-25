-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

-- Revert: truncate so the next MV refresh repopulates using the old logic
-- (restored in mvproc.go by reverting the Go code change).

TRUNCATE TABLE mvCountIssueRatingsComponentVersion;

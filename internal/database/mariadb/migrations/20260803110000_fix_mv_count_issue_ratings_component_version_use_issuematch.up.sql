-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

-- Fix: mvCountIssueRatingsComponentVersion was counting all issues linked via
-- ComponentVersionIssue + IssueVariant regardless of IssueMatch status, causing
-- vulnerabilityCounts to include risk_accepted/false_positive matches while
-- the vulnerability edges query filters strictly on issuematch_status = 'new'.
--
-- The refresh logic has been moved to Go (RefreshMVCountIssueRatingsComponentVersion
-- in mvproc.go). This migration truncates the stale MV data so the next scheduled
-- MV refresh repopulates it with the corrected query.

TRUNCATE TABLE mvCountIssueRatingsComponentVersion;

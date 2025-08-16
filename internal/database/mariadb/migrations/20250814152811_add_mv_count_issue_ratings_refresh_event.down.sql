-- SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

-- 1. Drop the event if exists
DROP EVENT IF EXISTS refresh_mvCountIssueRatings;

-- 2. Drop the stored procedure if exists
DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatings_proc;

-- 3. Drop the table if exists
DROP TABLE IF EXISTS mvCountIssueRatings1;
DROP TABLE IF EXISTS mvCountIssueRatings2;
DROP TABLE IF EXISTS mvCountIssueRatings2a;
DROP TABLE IF EXISTS mvCountIssueRatings3;
DROP TABLE IF EXISTS mvCountIssueRatings4;
DROP TABLE IF EXISTS mvCountIssueRatings5;
DROP TABLE IF EXISTS mvCountIssueRatings6;

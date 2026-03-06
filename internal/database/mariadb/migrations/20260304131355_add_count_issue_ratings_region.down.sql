-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

DROP EVENT IF EXISTS refresh_mvCountIssueRatingsRegion;

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsRegion_proc;

DROP TABLE IF EXISTS mvCountIssueRatingsRegion;
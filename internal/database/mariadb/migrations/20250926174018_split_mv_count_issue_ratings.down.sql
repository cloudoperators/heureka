-- SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

DROP EVENT IF EXISTS refresh_mvCountIssueRatings;
DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsUniqueService_proc;
DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsService_proc;
DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsServiceWithoutSupportGroup_proc;
DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsSupportGroup_proc;
DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsComponentVersion_proc;
DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsServiceId_proc;
DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsOther_proc;

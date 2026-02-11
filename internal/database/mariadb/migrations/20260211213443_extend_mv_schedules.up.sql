-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

ALTER EVENT refresh_mvServiceIssueCounts
ON SCHEDULE EVERY 2 HOUR;

ALTER EVENT refresh_mvSingleComponentByServiceVulnerabilityCounts
ON SCHEDULE EVERY 2 HOUR;

ALTER EVENT refresh_mvAllComponentsByServiceVulnerabilityCounts
ON SCHEDULE EVERY 2 HOUR;

ALTER EVENT refresh_mvCountIssueRatingsUniqueService
ON SCHEDULE EVERY 2 HOUR;

ALTER EVENT refresh_mvCountIssueRatingsService
ON SCHEDULE EVERY 2 HOUR;

ALTER EVENT refresh_mvCountIssueRatingsServiceWithoutSupportGroup
ON SCHEDULE EVERY 2 HOUR;

ALTER EVENT refresh_mvCountIssueRatingsSupportGroup
ON SCHEDULE EVERY 2 HOUR;

ALTER EVENT refresh_mvCountIssueRatingsComponentVersion
ON SCHEDULE EVERY 2 HOUR;

ALTER EVENT refresh_mvCountIssueRatingsServiceId
ON SCHEDULE EVERY 2 HOUR;

ALTER EVENT refresh_mvCountIssueRatingsOther
ON SCHEDULE EVERY 2 HOUR;

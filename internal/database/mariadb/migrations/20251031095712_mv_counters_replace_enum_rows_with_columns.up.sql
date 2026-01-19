-- SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE mvCountIssueRatingsUniqueService
    ADD COLUMN critical_count INT DEFAULT 0,
    ADD COLUMN high_count INT DEFAULT 0,
    ADD COLUMN medium_count INT DEFAULT 0,
    ADD COLUMN low_count INT DEFAULT 0,
    ADD COLUMN none_count INT DEFAULT 0;

UPDATE mvCountIssueRatingsUniqueService
SET critical_count = CASE WHEN issue_value = 'Critical' THEN issue_count ELSE 0 END,
    high_count     = CASE WHEN issue_value = 'High'     THEN issue_count ELSE 0 END,
    medium_count   = CASE WHEN issue_value = 'Medium'   THEN issue_count ELSE 0 END,
    low_count      = CASE WHEN issue_value = 'Low'      THEN issue_count ELSE 0 END,
    none_count     = CASE WHEN issue_value = 'None'     THEN issue_count ELSE 0 END;

ALTER TABLE mvCountIssueRatingsUniqueService
DROP COLUMN issue_value,
DROP COLUMN issue_count,
ADD COLUMN issue_count INT GENERATED ALWAYS AS (
    critical_count + high_count + medium_count + low_count + none_count
) STORED;

ALTER TABLE mvCountIssueRatingsUniqueService
  ADD COLUMN id INT NOT NULL PRIMARY KEY,
  ADD COLUMN is_active TINYINT(1) NOT NULL DEFAULT 1,
  ADD INDEX idx_mvCountIssueRatingsUniqueService_active (is_active);

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsUniqueService_proc;
CREATE PROCEDURE refresh_mvCountIssueRatingsUniqueService_proc()
BEGIN
    UPDATE mvCountIssueRatingsUniqueService
    SET is_active = 0
    WHERE is_active = 1;

    INSERT INTO mvCountIssueRatingsUniqueService
        (id, critical_count, high_count, medium_count, low_count, none_count, is_active)
    SELECT
        1,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical' THEN IV.issuevariant_issue_id END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High'     THEN IV.issuevariant_issue_id END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium'   THEN IV.issuevariant_issue_id END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low'      THEN IV.issuevariant_issue_id END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None'     THEN IV.issuevariant_issue_id END),
        1
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    WHERE I.issue_deleted_at IS NULL
    ON DUPLICATE KEY UPDATE
        critical_count = VALUES(critical_count),
        high_count     = VALUES(high_count),
        medium_count   = VALUES(medium_count),
        low_count      = VALUES(low_count),
        none_count     = VALUES(none_count),
        is_active      = VALUES(is_active);

    DELETE FROM mvCountIssueRatingsUniqueService
    WHERE is_active = 0;
END;


--

ALTER TABLE mvCountIssueRatingsService
    ADD COLUMN critical_count INT DEFAULT 0,
    ADD COLUMN high_count INT DEFAULT 0,
    ADD COLUMN medium_count INT DEFAULT 0,
    ADD COLUMN low_count INT DEFAULT 0,
    ADD COLUMN none_count INT DEFAULT 0;

UPDATE mvCountIssueRatingsService
SET critical_count = CASE WHEN issue_value = 'Critical' THEN issue_count ELSE 0 END,
    high_count     = CASE WHEN issue_value = 'High'     THEN issue_count ELSE 0 END,
    medium_count   = CASE WHEN issue_value = 'Medium'   THEN issue_count ELSE 0 END,
    low_count      = CASE WHEN issue_value = 'Low'      THEN issue_count ELSE 0 END,
    none_count     = CASE WHEN issue_value = 'None'     THEN issue_count ELSE 0 END;

ALTER TABLE mvCountIssueRatingsService
DROP COLUMN issue_value,
DROP COLUMN issue_count,
ADD COLUMN issue_count INT GENERATED ALWAYS AS (
    critical_count + high_count + medium_count + low_count + none_count
) STORED;

ALTER TABLE mvCountIssueRatingsService
  ADD PRIMARY KEY (supportgroup_ccrn),
  ADD COLUMN is_active TINYINT(1) NOT NULL DEFAULT 1,
  ADD INDEX idx_mvCountIssueRatingsService_active (is_active);

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsService_proc;
CREATE PROCEDURE refresh_mvCountIssueRatingsService_proc()
BEGIN
    UPDATE mvCountIssueRatingsService
    SET is_active = 0
    WHERE is_active = 1;

    INSERT INTO mvCountIssueRatingsService
        (supportgroup_ccrn, critical_count, high_count, medium_count, low_count, none_count, is_active)
    SELECT
        COALESCE(SG.supportgroup_ccrn, 'UNKNOWN'),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical' THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High'     THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium'   THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low'      THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None'     THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END),
        1
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    RIGHT JOIN IssueMatch IM ON I.issue_id = IM.issuematch_issue_id
    LEFT JOIN ComponentInstance CI ON CI.componentinstance_id = IM.issuematch_component_instance_id
    LEFT JOIN Service S ON S.service_id = CI.componentinstance_service_id
    LEFT JOIN SupportGroupService SGS ON SGS.supportgroupservice_service_id = CI.componentinstance_service_id
    LEFT JOIN SupportGroup SG ON SGS.supportgroupservice_support_group_id = SG.supportgroup_id
    LEFT JOIN Remediation R ON S.service_id = R.remediation_service_id
        AND I.issue_id = R.remediation_issue_id
        AND R.remediation_deleted_at IS NULL
    WHERE I.issue_deleted_at IS NULL
      AND (R.remediation_id IS NULL OR R.remediation_expiration_date < CURDATE())
    GROUP BY SG.supportgroup_ccrn
    ON DUPLICATE KEY UPDATE
        critical_count = VALUES(critical_count),
        high_count     = VALUES(high_count),
        medium_count   = VALUES(medium_count),
        low_count      = VALUES(low_count),
        none_count     = VALUES(none_count),
        is_active      = VALUES(is_active);

    DELETE FROM mvCountIssueRatingsService
    WHERE is_active = 0;
END;


--

ALTER TABLE mvCountIssueRatingsServiceWithoutSupportGroup
    ADD COLUMN critical_count INT DEFAULT 0,
    ADD COLUMN high_count INT DEFAULT 0,
    ADD COLUMN medium_count INT DEFAULT 0,
    ADD COLUMN low_count INT DEFAULT 0,
    ADD COLUMN none_count INT DEFAULT 0;

UPDATE mvCountIssueRatingsServiceWithoutSupportGroup
SET critical_count = CASE WHEN issue_value = 'Critical' THEN issue_count ELSE 0 END,
    high_count     = CASE WHEN issue_value = 'High'     THEN issue_count ELSE 0 END,
    medium_count   = CASE WHEN issue_value = 'Medium'   THEN issue_count ELSE 0 END,
    low_count      = CASE WHEN issue_value = 'Low'      THEN issue_count ELSE 0 END,
    none_count     = CASE WHEN issue_value = 'None'     THEN issue_count ELSE 0 END;
ALTER TABLE mvCountIssueRatingsServiceWithoutSupportGroup
DROP COLUMN issue_value,
DROP COLUMN issue_count,
ADD COLUMN issue_count INT GENERATED ALWAYS AS (
    critical_count + high_count + medium_count + low_count + none_count
) STORED;

ALTER TABLE mvCountIssueRatingsServiceWithoutSupportGroup
  ADD COLUMN id INT NOT NULL PRIMARY KEY,
  ADD COLUMN is_active TINYINT(1) NOT NULL DEFAULT 1,
  ADD INDEX idx_mvCountIssueRatingsServiceWithoutSupportGroup_active (is_active);

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsServiceWithoutSupportGroup_proc;
CREATE PROCEDURE refresh_mvCountIssueRatingsServiceWithoutSupportGroup_proc()
BEGIN
    UPDATE mvCountIssueRatingsServiceWithoutSupportGroup
    SET is_active = 0
    WHERE is_active = 1;

    INSERT INTO mvCountIssueRatingsServiceWithoutSupportGroup
        (id, critical_count, high_count, medium_count, low_count, none_count, is_active)
    SELECT
        1,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical' THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High'     THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium'   THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low'      THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None'     THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END),
        1
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    RIGHT JOIN IssueMatch IM ON I.issue_id = IM.issuematch_issue_id
    LEFT JOIN ComponentInstance CI ON CI.componentinstance_id = IM.issuematch_component_instance_id
    LEFT JOIN Service S ON S.service_id = CI.componentinstance_service_id
    LEFT JOIN Remediation R ON S.service_id = R.remediation_service_id
        AND I.issue_id = R.remediation_issue_id
        AND R.remediation_deleted_at IS NULL
    WHERE I.issue_deleted_at IS NULL
      AND (R.remediation_id IS NULL OR R.remediation_expiration_date < CURDATE())
    ON DUPLICATE KEY UPDATE
        critical_count = VALUES(critical_count),
        high_count     = VALUES(high_count),
        medium_count   = VALUES(medium_count),
        low_count      = VALUES(low_count),
        none_count     = VALUES(none_count),
        is_active      = VALUES(is_active);

    DELETE FROM mvCountIssueRatingsServiceWithoutSupportGroup
    WHERE is_active = 0;
END;


--

ALTER TABLE mvCountIssueRatingsSupportGroup
    ADD COLUMN critical_count INT DEFAULT 0,
    ADD COLUMN high_count INT DEFAULT 0,
    ADD COLUMN medium_count INT DEFAULT 0,
    ADD COLUMN low_count INT DEFAULT 0,
    ADD COLUMN none_count INT DEFAULT 0;

UPDATE mvCountIssueRatingsSupportGroup
SET critical_count = CASE WHEN issue_value = 'Critical' THEN issue_count ELSE 0 END,
    high_count     = CASE WHEN issue_value = 'High'     THEN issue_count ELSE 0 END,
    medium_count   = CASE WHEN issue_value = 'Medium'   THEN issue_count ELSE 0 END,
    low_count      = CASE WHEN issue_value = 'Low'      THEN issue_count ELSE 0 END,
    none_count     = CASE WHEN issue_value = 'None'     THEN issue_count ELSE 0 END;

ALTER TABLE mvCountIssueRatingsSupportGroup
DROP COLUMN issue_value,
DROP COLUMN issue_count,
ADD COLUMN issue_count INT GENERATED ALWAYS AS (
    critical_count + high_count + medium_count + low_count + none_count
) STORED;

ALTER TABLE mvCountIssueRatingsSupportGroup
  ADD PRIMARY KEY (supportgroup_ccrn),
  ADD COLUMN is_active TINYINT(1) NOT NULL DEFAULT 1,
  ADD INDEX idx_mvCountIssueRatingsSupportGroup_active (is_active);

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsSupportGroup_proc;
CREATE PROCEDURE refresh_mvCountIssueRatingsSupportGroup_proc()
BEGIN
    UPDATE mvCountIssueRatingsSupportGroup
    SET is_active = 0
    WHERE is_active = 1;

    INSERT INTO mvCountIssueRatingsSupportGroup
        (supportgroup_ccrn, critical_count, high_count, medium_count, low_count, none_count, is_active)
    SELECT
        COALESCE(SG.supportgroup_ccrn, 'UNKNOWN'),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', SGS.supportgroupservice_service_id, ',', SG.supportgroup_id)
        END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', SGS.supportgroupservice_service_id, ',', SG.supportgroup_id)
        END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', SGS.supportgroupservice_service_id, ',', SG.supportgroup_id)
        END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', SGS.supportgroupservice_service_id, ',', SG.supportgroup_id)
        END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', SGS.supportgroupservice_service_id, ',', SG.supportgroup_id)
        END),
        1
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    LEFT JOIN IssueMatch IM ON I.issue_id = IM.issuematch_issue_id
    LEFT JOIN ComponentInstance CI ON CI.componentinstance_id = IM.issuematch_component_instance_id
    LEFT JOIN SupportGroupService SGS ON SGS.supportgroupservice_service_id = CI.componentinstance_service_id
    LEFT JOIN SupportGroup SG ON SGS.supportgroupservice_support_group_id = SG.supportgroup_id
    LEFT JOIN Remediation R ON SGS.supportgroupservice_service_id = R.remediation_service_id
        AND I.issue_id = R.remediation_issue_id
        AND R.remediation_deleted_at IS NULL
    WHERE I.issue_deleted_at IS NULL
      AND CI.componentinstance_deleted_at IS NULL
      AND (R.remediation_id IS NULL OR R.remediation_expiration_date < CURDATE())
    GROUP BY SG.supportgroup_ccrn
    ON DUPLICATE KEY UPDATE
        critical_count = VALUES(critical_count),
        high_count     = VALUES(high_count),
        medium_count   = VALUES(medium_count),
        low_count      = VALUES(low_count),
        none_count     = VALUES(none_count),
        is_active      = VALUES(is_active);

    DELETE FROM mvCountIssueRatingsSupportGroup
    WHERE is_active = 0;
END;


--

ALTER TABLE mvCountIssueRatingsComponentVersion
    ADD COLUMN critical_count INT DEFAULT 0,
    ADD COLUMN high_count INT DEFAULT 0,
    ADD COLUMN medium_count INT DEFAULT 0,
    ADD COLUMN low_count INT DEFAULT 0,
    ADD COLUMN none_count INT DEFAULT 0;

UPDATE mvCountIssueRatingsComponentVersion
SET critical_count = CASE WHEN issue_value = 'Critical' THEN issue_count ELSE 0 END,
    high_count     = CASE WHEN issue_value = 'High'     THEN issue_count ELSE 0 END,
    medium_count   = CASE WHEN issue_value = 'Medium'   THEN issue_count ELSE 0 END,
    low_count      = CASE WHEN issue_value = 'Low'      THEN issue_count ELSE 0 END,
    none_count     = CASE WHEN issue_value = 'None'     THEN issue_count ELSE 0 END;

ALTER TABLE mvCountIssueRatingsComponentVersion
ADD COLUMN service_id INT DEFAULT NULL,
ADD COLUMN service_ccrn VARCHAR(255) DEFAULT NULL,
DROP COLUMN issue_value,
DROP COLUMN issue_count,
ADD COLUMN issue_count INT GENERATED ALWAYS AS (
    critical_count + high_count + medium_count + low_count + none_count
) STORED;

ALTER TABLE mvCountIssueRatingsComponentVersion
  ADD PRIMARY KEY (component_version_id),
  ADD COLUMN is_active TINYINT(1) NOT NULL DEFAULT 1,
  ADD INDEX idx_mvCountIssueRatingsComponentVersion_active (is_active);

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsComponentVersion_proc;
CREATE PROCEDURE refresh_mvCountIssueRatingsComponentVersion_proc()
BEGIN
    UPDATE mvCountIssueRatingsComponentVersion
    SET is_active = 0
    WHERE is_active = 1;

    INSERT INTO mvCountIssueRatingsComponentVersion
        (component_version_id, service_id, service_ccrn, critical_count, high_count, medium_count, low_count, none_count, is_active)
    SELECT
        CVI.componentversionissue_component_version_id,
        CI.componentinstance_service_id,
        S.service_ccrn,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical' THEN CONCAT(CVI.componentversionissue_component_version_id, ',', CVI.componentversionissue_issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High'     THEN CONCAT(CVI.componentversionissue_component_version_id, ',', CVI.componentversionissue_issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium'   THEN CONCAT(CVI.componentversionissue_component_version_id, ',', CVI.componentversionissue_issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low'      THEN CONCAT(CVI.componentversionissue_component_version_id, ',', CVI.componentversionissue_issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None'     THEN CONCAT(CVI.componentversionissue_component_version_id, ',', CVI.componentversionissue_issue_id) END),
        1
    FROM ComponentVersionIssue CVI
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = CVI.componentversionissue_issue_id
    INNER JOIN ComponentInstance CI ON CVI.componentversionissue_component_version_id = CI.componentinstance_component_version_id
    INNER JOIN Service S ON CI.componentinstance_service_id = S.service_id
    LEFT JOIN Remediation R ON CI.componentinstance_service_id = R.remediation_service_id
        AND CVI.componentversionissue_issue_id = R.remediation_issue_id
        AND R.remediation_deleted_at IS NULL
    WHERE (R.remediation_id IS NULL OR R.remediation_expiration_date < CURDATE())
    GROUP BY CVI.componentversionissue_component_version_id
    ON DUPLICATE KEY UPDATE
        service_id     = VALUES(service_id),
        service_ccrn   = VALUES(service_ccrn),
        critical_count = VALUES(critical_count),
        high_count     = VALUES(high_count),
        medium_count   = VALUES(medium_count),
        low_count      = VALUES(low_count),
        none_count     = VALUES(none_count),
        is_active      = VALUES(is_active);

    DELETE FROM mvCountIssueRatingsComponentVersion
    WHERE is_active = 0;
END;

--

ALTER TABLE mvCountIssueRatingsServiceId
    ADD COLUMN critical_count INT DEFAULT 0,
    ADD COLUMN high_count INT DEFAULT 0,
    ADD COLUMN medium_count INT DEFAULT 0,
    ADD COLUMN low_count INT DEFAULT 0,
    ADD COLUMN none_count INT DEFAULT 0;

UPDATE mvCountIssueRatingsServiceId
SET critical_count = CASE WHEN issue_value = 'Critical' THEN issue_count ELSE 0 END,
    high_count     = CASE WHEN issue_value = 'High'     THEN issue_count ELSE 0 END,
    medium_count   = CASE WHEN issue_value = 'Medium'   THEN issue_count ELSE 0 END,
    low_count      = CASE WHEN issue_value = 'Low'      THEN issue_count ELSE 0 END,
    none_count     = CASE WHEN issue_value = 'None'     THEN issue_count ELSE 0 END;

ALTER TABLE mvCountIssueRatingsServiceId
DROP COLUMN issue_value,
DROP COLUMN issue_count,
ADD COLUMN issue_count INT GENERATED ALWAYS AS (
    critical_count + high_count + medium_count + low_count + none_count
) STORED,
ADD COLUMN service_ccrn VARCHAR(255) DEFAULT NULL;

ALTER TABLE mvCountIssueRatingsServiceId
  ADD PRIMARY KEY (service_id),
  ADD COLUMN is_active TINYINT(1) NOT NULL DEFAULT 1,
  ADD INDEX idx_mvCountIssueRatingsServiceId_active (is_active);

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsServiceId_proc;
CREATE PROCEDURE refresh_mvCountIssueRatingsServiceId_proc()
BEGIN
    UPDATE mvCountIssueRatingsServiceId
    SET is_active = 0
    WHERE is_active = 1;

    INSERT INTO mvCountIssueRatingsServiceId
        (service_id, service_ccrn, critical_count, high_count, medium_count, low_count, none_count, is_active)
    SELECT
        CI.componentinstance_service_id,
        S.service_ccrn,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical' THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High'     THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium'   THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low'      THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None'     THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id) END),
        1
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    LEFT JOIN IssueMatch IM ON I.issue_id = IM.issuematch_issue_id
    LEFT JOIN ComponentInstance CI ON CI.componentinstance_id = IM.issuematch_component_instance_id
        AND CI.componentinstance_deleted_at IS NULL
    LEFT JOIN Service S ON S.service_id = CI.componentinstance_service_id
    LEFT JOIN Remediation R ON CI.componentinstance_service_id = R.remediation_service_id
        AND I.issue_id = R.remediation_issue_id
        AND R.remediation_deleted_at IS NULL
    WHERE I.issue_deleted_at IS NULL
      AND (R.remediation_id IS NULL OR R.remediation_expiration_date < CURDATE())
    GROUP BY CI.componentinstance_service_id
    ON DUPLICATE KEY UPDATE
        service_ccrn   = VALUES(service_ccrn),
        critical_count = VALUES(critical_count),
        high_count     = VALUES(high_count),
        medium_count   = VALUES(medium_count),
        low_count      = VALUES(low_count),
        none_count     = VALUES(none_count),
        is_active      = VALUES(is_active);

    DELETE FROM mvCountIssueRatingsServiceId
    WHERE is_active = 0;
END;


--

ALTER TABLE mvCountIssueRatingsOther
    ADD COLUMN critical_count INT DEFAULT 0,
    ADD COLUMN high_count INT DEFAULT 0,
    ADD COLUMN medium_count INT DEFAULT 0,
    ADD COLUMN low_count INT DEFAULT 0,
    ADD COLUMN none_count INT DEFAULT 0;

UPDATE mvCountIssueRatingsOther
SET critical_count = CASE WHEN issue_value = 'Critical' THEN issue_count ELSE 0 END,
    high_count     = CASE WHEN issue_value = 'High'     THEN issue_count ELSE 0 END,
    medium_count   = CASE WHEN issue_value = 'Medium'   THEN issue_count ELSE 0 END,
    low_count      = CASE WHEN issue_value = 'Low'      THEN issue_count ELSE 0 END,
    none_count     = CASE WHEN issue_value = 'None'     THEN issue_count ELSE 0 END;

ALTER TABLE mvCountIssueRatingsOther
DROP COLUMN issue_value,
DROP COLUMN issue_count,
ADD COLUMN issue_count INT GENERATED ALWAYS AS (
    critical_count + high_count + medium_count + low_count + none_count
) STORED;

ALTER TABLE mvCountIssueRatingsOther
  ADD COLUMN id INT NOT NULL PRIMARY KEY,
  ADD COLUMN is_active TINYINT(1) NOT NULL DEFAULT 1,
  ADD INDEX idx_mvCountIssueRatingsOther_active (is_active);

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsOther_proc;
CREATE PROCEDURE refresh_mvCountIssueRatingsOther_proc()
BEGIN
    UPDATE mvCountIssueRatingsOther
    SET is_active = 0
    WHERE is_active = 1;

    INSERT INTO mvCountIssueRatingsOther
        (id, critical_count, high_count, medium_count, low_count, none_count, is_active)
    SELECT
        1,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical' THEN IV.issuevariant_issue_id END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High'     THEN IV.issuevariant_issue_id END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium'   THEN IV.issuevariant_issue_id END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low'      THEN IV.issuevariant_issue_id END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None'     THEN IV.issuevariant_issue_id END),
        1
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    WHERE I.issue_deleted_at IS NULL
    ON DUPLICATE KEY UPDATE
        critical_count = VALUES(critical_count),
        high_count     = VALUES(high_count),
        medium_count   = VALUES(medium_count),
        low_count      = VALUES(low_count),
        none_count     = VALUES(none_count),
        is_active      = VALUES(is_active);

    DELETE FROM mvCountIssueRatingsOther
    WHERE is_active = 0;
END;

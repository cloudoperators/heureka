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

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsUniqueService_proc;
CREATE PROCEDURE refresh_mvCountIssueRatingsUniqueService_proc()
BEGIN
    TRUNCATE TABLE mvCountIssueRatingsUniqueService;
    INSERT INTO mvCountIssueRatingsUniqueService (
        critical_count,
        high_count,
        medium_count,
        low_count,
        none_count
    )
    SELECT
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical' THEN IV.issuevariant_issue_id END) AS critical_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High' THEN IV.issuevariant_issue_id END) AS high_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium' THEN IV.issuevariant_issue_id END) AS medium_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low' THEN IV.issuevariant_issue_id END) AS low_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None' THEN IV.issuevariant_issue_id END) AS none_count
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    WHERE I.issue_deleted_at IS NULL;
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

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsService_proc;
CREATE PROCEDURE refresh_mvCountIssueRatingsService_proc()
BEGIN
    TRUNCATE TABLE mvCountIssueRatingsService;
    INSERT INTO mvCountIssueRatingsService (
        supportgroup_ccrn,
        critical_count,
        high_count,
        medium_count,
        low_count,
        none_count
    )
    SELECT
        COALESCE(SG.supportgroup_ccrn, 'UNKNOWN') AS supportgroup_ccrn,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical' THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END) AS critical_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High' THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END) AS high_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium' THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END) AS medium_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low' THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END) AS low_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None' THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END) AS none_count
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    RIGHT JOIN IssueMatch IM ON I.issue_id = IM.issuematch_issue_id
    LEFT JOIN ComponentInstance CI ON CI.componentinstance_id = IM.issuematch_component_instance_id
    LEFT JOIN ComponentVersion CV ON CI.componentinstance_component_version_id = CV.componentversion_id
    LEFT JOIN Service S ON S.service_id = CI.componentinstance_service_id
    LEFT JOIN SupportGroupService SGS ON SGS.supportgroupservice_service_id = CI.componentinstance_service_id
    LEFT JOIN SupportGroup SG ON SGS.supportgroupservice_support_group_id = SG.supportgroup_id
    WHERE I.issue_deleted_at IS NULL
    GROUP BY SG.supportgroup_ccrn;
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

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsServiceWithoutSupportGroup_proc;
CREATE PROCEDURE refresh_mvCountIssueRatingsServiceWithoutSupportGroup_proc()
BEGIN
    TRUNCATE TABLE mvCountIssueRatingsServiceWithoutSupportGroup;
    INSERT INTO mvCountIssueRatingsServiceWithoutSupportGroup (
        critical_count,
        high_count,
        medium_count,
        low_count,
        none_count
    )
    SELECT
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical' THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END) AS critical_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High' THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END) AS high_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium' THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END) AS medium_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low' THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END) AS low_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None' THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END) AS none_count
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    RIGHT JOIN IssueMatch IM ON I.issue_id = IM.issuematch_issue_id
    LEFT JOIN ComponentInstance CI ON CI.componentinstance_id = IM.issuematch_component_instance_id
    LEFT JOIN ComponentVersion CV ON CI.componentinstance_component_version_id = CV.componentversion_id
    LEFT JOIN Service S ON S.service_id = CI.componentinstance_service_id
    WHERE I.issue_deleted_at IS NULL;
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

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsSupportGroup_proc;
CREATE PROCEDURE refresh_mvCountIssueRatingsSupportGroup_proc()
BEGIN
    TRUNCATE TABLE mvCountIssueRatingsSupportGroup;
    INSERT INTO mvCountIssueRatingsSupportGroup (
        supportgroup_ccrn,
        critical_count,
        high_count,
        medium_count,
        low_count,
        none_count
    )
    SELECT
        COALESCE(SG.supportgroup_ccrn, 'UNKNOWN') AS supportgroup_ccrn,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', SGS.supportgroupservice_service_id, ',', SG.supportgroup_id)
        END) AS critical_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', SGS.supportgroupservice_service_id, ',', SG.supportgroup_id)
        END) AS high_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', SGS.supportgroupservice_service_id, ',', SG.supportgroup_id)
        END) AS medium_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', SGS.supportgroupservice_service_id, ',', SG.supportgroup_id)
        END) AS low_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', SGS.supportgroupservice_service_id, ',', SG.supportgroup_id)
        END) AS none_count
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    LEFT JOIN IssueMatch IM ON I.issue_id = IM.issuematch_issue_id
    LEFT JOIN ComponentInstance CI ON CI.componentinstance_id = IM.issuematch_component_instance_id
    LEFT JOIN SupportGroupService SGS ON SGS.supportgroupservice_service_id = CI.componentinstance_service_id
    LEFT JOIN SupportGroup SG ON SGS.supportgroupservice_support_group_id = SG.supportgroup_id
    WHERE I.issue_deleted_at IS NULL
    GROUP BY SG.supportgroup_ccrn;
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
DROP COLUMN issue_value,
DROP COLUMN issue_count,
ADD COLUMN issue_count INT GENERATED ALWAYS AS (
    critical_count + high_count + medium_count + low_count + none_count
) STORED;

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsComponentVersion_proc;
CREATE PROCEDURE refresh_mvCountIssueRatingsComponentVersion_proc()
BEGIN
    TRUNCATE TABLE mvCountIssueRatingsComponentVersion;
    INSERT INTO mvCountIssueRatingsComponentVersion (
        component_version_id,
        critical_count,
        high_count,
        medium_count,
        low_count,
        none_count
    )
    SELECT
        CVI.componentversionissue_component_version_id AS component_version_id,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical' THEN CONCAT(CVI.componentversionissue_component_version_id, ',', CVI.componentversionissue_issue_id) END) AS critical_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High' THEN CONCAT(CVI.componentversionissue_component_version_id, ',', CVI.componentversionissue_issue_id) END) AS high_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium' THEN CONCAT(CVI.componentversionissue_component_version_id, ',', CVI.componentversionissue_issue_id) END) AS medium_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low' THEN CONCAT(CVI.componentversionissue_component_version_id, ',', CVI.componentversionissue_issue_id) END) AS low_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None' THEN CONCAT(CVI.componentversionissue_component_version_id, ',', CVI.componentversionissue_issue_id) END) AS none_count
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    LEFT JOIN ComponentVersionIssue CVI ON I.issue_id = CVI.componentversionissue_issue_id
    WHERE I.issue_deleted_at IS NULL
    GROUP BY CVI.componentversionissue_component_version_id;
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
) STORED;

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsServiceId_proc;
CREATE PROCEDURE refresh_mvCountIssueRatingsServiceId_proc()
BEGIN
    TRUNCATE TABLE mvCountIssueRatingsServiceId;
    INSERT INTO mvCountIssueRatingsServiceId (
        service_id,
        critical_count,
        high_count,
        medium_count,
        low_count,
        none_count
    )
    SELECT
        CI.componentinstance_service_id AS service_id,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical' THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id) END) AS critical_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High' THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id) END) AS high_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium' THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id) END) AS medium_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low' THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id) END) AS low_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None' THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id) END) AS none_count
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    LEFT JOIN IssueMatch IM ON I.issue_id = IM.issuematch_issue_id
    LEFT JOIN ComponentInstance CI ON CI.componentinstance_id = IM.issuematch_component_instance_id
    WHERE I.issue_deleted_at IS NULL
    GROUP BY CI.componentinstance_service_id;
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

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsOther_proc;
CREATE PROCEDURE refresh_mvCountIssueRatingsOther_proc()
BEGIN
    TRUNCATE TABLE mvCountIssueRatingsOther;

    INSERT INTO mvCountIssueRatingsOther (
        critical_count,
        high_count,
        medium_count,
        low_count,
        none_count
    )
    SELECT
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical' THEN IV.issuevariant_issue_id END) AS critical_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High' THEN IV.issuevariant_issue_id END) AS high_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium' THEN IV.issuevariant_issue_id END) AS medium_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low' THEN IV.issuevariant_issue_id END) AS low_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None' THEN IV.issuevariant_issue_id END) AS none_count
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    WHERE I.issue_deleted_at IS NULL;
END;

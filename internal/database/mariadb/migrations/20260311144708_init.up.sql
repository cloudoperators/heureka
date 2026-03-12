-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS User (
    user_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_name VARCHAR(255) NOT NULL,
    user_unique_user_id VARCHAR(64) NOT NULL,
    user_email VARCHAR(255) NOT NULL,
    user_type INT UNSIGNED,
    user_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    user_created_by INT UNSIGNED NULL CHECK (user_created_by <> 0),
    user_deleted_at TIMESTAMP NULL,
    user_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    user_updated_by INT UNSIGNED NULL CHECK (user_updated_by <> 0),
    user_active_unique_user_id VARCHAR(64) GENERATED ALWAYS AS (
        IF(
            user_deleted_at IS NULL,
            user_unique_user_id,
            NULL
        )
    ) STORED INVISIBLE,
    CONSTRAINT user_id_UNIQUE UNIQUE (user_id),
    CONSTRAINT user_unique_active_unique_user_id UNIQUE (user_active_unique_user_id),
    CONSTRAINT fk_user_created_by FOREIGN KEY (user_created_by) REFERENCES User (user_id),
    CONSTRAINT fk_user_updated_by FOREIGN KEY (user_updated_by) REFERENCES User (user_id)
);

SET @TechnicalUserType = 2;

SET @SystemUserId = 1;

SET @SystemUserName = 'systemuser';

SET @SystemUserUniqueUserId = 'S0000000';

SET @SystemUserEmail = '';

INSERT IGNORE INTO
    User (
        user_id,
        user_name,
        user_unique_user_id,
        user_email,
        user_type,
        user_created_at,
        user_created_by
    )
VALUES (
        @SystemUserId,
        @SystemUserName,
        @SystemUserUniqueUserId,
        @SystemUserEmail,
        @TechnicalUserType,
        CURRENT_TIMESTAMP(),
        @SystemUserId
    );

CREATE TABLE IF NOT EXISTS Component (
    component_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    component_ccrn VARCHAR(255) NOT NULL,
    component_type VARCHAR(255) NOT NULL,
    component_url VARCHAR(255) NULL,
    component_repository VARCHAR(255) NULL,
    component_organization VARCHAR(255) NULL,
    component_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    component_created_by INT UNSIGNED NULL,
    component_deleted_at TIMESTAMP NULL,
    component_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    component_updated_by INT UNSIGNED NULL,
    component_active_ccrn VARCHAR(255) GENERATED ALWAYS AS (
        IF(
            component_deleted_at IS NULL,
            component_ccrn,
            NULL
        )
    ) STORED INVISIBLE,
    CONSTRAINT id_UNIQUE UNIQUE (component_id),
    CONSTRAINT component_unique_active_ccrn UNIQUE (component_active_ccrn),
    CONSTRAINT fk_component_created_by FOREIGN KEY (component_created_by) REFERENCES User (user_id),
    CONSTRAINT fk_component_updated_by FOREIGN KEY (component_updated_by) REFERENCES User (user_id)
);

CREATE TABLE IF NOT EXISTS ComponentVersion (
    componentversion_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    componentversion_version VARCHAR(255) NOT NULL,
    componentversion_component_id INT UNSIGNED NOT NULL,
    componentversion_tag VARCHAR(255) NULL,
    componentversion_repository VARCHAR(255) NULL,
    componentversion_organization VARCHAR(255) NULL,
    componentversion_end_of_life BOOLEAN NOT NULL DEFAULT FALSE,
    componentversion_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    componentversion_created_by INT UNSIGNED NULL,
    componentversion_deleted_at TIMESTAMP NULL,
    componentversion_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    componentversion_updated_by INT UNSIGNED NULL,
    componentversion_active_version VARCHAR(255) GENERATED ALWAYS AS (
        IF(
            componentversion_deleted_at IS NULL,
            componentversion_version,
            NULL
        )
    ) STORED INVISIBLE,
    componentversion_active_component_id INT UNSIGNED GENERATED ALWAYS AS (
        IF(
            componentversion_deleted_at IS NULL,
            componentversion_component_id,
            NULL
        )
    ) VIRTUAL INVISIBLE,
    CONSTRAINT componentversion_id_UNIQUE UNIQUE (componentversion_id),
    CONSTRAINT componentversion_unique_active_version UNIQUE (
        componentversion_active_version,
        componentversion_active_component_id
    ),
    CONSTRAINT fk_component_version_component FOREIGN KEY (componentversion_component_id) REFERENCES Component (component_id) ON UPDATE CASCADE,
    CONSTRAINT fk_componentversion_created_by FOREIGN KEY (componentversion_created_by) REFERENCES User (user_id),
    CONSTRAINT fk_componentversion_updated_by FOREIGN KEY (componentversion_updated_by) REFERENCES User (user_id)
);

CREATE TABLE IF NOT EXISTS SupportGroup (
    supportgroup_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    supportgroup_ccrn VARCHAR(255) NOT NULL,
    supportgroup_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    supportgroup_created_by INT UNSIGNED NULL,
    supportgroup_deleted_at TIMESTAMP NULL,
    supportgroup_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    supportgroup_updated_by INT UNSIGNED NULL,
    CONSTRAINT supportgroup_id_UNIQUE UNIQUE (supportgroup_id),
    CONSTRAINT supportgroup_ccrn_UNIQUE UNIQUE (supportgroup_ccrn),
    CONSTRAINT fk_supportgroup_created_by FOREIGN KEY (supportgroup_created_by) REFERENCES User (user_id),
    CONSTRAINT fk_supportgroup_updated_by FOREIGN KEY (supportgroup_updated_by) REFERENCES User (user_id)
);

CREATE TABLE IF NOT EXISTS Service (
    service_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    service_ccrn VARCHAR(255) NOT NULL,
    service_domain VARCHAR(255) NULL,
    service_region VARCHAR(255) NULL,
    service_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    service_created_by INT UNSIGNED NULL,
    service_deleted_at TIMESTAMP NULL,
    service_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    service_updated_by INT UNSIGNED NULL,
    service_active_ccrn VARCHAR(255) GENERATED ALWAYS AS (
        IF(
            service_deleted_at IS NULL,
            service_ccrn,
            NULL
        )
    ) STORED INVISIBLE,
    CONSTRAINT service_id_UNIQUE UNIQUE (service_id),
    CONSTRAINT service_unique_active_ccrn UNIQUE (service_active_ccrn),
    CONSTRAINT fk_service_created_by FOREIGN KEY (service_created_by) REFERENCES User (user_id),
    CONSTRAINT fk_service_updated_by FOREIGN KEY (service_updated_by) REFERENCES User (user_id)
);

CREATE TABLE IF NOT EXISTS SupportGroupService (
    supportgroupservice_service_id INT UNSIGNED NOT NULL,
    supportgroupservice_support_group_id INT UNSIGNED NOT NULL,
    supportgroupservice_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    supportgroupservice_deleted_at TIMESTAMP NULL,
    supportgroupservice_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    PRIMARY KEY (
        supportgroupservice_service_id,
        supportgroupservice_support_group_id
    ),
    CONSTRAINT fk_support_group_service FOREIGN KEY (
        supportgroupservice_support_group_id
    ) REFERENCES SupportGroup (supportgroup_id) ON UPDATE CASCADE,
    CONSTRAINT fk_service_support_group FOREIGN KEY (
        supportgroupservice_service_id
    ) REFERENCES Service (service_id) ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS ComponentInstance (
    componentinstance_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    componentinstance_ccrn VARCHAR(2048) NOT NULL,
    componentinstance_region VARCHAR(1024) NULL,
    componentinstance_cluster VARCHAR(1024) NULL,
    componentinstance_namespace VARCHAR(1024) NULL,
    componentinstance_domain VARCHAR(1024) NULL,
    componentinstance_project VARCHAR(1024) NULL,
    componentinstance_pod VARCHAR(1024) NULL,
    componentinstance_container VARCHAR(1024) NULL,
    componentinstance_type ENUM(
        'Unknown',
        'Project',
        'Server',
        'SecurityGroup',
        'SecurityGroupRule',
        'DnsZone',
        'FloatingIp',
        'RbacPolicy',
        'User',
        'Container',
        'RecordSet',
        'ProjectConfiguration'
    ) NULL DEFAULT 'Unknown',
    componentinstance_context JSON NULL DEFAULT "{}" CHECK (
        JSON_TYPE(componentinstance_context) != 'ARRAY'
    ),
    componentinstance_count INT NOT NULL DEFAULT 0,
    componentinstance_component_version_id INT UNSIGNED NULL,
    componentinstance_service_id INT UNSIGNED NOT NULL,
    componentinstance_parent_id INT UNSIGNED NULL,
    componentinstance_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    componentinstance_created_by INT UNSIGNED NULL,
    componentinstance_deleted_at TIMESTAMP NULL,
    componentinstance_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    componentinstance_updated_by INT UNSIGNED NULL,
    componentinstance_active_ccrn VARCHAR(2048) GENERATED ALWAYS AS (
        IF(
            componentinstance_deleted_at IS NULL,
            componentinstance_ccrn,
            NULL
        )
    ) STORED INVISIBLE,
    componentinstance_active_service_id INT UNSIGNED GENERATED ALWAYS AS (
        IF(
            componentinstance_deleted_at IS NULL,
            componentinstance_service_id,
            NULL
        )
    ) VIRTUAL INVISIBLE,
    CONSTRAINT componentinstance_id_UNIQUE UNIQUE (componentinstance_id),
    CONSTRAINT componentinstance_unique_active_ccrn UNIQUE (
        componentinstance_active_ccrn,
        componentinstance_active_service_id
    ),
    CONSTRAINT fk_component_instance_component_version FOREIGN KEY (
        componentinstance_component_version_id
    ) REFERENCES ComponentVersion (componentversion_id) ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_component_instance_service FOREIGN KEY (componentinstance_service_id) REFERENCES Service (service_id) ON UPDATE CASCADE,
    CONSTRAINT fk_componentinstance_parent_id FOREIGN KEY (componentinstance_parent_id) REFERENCES ComponentInstance (componentinstance_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_componentinstance_created_by FOREIGN KEY (componentinstance_created_by) REFERENCES User (user_id),
    CONSTRAINT fk_componentinstance_updated_by FOREIGN KEY (componentinstance_updated_by) REFERENCES User (user_id)
);

CREATE TABLE IF NOT EXISTS Owner (
    owner_service_id INT UNSIGNED NOT NULL,
    owner_user_id INT UNSIGNED NOT NULL,
    owner_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    owner_deleted_at TIMESTAMP NULL,
    owner_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    PRIMARY KEY (
        owner_service_id,
        owner_user_id
    ),
    CONSTRAINT fk_service_user FOREIGN KEY (owner_service_id) REFERENCES Service (service_id) ON UPDATE CASCADE,
    CONSTRAINT fk_user_service FOREIGN KEY (owner_user_id) REFERENCES User (user_id) ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS SupportGroupUser (
    supportgroupuser_user_id INT UNSIGNED NOT NULL,
    supportgroupuser_support_group_id INT UNSIGNED NOT NULL,
    supportgroupuser_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    supportgroupuser_deleted_at TIMESTAMP NULL,
    supportgroupuser_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    PRIMARY KEY (
        supportgroupuser_user_id,
        supportgroupuser_support_group_id
    ),
    CONSTRAINT fk_support_group_user FOREIGN KEY (
        supportgroupuser_support_group_id
    ) REFERENCES SupportGroup (supportgroup_id) ON UPDATE CASCADE,
    CONSTRAINT fk_user_support_group FOREIGN KEY (supportgroupuser_user_id) REFERENCES User (user_id) ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS IssueRepository (
    issuerepository_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    issuerepository_name VARCHAR(2048) NOT NULL,
    issuerepository_url VARCHAR(2048) NOT NULL,
    issuerepository_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    issuerepository_created_by INT UNSIGNED NULL,
    issuerepository_deleted_at TIMESTAMP NULL,
    issuerepository_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    issuerepository_updated_by INT UNSIGNED NULL,
    issuerepository_active_name VARCHAR(2048) GENERATED ALWAYS AS (
        IF(
            issuerepository_deleted_at IS NULL,
            issuerepository_name,
            NULL
        )
    ) STORED INVISIBLE,
    CONSTRAINT issuerepository_id_UNIQUE UNIQUE (issuerepository_id),
    CONSTRAINT issuerepository_unique_active_name UNIQUE (issuerepository_active_name),
    CONSTRAINT fk_issuerepository_created_by FOREIGN KEY (issuerepository_created_by) REFERENCES User (user_id),
    CONSTRAINT fk_issuerepository_updated_by FOREIGN KEY (issuerepository_updated_by) REFERENCES User (user_id)
);

CREATE TABLE IF NOT EXISTS Issue (
    issue_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    issue_type ENUM(
        'Vulnerability',
        'PolicyViolation',
        'SecurityEvent'
    ) NOT NULL,
    issue_primary_name VARCHAR(255) NOT NULL,
    issue_description LONGTEXT NOT NULL,
    issue_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    issue_created_by INT UNSIGNED NULL,
    issue_deleted_at TIMESTAMP NULL,
    issue_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    issue_updated_by INT UNSIGNED NULL,
    issue_active_primary_name VARCHAR(255) GENERATED ALWAYS AS (
        IF(
            issue_deleted_at IS NULL,
            issue_primary_name,
            NULL
        )
    ) STORED INVISIBLE,
    CONSTRAINT issue_id_UNIQUE UNIQUE (issue_id),
    CONSTRAINT issue_unique_active_primary_name UNIQUE (issue_active_primary_name),
    CONSTRAINT fk_issue_created_by FOREIGN KEY (issue_created_by) REFERENCES User (user_id),
    CONSTRAINT fk_issue_updated_by FOREIGN KEY (issue_updated_by) REFERENCES User (user_id)
);

CREATE INDEX idx_issue_issue_type ON Issue (issue_type);

CREATE TABLE IF NOT EXISTS IssueVariant (
    issuevariant_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    issuevariant_issue_id INT UNSIGNED NOT NULL,
    issuevariant_repository_id INT UNSIGNED NOT NULL,
    issuevariant_vector VARCHAR(512) NULL,
    issuevariant_rating ENUM(
        'None',
        'Low',
        'Medium',
        'High',
        'Critical'
    ) NOT NULL,
    issuevariant_secondary_name VARCHAR(255) NOT NULL,
    issuevariant_description LONGTEXT NOT NULL,
    issuevariant_external_url VARCHAR(2048) NULL,
    issuevariant_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    issuevariant_created_by INT UNSIGNED NULL,
    issuevariant_deleted_at TIMESTAMP NULL,
    issuevariant_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    issuevariant_updated_by INT UNSIGNED NULL,
    issuevariant_active_secondary_name VARCHAR(255) GENERATED ALWAYS AS (
        IF(
            issuevariant_deleted_at IS NULL,
            issuevariant_secondary_name,
            NULL
        )
    ) STORED INVISIBLE,
    CONSTRAINT issuevariant_id_UNIQUE UNIQUE (issuevariant_id),
    CONSTRAINT issuevariant_unique_active_secondary_name UNIQUE (
        issuevariant_active_secondary_name
    ),
    CONSTRAINT fk_issuevariant_issue FOREIGN KEY (issuevariant_issue_id) REFERENCES Issue (issue_id),
    CONSTRAINT fk_issuevariant_issuerepository FOREIGN KEY (issuevariant_repository_id) REFERENCES IssueRepository (issuerepository_id) ON UPDATE CASCADE,
    CONSTRAINT fk_issuevariant_created_by FOREIGN KEY (issuevariant_created_by) REFERENCES User (user_id),
    CONSTRAINT fk_issuevariant_updated_by FOREIGN KEY (issuevariant_updated_by) REFERENCES User (user_id)
);

CREATE TABLE IF NOT EXISTS IssueMatch (
    issuematch_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    issuematch_status VARCHAR(128) NOT NULL,
    issuematch_remediation_date DATETIME NULL,
    issuematch_target_remediation_date DATETIME NOT NULL,
    issuematch_vector VARCHAR(512) NULL,
    issuematch_rating ENUM(
        'None',
        'Low',
        'Medium',
        'High',
        'Critical'
    ) NOT NULL,
    issuematch_user_id INT UNSIGNED NOT NULL,
    issuematch_issue_id INT UNSIGNED NOT NULL,
    issuematch_component_instance_id INT UNSIGNED NOT NULL,
    issuematch_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    issuematch_created_by INT UNSIGNED NULL,
    issuematch_deleted_at TIMESTAMP NULL,
    issuematch_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    issuematch_updated_by INT UNSIGNED NULL,
    CONSTRAINT issuematch_id_UNIQUE UNIQUE (issuematch_id),
    CONSTRAINT fk_issue_match_user_id FOREIGN KEY (issuematch_user_id) REFERENCES User (user_id) ON UPDATE CASCADE,
    CONSTRAINT fk_issue_match_component_instance FOREIGN KEY (
        issuematch_component_instance_id
    ) REFERENCES ComponentInstance (componentinstance_id) ON UPDATE CASCADE,
    CONSTRAINT fk_issue_match_issue FOREIGN KEY (issuematch_issue_id) REFERENCES Issue (issue_id) ON UPDATE CASCADE,
    CONSTRAINT fk_issuematch_created_by FOREIGN KEY (issuematch_created_by) REFERENCES User (user_id),
    CONSTRAINT fk_issuematch_updated_by FOREIGN KEY (issuematch_updated_by) REFERENCES User (user_id)
);

CREATE TABLE IF NOT EXISTS ComponentVersionIssue (
    componentversionissue_component_version_id INT UNSIGNED NOT NULL,
    componentversionissue_issue_id INT UNSIGNED NOT NULL,
    componentversionissue_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    componentversionissue_deleted_at TIMESTAMP NULL,
    componentversionissue_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    PRIMARY KEY (
        componentversionissue_component_version_id,
        componentversionissue_issue_id
    ),
    CONSTRAINT fk_componentversionissue_issue FOREIGN KEY (
        componentversionissue_issue_id
    ) REFERENCES Issue (issue_id),
    CONSTRAINT fk_componentversionissue_component_version FOREIGN KEY (
        componentversionissue_component_version_id
    ) REFERENCES ComponentVersion (componentversion_id)
);

CREATE TABLE IF NOT EXISTS IssueRepositoryService (
    issuerepositoryservice_service_id INT UNSIGNED NOT NULL,
    issuerepositoryservice_issue_repository_id INT UNSIGNED NOT NULL,
    issuerepositoryservice_priority INT UNSIGNED NOT NULL,
    PRIMARY KEY (
        issuerepositoryservice_service_id,
        issuerepositoryservice_issue_repository_id
    ),
    CONSTRAINT fk_issue_repository_service FOREIGN KEY (
        issuerepositoryservice_issue_repository_id
    ) REFERENCES IssueRepository (issuerepository_id) ON UPDATE CASCADE,
    CONSTRAINT fk_service_issue_repository FOREIGN KEY (
        issuerepositoryservice_service_id
    ) REFERENCES Service (service_id) ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS Patch (
    patch_id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    patch_service_id INT UNSIGNED NOT NULL,
    patch_service_name VARCHAR(255) NOT NULL,
    patch_component_version_id INT UNSIGNED NOT NULL,
    patch_component_version_name VARCHAR(255) NOT NULL,
    patch_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    patch_created_by INT UNSIGNED NULL,
    patch_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    patch_updated_by INT UNSIGNED NULL,
    patch_deleted_at TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (patch_id),
    CONSTRAINT fk_patch_service FOREIGN KEY (patch_service_id) REFERENCES Service (service_id) ON DELETE CASCADE,
    CONSTRAINT fk_patch_component_version FOREIGN KEY (patch_component_version_id) REFERENCES ComponentVersion (componentversion_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS Remediation (
    remediation_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    remediation_type ENUM(
        'false_positive',
        'risk_accepted',
        'mitigation',
        'rescore'
    ) NOT NULL,
    remediation_severity ENUM(
        'None',
        'Low',
        'Medium',
        'High',
        'Critical'
    ) NOT NULL DEFAULT 'None',
    remediation_description LONGTEXT NOT NULL,
    remediation_remediation_date DATETIME NULL,
    remediation_expiration_date DATETIME NULL,
    remediation_remediated_by VARCHAR(255) NULL,
    remediation_remediated_by_id INT UNSIGNED NULL,
    remediation_service VARCHAR(255) NOT NULL,
    remediation_service_id INT UNSIGNED NOT NULL,
    remediation_component VARCHAR(255) NULL,
    remediation_component_id INT UNSIGNED NULL,
    remediation_issue VARCHAR(255) NOT NULL,
    remediation_issue_id INT UNSIGNED NOT NULL,
    remediation_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    remediation_created_by INT UNSIGNED NULL,
    remediation_deleted_at TIMESTAMP NULL,
    remediation_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    remediation_updated_by INT UNSIGNED NULL,
    CONSTRAINT fk_remediation_created_by FOREIGN KEY (remediation_created_by) REFERENCES User (user_id),
    CONSTRAINT fk_remediation_updated_by FOREIGN KEY (remediation_updated_by) REFERENCES User (user_id),
    CONSTRAINT fk_remediation_service_id FOREIGN KEY (remediation_service_id) REFERENCES Service (service_id),
    CONSTRAINT fk_remediation_component_id FOREIGN KEY (remediation_component_id) REFERENCES Component (component_id),
    CONSTRAINT fk_remediation_issue_id FOREIGN KEY (remediation_issue_id) REFERENCES Issue (issue_id)
);

CREATE TABLE IF NOT EXISTS ScannerRun (
    scannerrun_run_id INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    scannerrun_uuid UUID NOT NULL UNIQUE,
    scannerrun_tag VARCHAR(255) NOT NULL,
    scannerrun_start_run TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    scannerrun_end_run TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    scannerrun_is_completed BOOLEAN NOT NULL DEFAULT FALSE,
    scannerrun_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    scannerrun_created_by INT UNSIGNED NULL,
    scannerrun_deleted_at TIMESTAMP NULL,
    scannerrun_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    scannerrun_updated_by INT UNSIGNED NULL
);

CREATE TABLE IF NOT EXISTS ScannerRunComponentInstanceTracker (
    scannerruncomponentinstancetracker_scannerrun_run_id INT UNSIGNED NOT NULL,
    scannerruncomponentinstancetracker_component_instance_id INT UNSIGNED NOT NULL,
    CONSTRAINT fk_srcit_sr_id FOREIGN KEY (
        scannerruncomponentinstancetracker_scannerrun_run_id
    ) REFERENCES ScannerRun (scannerrun_run_id) ON UPDATE CASCADE,
    CONSTRAINT fk_srcit_ci_id FOREIGN KEY (
        scannerruncomponentinstancetracker_component_instance_id
    ) REFERENCES ComponentInstance (componentinstance_id) ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS ScannerRunError (
    scannerrunerror_scannerrun_run_id INT UNSIGNED NOT NULL,
    error TEXT NOT NULL,
    CONSTRAINT fk_sre_sr_id FOREIGN KEY (
        scannerrunerror_scannerrun_run_id
    ) REFERENCES ScannerRun (scannerrun_run_id) ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS post_migration_procedure_registry (name VARCHAR(128) PRIMARY KEY);

CREATE TABLE IF NOT EXISTS mvServiceIssueCounts (
    service_id INT NOT NULL PRIMARY KEY,
    critical_count INT DEFAULT 0,
    high_count INT DEFAULT 0,
    medium_count INT DEFAULT 0,
    low_count INT DEFAULT 0,
    none_count INT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS mvSingleComponentByServiceVulnerabilityCounts (
    service_id INT UNSIGNED NOT NULL,
    component_id INT UNSIGNED NOT NULL,
    critical_count INT DEFAULT 0,
    high_count INT DEFAULT 0,
    medium_count INT DEFAULT 0,
    low_count INT DEFAULT 0,
    none_count INT DEFAULT 0,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    PRIMARY KEY (service_id, component_id),
    KEY idx_mvSingleComponent_active (is_active),
    CONSTRAINT fk_mvsinglecomponentbyservicevulnerabilitycounts_service_id FOREIGN KEY (service_id) REFERENCES Service (service_id),
    CONSTRAINT fk_mvsinglecomponentbyservicevulnerabilitycounts_component_id FOREIGN KEY (component_id) REFERENCES Component (component_id)
);

CREATE TABLE IF NOT EXISTS mvAllComponentsByServiceVulnerabilityCounts (
    service_id INT UNSIGNED NOT NULL,
    critical_count INT DEFAULT 0,
    high_count INT DEFAULT 0,
    medium_count INT DEFAULT 0,
    low_count INT DEFAULT 0,
    none_count INT DEFAULT 0,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    PRIMARY KEY (service_id),
    KEY idx_mvAllComponents_active (is_active),
    CONSTRAINT fk_mvallcomponentsbyservicevulnerabilitycounts_service_id FOREIGN KEY (service_id) REFERENCES Service (service_id)
);

CREATE TABLE IF NOT EXISTS mvCountIssueRatingsUniqueService (
    id INT NOT NULL PRIMARY KEY,
    critical_count INT DEFAULT 0,
    high_count INT DEFAULT 0,
    medium_count INT DEFAULT 0,
    low_count INT DEFAULT 0,
    none_count INT DEFAULT 0,
    issue_count INT GENERATED ALWAYS AS (
        critical_count + high_count + medium_count + low_count + none_count
    ) STORED,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    INDEX idx_mvCountIssueRatingsUniqueService_active (is_active)
);

CREATE TABLE IF NOT EXISTS mvCountIssueRatingsService (
    supportgroup_ccrn VARCHAR(255) NOT NULL PRIMARY KEY,
    critical_count INT DEFAULT 0,
    high_count INT DEFAULT 0,
    medium_count INT DEFAULT 0,
    low_count INT DEFAULT 0,
    none_count INT DEFAULT 0,
    issue_count INT GENERATED ALWAYS AS (
        critical_count + high_count + medium_count + low_count + none_count
    ) STORED,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    INDEX idx_mvCountIssueRatingsService_active (is_active)
);

CREATE TABLE IF NOT EXISTS mvCountIssueRatingsServiceWithoutSupportGroup (
    id INT NOT NULL PRIMARY KEY,
    critical_count INT DEFAULT 0,
    high_count INT DEFAULT 0,
    medium_count INT DEFAULT 0,
    low_count INT DEFAULT 0,
    none_count INT DEFAULT 0,
    issue_count INT GENERATED ALWAYS AS (
        critical_count + high_count + medium_count + low_count + none_count
    ) STORED,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    INDEX idx_mvCountIssueRatingsServiceWithoutSupportGroup_active (is_active)
);

CREATE TABLE IF NOT EXISTS mvCountIssueRatingsSupportGroup (
    supportgroup_ccrn VARCHAR(255) NOT NULL PRIMARY KEY,
    critical_count INT DEFAULT 0,
    high_count INT DEFAULT 0,
    medium_count INT DEFAULT 0,
    low_count INT DEFAULT 0,
    none_count INT DEFAULT 0,
    issue_count INT GENERATED ALWAYS AS (
        critical_count + high_count + medium_count + low_count + none_count
    ) STORED,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    INDEX idx_mvCountIssueRatingsSupportGroup_active (is_active)
);

CREATE TABLE IF NOT EXISTS mvCountIssueRatingsComponentVersion (
    component_version_id INT UNSIGNED NOT NULL PRIMARY KEY,
    service_id INT DEFAULT NULL,
    service_ccrn VARCHAR(255) DEFAULT NULL,
    critical_count INT DEFAULT 0,
    high_count INT DEFAULT 0,
    medium_count INT DEFAULT 0,
    low_count INT DEFAULT 0,
    none_count INT DEFAULT 0,
    issue_count INT GENERATED ALWAYS AS (
        critical_count + high_count + medium_count + low_count + none_count
    ) STORED,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    INDEX idx_mvCountIssueRatingsComponentVersion_active (is_active)
);

CREATE TABLE IF NOT EXISTS mvCountIssueRatingsServiceId (
    service_id INT NOT NULL PRIMARY KEY,
    service_ccrn VARCHAR(255) DEFAULT NULL,
    critical_count INT DEFAULT 0,
    high_count INT DEFAULT 0,
    medium_count INT DEFAULT 0,
    low_count INT DEFAULT 0,
    none_count INT DEFAULT 0,
    issue_count INT GENERATED ALWAYS AS (
        critical_count + high_count + medium_count + low_count + none_count
    ) STORED,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    INDEX idx_mvCountIssueRatingsServiceId_active (is_active)
);

CREATE TABLE IF NOT EXISTS mvCountIssueRatingsOther (
    id INT NOT NULL PRIMARY KEY,
    critical_count INT DEFAULT 0,
    high_count INT DEFAULT 0,
    medium_count INT DEFAULT 0,
    low_count INT DEFAULT 0,
    none_count INT DEFAULT 0,
    issue_count INT GENERATED ALWAYS AS (
        critical_count + high_count + medium_count + low_count + none_count
    ) STORED,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    INDEX idx_mvCountIssueRatingsOther_active (is_active)
);

DROP PROCEDURE IF EXISTS call_registered_post_migration_procedures;

CREATE PROCEDURE call_registered_post_migration_procedures()
BEGIN
    DECLARE done INT DEFAULT 0;
    DECLARE pname VARCHAR(128);
    DECLARE cur CURSOR FOR SELECT name FROM post_migration_procedure_registry;
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = 1;

    OPEN cur;

    read_loop: LOOP
        FETCH cur INTO pname;
        IF done THEN
            LEAVE read_loop;
        END IF;

        SET @sql = CONCAT('CALL ', pname, '();');
        PREPARE stmt FROM @sql;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
    END LOOP;

    CLOSE cur;
END;

DROP PROCEDURE IF EXISTS add_post_migration_procedure;

CREATE PROCEDURE add_post_migration_procedure(IN p_name VARCHAR(128))
BEGIN
    INSERT IGNORE INTO post_migration_procedure_registry (name)
    VALUES (p_name);
END
;

DROP PROCEDURE IF EXISTS remove_post_migration_procedure;

CREATE PROCEDURE remove_post_migration_procedure(IN p_name VARCHAR(128))
BEGIN
    DELETE FROM post_migration_procedure_registry
    WHERE name = p_name;
END;

DROP PROCEDURE IF EXISTS refresh_mvServiceIssueCounts_proc;

CREATE PROCEDURE refresh_mvServiceIssueCounts_proc()
BEGIN
    CREATE TABLE IF NOT EXISTS mvServiceIssueCounts_tmp LIKE mvServiceIssueCounts;
    DELETE FROM mvServiceIssueCounts_tmp;

    INSERT INTO mvServiceIssueCounts_tmp
    SELECT
        S.service_id,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical'
            THEN CONCAT(CV.componentversion_id, ',', IV.issuevariant_issue_id) END) AS critical_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High'
            THEN CONCAT(CV.componentversion_id, ',', IV.issuevariant_issue_id) END) AS high_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium'
            THEN CONCAT(CV.componentversion_id, ',', IV.issuevariant_issue_id) END) AS medium_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low'
            THEN CONCAT(CV.componentversion_id, ',', IV.issuevariant_issue_id) END) AS low_count,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None'
            THEN CONCAT(CV.componentversion_id, ',', IV.issuevariant_issue_id) END) AS none_count
    FROM Service S
    LEFT JOIN ComponentInstance CI ON S.service_id = CI.componentinstance_service_id
    LEFT JOIN ComponentVersion CV ON CV.componentversion_id = CI.componentinstance_component_version_id
    LEFT JOIN ComponentVersionIssue CVI ON CV.componentversion_id = CVI.componentversionissue_component_version_id
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = CVI.componentversionissue_issue_id
    WHERE S.service_deleted_at IS NULL
    GROUP BY S.service_id;

    RENAME TABLE
        mvServiceIssueCounts TO mvServiceIssueCounts_old,
        mvServiceIssueCounts_tmp TO mvServiceIssueCounts;

    DROP TABLE mvServiceIssueCounts_old;
END;

DROP PROCEDURE IF EXISTS refresh_mvSingleComponentByServiceVulnerabilityCounts_proc;

CREATE PROCEDURE refresh_mvSingleComponentByServiceVulnerabilityCounts_proc()
BEGIN
    UPDATE mvSingleComponentByServiceVulnerabilityCounts
    SET is_active = 0;

    INSERT INTO mvSingleComponentByServiceVulnerabilityCounts (
        service_id,
        component_id,
        critical_count,
        high_count,
        medium_count,
        low_count,
        none_count,
        is_active
    )
    SELECT
        CI.componentinstance_service_id,
        CV.componentversion_component_id,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical'
            THEN CONCAT(CV.componentversion_component_id, ',', IV.issuevariant_issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High'
            THEN CONCAT(CV.componentversion_component_id, ',', IV.issuevariant_issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium'
            THEN CONCAT(CV.componentversion_component_id, ',', IV.issuevariant_issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low'
            THEN CONCAT(CV.componentversion_component_id, ',', IV.issuevariant_issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None'
            THEN CONCAT(CV.componentversion_component_id, ',', IV.issuevariant_issue_id) END),
        1
    FROM IssueMatch IM
    JOIN ComponentInstance CI ON CI.componentinstance_id = IM.issuematch_component_instance_id
    JOIN ComponentVersion CV ON CV.componentversion_id = CI.componentinstance_component_version_id
    JOIN IssueVariant IV ON IV.issuevariant_issue_id = IM.issuematch_issue_id
    JOIN Issue I ON I.issue_id = IV.issuevariant_issue_id
    LEFT JOIN Remediation R
        ON CI.componentinstance_service_id = R.remediation_service_id
       AND I.issue_id = R.remediation_issue_id
       AND R.remediation_deleted_at IS NULL
    WHERE IM.issuematch_status = 'new'
      AND I.issue_type = 'Vulnerability'
      AND IM.issuematch_deleted_at IS NULL
      AND I.issue_deleted_at IS NULL
      AND CI.componentinstance_deleted_at IS NULL
      AND CV.componentversion_deleted_at IS NULL
      AND (R.remediation_id IS NULL OR R.remediation_expiration_date < CURDATE())
    GROUP BY CI.componentinstance_service_id, CV.componentversion_component_id
    ON DUPLICATE KEY UPDATE
        critical_count = VALUES(critical_count),
        high_count     = VALUES(high_count),
        medium_count   = VALUES(medium_count),
        low_count      = VALUES(low_count),
        none_count     = VALUES(none_count),
        is_active      = 1;
END;

DROP PROCEDURE IF EXISTS refresh_mvAllComponentsByServiceVulnerabilityCounts_proc;

CREATE PROCEDURE refresh_mvAllComponentsByServiceVulnerabilityCounts_proc()
BEGIN
    UPDATE mvAllComponentsByServiceVulnerabilityCounts
    SET is_active = 0;

    INSERT INTO mvAllComponentsByServiceVulnerabilityCounts (
        service_id,
        critical_count,
        high_count,
        medium_count,
        low_count,
        none_count,
        is_active
    )
    SELECT
        CI.componentinstance_service_id,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical'
            THEN CONCAT(CV.componentversion_component_id, ',', IV.issuevariant_issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High'
            THEN CONCAT(CV.componentversion_component_id, ',', IV.issuevariant_issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium'
            THEN CONCAT(CV.componentversion_component_id, ',', IV.issuevariant_issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low'
            THEN CONCAT(CV.componentversion_component_id, ',', IV.issuevariant_issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None'
            THEN CONCAT(CV.componentversion_component_id, ',', IV.issuevariant_issue_id) END),
        1
    FROM IssueMatch IM
    JOIN ComponentInstance CI ON CI.componentinstance_id = IM.issuematch_component_instance_id
    JOIN ComponentVersion CV ON CV.componentversion_id = CI.componentinstance_component_version_id
    JOIN IssueVariant IV ON IV.issuevariant_issue_id = IM.issuematch_issue_id
    JOIN Issue I ON I.issue_id = IV.issuevariant_issue_id
    LEFT JOIN Remediation R
        ON CI.componentinstance_service_id = R.remediation_service_id
       AND I.issue_id = R.remediation_issue_id
       AND R.remediation_deleted_at IS NULL
    WHERE IM.issuematch_status = 'new'
      AND I.issue_type = 'Vulnerability'
      AND IM.issuematch_deleted_at IS NULL
      AND I.issue_deleted_at IS NULL
      AND CI.componentinstance_deleted_at IS NULL
      AND CV.componentversion_deleted_at IS NULL
      AND (R.remediation_id IS NULL OR R.remediation_expiration_date < CURDATE())
    GROUP BY CI.componentinstance_service_id
    ON DUPLICATE KEY UPDATE
        critical_count = VALUES(critical_count),
        high_count     = VALUES(high_count),
        medium_count   = VALUES(medium_count),
        low_count      = VALUES(low_count),
        none_count     = VALUES(none_count),
        is_active      = 1;
END;

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsUniqueService_proc;

CREATE PROCEDURE refresh_mvCountIssueRatingsUniqueService_proc()
BEGIN
    UPDATE mvCountIssueRatingsUniqueService
    SET is_active = 0
    WHERE is_active = 1;

    INSERT INTO mvCountIssueRatingsUniqueService (
        id,
        critical_count,
        high_count,
        medium_count,
        low_count,
        none_count,
        is_active
    )
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
        is_active      = 1;

    DELETE FROM mvCountIssueRatingsUniqueService
    WHERE is_active = 0;
END;

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsService_proc;

CREATE PROCEDURE refresh_mvCountIssueRatingsService_proc()
BEGIN
    UPDATE mvCountIssueRatingsService
    SET is_active = 0
    WHERE is_active = 1;

    INSERT INTO mvCountIssueRatingsService (
        supportgroup_ccrn,
        critical_count,
        high_count,
        medium_count,
        low_count,
        none_count,
        is_active
    )
    SELECT
        COALESCE(SG.supportgroup_ccrn, 'UNKNOWN'),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END),
        1
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    RIGHT JOIN IssueMatch IM ON I.issue_id = IM.issuematch_issue_id
    LEFT JOIN ComponentInstance CI ON CI.componentinstance_id = IM.issuematch_component_instance_id
    LEFT JOIN Service S ON S.service_id = CI.componentinstance_service_id
    LEFT JOIN SupportGroupService SGS ON SGS.supportgroupservice_service_id = CI.componentinstance_service_id
    LEFT JOIN SupportGroup SG ON SGS.supportgroupservice_support_group_id = SG.supportgroup_id
    LEFT JOIN Remediation R
        ON S.service_id = R.remediation_service_id
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
        is_active      = 1;

    DELETE FROM mvCountIssueRatingsService
    WHERE is_active = 0;
END;

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsServiceWithoutSupportGroup_proc;

CREATE PROCEDURE refresh_mvCountIssueRatingsServiceWithoutSupportGroup_proc()
BEGIN
    UPDATE mvCountIssueRatingsServiceWithoutSupportGroup
    SET is_active = 0
    WHERE is_active = 1;

    INSERT INTO mvCountIssueRatingsServiceWithoutSupportGroup (
        id,
        critical_count,
        high_count,
        medium_count,
        low_count,
        none_count,
        is_active
    )
    SELECT
        1,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id) END),
        1
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    RIGHT JOIN IssueMatch IM ON I.issue_id = IM.issuematch_issue_id
    LEFT JOIN ComponentInstance CI ON CI.componentinstance_id = IM.issuematch_component_instance_id
    LEFT JOIN Service S ON S.service_id = CI.componentinstance_service_id
    LEFT JOIN Remediation R
        ON S.service_id = R.remediation_service_id
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
        is_active      = 1;

    DELETE FROM mvCountIssueRatingsServiceWithoutSupportGroup
    WHERE is_active = 0;
END;

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsSupportGroup_proc;

CREATE PROCEDURE refresh_mvCountIssueRatingsSupportGroup_proc()
BEGIN
    UPDATE mvCountIssueRatingsSupportGroup
    SET is_active = 0
    WHERE is_active = 1;

    INSERT INTO mvCountIssueRatingsSupportGroup (
        supportgroup_ccrn,
        critical_count,
        high_count,
        medium_count,
        low_count,
        none_count,
        is_active
    )
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
    LEFT JOIN Remediation R
        ON SGS.supportgroupservice_service_id = R.remediation_service_id
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
        is_active      = 1;

    DELETE FROM mvCountIssueRatingsSupportGroup
    WHERE is_active = 0;
END;

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsComponentVersion_proc;

CREATE PROCEDURE refresh_mvCountIssueRatingsComponentVersion_proc()
BEGIN
    UPDATE mvCountIssueRatingsComponentVersion
    SET is_active = 0
    WHERE is_active = 1;

    INSERT INTO mvCountIssueRatingsComponentVersion (
        component_version_id,
        service_id,
        service_ccrn,
        critical_count,
        high_count,
        medium_count,
        low_count,
        none_count,
        is_active
    )
    SELECT
        CVI.componentversionissue_component_version_id,
        CI.componentinstance_service_id,
        S.service_ccrn,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical'
            THEN CONCAT(CVI.componentversionissue_component_version_id, ',', CVI.componentversionissue_issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High'
            THEN CONCAT(CVI.componentversionissue_component_version_id, ',', CVI.componentversionissue_issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium'
            THEN CONCAT(CVI.componentversionissue_component_version_id, ',', CVI.componentversionissue_issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low'
            THEN CONCAT(CVI.componentversionissue_component_version_id, ',', CVI.componentversionissue_issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None'
            THEN CONCAT(CVI.componentversionissue_component_version_id, ',', CVI.componentversionissue_issue_id) END),
        1
    FROM ComponentVersionIssue CVI
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = CVI.componentversionissue_issue_id
    INNER JOIN ComponentInstance CI ON CVI.componentversionissue_component_version_id = CI.componentinstance_component_version_id
    INNER JOIN Service S ON CI.componentinstance_service_id = S.service_id
    LEFT JOIN Remediation R
        ON CI.componentinstance_service_id = R.remediation_service_id
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
        is_active      = 1;

    DELETE FROM mvCountIssueRatingsComponentVersion
    WHERE is_active = 0;
END;

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsServiceId_proc;

CREATE PROCEDURE refresh_mvCountIssueRatingsServiceId_proc()
BEGIN
    UPDATE mvCountIssueRatingsServiceId
    SET is_active = 0
    WHERE is_active = 1;

    INSERT INTO mvCountIssueRatingsServiceId (
        service_id,
        service_ccrn,
        critical_count,
        high_count,
        medium_count,
        low_count,
        none_count,
        is_active
    )
    SELECT
        CI.componentinstance_service_id,
        S.service_ccrn,
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Critical'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'High'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Medium'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'Low'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id) END),
        COUNT(DISTINCT CASE WHEN IV.issuevariant_rating = 'None'
            THEN CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id) END),
        1
    FROM Issue I
    LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id
    LEFT JOIN IssueMatch IM ON I.issue_id = IM.issuematch_issue_id
    LEFT JOIN ComponentInstance CI
        ON CI.componentinstance_id = IM.issuematch_component_instance_id
       AND CI.componentinstance_deleted_at IS NULL
    LEFT JOIN Service S ON S.service_id = CI.componentinstance_service_id
    LEFT JOIN Remediation R
        ON CI.componentinstance_service_id = R.remediation_service_id
       AND I.issue_id = R.remediation_issue_id
       AND R.remediation_deleted_at IS NULL
    WHERE I.issue_deleted_at IS NULL
      AND IM.issuematch_id IS NOT NULL
      AND (R.remediation_id IS NULL OR R.remediation_expiration_date < CURDATE())
    GROUP BY CI.componentinstance_service_id
    ON DUPLICATE KEY UPDATE
        service_ccrn   = VALUES(service_ccrn),
        critical_count = VALUES(critical_count),
        high_count     = VALUES(high_count),
        medium_count   = VALUES(medium_count),
        low_count      = VALUES(low_count),
        none_count     = VALUES(none_count),
        is_active      = 1;

    DELETE FROM mvCountIssueRatingsServiceId
    WHERE is_active = 0;
END;

DROP PROCEDURE IF EXISTS refresh_mvCountIssueRatingsOther_proc;

CREATE PROCEDURE refresh_mvCountIssueRatingsOther_proc()
BEGIN
    UPDATE mvCountIssueRatingsOther
    SET is_active = 0
    WHERE is_active = 1;

    INSERT INTO mvCountIssueRatingsOther (
        id,
        critical_count,
        high_count,
        medium_count,
        low_count,
        none_count,
        is_active
    )
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
        is_active      = 1;

    DELETE FROM mvCountIssueRatingsOther
    WHERE is_active = 0;
END;

INSERT IGNORE INTO
    post_migration_procedure_registry (name)
VALUES (
        'refresh_mvServiceIssueCounts_proc'
    ),
    (
        'refresh_mvSingleComponentByServiceVulnerabilityCounts_proc'
    ),
    (
        'refresh_mvAllComponentsByServiceVulnerabilityCounts_proc'
    ),
    (
        'refresh_mvCountIssueRatingsUniqueService_proc'
    ),
    (
        'refresh_mvCountIssueRatingsService_proc'
    ),
    (
        'refresh_mvCountIssueRatingsServiceWithoutSupportGroup_proc'
    ),
    (
        'refresh_mvCountIssueRatingsSupportGroup_proc'
    ),
    (
        'refresh_mvCountIssueRatingsComponentVersion_proc'
    ),
    (
        'refresh_mvCountIssueRatingsServiceId_proc'
    ),
    (
        'refresh_mvCountIssueRatingsOther_proc'
    );

--
-- SET GLOBAL event_scheduler = ON;
--

DROP EVENT IF EXISTS refresh_mvServiceIssueCounts;

CREATE EVENT refresh_mvServiceIssueCounts
ON SCHEDULE EVERY 2 HOUR
DO
    CALL refresh_mvServiceIssueCounts_proc();

DROP EVENT IF EXISTS refresh_mvSingleComponentByServiceVulnerabilityCounts;

CREATE EVENT refresh_mvSingleComponentByServiceVulnerabilityCounts
ON SCHEDULE EVERY 2 HOUR
DO
    CALL refresh_mvSingleComponentByServiceVulnerabilityCounts_proc();

DROP EVENT IF EXISTS refresh_mvAllComponentsByServiceVulnerabilityCounts;

CREATE EVENT refresh_mvAllComponentsByServiceVulnerabilityCounts
ON SCHEDULE EVERY 2 HOUR
DO
    CALL refresh_mvAllComponentsByServiceVulnerabilityCounts_proc();

DROP EVENT IF EXISTS refresh_mvCountIssueRatingsUniqueService;

CREATE EVENT refresh_mvCountIssueRatingsUniqueService
ON SCHEDULE EVERY 2 HOUR
DO
    CALL refresh_mvCountIssueRatingsUniqueService_proc();

DROP EVENT IF EXISTS refresh_mvCountIssueRatingsService;

CREATE EVENT refresh_mvCountIssueRatingsService
ON SCHEDULE EVERY 2 HOUR
DO
    CALL refresh_mvCountIssueRatingsService_proc();

DROP EVENT IF EXISTS refresh_mvCountIssueRatingsServiceWithoutSupportGroup;

CREATE EVENT refresh_mvCountIssueRatingsServiceWithoutSupportGroup
ON SCHEDULE EVERY 2 HOUR
DO
    CALL refresh_mvCountIssueRatingsServiceWithoutSupportGroup_proc();

DROP EVENT IF EXISTS refresh_mvCountIssueRatingsSupportGroup;

CREATE EVENT refresh_mvCountIssueRatingsSupportGroup
ON SCHEDULE EVERY 2 HOUR
DO
    CALL refresh_mvCountIssueRatingsSupportGroup_proc();

DROP EVENT IF EXISTS refresh_mvCountIssueRatingsComponentVersion;

CREATE EVENT refresh_mvCountIssueRatingsComponentVersion
ON SCHEDULE EVERY 2 HOUR
DO
    CALL refresh_mvCountIssueRatingsComponentVersion_proc();

DROP EVENT IF EXISTS refresh_mvCountIssueRatingsServiceId;

CREATE EVENT refresh_mvCountIssueRatingsServiceId
ON SCHEDULE EVERY 2 HOUR
DO
    CALL refresh_mvCountIssueRatingsServiceId_proc();

DROP EVENT IF EXISTS refresh_mvCountIssueRatingsOther;

CREATE EVENT refresh_mvCountIssueRatingsOther
ON SCHEDULE EVERY 2 HOUR
DO
    CALL refresh_mvCountIssueRatingsOther_proc();
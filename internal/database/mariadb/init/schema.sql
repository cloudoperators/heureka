--  SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
--  SPDX-License-Identifier: Apache-2.0

create schema if not exists heureka;

use heureka;

create table if not exists User
(
    user_id         int unsigned auto_increment
        primary key,
    user_name           varchar(255)                          not null,
    user_unique_user_id varchar(64)                           not null,
    user_type           int unsigned,
    user_created_at     timestamp default current_timestamp() not null,
    user_created_by     int unsigned                          null check (user_created_by <> 0),
    user_deleted_at     timestamp                             null,
    user_updated_at     timestamp default current_timestamp() not null on update current_timestamp(),
    user_updated_by     int unsigned                          null check (user_updated_by <> 0),
    constraint user_id_UNIQUE
        unique (user_id),
    constraint unique_user_id_UNIQUE
        unique (user_unique_user_id),
    constraint fk_user_created_by
        foreign key (user_created_by) references User (user_id),
    constraint fk_user_updated_by
        foreign key (user_updated_by) references User (user_id)
);

set @TechnicalUserType = 2;
set @SystemUserId = 1;
set @SystemUserName = 'systemuser';
set @SystemUserUniqueUserId = 'S0000000';
insert ignore into User (user_id, user_name, user_unique_user_id,  user_type, user_created_at, user_created_by)
    values
    (@SystemUserId, @SystemUserName, @SystemUserUniqueUserId, @TechnicalUserType, current_timestamp(), @SystemUserId);

create table if not exists Component
(
    component_id         int unsigned auto_increment
        primary key,
    component_ccrn       varchar(255)                          not null,
    component_type       varchar(255)                          not null,
    component_created_at timestamp default current_timestamp() not null,
    component_created_by int unsigned                          null,
    component_deleted_at timestamp                             null,
    component_updated_at timestamp default current_timestamp() not null on update current_timestamp(),
    component_updated_by int unsigned                          null,
    constraint id_UNIQUE
        unique (component_id),
    constraint component_ccrn_UNIQUE
        unique (component_ccrn),
    constraint fk_component_created_by
        foreign key (component_created_by) references User (user_id),
    constraint fk_component_updated_by
        foreign key (component_updated_by) references User (user_id)
);

create table if not exists ComponentVersion
(
    componentversion_id           int unsigned auto_increment
        primary key,
    componentversion_version      varchar(255)                          not null,
    componentversion_component_id int unsigned                          not null,
    componentversion_tag          varchar(255)                          null,
    componentversion_repository   varchar(255)                          null,
    componentversion_organization varchar(255)                          null,
    componentversion_created_at   timestamp default current_timestamp() not null,
    componentversion_created_by   int unsigned                          null,
    componentversion_deleted_at   timestamp                             null,
    componentversion_updated_at   timestamp default current_timestamp() not null on update current_timestamp(),
    componentversion_updated_by   int unsigned                          null,
    constraint componentversion_id_UNIQUE
        unique (componentversion_id),
    constraint version_component_UNIQUE
        unique (componentversion_version, componentversion_component_id),
    constraint fk_component_version_component
        foreign key (componentversion_component_id) references Component (component_id)
            on update cascade,
    constraint fk_componentversion_created_by
        foreign key (componentversion_created_by) references User (user_id),
    constraint fk_componentversion_updated_by
        foreign key (componentversion_updated_by) references User (user_id)
);

create table if not exists SupportGroup
(
    supportgroup_id         int unsigned auto_increment
        primary key,
    supportgroup_ccrn       varchar(255)                          not null,
    supportgroup_created_at timestamp default current_timestamp() not null,
    supportgroup_created_by int unsigned                          null,
    supportgroup_deleted_at timestamp                             null,
    supportgroup_updated_at timestamp default current_timestamp() not null on update current_timestamp(),
    supportgroup_updated_by int unsigned                          null,
    constraint supportgroup_id_UNIQUE
        unique (supportgroup_id),
    constraint supportgroup_ccrn_UNIQUE
        unique (supportgroup_ccrn),
    constraint fk_supportgroup_created_by
        foreign key (supportgroup_created_by) references User (user_id),
    constraint fk_supportgroup_updated_by
        foreign key (supportgroup_updated_by) references User (user_id)
);

create table if not exists Service
(
    service_id         int unsigned auto_increment
        primary key,
    service_ccrn       varchar(255)                          not null,
    service_created_at timestamp default current_timestamp() not null,
    service_created_by int unsigned                          null,
    service_deleted_at timestamp                             null,
    service_updated_at timestamp default current_timestamp() not null on update current_timestamp(),
    service_updated_by int unsigned                          null,
    constraint service_id_UNIQUE
        unique (service_id),
    constraint service_ccrn_UNIQUE
        unique (service_ccrn),
    constraint fk_service_created_by
        foreign key (service_created_by) references User (user_id),
    constraint fk_service_updated_by
        foreign key (service_updated_by) references User (user_id)
);

create table if not exists SupportGroupService
(
    supportgroupservice_service_id       int unsigned                          not null,
    supportgroupservice_support_group_id int unsigned                          not null,
    supportgroupservice_created_at       timestamp default current_timestamp() not null,
    supportgroupservice_deleted_at       timestamp                             null,
    supportgroupservice_updated_at       timestamp default current_timestamp() not null on update current_timestamp(),
    primary key (supportgroupservice_service_id, supportgroupservice_support_group_id),
    constraint fk_support_group_service
        foreign key (supportgroupservice_support_group_id) references SupportGroup (supportgroup_id)
            on update cascade,
    constraint fk_service_support_group
        foreign key (supportgroupservice_service_id) references Service (service_id)
            on update cascade
);

create table if not exists Activity
(
    activity_id         int unsigned auto_increment
        primary key,
    activity_status     enum ('open','closed','in_progress')  not null,
    activity_created_at timestamp default current_timestamp() not null,
    activity_created_by int unsigned                          null,
    activity_deleted_at timestamp                             null,
    activity_updated_at timestamp default current_timestamp() not null on update current_timestamp(),
    activity_updated_by int unsigned                          null,
    constraint activity_id_UNIQUE
        unique (activity_id),
    constraint fk_activity_created_by
        foreign key (activity_created_by) references User (user_id),
    constraint fk_activity_updated_by
        foreign key (activity_updated_by) references User (user_id)
);

create table if not exists ComponentInstance
(
    componentinstance_id                   int unsigned auto_increment
        primary key,
    componentinstance_ccrn                 varchar(2048)                         not null,
    componentinstance_region               varchar(2048)                         null,
    componentinstance_cluster              varchar(2048)                         null,
    componentinstance_namespace            varchar(2048)                         null,
    componentinstance_domain               varchar(2048)                         null,
    componentinstance_project              varchar(2048)                         null,
    componentinstance_count                int       default 0                   not null,
    componentinstance_component_version_id int unsigned                          not null,
    componentinstance_service_id           int unsigned                          not null,
    componentinstance_created_at           timestamp default current_timestamp() not null,
    componentinstance_created_by           int unsigned                          null,
    componentinstance_deleted_at           timestamp                             null,
    componentinstance_updated_at           timestamp default current_timestamp() not null on update current_timestamp(),
    componentinstance_updated_by           int unsigned                          null,
    constraint componentinstance_id_UNIQUE
        unique (componentinstance_id),
    constraint componentinstance_ccrn_service_id_unique
        unique (componentinstance_ccrn, componentinstance_service_id) using hash,
    constraint fk_component_instance_component_version
        foreign key (componentinstance_component_version_id) references ComponentVersion (componentversion_id)
            on update cascade,
    constraint fk_component_instance_service
        foreign key (componentinstance_service_id) references Service (service_id)
            on update cascade,
    constraint fk_componentinstance_created_by
        foreign key (componentinstance_created_by) references User (user_id),
    constraint fk_componentinstance_updated_by
        foreign key (componentinstance_updated_by) references User (user_id)
);

create table if not exists Evidence
(
    evidence_id          int unsigned auto_increment
        primary key,
    evidence_author_id   int unsigned                                                                         not null,
    evidence_activity_id int unsigned                                                                         not null,
    evidence_type        enum ('risk_accepted','mitigated','severity_adjustment', 'false_positive', 'reopen') not null,
    evidence_description longtext                                                                             not null,
    evidence_vector      varchar(512)                                                                         null,
    evidence_rating      enum ('None','Low','Medium','High','Critical')                                       null,
    evidence_raa_end     datetime                                                                             null,
    evidence_created_at  timestamp default current_timestamp()                                                not null,
    evidence_created_by  int unsigned                                                                         null,
    evidence_deleted_at  timestamp                                                                            null,
    evidence_updated_at  timestamp default current_timestamp()                                                not null on update current_timestamp(),
    evidence_updated_by  int unsigned                                                                         null,
    constraint evidence_id_UNIQUE
        unique (evidence_id),
    constraint fk_evidence_user
        foreign key (evidence_author_id) references User (user_id)
            on update cascade,
    constraint fk_evidience_activity
        foreign key (evidence_activity_id) references Activity (activity_id)
            on update cascade,
    constraint fk_evidence_created_by
        foreign key (evidence_created_by) references User (user_id),
    constraint fk_evidence_updated_by
        foreign key (evidence_updated_by) references User (user_id)
);


create table if not exists Owner
(
    owner_service_id int unsigned                          not null,
    owner_user_id    int unsigned                          not null,
    owner_created_at timestamp default current_timestamp() not null,
    owner_deleted_at timestamp                             null,
    owner_updated_at timestamp default current_timestamp() not null on update current_timestamp(),
    primary key (owner_service_id, owner_user_id),
    constraint fk_service_user
        foreign key (owner_service_id) references Service (service_id)
            on update cascade,
    constraint fk_user_service
        foreign key (owner_user_id) references User (user_id)
            on update cascade
);



create table if not exists SupportGroupUser
(
    supportgroupuser_user_id          int unsigned                          not null,
    supportgroupuser_support_group_id int unsigned                          not null,
    supportgroupuser_created_at       timestamp default current_timestamp() not null,
    supportgroupuser_deleted_at       timestamp                             null,
    supportgroupuser_updated_at       timestamp default current_timestamp() not null on update current_timestamp(),
    primary key (supportgroupuser_user_id, supportgroupuser_support_group_id),
    constraint fk_support_group_user
        foreign key (supportgroupuser_support_group_id) references SupportGroup (supportgroup_id)
            on update cascade,
    constraint fk_user_support_group
        foreign key (supportgroupuser_user_id) references User (user_id)
            on update cascade
);


create table if not exists IssueRepository
(
    issuerepository_id         int unsigned auto_increment primary key,
    issuerepository_name       varchar(2048)                         not null,
    issuerepository_url        varchar(2048)                         not null,
    issuerepository_created_at timestamp default current_timestamp() not null,
    issuerepository_created_by int unsigned                          null,
    issuerepository_deleted_at timestamp                             null,
    issuerepository_updated_at timestamp default current_timestamp() not null on update current_timestamp(),
    issuerepository_updated_by int unsigned                          null,
    constraint issuerepository_id_UNIQUE
        unique (issuerepository_id),
    constraint issuerepository_name_UNIQUE
        unique (issuerepository_name),
    constraint fk_issuerepository_created_by
        foreign key (issuerepository_created_by) references User (user_id),
    constraint fk_issuerepository_updated_by
        foreign key (issuerepository_updated_by) references User (user_id)
);

create table if not exists Issue
(
    issue_id           int unsigned auto_increment
        primary key,
    issue_type         enum ('Vulnerability','PolicyViolation','SecurityEvent') not null,
    issue_primary_name varchar(255)                                             not null,
    issue_description  longtext                                                 not null,
    issue_created_at   timestamp default current_timestamp()                    not null,
    issue_created_by   int unsigned                                             null,
    issue_deleted_at   timestamp                                                null,
    issue_updated_at   timestamp default current_timestamp()                    not null on update current_timestamp(),
    issue_updated_by   int unsigned                                             null,
    constraint issue_id_UNIQUE
        unique (issue_id),
    constraint issue_primary_name_UNIQUE
        unique (issue_primary_name),
    constraint fk_issue_created_by
        foreign key (issue_created_by) references User (user_id),
    constraint fk_issue_updated_by
        foreign key (issue_updated_by) references User (user_id)
);

create table if not exists IssueVariant
(
    issuevariant_id             int unsigned auto_increment
        primary key,
    issuevariant_issue_id       int unsigned                                     not null,
    issuevariant_repository_id  int unsigned                                     not null,
    issuevariant_vector         varchar(512)                                     null,
    issuevariant_rating         enum ('None','Low','Medium', 'High', 'Critical') not null,
    issuevariant_secondary_name varchar(255)                                     not null,
    issuevariant_description    longtext                                         not null,
    issuevariant_external_url   varchar(512)                                     null,
    issuevariant_created_at     timestamp default current_timestamp()            not null,
    issuevariant_created_by     int unsigned                                     null,
    issuevariant_deleted_at     timestamp                                        null,
    issuevariant_updated_at     timestamp default current_timestamp()            not null on update current_timestamp(),
    issuevariant_updated_by     int unsigned                                     null,
    constraint issuevariant_id_UNIQUE
        unique (issuevariant_id),
    constraint issuevariant_secondary_name_UNIQUE
        unique (issuevariant_secondary_name),
    constraint fk_issuevariant_issue
        foreign key (issuevariant_issue_id) references Issue (issue_id),
    constraint fk_issuevariant_issuerepository
        foreign key (issuevariant_repository_id) references IssueRepository (issuerepository_id)
            on update cascade,
    constraint fk_issuevariant_created_by
        foreign key (issuevariant_created_by) references User (user_id),
    constraint fk_issuevariant_updated_by
        foreign key (issuevariant_updated_by) references User (user_id)
);

create table if not exists ActivityHasIssue
(
    activityhasissue_activity_id int unsigned                          not null,
    activityhasissue_issue_id    int unsigned                          not null,
    activityhasissue_created_at  timestamp default current_timestamp() not null,
    activityhasissue_deleted_at  timestamp                             null,
    activityhasissue_updated_at  timestamp default current_timestamp() not null on update current_timestamp(),
    primary key (activityhasissue_activity_id, activityhasissue_issue_id),
    constraint fk_activity_issue
        foreign key (activityhasissue_issue_id) references Issue (issue_id),
    constraint fk_issue_activity
        foreign key (activityhasissue_activity_id) references Activity (activity_id)
);

create table if not exists ActivityHasService
(
    activityhasservice_activity_id int unsigned                          not null,
    activityhasservice_service_id  int unsigned                          not null,
    activityhasservice_created_at  timestamp default current_timestamp() not null,
    activityhasservice_deleted_at  timestamp                             null,
    activityhasservice_updated_at  timestamp default current_timestamp() not null on update current_timestamp(),
    primary key (activityhasservice_activity_id, activityhasservice_service_id),
    constraint fk_activity_service
        foreign key (activityhasservice_service_id) references Service (service_id),
    constraint fk_service_activity
        foreign key (activityhasservice_activity_id) references Activity (activity_id)
);

create table if not exists IssueMatch
(
    issuematch_id                      int unsigned auto_increment
        primary key,
    issuematch_status                  varchar(128)                                     not null,
    issuematch_remediation_date        datetime                                         null,
    issuematch_target_remediation_date datetime                                         not null,

    issuematch_vector                  varchar(512)                                     null,
    issuematch_rating                  enum ('None','Low','Medium', 'High', 'Critical') not null,

    issuematch_user_id                 int unsigned                                     not null,
    issuematch_issue_id                int unsigned                                     not null,
    issuematch_component_instance_id   int unsigned                                     not null,

    issuematch_created_at              timestamp default current_timestamp()            not null,
    issuematch_created_by              int unsigned                                     null,
    issuematch_deleted_at              timestamp                                        null,
    issuematch_updated_at              timestamp default current_timestamp()            not null on update current_timestamp(),
    issuematch_updated_by              int unsigned                                     null,
    constraint issuematch_id_UNIQUE
        unique (issuematch_id),
    constraint fk_issue_match_user_id
        foreign key (issuematch_user_id) references User (user_id)
            on update cascade,
    constraint fk_issue_match_component_instance
        foreign key (issuematch_component_instance_id) references ComponentInstance (componentinstance_id)
            on update cascade,
    constraint fk_issue_match_issue
        foreign key (issuematch_issue_id) references Issue (issue_id)
            on update cascade,
    constraint fk_issuematch_created_by
        foreign key (issuematch_created_by) references User (user_id),
    constraint fk_issuematch_updated_by
        foreign key (issuematch_updated_by) references User (user_id)
);

create table if not exists IssueMatchChange
(
    issuematchchange_id             int unsigned auto_increment
        primary key,
    issuematchchange_activity_id    int unsigned                          not null,
    issuematchchange_issue_match_id int unsigned                          not null,
    issuematchchange_action         enum ('add','remove')                 not null,
    issuematchchange_created_at     timestamp default current_timestamp() not null,
    issuematchchange_created_by     int unsigned                          null,
    issuematchchange_deleted_at     timestamp                             null,
    issuematchchange_updated_at     timestamp default current_timestamp() not null on update current_timestamp(),
    issuematchchange_updated_by     int unsigned                          null,
    constraint fk_issuematchchange_activity
        foreign key (issuematchchange_activity_id) references Activity (activity_id),
    constraint fk_issuematchchange_issue_match
        foreign key (issuematchchange_issue_match_id) references IssueMatch (issuematch_id),
    constraint fk_issuematchchange_created_by
        foreign key (issuematchchange_created_by) references User (user_id),
    constraint fk_issuematchchange_updated_by
        foreign key (issuematchchange_updated_by) references User (user_id)
);

create table if not exists ComponentVersionIssue
(
    componentversionissue_component_version_id int unsigned                          not null,
    componentversionissue_issue_id             int unsigned                          not null,
    componentversionissue_created_at           timestamp default current_timestamp() not null,
    componentversionissue_deleted_at           timestamp                             null,
    componentversionissue_updated_at           timestamp default current_timestamp() not null on update current_timestamp(),
    primary key (componentversionissue_component_version_id, componentversionissue_issue_id),
    constraint fk_componentversionissue_issue
        foreign key (componentversionissue_issue_id) references Issue (issue_id),
    constraint fk_componentversionissue_component_version
        foreign key (componentversionissue_component_version_id) references ComponentVersion (componentversion_id)
);


create table if not exists IssueMatchEvidence
(
    issuematchevidence_evidence_id    int unsigned                          not null,
    issuematchevidence_issue_match_id int unsigned                          not null,
    issuematchevidence_created_at     timestamp default current_timestamp() not null,
    issuematchevidence_deleted_at     timestamp                             null,
    issuematchevidence_updated_at     timestamp default current_timestamp() not null on update current_timestamp(),
    primary key (issuematchevidence_evidence_id, issuematchevidence_issue_match_id),
    constraint fk_issue_match_evidence
        foreign key (issuematchevidence_evidence_id) references Evidence (evidence_id),
    constraint fk_evidence_issue_match
        foreign key (issuematchevidence_issue_match_id) references IssueMatch (issuematch_id)
);

create table if not exists IssueRepositoryService
(
    issuerepositoryservice_service_id          int unsigned                          not null,
    issuerepositoryservice_issue_repository_id int unsigned                          not null,
    issuerepositoryservice_priority            int unsigned                          not null,
    primary key (issuerepositoryservice_service_id, issuerepositoryservice_issue_repository_id),
    constraint fk_issue_repository_service
        foreign key (issuerepositoryservice_issue_repository_id) references IssueRepository (issuerepository_id)
            on update cascade,
    constraint fk_service_issue_repository
        foreign key (issuerepositoryservice_service_id) references Service (service_id)
            on update cascade
);



create table if not exists ScannerRun
(
    scannerrun_run_id          int unsigned primary key auto_increment,
    scannerrun_uuid            UUID not null unique,
    scannerrun_tag             varchar(255) not null,
    scannerrun_start_run       timestamp default current_timestamp() not null,
    
    scannerrun_end_run         timestamp default current_timestamp() not null,

    scannerrun_is_completed    boolean not null default false
);

create table if not exists ScannerRunIssueTracker
(
    scannerrunissuetracker_scannerrun_run_id int unsigned not null,
    scannerrunissuetracker_issue_id  int unsigned not null,

    constraint fk_srit_sr_id foreign key (scannerrunissuetracker_scannerrun_run_id) references ScannerRun (scannerrun_run_id) on update cascade,
    constraint fk_srit_i_id foreign key (scannerrunissuetracker_issue_id) references Issue (issue_id) on update cascade
);


create table if not exists ScannerRunError
(
    scannerrunerror_scannerrun_run_id int unsigned not null,
    error                             text not null,

    constraint fk_sre_sr_id foreign key (scannerrunerror_scannerrun_run_id) references ScannerRun (scannerrun_run_id) on update cascade
);

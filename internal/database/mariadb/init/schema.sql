--  SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
--  SPDX-License-Identifier: Apache-2.0

create schema if not exists heureka;

use heureka;

create table if not exists Component
(
    component_id         int unsigned auto_increment
        primary key,
    component_name       varchar(256)                          not null,
    component_type       varchar(256)                          not null,
    component_created_at timestamp default current_timestamp() not null,
    component_deleted_at timestamp                             null,
    component_updated_at timestamp default current_timestamp() not null on update current_timestamp(),
    constraint id_UNIQUE
        unique (component_id),
    constraint name_UNIQUE
        unique (component_name)
);

create table if not exists ComponentVersion
(
    componentversion_id           int unsigned auto_increment
        primary key,
    componentversion_version      varchar(256)                          not null,
    componentversion_component_id int unsigned                          not null,
    componentversion_created_at   timestamp default current_timestamp() not null,
    componentversion_deleted_at   timestamp                             null,
    componentversion_updated_at   timestamp default current_timestamp() not null on update current_timestamp(),
    constraint id_UNIQUE
        unique (componentversion_id),
    constraint version_component_unique
        unique (componentversion_version, componentversion_component_id),
    constraint fk_component_version_component
        foreign key (componentversion_component_id) references Component (component_id)
            on update cascade
);

create table if not exists SupportGroup
(
    supportgroup_id         int unsigned auto_increment
        primary key,
    supportgroup_name       varchar(256)                          not null,
    supportgroup_created_at timestamp default current_timestamp() not null,
    supportgroup_deleted_at timestamp                             null,
    supportgroup_updated_at timestamp default current_timestamp() not null on update current_timestamp(),
    constraint id_UNIQUE
        unique (supportgroup_id),
    constraint name_UNIQUE
        unique (supportgroup_name)
);

create table if not exists Service
(
    service_id         int unsigned auto_increment
        primary key,
    service_name       varchar(256)                          not null,
    service_created_at timestamp default current_timestamp() not null,
    service_deleted_at timestamp                             null,
    service_updated_at timestamp default current_timestamp() not null on update current_timestamp(),
    constraint id_UNIQUE
        unique (service_id),
    constraint name_UNIQUE
        unique (service_name)
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
    activity_status enum('open','closed','in_progress') not null,
    activity_created_at timestamp default current_timestamp() not null,
    activity_deleted_at timestamp                             null,
    activity_updated_at timestamp default current_timestamp() not null on update current_timestamp(),
    constraint id_UNIQUE
        unique (activity_id)
);


create table if not exists ComponentInstance
(
    componentinstance_id                   int unsigned auto_increment
        primary key,
    componentinstance_ccrn                 varchar(2048)                         not null,
    componentinstance_count                int       default 0                   not null,
    componentinstance_component_version_id int unsigned                          not null,
    componentinstance_service_id           int unsigned                          not null,
    componentinstance_created_at           timestamp default current_timestamp() not null,
    componentinstance_deleted_at           timestamp                             null,
    componentinstance_updated_at           timestamp default current_timestamp() not null on update current_timestamp(),
    constraint id_UNIQUE
        unique (componentinstance_id),
    constraint name_service_unique
        unique (componentinstance_ccrn, componentinstance_service_id) using hash,
    constraint fk_component_instance_component_version
        foreign key (componentinstance_component_version_id) references ComponentVersion (componentversion_id)
            on update cascade,
    constraint fk_component_instance_service
        foreign key (componentinstance_service_id) references Service (service_id)
            on update cascade
);

create table if not exists User
(
    user_id         int unsigned auto_increment
        primary key,
    user_name       varchar(256)                          not null,
    user_sapID      varchar(64)                           not null,
    user_created_at timestamp default current_timestamp() not null,
    user_deleted_at timestamp                             null,
    user_updated_at timestamp default current_timestamp() not null on update current_timestamp(),
    constraint id_UNIQUE
        unique (user_id),
    constraint sapID_UNIQUE
        unique (user_sapID)
);

create table if not exists Evidence
(
    evidence_id          int unsigned auto_increment
        primary key,
    evidence_author_id   int unsigned                          not null,
    evidence_activity_id int unsigned                          not null,
    evidence_type        enum('risk_accepted','mitigated','severity_adjustment', 'false_positive', 'reopen') not null,
    evidence_description longtext                              not null,
    evidence_vector      varchar(512)                          null,
    evidence_rating      enum('None','Low','Medium','High','Critical') null,
    evidence_raa_end     datetime                              null,
    evidence_created_at  timestamp default current_timestamp() not null,
    evidence_deleted_at  timestamp                             null,
    evidence_updated_at  timestamp default current_timestamp() not null on update current_timestamp(),
    constraint id_UNIQUE
        unique (evidence_id),
    constraint fk_evidence_user
        foreign key (evidence_author_id) references User (user_id)
            on update cascade,
    constraint fk_evidience_activity
        foreign key (evidence_activity_id) references Activity (activity_id)
            on update cascade
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
    issuerepository_deleted_at timestamp                             null,
    issuerepository_updated_at timestamp default current_timestamp() not null on update current_timestamp(),
    constraint id_UNIQUE
        unique (issuerepository_id),
    constraint name_UNIQUE
        unique (issuerepository_name)
);

create table if not exists Issue
(
    issue_id            int unsigned auto_increment
        primary key,
    issue_type          enum('Vulnerability','PolicyViolation','SecurityEvent') not null,
    issue_primary_name  varchar(256)                          not null,
    issue_description   longtext                              not null,
    issue_created_at    timestamp default current_timestamp() not null,
    issue_deleted_at    timestamp                             null,
    issue_updated_at    timestamp default current_timestamp() not null on update current_timestamp(),
    constraint id_UNIQUE
        unique (issue_id),
    constraint name_UNIQUE
        unique (issue_primary_name)
);

create table if not exists IssueVariant
(
    issuevariant_id                          int unsigned auto_increment
        primary key,
    issuevariant_issue_id int unsigned                          not null,
    issuevariant_repository_id               int unsigned                          not null,
    issuevariant_vector                      varchar(512)                          null,
    issuevariant_rating                      enum('None','Low','Medium', 'High', 'Critical')                             not null,
    issuevariant_secondary_name              varchar(256)                          not null,
    issuevariant_description                 longtext                              not null,
    issuevariant_created_at                  timestamp default current_timestamp() not null,
    issuevariant_deleted_at                  timestamp                             null,
    issuevariant_updated_at                  timestamp default current_timestamp() not null on update current_timestamp(),
    constraint id_UNIQUE
        unique (issuevariant_id),
    constraint name_UNIQUE
        unique (issuevariant_secondary_name),
    constraint fk_issuevariant_issue
        foreign key (issuevariant_issue_id) references Issue (issue_id),
    constraint fk_issuevariant_issuerepository
        foreign key (issuevariant_repository_id) references IssueRepository (issuerepository_id)
            on update cascade
);

create table if not exists ActivityHasIssue
(
    activityhasissue_activity_id                 int unsigned                          not null,
    activityhasissue_issue_id int unsigned                          not null,
    activityhasissue_created_at                  timestamp default current_timestamp() not null,
    activityhasissue_deleted_at                  timestamp                             null,
    activityhasissue_updated_at                  timestamp default current_timestamp() not null on update current_timestamp(),
    primary key (activityhasissue_activity_id, activityhasissue_issue_id),
    constraint fk_activity_issue
        foreign key (activityhasissue_issue_id) references Issue (issue_id),
    constraint fk_issue_activity
        foreign key (activityhasissue_activity_id) references Activity (activity_id)
);

create table if not exists ActivityHasService
(
    activityhasservice_activity_id                 int unsigned                          not null,
    activityhasservice_service_id                  int unsigned                          not null,
    activityhasservice_created_at                  timestamp default current_timestamp() not null,
    activityhasservice_deleted_at                  timestamp                             null,
    activityhasservice_updated_at                  timestamp default current_timestamp() not null on update current_timestamp(),
    primary key (activityhasservice_activity_id, activityhasservice_service_id),
    constraint fk_activity_service
        foreign key (activityhasservice_service_id) references Service (service_id),
    constraint fk_service_activity
        foreign key (activityhasservice_activity_id) references Activity (activity_id)
);

create table if not exists IssueMatch
(
    issuematch_id                          int unsigned auto_increment
        primary key,
    issuematch_status                      varchar(128)                          not null,
    issuematch_remediation_date            datetime                              null,
    issuematch_target_remediation_date     datetime                              not null,

    issuematch_vector                      varchar(512)                          null,
    issuematch_rating                      enum('None','Low','Medium', 'High', 'Critical') not null,

    issuematch_user_id                     int unsigned                          not null,
    issuematch_issue_id int unsigned                          not null,
    issuematch_component_instance_id       int unsigned                          not null,

    issuematch_created_at                  timestamp default current_timestamp() not null,
    issuematch_deleted_at                  timestamp                             null,
    issuematch_updated_at                  timestamp default current_timestamp() not null on update current_timestamp(),
    constraint id_UNIQUE
        unique (issuematch_id),
    constraint fk_issue_match_user_id
        foreign key (issuematch_user_id) references User (user_id)
            on update cascade,
    constraint fk_issue_match_component_instance
        foreign key (issuematch_component_instance_id) references ComponentInstance (componentinstance_id)
            on update cascade,
    constraint fk_issue_match_issue
        foreign key (issuematch_issue_id) references Issue (issue_id)
            on update cascade
);

create table if not exists IssueMatchChange
(
    issuematchchange_id                                   int unsigned auto_increment
        primary key,
    issuematchchange_activity_id                          int unsigned                          not null,
    issuematchchange_issue_match_id               int unsigned                          not null,
    issuematchchange_action                               enum('add','remove')                  not null,
    issuematchchange_created_at                           timestamp default current_timestamp() not null,
    issuematchchange_deleted_at                           timestamp                             null,
    issuematchchange_updated_at                           timestamp default current_timestamp() not null on update current_timestamp(),
    constraint fk_issuematchchange_activity
        foreign key (issuematchchange_activity_id) references Activity (activity_id),
    constraint fk_issuematchchange_issue_match
        foreign key (issuematchchange_issue_match_id) references IssueMatch (issuematch_id)
);

create table if not exists ComponentVersionIssue
(
    componentversionissue_component_version_id                 int unsigned                          not null,
    componentversionissue_issue_id          int unsigned                          not null,
    componentversionissue_created_at                           timestamp default current_timestamp() not null,
    componentversionissue_deleted_at                           timestamp                             null,
    componentversionissue_updated_at                           timestamp default current_timestamp() not null on update current_timestamp(),
    primary key (componentversionissue_component_version_id, componentversionissue_issue_id),
    constraint fk_componentversionissue_issue
        foreign key (componentversionissue_issue_id) references Issue (issue_id),
    constraint fk_componentversionissue_component_version
        foreign key (componentversionissue_component_version_id) references ComponentVersion (componentversion_id)
);


create table if not exists IssueMatchEvidence
(
    issuematchevidence_evidence_id                 int unsigned                          not null,
    issuematchevidence_issue_match_id      int unsigned                          not null,
    issuematchevidence_created_at                  timestamp default current_timestamp() not null,
    issuematchevidence_deleted_at                  timestamp                             null,
    issuematchevidence_updated_at                  timestamp default current_timestamp() not null on update current_timestamp(),
    primary key (issuematchevidence_evidence_id, issuematchevidence_issue_match_id),
    constraint fk_issue_match_evidence
        foreign key (issuematchevidence_evidence_id) references Evidence (evidence_id),
    constraint fk_evidence_issue_match
        foreign key (issuematchevidence_issue_match_id) references IssueMatch (issuematch_id)
);

create table if not exists IssueRepositoryService
(
    issuerepositoryservice_service_id                int unsigned                          not null,
    issuerepositoryservice_issue_repository_id    int unsigned                          not null,
    issuerepositoryservice_priority                  int unsigned                          not null,
    issuerepositoryservice_created_at                timestamp default current_timestamp() not null,
    issuerepositoryservice_deleted_at                timestamp                             null,
    issuerepositoryservice_updated_at                timestamp default current_timestamp() not null on update current_timestamp(),
    primary key (issuerepositoryservice_service_id, issuerepositoryservice_issue_repository_id),
    constraint fk_issue_repository_service
        foreign key (issuerepositoryservice_issue_repository_id) references IssueRepository (issuerepository_id)
            on update cascade,
    constraint fk_service_issue_repository
        foreign key (issuerepositoryservice_service_id) references Service (service_id)
            on update cascade
);



create index if not exists fk_issuevariant_issuerepository
    on IssueVariant (issuevariant_repository_id);

create index if not exists fk_issue_variant_issue 
    on IssueVariant (issuevariant_issue_id);

create index if not exists fk_component_version_component 
    on ComponentVersion (componentversion_component_id);

create index if not exists fk_component_instance_component_version 
    on ComponentInstance (componentinstance_component_version_id);

create index if not exists fk_component_instance_service 
    on ComponentInstance (componentinstance_service_id);

create index if not exists fk_evidence_activity 
    on Evidence (evidence_activity_id);

create index if not exists fk_evidence_user 
    on Evidence (evidence_author_id);

create index if not exists fk_service_user 
    on Owner (owner_service_id);

create index if not exists fk_user_service 
    on Owner (owner_user_id);

create index if not exists fk_issue_match_user
    on IssueMatch (issuematch_user_id);

create index if not exists fk_issue_match_component_instance 
    on IssueMatch (issue_component_instance_id);

create index if not exists fk_issue_match_issue 
    on IssueMatch (issuematch_issue_id);

create index if not exists fk_support_group_user 
    on SupportGroupUser (supportgroupuser_support_group_id);

create index if not exists fk_user_support_group 
    on SupportGroupUser (supportgroupuser_user_id);

create index if not exists fk_support_group_service 
    on SupportGroupService (supportgroupservice_support_group_id);

create index if not exists fk_service_support_group 
    on SupportGroupService (supportgroupservice_service_id);

create index if not exists fk_activity_issue 
    on ActivityHasIssue (activityhasissue_issue_id);

create index if not exists fk_issue_activity 
    on ActivityHasIssue (activityhasissue_activity_id);

create index if not exists fk_activity_service
    on ActivityHasService (activityhasservice_service_id);

create index if not exists fk_service_activity
    on ActivityHasService (activityhasservice_activity_id);

create index if not exists fk_componentversionissue_issue
    on ComponentVersionIssue (componentversionissue_issue_id);

create index if not exists fk_componentversionissue_component_version
    on ComponentVersionIssue (componentversionissue_component_version_id);

create index if not exists fk_issue_match_evidence
    on IssueMatchEvidence (issuematchevidence_evidence_id);

create index if not exists fk_evidence_issue_match
    on IssueMatchEvidence (issuematchevidence_issue_match_id);

create index if not exists issuematch_rating_idx
    on IssueMatch (issuematch_rating);

create index if not exists issuevariant_rating_idx
    on IssueVariant (issuevariant_rating);

create index if not exists fk_issue_repository_service 
    on IssueRepositoryService (issuerepositoryservice_issue_repository_id);

create index if not exists fk_service_issue_repository 
    on IssueRepositoryService (issuerepositoryservice_service_id);

create index if not exists fk_issuematchchange_activity
    on IssueMatchChange (issuematchchange_activity_id);

create index if not exists fk_issuematchchange_issue_match
    on IssueMatchChange (issuematchchange_issue_match_id);
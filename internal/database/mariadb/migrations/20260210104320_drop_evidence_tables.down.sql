--  SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
--  SPDX-License-Identifier: Apache-2.0

create table if not exists Evidence (
    evidence_id int unsigned auto_increment primary key,
    evidence_author_id int unsigned not null,
    evidence_activity_id int unsigned not null,
    evidence_type enum(
        'risk_accepted',
        'mitigated',
        'severity_adjustment',
        'false_positive',
        'reopen'
    ) not null,
    evidence_description longtext not null,
    evidence_vector varchar(512) null,
    evidence_rating enum(
        'None',
        'Low',
        'Medium',
        'High',
        'Critical'
    ) null,
    evidence_raa_end datetime null,
    evidence_created_at timestamp default current_timestamp() not null,
    evidence_created_by int unsigned null,
    evidence_deleted_at timestamp null,
    evidence_updated_at timestamp default current_timestamp() not null on update current_timestamp(),
    evidence_updated_by int unsigned null,
    constraint evidence_id_UNIQUE unique (evidence_id),
    constraint fk_evidence_user foreign key (evidence_author_id) references User (user_id) on update cascade,
    constraint fk_evidence_created_by foreign key (evidence_created_by) references User (user_id),
    constraint fk_evidence_updated_by foreign key (evidence_updated_by) references User (user_id)
);

create table if not exists IssueMatchEvidence (
    issuematchevidence_evidence_id int unsigned not null,
    issuematchevidence_issue_match_id int unsigned not null,
    issuematchevidence_created_at timestamp default current_timestamp() not null,
    issuematchevidence_deleted_at timestamp null,
    issuematchevidence_updated_at timestamp default current_timestamp() not null on update current_timestamp(),
    primary key (
        issuematchevidence_evidence_id,
        issuematchevidence_issue_match_id
    ),
    constraint fk_issue_match_evidence foreign key (
        issuematchevidence_evidence_id
    ) references Evidence (evidence_id),
    constraint fk_evidence_issue_match foreign key (
        issuematchevidence_issue_match_id
    ) references IssueMatch (issuematch_id)
);
--  SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
--  SPDX-License-Identifier: Apache-2.0

create table if not exists Activity (
    activity_id int unsigned auto_increment primary key,
    activity_status enum(
        'open',
        'closed',
        'in_progress'
    ) not null,
    activity_created_at timestamp default current_timestamp() not null,
    activity_created_by int unsigned null,
    activity_deleted_at timestamp null,
    activity_updated_at timestamp default current_timestamp() not null on update current_timestamp(),
    activity_updated_by int unsigned null,
    constraint activity_id_UNIQUE unique (activity_id),
    constraint fk_activity_created_by foreign key (activity_created_by) references User (user_id),
    constraint fk_activity_updated_by foreign key (activity_updated_by) references User (user_id)
);

ALTER TABLE Evidence
ADD COLUMN IF NOT EXISTS evidence_activity_id INT UNSIGNED;

ALTER TABLE Evidence
ADD CONSTRAINT fk_evidience_activity FOREIGN KEY IF NOT EXISTS (evidence_activity_id) REFERENCES Activity (activity_id) ON UPDATE CASCADE;

create table if not exists ActivityHasIssue (
    activityhasissue_activity_id int unsigned not null,
    activityhasissue_issue_id int unsigned not null,
    activityhasissue_created_at timestamp default current_timestamp() not null,
    activityhasissue_deleted_at timestamp null,
    activityhasissue_updated_at timestamp default current_timestamp() not null on update current_timestamp(),
    primary key (
        activityhasissue_activity_id,
        activityhasissue_issue_id
    ),
    constraint fk_activity_issue foreign key (activityhasissue_issue_id) references Issue (issue_id),
    constraint fk_issue_activity foreign key (activityhasissue_activity_id) references Activity (activity_id)
);

create table if not exists ActivityHasService (
    activityhasservice_activity_id int unsigned not null,
    activityhasservice_service_id int unsigned not null,
    activityhasservice_created_at timestamp default current_timestamp() not null,
    activityhasservice_deleted_at timestamp null,
    activityhasservice_updated_at timestamp default current_timestamp() not null on update current_timestamp(),
    primary key (
        activityhasservice_activity_id,
        activityhasservice_service_id
    ),
    constraint fk_activity_service foreign key (activityhasservice_service_id) references Service (service_id),
    constraint fk_service_activity foreign key (
        activityhasservice_activity_id
    ) references Activity (activity_id)
);
-- SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

create table if not exists Remediation
(
    remediation_id         int unsigned auto_increment
        primary key,
    remediation_type                enum('false_positive')                  not null,
    remediation_description         longtext                                not null,
    remediation_remediation_date    datetime                                null,
    remediation_expiration_date     datetime                                null,
    remediation_remediated_by       varchar(255)                            null,
    remediation_remediated_by_id    int unsigned                            null,
    remediation_service             varchar(255)                            not null,
    remediation_service_id          int unsigned                            not null,
    remediation_component           varchar(255)                            null,
    remediation_component_id        int unsigned                            null,
    remediation_issue               varchar(255)                            not null,
    remediation_issue_id            int unsigned                            not null,
    remediation_created_at          timestamp default current_timestamp()   not null,
    remediation_created_by          int unsigned                            null,
    remediation_deleted_at          timestamp                               null,
    remediation_updated_at          timestamp default current_timestamp()   not null on update current_timestamp(),
    remediation_updated_by          int unsigned                            null,
    constraint fk_remediation_created_by
        foreign key (remediation_created_by) references User (user_id),
    constraint fk_remediation_updated_by
        foreign key (remediation_updated_by) references User (user_id),
    constraint fk_remediation_service_id
        foreign key (remediation_service_id) references Service (service_id),
    constraint fk_remediation_component_id
        foreign key (remediation_component_id) references Component (component_id),
    constraint fk_remediation_issue_id
        foreign key (remediation_issue_id) references Issue (issue_id)
);
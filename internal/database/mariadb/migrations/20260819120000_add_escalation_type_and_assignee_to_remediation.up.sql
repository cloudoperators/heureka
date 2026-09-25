-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE Remediation
    MODIFY COLUMN remediation_type ENUM(
        'false_positive',
        'risk_accepted',
        'mitigation',
        'rescore',
        'escalation'
    ) NOT NULL,
    ADD COLUMN remediation_assignee    VARCHAR(255)  NULL AFTER remediation_remediated_by_id,
    ADD COLUMN remediation_assignee_id INT UNSIGNED  NULL AFTER remediation_assignee,
    ADD CONSTRAINT fk_remediation_assignee FOREIGN KEY (remediation_assignee_id) REFERENCES User(user_id);

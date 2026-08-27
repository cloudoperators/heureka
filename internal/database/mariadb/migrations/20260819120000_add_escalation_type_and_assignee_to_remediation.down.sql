-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE Remediation
    DROP FOREIGN KEY fk_remediation_assignee,
    DROP COLUMN remediation_assignee_id,
    DROP COLUMN remediation_assignee,
    MODIFY COLUMN remediation_type ENUM(
        'false_positive',
        'risk_accepted',
        'mitigation',
        'rescore'
    ) NOT NULL;

--  SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
--  SPDX-License-Identifier: Apache-2.0

ALTER TABLE Remediation
MODIFY remediation_type ENUM(
    'false_positive',
    'risk_accepted',
    'mitigation',
    'rescore'
);
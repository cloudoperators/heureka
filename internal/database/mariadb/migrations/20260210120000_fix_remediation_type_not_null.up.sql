-- SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

-- Fix remediation_type NULLability regression

-- Step 1: replace NULL values
UPDATE Remediation
SET remediation_type = 'false_positive'
WHERE remediation_type IS NULL;

-- Step 2: restore NOT NULL constraint
ALTER TABLE Remediation
MODIFY remediation_type ENUM(
    'false_positive',
    'risk_accepted',
    'mitigation',
    'rescore'
) NOT NULL;


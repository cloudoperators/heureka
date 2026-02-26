-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE Remediation
ADD COLUMN remediation_severity ENUM ('None','Low','Medium', 'High', 'Critical') NOT NULL DEFAULT 'None';
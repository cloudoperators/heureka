-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE IssueMatch
    ADD COLUMN issuematch_acknowledged BOOLEAN NOT NULL DEFAULT FALSE;

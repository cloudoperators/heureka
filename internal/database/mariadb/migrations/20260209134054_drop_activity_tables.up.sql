--  SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
--  SPDX-License-Identifier: Apache-2.0

DROP TABLE IF EXISTS activityhasservice;

DROP TABLE IF EXISTS activityhasissue;

ALTER TABLE evidence DROP CONSTRAINT IF EXISTS fk_evidience_activity;

ALTER TABLE evidence DROP COLUMN IF EXISTS evidence_activity_id;

DROP TABLE IF EXISTS activity;
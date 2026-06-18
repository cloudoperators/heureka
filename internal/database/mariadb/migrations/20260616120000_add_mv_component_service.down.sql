-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

DROP PROCEDURE IF EXISTS refresh_mvComponentService_proc;
DROP TABLE IF EXISTS mvComponentService;
DROP TABLE IF EXISTS mvComponentService_tmp;
DROP TABLE IF EXISTS mvComponentService_old;

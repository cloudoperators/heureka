-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE ScannerRun
    ADD COLUMN scannerrun_error TEXT NULL AFTER scannerrun_is_completed;

UPDATE ScannerRun sr
JOIN ScannerRunError sre
    ON sre.scannerrunerror_scannerrun_run_id = sr.scannerrun_run_id
SET sr.scannerrun_error = sre.error;

DROP TABLE ScannerRunError;

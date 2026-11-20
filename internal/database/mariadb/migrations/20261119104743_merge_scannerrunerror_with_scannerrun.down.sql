-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS ScannerRunError (
    scannerrunerror_scannerrun_run_id INT UNSIGNED NOT NULL,
    error TEXT NOT NULL,
    CONSTRAINT fk_sre_sr_id FOREIGN KEY (
        scannerrunerror_scannerrun_run_id
    ) REFERENCES ScannerRun (scannerrun_run_id) ON UPDATE CASCADE
);

INSERT INTO ScannerRunError (
    scannerrunerror_scannerrun_run_id,
    error
)
SELECT
    scannerrun_run_id,
    scannerrun_error
FROM ScannerRun
WHERE scannerrun_error IS NOT NULL;

ALTER TABLE ScannerRun
    DROP COLUMN scannerrun_error;

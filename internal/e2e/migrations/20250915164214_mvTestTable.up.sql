-- SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE mvTestData(
    test_id INT NOT NULL PRIMARY KEY
);

CREATE PROCEDURE refresh_mvTestData_proc()
BEGIN
    INSERT INTO mvTestData (test_id) VALUES (10);
END;


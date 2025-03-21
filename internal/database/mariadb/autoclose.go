// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

func (s *SqlDatabase) Autoclose() (bool, error) {
	var count int
	var err error

	row := s.db.QueryRow("SELECT COUNT(*) FROM ScannerRun WHERE scannerrun_is_completed = TRUE")

	err = row.Scan(&count)
	if err != nil {
		return false, err
	}

	if count >= 2 {
		row := s.db.QueryRow("SELECT COUNT(*) FROM Issue")

		err = row.Scan(&count)
		if err != nil {
			return false, err
		}
		if count > 0 {
			return true, nil
		}

	}

	return false, nil
}

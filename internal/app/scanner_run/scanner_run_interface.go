// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner_run

import "github.com/cloudoperators/heureka/internal/entity"

type ScannerRunHandler interface {
	CreateScannerRunHandler(*entity.ScannerRun) (*entity.ScannerRun, error)
}

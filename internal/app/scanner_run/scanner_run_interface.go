// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner_run

import "github.com/cloudoperators/heureka/internal/entity"

type ScannerRunHandler interface {
	CompleteScannerRun(string) (bool, error)
	CreateScannerRun(*entity.ScannerRun) (bool, error)
	FailScannerRun(string, string) (bool, error)
	GetScannerRuns(*entity.ScannerRunFilter, *entity.ListOptions) ([]entity.ScannerRun, error)
	GetScannerRunTags() ([]string, error)
	ScannerRunsTotalCount() (int, error)
}

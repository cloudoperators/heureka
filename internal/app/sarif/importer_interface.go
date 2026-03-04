// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package sarif

import (
	"context"
)

type Importer interface {
	ImportSARIF(ctx context.Context, input *ImportInput) (*ImportResult, error)
}

type ImportInput struct {
	SARIFDocument      string
	ScannerName        string
	ServiceId          int64
	Tag                string
}

type ImportResult struct {
	ScannerRunId       int64
	IssuesCreated      int
	IssueMatchesCreated int
	AssetsCreated      int
	Errors             []ImportError
}

type ImportError struct {
	Line     int
	Message  string
	Severity string
}

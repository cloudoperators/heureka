// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

import "time"

type ScanType string

const (
	ScanTypeInProgress ScanType = "InProgress"
	ScanTypeFail       ScanType = "Fail"
	ScanTypeSuccess    ScanType = "Success"
)

type Scan struct {
	Id         int64     `json:"id"`
	Type       ScanType  `json:"type"`
	Scope      string    `json:"scope"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
}

type ScanResult struct {
	WithCursor
	*Scan
}

type ScanFilter struct {
	Paginated
	Id    []*int64  `json:"id"`
	Scope []*string `json:"scope"`
}

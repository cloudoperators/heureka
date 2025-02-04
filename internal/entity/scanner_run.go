// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

import "time"

type ScannerRun struct {
	RunID     int64     `json:"run_id"`
	UUID      string    `json:"uuid"`
	Tag       string    `json:"tag"`
	StartRun  time.Time `json:"start_run"`
	EndRun    time.Time `json:"end_run"`
	Completed bool      `json:"is_completed"`
}

func (sc ScannerRun) IsCompleted() bool {
	return sc.Completed
}

type ScannerRunFilter struct {
	Paginated

	RunID     []int64   `json:"run_id"`
	UUID      []string  `json:"uuid"`
	Tag       []string  `json:"tag"`
	Completed bool      `json:"is_completed"`
	Search    []*string `json:"search"`
}

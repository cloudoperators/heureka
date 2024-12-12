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

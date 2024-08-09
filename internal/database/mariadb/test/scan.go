package test

import (
	"database/sql"

	"github.com/brianvoe/gofakeit/v7"
	"github.wdf.sap.corp/cc/heureka/internal/database/mariadb"
)

func NewFakeScan() mariadb.ScanRow {
	scanTypes := []string{"inProgress", "fail", "success"}
	// TODO: Add missing fields here

	return mariadb.ScanRow{
		ScanId: sql.NullInt64{Int64: gofakeit.Int64(), Valid: true},
		ScanType: sql.NullString{
			String: gofakeit.RandomString(scanTypes),
			Valid:  true,
		},
		Scope: sql.NullString{String: gofakeit.RandomString(scanTypes), Valid: true},
	}
}

// InsertFakeScan inserts a new Scan into the tables
func (s *DatabaseSeeder) InsertFakeScan(scan mariadb.ScanRow) (int64, error) {
	query := `
		INSERT INTO Scans (
			scan_id,
		) VALUES (
			:scan_id,
		)`
	return s.ExecPreparedNamed(query, scan)
}

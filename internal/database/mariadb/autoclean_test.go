// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

var scannerruns = map[string][]test.ScannerRunWithAssociatedData{
	"ScannerRunTag1": {
		{
			Timestamp:   time.Now().Add(time.Duration(0) * time.Minute).Truncate(time.Second).UTC(),
			ErrorMsgs:   []string{"dummy error message1 for ScannerRunTag1"},
			TrackersCnt: 4,
		},
		{
			Timestamp:   time.Now().Add(time.Duration(1) * time.Minute).Truncate(time.Second).UTC(),
			ErrorMsgs:   []string{"dummy error message2 for ScannerRunTag1"},
			TrackersCnt: 3,
		},
		{
			Timestamp:   time.Now().Add(time.Duration(2) * time.Minute).Truncate(time.Second).UTC(),
			ErrorMsgs:   []string{"dummy error message3 for ScannerRunTag1"},
			TrackersCnt: 2,
		},
	},
	"ScannerRunTag2": {
		{
			Timestamp:   time.Now().Add(time.Duration(0) * time.Minute).Truncate(time.Second).UTC(),
			ErrorMsgs:   []string{"dummy error message1 for ScannerRunTag2"},
			TrackersCnt: 5,
		},
		{
			Timestamp:   time.Now().Add(time.Duration(1) * time.Minute).Truncate(time.Second).UTC(),
			ErrorMsgs:   []string{"dummy error message2 for ScannerRunTag2"},
			TrackersCnt: 6,
		},
		{
			Timestamp:   time.Now().Add(time.Duration(2) * time.Minute).Truncate(time.Second).UTC(),
			ErrorMsgs:   []string{"dummy error message3 for ScannerRunTag2"},
			TrackersCnt: 7,
		},
	},
}

var _ = Describe("Autoclean", Label("database", "Autoclean"), func() {
	var db *mariadb.SqlDatabase

	BeforeEach(func() {
		db = dbm.NewTestSchema()
		dbSeeder, err := test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")

		err = dbSeeder.SeedScannerRunWithAssociatedData(scannerruns)
		Expect(err).To(BeNil(), "Seed Scanner Runs error")
	})

	AfterEach(func() {
		dbm.TestTearDown(db)
	})

	When("Running autoclean with retention set to 1", func() {
		It("Should leave only one (latest) scannerrun per tag in database", func() {
			err := db.Autoclean(context.Background(), 1)
			Expect(err).To(BeNil())

			scannerRuns := fetchScannerRuns(db.GetDb())
			for tag, srs := range scannerRuns {
				Expect(srs).To(HaveLen(1))

				ts := getScannerRunTimestamp(db.GetDb(), srs[0])
				Expect(*ts).To(BeEquivalentTo(scannerruns[tag][len(scannerruns[tag])-1].Timestamp))
			}
		})
	})
	When("Running autoclean with retention set to 0", func() {
		It("Should not remove any scannerrun from database", func() {
			err := db.Autoclean(context.Background(), 0)
			Expect(err).To(BeNil())

			scannerRuns := fetchScannerRuns(db.GetDb())
			for tag, srs := range scannerRuns {
				expectedTagTimestamps := lo.Map(scannerruns[tag], func(sr test.ScannerRunWithAssociatedData, _ int) time.Time { return sr.Timestamp })
				actualTagTimestamps := []time.Time{}

				for _, sr := range srs {
					ts := getScannerRunTimestamp(db.GetDb(), sr)
					actualTagTimestamps = append(actualTagTimestamps, *ts)
				}

				Expect(actualTagTimestamps).To(ConsistOf(expectedTagTimestamps))
			}
		})
	})
	When("Runing autoclean for tracker", func() {
		It("Should remove associated scannerrun trackers of removed run from database", func() {
			err := db.Autoclean(context.Background(), 1)
			Expect(err).To(BeNil())

			scannerRuns := fetchScannerRuns(db.GetDb())

			allExistingScannerRunIds := []int64{}
			for _, srs := range scannerRuns {
				allExistingScannerRunIds = append(allExistingScannerRunIds, srs...)
			}

			scannerRunIdsForAllTrackers := getScannerRunIdsForAllTrackers(db.GetDb())

			expectedNumberOfTrackers := 0

			for _, srs := range scannerruns {
				if len(srs) > 0 {
					expectedNumberOfTrackers += srs[len(srs)-1].TrackersCnt
				}
			}

			Expect(scannerRunIdsForAllTrackers).To(HaveLen(expectedNumberOfTrackers))

			for _, tid := range scannerRunIdsForAllTrackers {
				Expect(tid).To(BeElementOf(allExistingScannerRunIds))
			}
		})
	})
	When("Running autoclean for errors", func() {
		It("Should remove associated scannerrun errors of removed run from database", func() {
			err := db.Autoclean(context.Background(), 1)
			Expect(err).To(BeNil())

			expectedErrors := []string{}
			for _, sr := range scannerruns {
				expectedErrors = append(expectedErrors, sr[len(sr)-1].ErrorMsgs...)
			}

			errs := getAllScannerRunErrors(db.GetDb())

			Expect(errs).To(ConsistOf(expectedErrors))
		})
	})
})

func fetchScannerRuns(db mariadb.Db) map[string][]int64 {
	query, args, err := sq.Select("scannerrun_tag", "scannerrun_run_id").
		From("ScannerRun").
		Where("scannerrun_is_completed = TRUE AND scannerrun_deleted_at IS NULL").
		ToSql()
	Expect(err).To(BeNil())

	rows, err := db.Query(query, args...)
	Expect(err).To(BeNil())

	defer rows.Close()

	runs := map[string][]int64{}

	for rows.Next() {
		var (
			tag string
			id  int64
		)

		err := rows.Scan(&tag, &id)
		Expect(err).To(BeNil())

		runs[tag] = append(runs[tag], id)
	}

	Expect(rows.Err()).To(BeNil())

	return runs
}

func getScannerRunTimestamp(db mariadb.Db, srid int64) *time.Time {
	query, args, err := sq.Select("scannerrun_end_run").
		From("ScannerRun").
		Where(sq.Eq{"scannerrun_run_id": srid}).
		ToSql()
	Expect(err).To(BeNil())

	rows, err := db.Query(query, args...)
	Expect(err).To(BeNil())

	defer rows.Close()

	var ts time.Time

	err = db.QueryRow(query, args...).Scan(&ts)
	Expect(err).To(BeNil())

	return &ts
}

func getAllScannerRunErrors(db mariadb.Db) []string {
	query, args, err := sq.Select("error").From("ScannerRunError").ToSql()
	Expect(err).To(BeNil())

	rows, err := db.Query(query, args...)
	Expect(err).To(BeNil())

	defer rows.Close()

	var errorTexts []string

	for rows.Next() {
		var errMsg string

		err := rows.Scan(&errMsg)
		Expect(err).To(BeNil())

		errorTexts = append(errorTexts, errMsg)
	}

	Expect(rows.Err()).To(BeNil())

	return errorTexts
}

func getScannerRunIdsForAllTrackers(db mariadb.Db) []int64 {
	query, args, err := sq.Select("scannerruncomponentinstancetracker_scannerrun_run_id").
		From("ScannerRunComponentInstanceTracker").
		ToSql()
	Expect(err).To(BeNil())

	rows, err := db.Query(query, args...)
	Expect(err).To(BeNil())

	defer rows.Close()

	var ids []int64

	for rows.Next() {
		var id int64

		err := rows.Scan(&id)
		Expect(err).To(BeNil())

		ids = append(ids, id)
	}

	Expect(rows.Err()).To(BeNil())

	return ids
}

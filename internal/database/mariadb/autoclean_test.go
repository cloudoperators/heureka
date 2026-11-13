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
	// All completed runs
	"ScannerRunTag1": {
		{
			Timestamp:   time.Now().Add(time.Duration(0) * time.Minute).Truncate(time.Second).UTC(),
			TrackersCnt: 4,
			Completed:   true,
		},
		{
			Timestamp:   time.Now().Add(time.Duration(1) * time.Minute).Truncate(time.Second).UTC(),
			TrackersCnt: 3,
			Completed:   true,
		},
		{
			Timestamp:   time.Now().Add(time.Duration(2) * time.Minute).Truncate(time.Second).UTC(),
			TrackersCnt: 2,
			Completed:   true,
		},
	},
	// No completed runs
	"ScannerRunTag2": {
		{
			Timestamp: time.Now().Add(time.Duration(0) * time.Minute).Truncate(time.Second).UTC(),
			ErrorMsgs: []string{"dummy error message1 for ScannerRunTag2"},
			Completed: false,
		},
		{
			Timestamp: time.Now().Add(time.Duration(1) * time.Minute).Truncate(time.Second).UTC(),
			ErrorMsgs: []string{"dummy error message2 for ScannerRunTag2"},
			Completed: false,
		},
		{
			Timestamp: time.Now().Add(time.Duration(2) * time.Minute).Truncate(time.Second).UTC(),
			ErrorMsgs: []string{"dummy error message3 for ScannerRunTag2"},
			Completed: false,
		},
	},
	// One completed run
	"ScannerRunTag3": {
		{
			Timestamp: time.Now().Add(time.Duration(0) * time.Minute).Truncate(time.Second).UTC(),
			Completed: false,
			ErrorMsgs: []string{"dummy error message1 for ScannerRunTag3"},
		},
		{
			Timestamp: time.Now().Add(time.Duration(1) * time.Minute).Truncate(time.Second).UTC(),
			Completed: true,
		},
		{
			Timestamp: time.Now().Add(time.Duration(2) * time.Minute).Truncate(time.Second).UTC(),
			ErrorMsgs: []string{"dummy error message3 for ScannerRunTag3"},
			Completed: false,
		},
	},
	// One completed run in total
	"ScannerRunTag4": {
		{
			Timestamp: time.Now().Add(time.Duration(1) * time.Minute).Truncate(time.Second).UTC(),
			Completed: true,
		},
	},
	// Many completed runs
	"ScannerRunTag5": {
		{
			Timestamp: time.Now().Add(time.Duration(0) * time.Minute).Truncate(time.Second).UTC(),
			Completed: true,
		},
		{
			Timestamp: time.Now().Add(time.Duration(1) * time.Minute).Truncate(time.Second).UTC(),
			Completed: true,
		},
		{
			Timestamp: time.Now().Add(time.Duration(2) * time.Minute).Truncate(time.Second).UTC(),
			Completed: false,
			ErrorMsgs: []string{"dummy error message3 for ScannerRunTag5"},
		},
		{
			Timestamp: time.Now().Add(time.Duration(3) * time.Minute).Truncate(time.Second).UTC(),
			Completed: true,
		},
		{
			Timestamp: time.Now().Add(time.Duration(4) * time.Minute).Truncate(time.Second).UTC(),
			Completed: true,
		},
	},
	// Two completed but old ones
	"ScannerRunTag6": {
		{
			Timestamp: time.Now().Add(time.Duration(0) * time.Minute).Truncate(time.Second).UTC(),
			Completed: true,
		},
		{
			Timestamp: time.Now().Add(time.Duration(1) * time.Minute).Truncate(time.Second).UTC(),
			Completed: true,
		},
		{
			Timestamp: time.Now().Add(time.Duration(2) * time.Minute).Truncate(time.Second).UTC(),
			Completed: false,
			ErrorMsgs: []string{"dummy error message3 for ScannerRunTag6"},
		},
		{
			Timestamp: time.Now().Add(time.Duration(3) * time.Minute).Truncate(time.Second).UTC(),
			Completed: false,
			ErrorMsgs: []string{"dummy error message4 for ScannerRunTag6"},
		},
		{
			Timestamp: time.Now().Add(time.Duration(4) * time.Minute).Truncate(time.Second).UTC(),
			Completed: false,
			ErrorMsgs: []string{"dummy error message5 for ScannerRunTag6"},
		},
	},
}

type autocleanTest struct {
	db *mariadb.SqlDatabase
}

func newAutocleanTest() *autocleanTest {
	at := autocleanTest{db: dbm.NewTestSchema()}
	dbSeeder, err := test.NewDatabaseSeeder(dbm.DbConfig())
	Expect(err).To(BeNil(), "Database Seeder Setup should work")
	err = dbSeeder.SeedScannerRunWithAssociatedData(scannerruns)
	Expect(err).To(BeNil(), "Seed Scanner Runs error")

	return &at
}

func (at *autocleanTest) teardown() {
	dbm.TestTearDown(at.db)
}

func (at *autocleanTest) callAutocleanAndExpectNoErrors(keepNRuns uint64) {
	err := at.db.Autoclean(context.Background(), keepNRuns)
	Expect(err).To(BeNil())
}

func (at *autocleanTest) fetchNotDeletedScannerRuns() map[string][]int64 {
	query, args, err := sq.Select("scannerrun_tag", "scannerrun_run_id").
		From("ScannerRun").
		Where("scannerrun_deleted_at IS NULL").
		ToSql()
	Expect(err).To(BeNil())

	rows, err := at.db.GetDb().Query(query, args...)
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

func (at *autocleanTest) GetDb() mariadb.Db {
	return at.db.GetDb()
}

func (at *autocleanTest) expectTagScannerRuns(activeScannerRuns map[string][]int64, tag string, args ...int) {
	activeRunsIds := activeScannerRuns[tag]
	Expect(activeRunsIds).To(HaveLen(len(args)))

	testRuns := scannerruns[tag]

	lo.ForEach(args, func(v int, i int) {
		idx := (len(testRuns) + v) % len(testRuns)
		Expect(*at.getScannerRunTimestamp(activeRunsIds[i])).To(BeEquivalentTo(testRuns[idx].Timestamp))
	})
}

func (at *autocleanTest) expectAllScannerRuns(activeScannerRuns map[string][]int64) {
	for tag, srs := range activeScannerRuns {
		expectedTagTimestamps := lo.Map(scannerruns[tag], func(sr test.ScannerRunWithAssociatedData, _ int) time.Time { return sr.Timestamp })
		actualTagTimestamps := lo.Map(srs, func(sr int64, _ int) time.Time { return *at.getScannerRunTimestamp(sr) })
		Expect(actualTagTimestamps).To(ConsistOf(expectedTagTimestamps))
	}
}

func (at *autocleanTest) expectTagScannerRunsTrackers(activeScannerRuns map[string][]int64, tag string, args ...int) {
	activeRunsIds := activeScannerRuns[tag]
	testRuns := scannerruns[tag]

	scannerRunIdsForAllTrackers := at.getScannerRunIdsForAllTrackers()

	expectedNumberOfTrackers := lo.SumBy(args, func(v int) int {
		idx := (len(testRuns) + v) % len(testRuns)
		return testRuns[idx].TrackersCnt
	})
	Expect(scannerRunIdsForAllTrackers).To(HaveLen(expectedNumberOfTrackers))

	lo.ForEach(scannerRunIdsForAllTrackers, func(tid int64, _ int) {
		Expect(tid).To(BeElementOf(activeRunsIds))
	})
}

func (at *autocleanTest) getScannerRunTimestamp(srid int64) *time.Time {
	query, args, err := sq.Select("scannerrun_end_run").
		From("ScannerRun").
		Where(sq.Eq{"scannerrun_run_id": srid}).
		ToSql()
	Expect(err).To(BeNil())

	rows, err := at.db.GetDb().Query(query, args...)
	Expect(err).To(BeNil())

	defer rows.Close()

	var ts time.Time

	err = at.db.GetDb().QueryRow(query, args...).Scan(&ts)
	Expect(err).To(BeNil())

	return &ts
}

func (at *autocleanTest) getScannerRunIdsForAllTrackers() []int64 {
	query, args, err := sq.Select("scannerruncomponentinstancetracker_scannerrun_run_id").
		From("ScannerRunComponentInstanceTracker").
		ToSql()
	Expect(err).To(BeNil())

	rows, err := at.db.GetDb().Query(query, args...)
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

func (at *autocleanTest) getAllScannerRunErrors() []string {
	query, args, err := sq.Select("error").From("ScannerRunError").ToSql()
	Expect(err).To(BeNil())

	rows, err := at.db.GetDb().Query(query, args...)
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

var _ = Describe("Autoclean", Label("database", "Autoclean"), func() {
	var at *autocleanTest

	BeforeEach(func() {
		at = newAutocleanTest()
	})

	AfterEach(func() {
		at.teardown()
	})

	When("Running autoclean with retention set to 1", func() {
		It("Should leave at least two completed scanner runs if possible", func() {
			at.callAutocleanAndExpectNoErrors(1)
			activeScannerRuns := at.fetchNotDeletedScannerRuns()

			// Should leave at least two (latest completed) scannerrun per tag containing at least two completed runs
			at.expectTagScannerRuns(activeScannerRuns, "ScannerRunTag1", -2, -1)
			at.expectTagScannerRuns(activeScannerRuns, "ScannerRunTag5", -2, -1)

			// Should leave one scannerrun for tags with none completed run
			at.expectTagScannerRuns(activeScannerRuns, "ScannerRunTag2", -1)

			// Should leave at least one scannerrun for tags with one completed run
			at.expectTagScannerRuns(activeScannerRuns, "ScannerRunTag3", 1, -1)
			at.expectTagScannerRuns(activeScannerRuns, "ScannerRunTag4", -1)

			// Should leave one not completed scannerrun and two extra old completed scanner runs for tags with only old completed scanner runs
			at.expectTagScannerRuns(activeScannerRuns, "ScannerRunTag6", 0, 1, -1)
		})
	})
	When("Running autoclean with retention set to 3", func() {
		It("Should leave at least two completed scanner runs if possible", func() {
			at.callAutocleanAndExpectNoErrors(3)
			activeScannerRuns := at.fetchNotDeletedScannerRuns()

			// Should leave one run from tag with only one run
			at.expectTagScannerRuns(activeScannerRuns, "ScannerRunTag4", -1)

			// Should leave 3 least runs from tag with many completed runs
			at.expectTagScannerRuns(activeScannerRuns, "ScannerRunTag5", -3, -2, -1)

			// Should leave 3 not completed scanner runs and two extra old completed scanner runs for tags with only old completed scanner runs
			at.expectTagScannerRuns(activeScannerRuns, "ScannerRunTag6", 0, 1, -3, -2, -1)
		})
	})
	When("Running autoclean with retention set to 0", func() {
		It("Should not remove any scannerrun from database", func() {
			at.callAutocleanAndExpectNoErrors(0)
			activeScannerRuns := at.fetchNotDeletedScannerRuns()

			at.expectAllScannerRuns(activeScannerRuns)
		})
	})
	When("Runing autoclean for tracker and errors", func() {
		It("Should remove associated scannerrun trackers and errors of removed run from database", func() {
			at.callAutocleanAndExpectNoErrors(2)
			activeScannerRuns := at.fetchNotDeletedScannerRuns()

			at.expectTagScannerRuns(activeScannerRuns, "ScannerRunTag1", -2, -1)
			at.expectTagScannerRunsTrackers(activeScannerRuns, "ScannerRunTag1", -2, -1)

			errs := at.getAllScannerRunErrors()
			Expect(errs).To(HaveLen(5))
			Expect(errs).To(ContainElement(scannerruns["ScannerRunTag2"][len(scannerruns["ScannerRunTag2"])-1].ErrorMsgs[0]))
			Expect(errs).To(ContainElement(scannerruns["ScannerRunTag2"][len(scannerruns["ScannerRunTag2"])-2].ErrorMsgs[0]))
			Expect(errs).To(ContainElement(scannerruns["ScannerRunTag3"][len(scannerruns["ScannerRunTag3"])-1].ErrorMsgs[0]))
			Expect(errs).To(ContainElement(scannerruns["ScannerRunTag6"][len(scannerruns["ScannerRunTag6"])-1].ErrorMsgs[0]))
			Expect(errs).To(ContainElement(scannerruns["ScannerRunTag6"][len(scannerruns["ScannerRunTag6"])-2].ErrorMsgs[0]))
		})
	})
})

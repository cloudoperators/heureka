// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"fmt"
	"strings"
	"time"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

type autoPatchTestInfo struct {
	description                          string
	dbSeeds                              test.DbSeeds
	scannerRuns                          [][]string
	expectedResults                      []bool
	expectedPatchCount                   int
	patchedComponents                    []string
	expectDeletedIssueMatchesByIssueName []string
	expectDeletedVersions                *[]string
	expectDeletedComponentVersionIssues  []test.ComponentVersionIssue
}

type autoPatchTest struct {
	db             *mariadb.SqlDatabase
	databaseSeeder *test.DatabaseSeeder
}

func newAutoPatchTest() *autoPatchTest {
	db := dbm.NewTestSchema()
	dbSeeder, err := test.NewDatabaseSeeder(dbm.DbConfig())
	Expect(err).To(BeNil(), "Database Seeder Setup should work")
	return &autoPatchTest{db: db, databaseSeeder: dbSeeder}
}

func (apt *autoPatchTest) TearDown() {
	dbm.TestTearDown(apt.db)
}

func (apt *autoPatchTest) Run(tag string, info autoPatchTestInfo, fn func(db *mariadb.SqlDatabase) (bool, error)) {
	scannerRuns := lo.Map(info.scannerRuns, func(r []string, i int) test.ScannerRunDef {
		return test.ScannerRunDef{
			Tag:                        tag,
			IsCompleted:                true,
			Timestamp:                  time.Now().Add(time.Duration(i) * time.Minute),
			DetectedComponentInstances: r,
		}
	})

	scannerRunsSeeder := test.NewScannerRunsSeeder(apt.databaseSeeder)
	Expect(len(info.expectedResults)).To(BeEquivalentTo(len(scannerRuns)), "Number of expected results should be equal to number of scans (%s)", info.description)

	err := scannerRunsSeeder.SeedDbSeeds(info.dbSeeds)
	Expect(err).To(BeNil())

	// Expect ComponentVersionIssues not to be deleted yet
	apt.ExpectComponentVersionIssuesToBeStoredInDatabase(info.expectDeletedComponentVersionIssues, info.description)

	// run empty db autopatch
	res, err := fn(apt.db)
	Expect(err).To(BeNil(), "WHEN db has no scans THEN it should return no error")
	Expect(res).To(BeEquivalentTo(false), "WHEN db has no scans THEN it should return false")

	for i, sr := range scannerRuns {
		err := scannerRunsSeeder.SeedScannerRun(sr)
		Expect(err).To(BeNil())
		res, err := fn(apt.db)
		// Expect no error in autopatch
		Expect(err).To(BeNil(), fmt.Sprintf("%s THEN it should return no error", info.description))
		// Expect boolean result
		Expect(res).To(BeEquivalentTo(info.expectedResults[i]), fmt.Sprintf("%s THEN it should return %t (for scan number: %d)", info.description, info.expectedResults[i], i))
	}

	// Expect patch count
	apt.ExpectNumberOfPatches(info.expectedPatchCount, info.description)

	for _, patchedComponent := range info.patchedComponents {
		// Expect component to be patched
		apt.ExpectPatchForComponent(patchedComponent, info.description)
		// Expect component to be deleted
		apt.ExpectComponentToBeDeleted(patchedComponent, info.description)
	}

	// Expect issueMatches to be deleted
	apt.ExpectIssueMatchesToBeDeleted(info.expectDeletedIssueMatchesByIssueName, info.description)

	// Expect versions to be deleted
	if info.expectDeletedVersions != nil {
		apt.ExpectVersionsToBeDeleted(*info.expectDeletedVersions, info.description)
	}

	apt.ExpectComponentVersionIssuesToBeDeleted(info.expectDeletedComponentVersionIssues, info.description)

	err = apt.databaseSeeder.CleanupScannerRuns()
	Expect(err).To(BeNil())
}

func (apt *autoPatchTest) ExpectNumberOfPatches(n int, when string) {
	patchCount, err := apt.databaseSeeder.GetCountOfPatches()
	Expect(err).To(BeNil(), "Could not get patch count")
	Expect(patchCount).To(BeEquivalentTo(int64(n)), "%s THEN it should create %d patches", when, n)
}

func (apt *autoPatchTest) ExpectPatchForComponent(componentName string, when string) {
	patches, err := apt.databaseSeeder.FetchPatchesByComponentInstanceCCRN(componentName)
	Expect(err).To(BeNil(), "%s THEN %s component should have one matching patch (%s)", when, componentName, err)
	Expect(patches).NotTo(BeEmpty(), "%s THEN %s component should have at least one patch found", when, componentName)
}

func (apt *autoPatchTest) ExpectComponentToBeDeleted(componentName string, when string) {
	component, err := apt.databaseSeeder.FetchComponentInstanceByCCRN(componentName)
	Expect(err).To(BeNil(), "%s THEN %s component should be deleted (%s)", when, componentName, err)
	Expect(component.DeletedAt.Valid).To(BeEquivalentTo(true), "%s THEN %s component should be deleted", when, componentName)
	Expect(component.DeletedAt.Time.IsZero()).To(BeEquivalentTo(false), "%s THEN %s component should be deleted", when, componentName)
}

func (apt *autoPatchTest) ExpectIssueMatchesToBeDeleted(expectDeletedIssueMatchesByIssueName []string, when string) {
	deletedIssues, err := apt.databaseSeeder.FetchAllNamesOfDeletedIssueMatches()
	Expect(err).To(BeNil(), "%s THEN issue matches: '%s' should be deleted (%s)", when, strings.Join(expectDeletedIssueMatchesByIssueName, ", "), err)
	Expect(len(deletedIssues)).To(BeEquivalentTo(len(expectDeletedIssueMatchesByIssueName)), "%s THEN issue matches: '%s' should be deleted", when, strings.Join(expectDeletedIssueMatchesByIssueName, ", "))
	Expect(deletedIssues).To(ConsistOf(expectDeletedIssueMatchesByIssueName), "%s THEN issue matches: '%s' should be deleted", when, strings.Join(expectDeletedIssueMatchesByIssueName, ", "))
}

func (apt *autoPatchTest) ExpectVersionsToBeDeleted(expectDeletedVersions []string, when string) {
	deletedVersions, err := apt.databaseSeeder.FetchAllNamesOfDeletedVersions()
	Expect(err).To(BeNil(), "%s THEN versions: '%s' should be deleted (%s)", when, strings.Join(expectDeletedVersions, ", "), err)
	Expect(len(deletedVersions)).To(BeEquivalentTo(len(expectDeletedVersions)), "%s THEN versions: '%s' should be deleted", when, strings.Join(expectDeletedVersions, ", "))
	Expect(deletedVersions).To(ConsistOf(expectDeletedVersions), "%s THEN versions: '%s' should be deleted", when, strings.Join(expectDeletedVersions, ", "))
}

func (apt *autoPatchTest) ExpectComponentVersionIssuesToBeStoredInDatabase(expectComponentVersionIssues []test.ComponentVersionIssue, when string) {
	componentVersionIssues, err := apt.databaseSeeder.FetchAllNamesOfComponentVersionIssues()
	Expect(err).To(BeNil(), "%s THEN component version issues: '%s' should not be deleted yet (%s)", when, strings.Join(
		lo.Map(expectComponentVersionIssues, func(cvi test.ComponentVersionIssue, _ int) string {
			return fmt.Sprintf("i:%s cv:%s", cvi.Issue, cvi.ComponentVersion)
		}), ", "), err)
	Expect(componentVersionIssues).To(test.ContainAll(expectComponentVersionIssues), "%s THEN component version issues: '%s' should be deleted yet", when, strings.Join(
		lo.Map(expectComponentVersionIssues, func(cvi test.ComponentVersionIssue, _ int) string {
			return fmt.Sprintf("i:%s cv:%s", cvi.Issue, cvi.ComponentVersion)
		}), ", "))
}

func (apt *autoPatchTest) ExpectComponentVersionIssuesToBeDeleted(expectDeletedComponentVersionIssues []test.ComponentVersionIssue, when string) {
	componentVersionIssues, err := apt.databaseSeeder.FetchAllNamesOfComponentVersionIssues()
	Expect(err).To(BeNil(), "%s THEN component version issues: '%s' should be deleted (%s)", when, strings.Join(
		lo.Map(expectDeletedComponentVersionIssues, func(cvi test.ComponentVersionIssue, _ int) string {
			return fmt.Sprintf("i:%s cv:%s", cvi.Issue, cvi.ComponentVersion)
		}), ", "), err)
	Expect(componentVersionIssues).To(test.ContainNone(expectDeletedComponentVersionIssues), "%s THEN component version issues: '%s' should be deleted", when, strings.Join(
		lo.Map(expectDeletedComponentVersionIssues, func(cvi test.ComponentVersionIssue, _ int) string {
			return fmt.Sprintf("i:%s cv:%s", cvi.Issue, cvi.ComponentVersion)
		}), ", "))
}

var _ = Describe("Autopatch", Label("database", "Autopatch"), func() {
	var apt *autoPatchTest
	BeforeEach(func() {
		apt = newAutoPatchTest()
	})
	AfterEach(func() {
		apt.TearDown()
	})

	autopatchTests := []autoPatchTestInfo{
		{
			description:     "WHEN single completed scan with no components",
			scannerRuns:     [][]string{{}},
			expectedResults: []bool{false},
		},
		{
			description:     "WHEN single completed scan with components",
			scannerRuns:     [][]string{{"C1"}},
			expectedResults: []bool{false},
		},
		{
			description:     "WHEN two completed empty scans",
			scannerRuns:     [][]string{{}, {}},
			expectedResults: []bool{false, false},
		},
		{
			description:        "WHEN two completed scans: first has component, second no longer has that component",
			scannerRuns:        [][]string{{"C1"}, {}},
			expectedResults:    []bool{false, true},
			expectedPatchCount: 1,
			patchedComponents:  []string{"C1"},
		},
		{
			description:     "WHEN two completed scans: both have the same component",
			scannerRuns:     [][]string{{"C1"}, {"C1"}},
			expectedResults: []bool{false, false},
		},
		{
			description:        "WHEN two completed scans: first has C1, second has C2",
			scannerRuns:        [][]string{{"C1"}, {"C2"}},
			expectedResults:    []bool{false, true}, // C1 disappeared -> autopatch
			expectedPatchCount: 1,
			patchedComponents:  []string{"C1"},
		},
		{
			description: "WHEN component instance is tied to IssueMatch and disappears",
			scannerRuns: [][]string{{"C1"}, {"C2"}},
			dbSeeds: test.DbSeeds{
				IssueMatchComponents: []test.IssueMatchComponent{{Issue: "Issue1", ComponentInstance: "C1"}},
				Issues:               []string{"Issue1"},
				Components:           []test.ComponentData{{Name: "C1"}},
			},
			expectedResults:                      []bool{false, true},
			expectedPatchCount:                   1,
			patchedComponents:                    []string{"C1"},
			expectDeletedIssueMatchesByIssueName: []string{"Issue1"},
		},
		{
			description: "WHEN component instance is tied to IssueMatch and stays present in next run",
			scannerRuns: [][]string{{"C1"}, {"C1"}},
			dbSeeds: test.DbSeeds{
				IssueMatchComponents: []test.IssueMatchComponent{{Issue: "Issue1", ComponentInstance: "C1"}},
				Issues:               []string{"Issue1"},
				Components:           []test.ComponentData{{Name: "C1"}},
			},
			expectedResults: []bool{false, false},
		},
		{
			description:     "WHEN three completed empty scans",
			scannerRuns:     [][]string{{}, {}, {}},
			expectedResults: []bool{false, false, false},
		},
		{
			description:        "WHEN 3 scans: <C1>, <no component>, <no component>",
			scannerRuns:        [][]string{{"C1"}, {}, {}},
			expectedResults:    []bool{false, true, false},
			expectedPatchCount: 1,
		},
		{
			description:     "WHEN 3 scans: <C1>, <C1>, <C1>",
			scannerRuns:     [][]string{{"C1"}, {"C1"}, {"C1"}},
			expectedResults: []bool{false, false, false},
		},
		{
			description:        "WHEN 3 scans: <C1>, <no component>, <C1>",
			scannerRuns:        [][]string{{"C1"}, {}, {"C1"}},
			expectedResults:    []bool{false, true, false},
			expectedPatchCount: 1,
			patchedComponents:  []string{"C1"},
		},
		{
			description:        "WHEN 3 scans: <no component>, <C1>, <no component>",
			scannerRuns:        [][]string{{}, {"C1"}, {}},
			expectedResults:    []bool{false, false, true},
			expectedPatchCount: 1,
			patchedComponents:  []string{"C1"},
		},
		{
			description: "WHEN 2 scans detect disappearance of 3 components with the same version and service id", // THEN only single patch should be created
			scannerRuns: [][]string{{"C1", "C2", "C3"}, {}},
			dbSeeds: test.DbSeeds{
				IssueMatchComponents: []test.IssueMatchComponent{
					{Issue: "IC1", ComponentInstance: "C1"},
					{Issue: "IC2a", ComponentInstance: "C2"},
					{Issue: "IC2b", ComponentInstance: "C2"},
					{Issue: "IC3", ComponentInstance: "C3"},
				},
				Issues: []string{"IC1", "IC2a", "IC2b", "IC3", "IX"},
				Components: []test.ComponentData{
					{Name: "C1", Version: "V1", Service: "S1"},
					{Name: "C2", Version: "V1", Service: "S1"},
					{Name: "C3", Version: "V1", Service: "S1"},
				},
			},
			expectedResults:                      []bool{false, true},
			expectedPatchCount:                   1,
			patchedComponents:                    []string{"C1", "C2", "C3"},
			expectDeletedIssueMatchesByIssueName: []string{"IC1", "IC2a", "IC2b", "IC3"},
			expectDeletedVersions:                lo.ToPtr([]string{"V1"}),
		},
		{
			description: "WHEN 1 component disappear from 2 components with the same version and service", // THEN patch should not be created
			scannerRuns: [][]string{{"C1", "C2"}, {"C2"}},
			dbSeeds: test.DbSeeds{
				Components: []test.ComponentData{
					{Name: "C1", Version: "V1", Service: "S1"},
					{Name: "C2", Version: "V1", Service: "S1"},
				},
			},
			expectedResults:       []bool{false, true},
			expectedPatchCount:    0,
			expectDeletedVersions: lo.ToPtr([]string{}),
		},
		{
			description: "WHEN 1 component disappear from 2 components with different version and service", // THEN patch should be created and not used version should be removed
			scannerRuns: [][]string{{"C1", "C2"}, {"C1"}},
			dbSeeds: test.DbSeeds{
				Components: []test.ComponentData{
					{Name: "C1", Version: "V10", Service: "S1"},
					{Name: "C2", Version: "V20", Service: "S2"},
				},
			},
			expectedResults:       []bool{false, true},
			expectedPatchCount:    1,
			patchedComponents:     []string{"C2"},
			expectDeletedVersions: lo.ToPtr([]string{"V20"}),
		},
		{
			description: "WHEN 1 component with component version issues disappear from 2 components with different version and service", //THEN patch should be created and not used version should be removed and component version issues should be removed
			scannerRuns: [][]string{{"C1", "C2"}, {"C1"}},
			dbSeeds: test.DbSeeds{
				Issues: []string{"Issue1", "Issue2", "Issue3", "Issue4"},
				Components: []test.ComponentData{
					{Name: "C1", Version: "V10", Service: "S1"},
					{Name: "C2", Version: "V20", Service: "S2"},
				},
				ComponentVersionIssues: []test.ComponentVersionIssue{
					{Issue: "Issue1", ComponentVersion: "V10"},
					{Issue: "Issue4", ComponentVersion: "V10"},
					{Issue: "Issue1", ComponentVersion: "V20"},
					{Issue: "Issue2", ComponentVersion: "V20"},
					{Issue: "Issue3", ComponentVersion: "V20"},
				},
			},
			expectedResults:                     []bool{false, true},
			expectedPatchCount:                  1,
			patchedComponents:                   []string{"C2"},
			expectDeletedVersions:               lo.ToPtr([]string{"V20"}),
			expectDeletedComponentVersionIssues: []test.ComponentVersionIssue{{"Issue1", "V20"}, {"Issue2", "V20"}, {"Issue3", "V20"}},
		},
		{
			description: "WHEN 1 component disappear from 2 components with the same version and service", //THEN patch should be created and version should not be removed
			scannerRuns: [][]string{{"C1", "C2"}, {"C1"}},
			dbSeeds: test.DbSeeds{
				Components: []test.ComponentData{
					{Name: "C1", Version: "V10", Service: "S1"},
					{Name: "C2", Version: "V10", Service: "S2"},
				},
			},
			expectedResults:       []bool{false, true},
			expectedPatchCount:    1,
			patchedComponents:     []string{"C2"},
			expectDeletedVersions: lo.ToPtr([]string{}),
		},
		{
			description: "WHEN 4 scans detect disappearance of 1 component and the component appear and disappear again", //THEN two patches should be created for the same service and version
			scannerRuns: [][]string{{"C0", "C1"}, {"C0"}, {"C0", "C2"}, {"C0"}},
			dbSeeds: test.DbSeeds{
				Components: []test.ComponentData{
					{Name: "C0", Version: "V1", Service: "S0"},
					{Name: "C1", Version: "V1", Service: "S1"},
					{Name: "C2", Version: "V1", Service: "S1"},
				},
			},
			expectedResults:       []bool{false, true, false, true},
			expectedPatchCount:    2,
			patchedComponents:     []string{"C1", "C2"},
			expectDeletedVersions: lo.ToPtr([]string{}),
		},
	}

	When("Running autopatch", Label("Autopatch"), func() {
		Context("all scenarios DDT", func() {
			It("Returns no error and expected result per scenario", func() {
				for it, t := range autopatchTests {
					tag := fmt.Sprintf("ScannerRunTag_%d", it)
					apt.Run(tag, t, func(db *mariadb.SqlDatabase) (bool, error) { return db.Autopatch() })
				}
			})
		})
	})
})

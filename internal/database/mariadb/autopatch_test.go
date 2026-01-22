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

type scannerRun struct {
	isNotCompleted       bool
	issues               []string
	components           []test.Component
	issueMatchComponents []test.IssueMatchComponent
}

type autoPatchTestInfo struct {
	description                          string
	scannerRuns                          []scannerRun
	expectedResults                      []bool
	expectedPatchCount                   int
	patchedComponents                    []string
	expectDeletedIssueMatchesByIssueName []string
	expectDeletedVersions                []string
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
	scannerRuns := lo.Map(info.scannerRuns, func(r scannerRun, i int) test.ScannerRunDef {
		return test.ScannerRunDef{
			Tag:                  tag,
			IsCompleted:          !r.isNotCompleted,
			Timestamp:            time.Now().Add(time.Duration(i) * time.Minute),
			Issues:               r.issues,
			Components:           r.components,
			IssueMatchComponents: r.issueMatchComponents,
		}
	})

	scannerRunsSeeder := test.NewScannerRunsSeeder(apt.databaseSeeder)
	Expect(len(info.expectedResults)).To(BeEquivalentTo(len(scannerRuns)), "Number of expected results should be equal to number of scans (%s)", info.description)

	// run empty db autopatch
	res, err := fn(apt.db)
	Expect(err).To(BeNil(), "WHEN db has no scans THEN it should return no error")
	Expect(res).To(BeEquivalentTo(false), "WHEN db has no scans THEN it should return false")

	for i, sr := range scannerRuns {
		err := scannerRunsSeeder.Seed(sr)
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
	apt.ExpectVersionsToBeDeleted(info.expectDeletedVersions, info.description)

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
	Expect(deletedVersions).To(ConsistOf(expectDeletedVersions), "%s THEN iversions: '%s' should be deleted", when, strings.Join(expectDeletedVersions, ", "))
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
			description: "WHEN there is one not-completed scan",
			scannerRuns: []scannerRun{
				{isNotCompleted: true},
			},
			expectedResults: []bool{false},
		},
		{
			description: "WHEN single completed scan with no components",
			scannerRuns: []scannerRun{
				{},
			},
			expectedResults: []bool{false},
		},
		{
			description: "WHEN single completed scan with components",
			scannerRuns: []scannerRun{
				{components: []test.Component{{Name: "C1"}}},
			},
			expectedResults: []bool{false},
		},
		{
			description: "WHEN two completed empty scans",
			scannerRuns: []scannerRun{
				{},
				{},
			},
			expectedResults: []bool{false, false},
		},
		{
			description: "WHEN first completed scan has components and second is not completed",
			scannerRuns: []scannerRun{
				{components: []test.Component{{Name: "C1"}}},
				{isNotCompleted: true},
			},
			expectedResults: []bool{false, false},
		},
		{
			description: "WHEN two completed scans: first has component, second no longer has that component",
			scannerRuns: []scannerRun{
				{components: []test.Component{{Name: "C1"}}},
				{},
			},
			expectedResults:    []bool{false, true},
			expectedPatchCount: 1,
			patchedComponents:  []string{"C1"},
		},
		{
			description: "WHEN two completed scans: both have the same component",
			scannerRuns: []scannerRun{
				{components: []test.Component{{Name: "C1"}}},
				{components: []test.Component{{Name: "C1"}}},
			},
			expectedResults: []bool{false, false},
		},
		{
			description: "WHEN two completed scans: first has C1, second has C2",
			scannerRuns: []scannerRun{
				{components: []test.Component{{Name: "C1"}}},
				{components: []test.Component{{Name: "C2"}}},
			},
			expectedResults:    []bool{false, true}, // C1 disappeared -> autopatch
			expectedPatchCount: 1,
			patchedComponents:  []string{"C1"},
		},
		{
			description: "WHEN component instance is tied to IssueMatch and disappears",
			scannerRuns: []scannerRun{
				{
					issues:               []string{"Issue1"},
					components:           []test.Component{{Name: "C1"}},
					issueMatchComponents: []test.IssueMatchComponent{{Issue: "Issue1", Component: "C1"}},
				},
				{components: []test.Component{{Name: "C2"}}},
			},
			expectedResults:                      []bool{false, true},
			expectedPatchCount:                   1,
			patchedComponents:                    []string{"C1"},
			expectDeletedIssueMatchesByIssueName: []string{"Issue1"},
		},
		{
			description: "WHEN component instance is tied to IssueMatch and stays present in next run",
			scannerRuns: []scannerRun{
				{
					issues:               []string{"Issue1"},
					components:           []test.Component{{Name: "C1"}},
					issueMatchComponents: []test.IssueMatchComponent{{Issue: "Issue1", Component: "C1"}},
				},
				{
					issues:               []string{"Issue1"},
					components:           []test.Component{{Name: "C1"}},
					issueMatchComponents: []test.IssueMatchComponent{{Issue: "Issue1", Component: "C1"}},
				},
			},
			expectedResults: []bool{false, false},
		},
		{
			description: "WHEN three completed empty scans",
			scannerRuns: []scannerRun{
				{},
				{},
				{},
			},
			expectedResults: []bool{false, false, false},
		},
		{
			description: "WHEN 3 scans: <C1>, <no component>, <no component>",
			scannerRuns: []scannerRun{
				{components: []test.Component{{Name: "C1"}}},
				{},
				{},
			},
			expectedResults:    []bool{false, true, false},
			expectedPatchCount: 1,
		},
		{
			description: "WHEN 3 scans: <C1>, <C1>, <C1>",
			scannerRuns: []scannerRun{
				{components: []test.Component{{Name: "C1"}}},
				{components: []test.Component{{Name: "C1"}}},
				{components: []test.Component{{Name: "C1"}}},
			},
			expectedResults: []bool{false, false, false},
		},
		{
			description: "WHEN 3 scans: <C1>, <no component>, <C1>",
			scannerRuns: []scannerRun{
				{components: []test.Component{{Name: "C1"}}},
				{},
				{components: []test.Component{{Name: "C1"}}},
			},
			expectedResults:    []bool{false, true, false},
			expectedPatchCount: 1,
			patchedComponents:  []string{"C1"},
		},
		{
			description: "WHEN 3 scans: <no component>, <C1>, <no component>",
			scannerRuns: []scannerRun{
				{},
				{components: []test.Component{{Name: "C1"}}},
				{},
			},
			expectedResults:    []bool{false, false, true},
			expectedPatchCount: 1,
			patchedComponents:  []string{"C1"},
		},
		{
			description: "WHEN 2 scans detect disappearance of 3 components with the same version and service id", //T HEN only single patch should be created
			scannerRuns: []scannerRun{
				{
					issues:     []string{"IC1", "IC2a", "IC2b", "IC3", "IX"},
					components: []test.Component{{Name: "C1"}, {Name: "C2"}, {Name: "C3"}},
					issueMatchComponents: []test.IssueMatchComponent{
						{Issue: "IC1", Component: "C1"},
						{Issue: "IC2a", Component: "C2"},
						{Issue: "IC2b", Component: "C2"},
						{Issue: "IC3", Component: "C3"},
					},
				},
				{},
			},
			expectedResults:                      []bool{false, true},
			expectedPatchCount:                   1,
			patchedComponents:                    []string{"C1", "C2", "C3"},
			expectDeletedIssueMatchesByIssueName: []string{"IC1", "IC2a", "IC2b", "IC3"},
		},
		{
			description: "WHEN 1 component disappear from 2 components with the same version and service", // THEN patch should not be created
			scannerRuns: []scannerRun{
				{components: []test.Component{{Name: "C1", Version: "V1", Service: "S1"}, {Name: "C2", Version: "V1", Service: "S1"}}},
				{components: []test.Component{{Name: "C2", Version: "V1", Service: "S1"}}},
			},
			expectedResults:    []bool{false, true},
			expectedPatchCount: 0,
		},
		{
			description: "WHEN 1 component disappear from 2 components with different version and service", // THEN patch should be created and not used version should be removed
			scannerRuns: []scannerRun{
				{components: []test.Component{{Name: "C1", Version: "V1", Service: "S1"}, {Name: "C2", Version: "V2", Service: "S2"}}},
				{components: []test.Component{{Name: "C2", Version: "V1", Service: "S1"}}},
			},
			expectedResults:       []bool{false, true},
			expectedPatchCount:    1,
			patchedComponents:     []string{"C1"},
			expectDeletedVersions: []string{"V2"},
		},
		{
			description: "WHEN 4 scans detect disappearance of 1 component and the component appear and disappear again", // THEN two patches should be created for the same service and version
			scannerRuns: []scannerRun{
				{components: []test.Component{{Name: "C1", Version: "V1", Service: "S1"}}},
				{},
				{components: []test.Component{{Name: "C2", Version: "V1", Service: "S1"}}},
				{},
			},
			expectedResults:    []bool{false, true, false, true},
			expectedPatchCount: 2,
			patchedComponents:  []string{"C1", "C2"},
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

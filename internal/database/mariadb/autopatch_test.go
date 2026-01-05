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
	components           []string
	issueMatchComponents []test.IssueMatchComponent
}

type autoPatchTestInfo struct {
	description                          string
	scannerRuns                          []scannerRun
	expectedResult                       bool
	expectedPatchCount                   int
	patchedComponents                    []string
	expectDeletedIssueMatchesByIssueName []string
}

type autoPatchTest struct {
	db             *mariadb.SqlDatabase
	databaseSeeder *test.DatabaseSeeder
}

func newAutoPatchTest() *autoPatchTest {
	db := dbm.NewTestSchema()
	seeder, err := test.NewDatabaseSeeder(dbm.DbConfig())
	Expect(err).To(BeNil(), "Database Seeder Setup should work")
	return &autoPatchTest{db: db, databaseSeeder: seeder}
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
	err := apt.databaseSeeder.SeedScannerRuns(scannerRuns...)
	Expect(err).To(BeNil())
	res, err := fn(apt.db)

	// Expect no error in autopatch
	Expect(err).To(BeNil(), fmt.Sprintf("%s THEN it should return no error", info.description))

	// Expect boolean result
	Expect(res).To(BeEquivalentTo(info.expectedResult), fmt.Sprintf("%s THEN it should return %t", info.description, info.expectedResult))

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
	Expect(len(patches)).To(BeEquivalentTo(1), "%s THEN %s component should have one matching patch", when, componentName)
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

var _ = Describe("Autopatch", Label("database", "Autopatch"), func() {
	var apt *autoPatchTest
	BeforeEach(func() {
		apt = newAutoPatchTest()
	})
	AfterEach(func() {
		apt.TearDown()
	})

	var autopatchTests = []autoPatchTestInfo{
		{
			description:    "WHEN db has no scans",
			scannerRuns:    []scannerRun{},
			expectedResult: false,
		},
		{
			description: "WHEN there is one not-completed scan",
			scannerRuns: []scannerRun{
				{isNotCompleted: true},
			},
			expectedResult: false,
		},
		{
			description: "WHEN single completed scan with no components",
			scannerRuns: []scannerRun{
				{},
			},
			expectedResult: false,
		},
		{
			description: "WHEN single completed scan with components",
			scannerRuns: []scannerRun{
				{components: []string{"C1"}},
			},
			expectedResult: false,
		},
		{
			description: "WHEN two completed empty scans",
			scannerRuns: []scannerRun{
				{},
				{},
			},
			expectedResult: false,
		},
		{
			description: "WHEN first completed scan has components and second is not completed",
			scannerRuns: []scannerRun{
				{components: []string{"C1"}},
				{isNotCompleted: true},
			},
			expectedResult: false,
		},
		{
			description: "WHEN two completed scans: first has component, second no longer has that component",
			scannerRuns: []scannerRun{
				{components: []string{"C1"}},
				{},
			},
			expectedResult:     true,
			expectedPatchCount: 1,
			patchedComponents:  []string{"C1"},
		},
		{
			description: "WHEN two completed scans: both have the same component",
			scannerRuns: []scannerRun{
				{components: []string{"C1"}},
				{components: []string{"C1"}},
			},
			expectedResult: false,
		},
		{
			description: "WHEN two completed scans: first has C1, second has C2",
			scannerRuns: []scannerRun{
				{components: []string{"C1"}},
				{components: []string{"C2"}},
			},
			expectedResult:     true, // C1 disappeared -> autopatch
			expectedPatchCount: 1,
			patchedComponents:  []string{"C1"},
		},
		{
			description: "WHEN component instance is tied to IssueMatch and disappears",
			scannerRuns: []scannerRun{
				{
					issues:               []string{"Issue1"},
					components:           []string{"C1"},
					issueMatchComponents: []test.IssueMatchComponent{{Issue: "Issue1", Component: "C1"}},
				},
				{components: []string{"C2"}},
			},
			expectedResult:                       true,
			expectedPatchCount:                   1,
			patchedComponents:                    []string{"C1"},
			expectDeletedIssueMatchesByIssueName: []string{"Issue1"},
		},
		{
			description: "WHEN component instance is tied to IssueMatch and stays present in next run",
			scannerRuns: []scannerRun{
				{
					issues:               []string{"Issue1"},
					components:           []string{"C1"},
					issueMatchComponents: []test.IssueMatchComponent{{Issue: "Issue1", Component: "C1"}},
				},
				{
					issues:               []string{"Issue1"},
					components:           []string{"C1"},
					issueMatchComponents: []test.IssueMatchComponent{{Issue: "Issue1", Component: "C1"}},
				},
			},
			expectedResult: false,
		},
		{
			description: "WHEN three completed empty scans",
			scannerRuns: []scannerRun{
				{},
				{},
				{},
			},
			expectedResult: false,
		},
		{
			description: "WHEN 3 scans: <C1>, <no component>, <no component>",
			scannerRuns: []scannerRun{
				{components: []string{"C1"}},
				{},
				{},
			},
			expectedResult: false, // autopatch only compares newest & second newest
		},
		{
			description: "WHEN 3 scans: <C1>, <C1>, <C1>",
			scannerRuns: []scannerRun{
				{components: []string{"C1"}},
				{components: []string{"C1"}},
				{components: []string{"C1"}},
			},
			expectedResult: false,
		},
		{
			description: "WHEN 3 scans: <C1>, <no component>, <C1>",
			scannerRuns: []scannerRun{
				{components: []string{"C1"}},
				{},
				{components: []string{"C1"}},
			},
			expectedResult: false,
		},
		{
			description: "WHEN 3 scans: <no component>, <C1>, <no component>",
			scannerRuns: []scannerRun{
				{},
				{components: []string{"C1"}},
				{},
			},
			expectedResult:     true, // latest (no components) vs second-latest (C1)
			expectedPatchCount: 1,
			patchedComponents:  []string{"C1"},
		},
		{
			description: "WHEN 2 scans detect disappearance of 3 components with the same version and service id", //THEN only single patch should be created
			scannerRuns: []scannerRun{
				{
					issues:     []string{"IC1", "IC2a", "IC2b", "IC3", "IX"},
					components: []string{"C1", "C2", "C3"},
					issueMatchComponents: []test.IssueMatchComponent{
						{Issue: "IC1", Component: "C1"},
						{Issue: "IC2a", Component: "C2"},
						{Issue: "IC2b", Component: "C2"},
						{Issue: "IC3", Component: "C3"},
					}},
				{},
			},
			expectedResult:                       true,
			expectedPatchCount:                   1,
			patchedComponents:                    []string{"C1", "C2", "C3"},
			expectDeletedIssueMatchesByIssueName: []string{"IC1", "IC2a", "IC2b", "IC3"},
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

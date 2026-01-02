// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"fmt"
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

type autoScanTestInfo struct {
	description        string
	scannerRuns        []scannerRun
	expectedResult     bool
	expectedPatchCount int
	patchedComponents  []string
}

type autoScanTest struct {
	db             *mariadb.SqlDatabase
	databaseSeeder *test.DatabaseSeeder
}

func newAutoScanTest() *autoScanTest {
	db := dbm.NewTestSchema()
	seeder, err := test.NewDatabaseSeeder(dbm.DbConfig())
	Expect(err).To(BeNil(), "Database Seeder Setup should work")
	return &autoScanTest{db: db, databaseSeeder: seeder}
}

func (ast *autoScanTest) TearDown() {
	dbm.TestTearDown(ast.db)
}

func (ast *autoScanTest) Run(tag string, info autoScanTestInfo, fn func(db *mariadb.SqlDatabase) (bool, error)) {
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
	err := ast.databaseSeeder.SeedScannerRuns(scannerRuns...)
	Expect(err).To(BeNil())
	res, err := fn(ast.db)

	// Expect no error in autopatch
	Expect(err).To(BeNil(), fmt.Sprintf("%s THEN it should return no error", info.description))

	// Expect boolean result
	Expect(res).To(BeEquivalentTo(info.expectedResult), fmt.Sprintf("%s THEN it should return %t", info.description, info.expectedResult))

	// Expect patch count
	ast.ExpectNumberOfPatches(info.expectedPatchCount, info.description)

	for _, patchedComponent := range info.patchedComponents {
		// Expect component to be patched
		ast.ExpectPatchForComponent(patchedComponent, info.description)
		// Expect component to be removed
		ast.ExpectComponentToBeRemoved(patchedComponent, info.description)
		// Expect issueMatches linked with component to be removed
		ast.ExpectAllIssueMatchesLinkedWithComponentToBeRemoved(patchedComponent, info.description)
	}

	err = ast.databaseSeeder.CleanupScannerRuns()
	Expect(err).To(BeNil())
}

func (ast *autoScanTest) ExpectNumberOfPatches(n int, when string) {
	patchCount, err := ast.databaseSeeder.GetCountOfPatches()
	Expect(err).To(BeNil(), "Could not get patch count")
	Expect(patchCount).To(BeEquivalentTo(int64(n)), "%s THEN it should create %d patches", when, n)
}

func (ast *autoScanTest) ExpectPatchForComponent(componentName string, when string) {
	patches, err := ast.databaseSeeder.FetchPatchesByComponentInstanceCCRN(componentName)
	Expect(err).To(BeNil(), "%s THEN %s component should have one matching patch (%s)", when, componentName, err)
	Expect(len(patches)).To(BeEquivalentTo(1), "%s THEN %s component should have one matching patch", when, componentName)
}

func (ast *autoScanTest) ExpectComponentToBeRemoved(componentName string, when string) {
	component, err := ast.databaseSeeder.FetchComponentInstanceByCCRN(componentName)
	Expect(err).To(BeNil(), "%s THEN %s component should be removed (%s)", when, componentName, err)
	Expect(component.DeletedAt.Valid).To(BeEquivalentTo(true), "%s THEN %s component should be removed", when, componentName)
	Expect(component.DeletedAt.Time.IsZero()).To(BeEquivalentTo(false), "%s THEN %s component should be removed", when, componentName)
}

func (ast *autoScanTest) ExpectAllIssueMatchesLinkedWithComponentToBeRemoved(componentName string, when string) {
	component, err := ast.databaseSeeder.FetchComponentInstanceByCCRN(componentName)
	Expect(err).To(BeNil(), "%s THEN issue matches for %s component should be removed (%s)", when, componentName, err)
	Expect(component.Id.Valid).To(BeEquivalentTo(true), "%s THEN issue matches for %s component should be removed", when, componentName)
	issueMatches, err := ast.databaseSeeder.FetchIssueMatchesByComponentInstance(component.Id.Int64)
	Expect(err).To(BeNil(), "%s THEN issue matches for %s component should be removed (%s)", when, componentName, err)
	for _, issueMatch := range issueMatches {
		Expect(issueMatch.DeletedAt.Valid).To(BeEquivalentTo(true), "%s THEN issue match %d for %s component should be removed", when, issueMatch.Id.Int64, componentName)
		Expect(issueMatch.DeletedAt.Time.IsZero()).To(BeEquivalentTo(false), "%s THEN issue match %s for %s component should be removed", when, issueMatch.Id.Int64, componentName)
	}

}

var _ = Describe("Autoclose", Label("database", "Autoclose"), Label("database", "AutoScanProcess"), func() {
	var ast *autoScanTest
	BeforeEach(func() {
		ast = newAutoScanTest()
	})
	AfterEach(func() {
		ast.TearDown()
	})

	var autocloseTests = []autoScanTestInfo{
		{
			description:    "WHEN db has no scans",
			scannerRuns:    []scannerRun{},
			expectedResult: false,
		},
		{
			description: "WHEN database has one not completed scan",
			scannerRuns: []scannerRun{
				{isNotCompleted: true},
			},
			expectedResult: false,
		},
		{
			description: "WHEN database has one completed empty scan",
			scannerRuns: []scannerRun{
				{},
			},
			expectedResult: false,
		},
		{
			description: "WHEN database has one completed scan with one issue",
			scannerRuns: []scannerRun{
				{issues: []string{"Issue1"}},
			},
			expectedResult: false,
		},
		{
			description: "WHEN database has two completed empty scans",
			scannerRuns: []scannerRun{
				{},
				{},
			},
			expectedResult: false,
		},
		{
			description: "WHEN database has one completed scan with an issue and the second one not-completed without the issue",
			scannerRuns: []scannerRun{
				{issues: []string{"Issue1"}},
				{isNotCompleted: true},
			},
			expectedResult: false,
		},
		{
			description: "WHEN database has two completed scans where the first run found an issue and the second one has no longer the issue from previous run",
			scannerRuns: []scannerRun{
				{issues: []string{"Issue1"}},
				{},
			},
			expectedResult: true,
		},
		{
			description: "WHEN database has two completed scans where the first run found an issue and the second one has the same issue",
			scannerRuns: []scannerRun{
				{issues: []string{"Issue1"}},
				{issues: []string{"Issue1"}},
			},
			expectedResult: false,
		},
		{
			description: "WHEN database has two completed scans where the first run found an issue and the second one has different issue",
			scannerRuns: []scannerRun{
				{issues: []string{"Issue1"}},
				{issues: []string{"Issue2"}},
			},
			expectedResult: true,
		},
		{
			description: "WHEN database has two completed scans where the first run found a componentInstance issue and the second one has no longer the issue from previous run",
			scannerRuns: []scannerRun{
				{
					issues:               []string{"Issue1"},
					components:           []string{"Component1"},
					issueMatchComponents: []test.IssueMatchComponent{{Issue: "Issue1", Component: "Component1"}}},
				{components: []string{"Component2"}},
			},
			expectedResult: true,
		},
		{
			description: "WHEN database has two completed scans where the first run found a componentInstance issue and the second one has the same issue",
			scannerRuns: []scannerRun{
				{
					issues:               []string{"Issue1"},
					components:           []string{"Component1"},
					issueMatchComponents: []test.IssueMatchComponent{{Issue: "Issue1", Component: "Component1"}}},
				{
					issues:               []string{"Issue1"},
					components:           []string{"Component1"},
					issueMatchComponents: []test.IssueMatchComponent{{Issue: "Issue1", Component: "Component1"}}},
			},
			expectedResult: false,
		},
		{
			description: "WHEN database has three completed empty scans",
			scannerRuns: []scannerRun{
				{},
				{},
				{},
			},
			expectedResult: false,
		},
		{
			description: "WHEN 3 scans: <issue>, <no issue>, <no issue>",
			scannerRuns: []scannerRun{
				{issues: []string{"Issue1"}},
				{},
				{},
			},
			expectedResult: false,
		},
		{
			description: "WHEN 3 scans: <issue>, <issue>, <issue>",
			scannerRuns: []scannerRun{
				{issues: []string{"Issue1"}},
				{issues: []string{"Issue1"}},
				{issues: []string{"Issue1"}},
			},
			expectedResult: false,
		},
		{
			description: "WHEN 3 scans: <issue>, <no issue>, <issue>",
			scannerRuns: []scannerRun{
				{issues: []string{"Issue1"}},
				{},
				{issues: []string{"Issue1"}},
			},
			expectedResult: false,
		},
		{
			description: "WHEN 3 scans: <no issue>, <issue>, <no issue>",
			scannerRuns: []scannerRun{
				{},
				{issues: []string{"Issue1"}},
				{},
			},
			expectedResult: true,
		},
	}
	When("Running autoclose", Label("Autoclose"), func() {
		Context("all scenarios DDT", func() {
			It("Returns no error and expected result per scenario", func() {
				for it, t := range autocloseTests {
					tag := fmt.Sprintf("ScannerRunTag_%d", it)
					ast.Run(tag, t, func(db *mariadb.SqlDatabase) (bool, error) { return db.Autoclose() })
				}
			})
		})
	})
})

var _ = Describe("Autopatch", Label("database", "Autopatch"), func() {
	var ast *autoScanTest
	BeforeEach(func() {
		ast = newAutoScanTest()
	})
	AfterEach(func() {
		ast.TearDown()
	})

	var autopatchTests = []autoScanTestInfo{
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
			expectedResult:     true,
			expectedPatchCount: 1,
			patchedComponents:  []string{"C1"},
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
				{components: []string{"C1", "C2", "C3"}},
				{},
			},
			expectedResult:     true,
			expectedPatchCount: 1,
			patchedComponents:  []string{"C1", "C2", "C3"},
		},
	}

	When("Running autopatch", Label("Autopatch"), Label("database", "AutoScanProcess"), func() {
		Context("all scenarios DDT", func() {
			It("Returns no error and expected result per scenario", func() {
				for it, t := range autopatchTests {
					tag := fmt.Sprintf("ScannerRunTag_%d", it)
					ast.Run(tag, t, func(db *mariadb.SqlDatabase) (bool, error) { return db.Autopatch() })
				}
			})
		})
	})
})

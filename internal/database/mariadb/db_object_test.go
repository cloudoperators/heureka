// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"strings"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type dummyEntity struct {
	Id int64
	A  string
}

func (de *dummyEntity) GetId() int64 {
	return de.Id
}

func (de *dummyEntity) SetId(id int64) {
	de.Id = id
}

type dummyEntityFilter struct {
	entity.Paginated
	Id []*int64
	A  []*string
	B  []*int
}

var _ = Describe("DbObject", Label("database", "DbObject"), func() {
	const (
		dummyId                      = 7
		dummyString                  = "ttt"
		anyVal                       = false
		noPagination                 = false
		pagination                   = true
		dummyCursorFieldVal          = 9
		defaultFirstEntriesForCursor = 1000
	)

	dummyCursorField := mariadb.Field{Name: entity.OrderByField(4), Value: dummyCursorFieldVal, Order: entity.OrderDirectionDesc}

	When("Insert query is called", func() {
		Context("Two properties are there in the object", func() {
			testObject := mariadb.DbObject[*dummyEntity]{
				TableName: "DummyTable",
				Properties: []*mariadb.Property{
					mariadb.NewProperty("id", mariadb.WrapAccess(func(de *dummyEntity) (int64, bool) { return de.Id, anyVal })),
					mariadb.NewProperty("a", mariadb.WrapAccess(func(de *dummyEntity) (string, bool) { return de.A, anyVal })),
				},
			}
			It("can build insert query and query parameters", func() {
				de := dummyEntity{Id: dummyId, A: dummyString}
				query, params, err := testObject.InsertQuery(&de)
				Expect(err).To(BeNil())
				Expect(query).To(BeEquivalentTo("INSERT INTO DummyTable (id,a) VALUES (?,?)"))
				Expect(params).To(HaveLen(2))
				Expect(params[0]).To(BeEquivalentTo(dummyId))
				Expect(params[1]).To(BeEquivalentTo(dummyString))
			})
		})
	})
	When("Filter query is called", func() {
		Context("Two filters are there in the filter object", func() {
			testObject := mariadb.DbObject[*dummyEntity]{
				TableName: "DummyTable",
				FilterProperties: []*mariadb.FilterProperty{
					mariadb.NewFilterProperty("DT.dummytable_id = ?", mariadb.WrapRetSlice(func(filter *dummyEntityFilter) []*int64 { return filter.Id })),
					mariadb.NewFilterProperty("DT.dummytable_a = ?", mariadb.WrapRetSlice(func(filter *dummyEntityFilter) []*string { return filter.A })),
				},
			}
			It("builds filter for one filter item", func() {
				def := dummyEntityFilter{A: []*string{lo.ToPtr(dummyString)}}
				By("returning correct filter query string", func() {
					query := testObject.GetFilterQuery(&def)
					noWSQuery := strings.ReplaceAll(query, " ", "")
					Expect(noWSQuery).To(BeEquivalentTo("((DT.dummytable_a=?))"))
				})
				By("returning correct parameter", func() {
					params := testObject.GetFilterParameters(&def, noPagination, nil)
					Expect(params).To(HaveLen(1))
					Expect(params[0]).To(HaveValue(Equal(dummyString)))
				})
				By("returning correct parameter with cursor parameters", func() {
					params := testObject.GetFilterParameters(&def, pagination, []mariadb.Field{dummyCursorField})
					Expect(params).To(HaveLen(3))
					Expect(params[0]).To(HaveValue(Equal(dummyString)))
					Expect(params[1]).To(HaveValue(Equal(dummyCursorFieldVal)))
					Expect(params[2]).To(HaveValue(Equal(defaultFirstEntriesForCursor)))
				})
			})

			It("builds filter for two filter items", func() {
				id := int64(dummyId)
				def := dummyEntityFilter{Id: []*int64{&id}, A: []*string{lo.ToPtr(dummyString)}}
				By("returning correct filter query string", func() {
					query := testObject.GetFilterQuery(&def)
					noWSQuery := strings.ReplaceAll(query, " ", "")
					Expect(noWSQuery).To(BeEquivalentTo("((DT.dummytable_id=?)AND(DT.dummytable_a=?))"))
				})
				By("returning correct parameters", func() {
					params := testObject.GetFilterParameters(&def, noPagination, nil)
					Expect(params).To(HaveLen(2))
					Expect(params[0]).To(HaveValue(Equal(int64(dummyId))))
					Expect(params[1]).To(HaveValue(Equal(dummyString)))
				})
			})

			It("builds empty filter query string for no filter", func() {
				def := dummyEntityFilter{}
				query := testObject.GetFilterQuery(&def)
				Expect(query).To(BeEmpty())
			})
		})
	})
	When("Joins query is build", func() {
		Context("no joins are there in the object", func() {
			testObject := mariadb.DbObject[*dummyEntity]{
				TableName: "DummyTable",
			}
			It("gets empty string as there is no join defined", func() {
				def := dummyEntityFilter{A: []*string{lo.ToPtr(dummyString)}}
				joins := testObject.GetJoins(&def, nil)
				Expect(joins).To(BeEmpty())
			})
		})
		Context("Two basic joins are there in the object", func() {
			testObject := mariadb.DbObject[*dummyEntity]{
				TableName: "DummyTable",
				JoinDefs: []*mariadb.JoinDef{
					{
						Name:      "X",
						Type:      mariadb.LeftJoin,
						Table:     "Xabc X",
						On:        "DT.dummytable_id = X.xabc_dummytable_id",
						Condition: mariadb.WrapJoinCondition(func(f *dummyEntityFilter, _ []entity.Order) bool { return len(f.A) > 0 }),
					},
					{
						Name:      "Y",
						Type:      mariadb.RightJoin,
						Table:     "Yabc Y",
						On:        "DT.dummytable_id = Y.yabc_dummytable_id",
						Condition: mariadb.WrapJoinCondition(func(f *dummyEntityFilter, _ []entity.Order) bool { return len(f.B) > 0 }),
					},
				},
			}
			It("gets empty string when there is empty filter used", func() {
				def := dummyEntityFilter{}
				joins := testObject.GetJoins(&def, nil)
				Expect(joins).To(BeEmpty())
			})
			It("gets one join with X when A mapping condition is met (there is at least one element to filter)", func() {
				def := dummyEntityFilter{A: []*string{lo.ToPtr(dummyString)}}
				joins := testObject.GetJoins(&def, nil)
				Expect(joins).To(BeEquivalentTo("LEFT JOIN Xabc X ON DT.dummytable_id = X.xabc_dummytable_id"))
			})
			It("gets one join with Y when B mapping condition is met (there is at least one element to filter)", func() {
				def := dummyEntityFilter{B: []*int{lo.ToPtr(10)}}
				joins := testObject.GetJoins(&def, nil)
				Expect(joins).To(BeEquivalentTo("RIGHT JOIN Yabc Y ON DT.dummytable_id = Y.yabc_dummytable_id"))
			})
			It("gets two joins with X and Y when A and B mapping conditions are met (there are at least one element for each filter)", func() {
				def := dummyEntityFilter{A: []*string{lo.ToPtr(dummyString)}, B: []*int{lo.ToPtr(10)}}
				joins := testObject.GetJoins(&def, nil)
				lines := strings.Split(strings.TrimSpace(joins), "\n")
				Expect(lines).To(HaveLen(2))
				Expect(lines).To(ConsistOf(
					MatchRegexp("^LEFT JOIN Xabc X ON DT.dummytable_id = X.xabc_dummytable_id$"),
					MatchRegexp("^RIGHT JOIN Yabc Y ON DT.dummytable_id = Y.yabc_dummytable_id$"),
				))
			})
		})
		Context("Join with dependency is there in the object", func() {
			testObject := mariadb.DbObject[*dummyEntity]{
				TableName: "DummyTable",
				JoinDefs: []*mariadb.JoinDef{
					{
						Name:      "X",
						Type:      mariadb.LeftJoin,
						Table:     "Xabc X",
						On:        "DT.dummytable_id = X.xabc_dummytable_id",
						Condition: mariadb.DependentJoin,
					},
					{
						Name:      "Y",
						Type:      mariadb.RightJoin,
						Table:     "Yabc Y",
						On:        "X.xabc_yabc_id = Y.yabc_id",
						DependsOn: []string{"X"},
						Condition: mariadb.WrapJoinCondition(func(f *dummyEntityFilter, _ []entity.Order) bool { return len(f.B) > 0 }),
					},
				},
			}
			It("gets two joins with X (and dependent Y) when B mapping condition is met (there is at least one element to filter)", func() {
				def := dummyEntityFilter{B: []*int{lo.ToPtr(10)}}
				joins := testObject.GetJoins(&def, nil)
				lines := strings.Split(strings.TrimSpace(joins), "\n")
				Expect(lines).To(HaveLen(2))
				Expect(lines[0]).To(BeEquivalentTo("LEFT JOIN Xabc X ON DT.dummytable_id = X.xabc_dummytable_id"))
				Expect(lines[1]).To(BeEquivalentTo("RIGHT JOIN Yabc Y ON X.xabc_yabc_id = Y.yabc_id"))
			})
		})
		Context("Two joins with the same dependency are there in the object", func() {
			testObject := mariadb.DbObject[*dummyEntity]{
				TableName: "DummyTable",
				JoinDefs: []*mariadb.JoinDef{
					{
						Name:      "X",
						Type:      mariadb.LeftJoin,
						Table:     "Xabc X",
						On:        "DT.dummytable_id = X.xabc_dummytable_id",
						Condition: mariadb.DependentJoin,
					},
					{
						Name:      "Y",
						Type:      mariadb.RightJoin,
						Table:     "Yabc Y",
						On:        "X.xabc_yabc_id = Y.yabc_id",
						DependsOn: []string{"X"},
						Condition: mariadb.WrapJoinCondition(func(f *dummyEntityFilter, _ []entity.Order) bool { return len(f.B) > 0 }),
					},
					{
						Name:      "Z",
						Type:      mariadb.RightJoin,
						Table:     "Zabc Z",
						On:        "X.xabc_zabc_id = Z.zabc_id",
						DependsOn: []string{"X"},
						Condition: mariadb.WrapJoinCondition(func(f *dummyEntityFilter, _ []entity.Order) bool { return len(f.A) > 0 }),
					},
				},
			}
			It("gets two joins and one dependent join when A and B mapping conditions are met (there is at least one element for each filter)", func() {
				def := dummyEntityFilter{A: []*string{lo.ToPtr(dummyString)}, B: []*int{lo.ToPtr(10)}}
				joins := testObject.GetJoins(&def, nil)
				lines := strings.Split(strings.TrimSpace(joins), "\n")
				Expect(lines).To(HaveLen(3))
				Expect(lines[0]).To(BeEquivalentTo("LEFT JOIN Xabc X ON DT.dummytable_id = X.xabc_dummytable_id"))
				Expect(lines[1:]).To(ConsistOf(
					MatchRegexp("^RIGHT JOIN Yabc Y ON X.xabc_yabc_id = Y.yabc_id$"),
					MatchRegexp("^RIGHT JOIN Zabc Z ON X.xabc_zabc_id = Z.zabc_id$"),
				))
			})
		})
		Context("Two joins on the same table are there in the object", func() {
			testObject := mariadb.DbObject[*dummyEntity]{
				TableName: "DummyTable",
				JoinDefs: []*mariadb.JoinDef{
					{
						Name:      "X_left",
						Type:      mariadb.LeftJoin,
						Table:     "Xabc X",
						On:        "DT.dummytable_id = X.xabc_dummytable_id",
						Condition: mariadb.WrapJoinCondition(func(f *dummyEntityFilter, _ []entity.Order) bool { return len(f.B) > 0 }),
					},
					{
						Name:      "X_right",
						Type:      mariadb.RightJoin,
						Table:     "Xabc X",
						On:        "DT.dummytable_id = X.xabc_dummytable_id",
						Condition: mariadb.WrapJoinCondition(func(f *dummyEntityFilter, _ []entity.Order) bool { return len(f.B) > 0 }),
					},
				},
			}
			It("gets one join even when both join defs have the same condition but only first one will be included to prevent join table name duplication", func() {
				def := dummyEntityFilter{B: []*int{lo.ToPtr(10)}}
				joins := testObject.GetJoins(&def, nil)
				Expect(joins).To(BeEquivalentTo("LEFT JOIN Xabc X ON DT.dummytable_id = X.xabc_dummytable_id"))
			})
		})
	})
})

/*import (
	"database/sql"
	"fmt"
	"sort"
	"time"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/entity"
	entity_test "github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/util"
	pkg_util "github.com/cloudoperators/heureka/pkg/util"
	"github.com/samber/lo"
)

var _ = Describe("Issue", Label("database", "Issue"), func() {
	var db *mariadb.SqlDatabase
	var seeder *test.DatabaseSeeder
	BeforeEach(func() {
		var err error
		db = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")
	})
	AfterEach(func() {
		dbm.TestTearDown(db)
	})

	When("Getting Issues", Label("GetIssues"), func() {
		Context("and the database is empty", func() {
			It("can perform the list query", func() {
				res, err := db.GetIssues(nil, nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 10 issues in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})

			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					res, err := db.GetIssues(nil, nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.IssueRows)))
					})

					By("returning the correct order", func() {
						var prev int64 = 0
						for _, r := range res {

							Expect(r.Issue.Id > prev).Should(BeTrue())
							prev = r.Issue.Id

						}
					})

					By("returning the correct fields", func() {
						for _, r := range res {
							for _, row := range seedCollection.IssueRows {
								if r.Issue.Id == row.Id.Int64 {
									Expect(
										r.Issue.PrimaryName,
									).Should(BeEquivalentTo(row.PrimaryName.String), "Name should match")
									Expect(
										r.Issue.Type,
									).Should(BeEquivalentTo(row.Type.String), "Type should match")
									Expect(
										r.Issue.Description,
									).Should(BeEquivalentTo(row.Description.String), "Description should match")
									Expect(
										r.Issue.CreatedAt,
									).ShouldNot(BeEquivalentTo(row.CreatedAt.Time), "CreatedAt matches")
									Expect(
										r.Issue.UpdatedAt,
									).ShouldNot(BeEquivalentTo(row.UpdatedAt.Time), "UpdatedAt matches")
								}
							}
						}
					})
				})
			})
			Context("and using a filter", func() {
				It("can filter by a single service name", func() {
					var row mariadb.BaseServiceRow
					searchingRow := true
					var issueRows []mariadb.IssueRow

					// get a service that should return at least 1 issue
					for searchingRow {
						row = test.PickOne(seedCollection.ServiceRows)
						issueRows = seedCollection.GetIssueByService(&row)
						searchingRow = len(issueRows) == 0
					}
					filter := &entity.IssueFilter{ServiceCCRN: []*string{&row.CCRN.String}}

					entries, err := db.GetIssues(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning some results", func() {
						Expect(entries).NotTo(BeEmpty())
					})
					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(len(issueRows)))
					})
				})
				It("can filter a non existing service name", func() {
					nonExistingName := pkg_util.GenerateRandomString(40, nil)
					filter := &entity.IssueFilter{ServiceCCRN: []*string{&nonExistingName}}

					entries, err := db.GetIssues(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning no results", func() {
						Expect(entries).To(BeEmpty())
					})
				})
				It("can filter by multiple existing service names", func() {
					serviceCcrns := make([]*string, len(seedCollection.ServiceRows))
					var expectedIssues []mariadb.IssueRow
					for i, row := range seedCollection.ServiceRows {
						x := row.CCRN.String
						expectedIssues = append(
							expectedIssues,
							seedCollection.GetIssueByService(&row)...)
						serviceCcrns[i] = &x
					}
					expectedIssues = lo.Uniq(expectedIssues)
					filter := &entity.IssueFilter{ServiceCCRN: serviceCcrns}

					entries, err := db.GetIssues(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(len(expectedIssues)))
					})
				})
				It("can filter by a single issue Id", func() {
					row := test.PickOne(seedCollection.IssueRows)
					filter := &entity.IssueFilter{Id: []*int64{&row.Id.Int64}}

					entries, err := db.GetIssues(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning exactly 1 element", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})

					By("returning the expected element", func() {
						Expect(entries[0].Issue.Id).To(BeEquivalentTo(row.Id.Int64))
					})
				})
				It("can filter by a single service Id", func() {
					serviceRow := test.PickOne(seedCollection.ServiceRows)
					ciIds := lo.FilterMap(seedCollection.ComponentInstanceRows, func(c mariadb.ComponentInstanceRow, _ int) (int64, bool) {
						return c.Id.Int64, serviceRow.Id.Int64 == c.ServiceId.Int64
					})
					issueIds := lo.FilterMap(seedCollection.IssueMatchRows, func(im mariadb.IssueMatchRow, _ int) (int64, bool) {
						return im.IssueId.Int64, lo.Contains(ciIds, im.ComponentInstanceId.Int64)
					})

					filter := &entity.IssueFilter{ServiceId: []*int64{&serviceRow.Id.Int64}}

					entries, err := db.GetIssues(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the expected element", func() {
						for _, entry := range entries {
							Expect(lo.Contains(issueIds, entry.Issue.Id)).To(BeTrue())
						}
					})
				})
				It("can filter by a single support group ccrn", func() {
					sgRow := test.PickOne(seedCollection.SupportGroupRows)
					serviceIds := lo.FilterMap(
						seedCollection.SupportGroupServiceRows,
						func(sgs mariadb.SupportGroupServiceRow, _ int) (int64, bool) {
							return sgs.ServiceId.Int64, sgRow.Id.Int64 == sgs.SupportGroupId.Int64
						},
					)
					ciIds := lo.FilterMap(
						seedCollection.ComponentInstanceRows,
						func(c mariadb.ComponentInstanceRow, _ int) (int64, bool) {
							return c.Id.Int64, lo.Contains(serviceIds, c.ServiceId.Int64)
						},
					)
					issueIds := lo.FilterMap(
						seedCollection.IssueMatchRows,
						func(im mariadb.IssueMatchRow, _ int) (int64, bool) {
							return im.IssueId.Int64, lo.Contains(
								ciIds,
								im.ComponentInstanceId.Int64,
							)
						},
					)

					filter := &entity.IssueFilter{SupportGroupCCRN: []*string{&sgRow.CCRN.String}}

					entries, err := db.GetIssues(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the expected element", func() {
						for _, entry := range entries {
							Expect(lo.Contains(issueIds, entry.Issue.Id)).To(BeTrue())
						}
					})
				})
				It("can filter by a single component version id", func() {
					// select a componentVersion
					cvRow := test.PickOne(seedCollection.ComponentVersionRows)

					// collect all issue ids that belong to the component version
					issueIds := []int64{}
					for _, cvvRow := range seedCollection.ComponentVersionIssueRows {
						if cvvRow.ComponentVersionId.Int64 == cvRow.Id.Int64 {
							issueIds = append(issueIds, cvvRow.IssueId.Int64)
						}
					}

					filter := &entity.IssueFilter{ComponentVersionId: []*int64{&cvRow.Id.Int64}}

					entries, err := db.GetIssues(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the expected elements", func() {
						for _, entry := range entries {
							Expect(issueIds).To(ContainElement(entry.Issue.Id))
						}
					})
				})
				It("can filter by a single component id", func() {
					// select a component
					cRow := test.PickOne(seedCollection.ComponentRows)

					// collect all componentVersion ids that belong to the component
					cvIds := []int64{}
					for _, cvRow := range seedCollection.ComponentVersionRows {
						if cvRow.ComponentId.Int64 == cRow.Id.Int64 {
							cvIds = append(cvIds, cvRow.Id.Int64)
						}
					}

					// collect all issue ids that belong to the component version ids
					issueIds := []int64{}
					for _, cviRow := range seedCollection.ComponentVersionIssueRows {
						if lo.Contains(cvIds, cviRow.ComponentVersionId.Int64) {
							issueIds = append(issueIds, cviRow.IssueId.Int64)
						}
					}

					filter := &entity.IssueFilter{ComponentId: []*int64{&cRow.Id.Int64}}

					entries, err := db.GetIssues(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the expected elements", func() {
						for _, entry := range entries {
							Expect(issueIds).To(ContainElement(entry.Issue.Id))
						}
					})
				})
				It("can filter by a single issueVariant id", func() {
					// select an issueVariant
					issueVariantRow := test.PickOne(seedCollection.IssueVariantRows)

					filter := &entity.IssueFilter{
						IssueVariantId: []*int64{&issueVariantRow.Id.Int64},
					}

					entries, err := db.GetIssues(filter, nil)

					issueIds := []int64{}
					for _, entry := range entries {
						issueIds = append(issueIds, entry.Issue.Id)
					}

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the expected elements", func() {
						Expect(issueIds).To(ContainElement(issueVariantRow.IssueId.Int64))
					})
				})
				It("can filter by a issueType", func() {
					issueType := test.PickOne(entity.AllIssueTypes)

					filter := &entity.IssueFilter{Type: []*string{&issueType}}

					entries, err := db.GetIssues(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					for _, entry := range entries {
						Expect(entry.Type).To(BeEquivalentTo(issueType))
					}
				})
				It("can filter by hasIssueMatches", func() {
					filter := &entity.IssueFilter{HasIssueMatches: true}

					entries, err := db.GetIssues(filter, nil)

					Expect(err).To(BeNil())
					for _, entry := range entries {
						hasMatch := lo.ContainsBy(
							seedCollection.IssueMatchRows,
							func(im mariadb.IssueMatchRow) bool {
								return im.IssueId.Int64 == entry.Issue.Id
							},
						)
						Expect(
							hasMatch,
						).To(BeTrue(), "Entry should have at least one matching IssueMatchRow")
					}
				})
				It("can filter by issueMatch severity", func() {
					for _, severity := range entity.AllSeverityValues {
						issueIds := lo.FilterMap(
							seedCollection.IssueMatchRows,
							func(im mariadb.IssueMatchRow, _ int) (int64, bool) {
								return im.IssueId.Int64, im.Rating.String == severity.String()
							},
						)

						filter := &entity.IssueFilter{
							IssueMatchSeverity: []*string{new(severity.String())},
						}

						entries, err := db.GetIssues(filter, nil)

						Expect(err).To(BeNil())
						for _, entry := range entries {
							Expect(
								lo.Contains(issueIds, entry.Issue.Id),
							).To(BeTrue(), "Entry should have severity %s", severity.String())
						}
					}
				})
				It("can filter issue PrimaryName using wild card search", func() {
					row := test.PickOne(seedCollection.IssueRows)

					searchStr := test.CutString(row.PrimaryName.String, 2, 2, 5)
					filter := &entity.IssueFilter{Search: []*string{&searchStr}}

					entries, err := db.GetIssues(filter, nil)

					issueIds := []int64{}
					for _, entry := range entries {
						issueIds = append(issueIds, entry.Issue.Id)
					}

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("at least one element was discarded (filtered)", func() {
						Expect(len(seedCollection.IssueRows) > len(issueIds)).To(BeTrue())
					})

					By("returning the expected elements", func() {
						Expect(issueIds).To(ContainElement(row.Id.Int64))
					})
				})
				It("can filter issue variant SecondaryName using wild card search", func() {
					// select an issueVariant
					issueVariantRow := test.PickOne(seedCollection.IssueVariantRows)

					searchStr := test.CutString(issueVariantRow.SecondaryName.String, 2, 2, 5)
					filter := &entity.IssueFilter{Search: []*string{&searchStr}}

					entries, err := db.GetIssues(filter, nil)

					issueIds := []int64{}
					for _, entry := range entries {
						issueIds = append(issueIds, entry.Issue.Id)
					}

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the expected elements", func() {
						Expect(issueIds).To(ContainElement(issueVariantRow.IssueId.Int64))
					})
				})
				When("filtered by IssueStatus", func() {
					var remediation entity.Remediation
					BeforeEach(func() {
						issueMatch := test.PickOne(seedCollection.IssueMatchRows)
						remediation = entity_test.NewFakeRemediationEntity()
						remediation.ExpirationDate = time.Now().Add(10 * 24 * time.Hour)
						remediation.IssueId = issueMatch.IssueId.Int64
						remediation.ComponentId = 0 // component is optional
						remediation.CreatedBy = util.SystemUserId
						remediation.UpdatedBy = util.SystemUserId

						ci, _ := lo.Find(
							seedCollection.ComponentInstanceRows,
							func(cir mariadb.ComponentInstanceRow) bool {
								return cir.Id.Int64 == issueMatch.ComponentInstanceId.Int64
							},
						)

						remediation.ServiceId = ci.ServiceId.Int64

						_, err := db.CreateRemediation(&remediation)

						Expect(err).To(BeNil())
					})
					It("can filter issue by IssueStatusOpen", func() {
						filter := &entity.IssueFilter{Status: entity.IssueStatusOpen}

						entries, err := db.GetIssues(filter, nil)

						By("throwing no error", func() {
							Expect(err).To(BeNil())
						})

						for _, entry := range entries {
							Expect(entry.Issue.Id).ToNot(BeEquivalentTo(remediation.IssueId))
						}
					})
					It("can filter issue by IssueStatusRemediated", func() {
						filter := &entity.IssueFilter{Status: entity.IssueStatusRemediated}

						entries, err := db.GetIssues(filter, nil)

						By("throwing no error", func() {
							Expect(err).To(BeNil())
						})

						issueIds := lo.Map(entries, func(e entity.IssueResult, _ int) int64 {
							return e.Issue.Id
						})

						Expect(lo.Contains(issueIds, remediation.IssueId)).To(BeTrue())
					})
				})
			})
			Context("and using pagination", func() {
				DescribeTable("can correctly paginate", func(pageSize int) {
					test.TestPaginationOfListWithOrder(
						db.GetIssues,
						func(first *int, after *string) *entity.IssueFilter {
							return &entity.IssueFilter{
								Paginated: entity.Paginated{First: first, After: after},
							}
						},
						[]entity.Order{},
						func(entries []entity.IssueResult) string {
							after, _ := mariadb.EncodeCursor(
								mariadb.WithIssue(
									[]entity.Order{},
									*entries[len(entries)-1].Issue,
									0,
								),
							)
							return after
						},
						len(seedCollection.IssueRows),
						pageSize,
					)
				},
					Entry("when pageSize is 1", 1),
					Entry("when pageSize is 3", 3),
					Entry("when pageSize is 5", 5),
					Entry("when pageSize is 11", 11),
					Entry("when pageSize is 100", 100),
				)
			})
		})
	})
	When("Getting Issues with Aggregations", Label("GetIssuesWithAggregations"), func() {
		Context("and the database contains service without aggregations", func() {
			BeforeEach(func() {
				newIssueRow := test.NewFakeIssue()
				newIssue := newIssueRow.AsIssue()
				db.CreateIssue(&newIssue)
			})
			It("returns the issues with aggregations", func() {
				entriesWithAggregations, err := db.GetIssuesWithAggregations(nil, nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				By("returning some aggregations", func() {
					for _, entryWithAggregations := range entriesWithAggregations {
						Expect(entryWithAggregations).NotTo(
							BeEquivalentTo(entity.IssueAggregations{}))
						Expect(
							entryWithAggregations.IssueAggregations.IssueMatches,
						).To(BeEquivalentTo(0))
						Expect(
							entryWithAggregations.IssueAggregations.AffectedServices,
						).To(BeEquivalentTo(0))
						Expect(
							entryWithAggregations.IssueAggregations.AffectedComponentInstances,
						).To(BeEquivalentTo(0))
						Expect(
							entryWithAggregations.IssueAggregations.ComponentVersions,
						).To(BeEquivalentTo(0))
					}
				})
				By("returning all issues", func() {
					Expect(len(entriesWithAggregations)).To(BeEquivalentTo(1))
				})
			})
		})
		Context("and and we have 10 elements in the database", func() {
			BeforeEach(func() {
				_ = seeder.SeedDbWithNFakeData(10)
			})
			It("returns the issues with aggregations", func() {
				entriesWithAggregations, err := db.GetIssuesWithAggregations(nil, nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				By("returning some aggregations", func() {
					for _, entryWithAggregations := range entriesWithAggregations {
						Expect(entryWithAggregations).NotTo(
							BeEquivalentTo(entity.IssueAggregations{}))
					}
				})
				By("returning all ld constraints exclude all Go files inservices", func() {
					Expect(len(entriesWithAggregations)).To(BeEquivalentTo(10))
				})
			})
			It("returns correct aggregation values", func() {
				//Should be filled with a check for each aggregation value,
				// this is currently skipped due to the complexity of the test implementation
				// as we would need to implement for each of the aggregations a manual aggregation
				// based on the seederCollection.
				//
				// This tests should therefore only get implemented in case we encourage errors in
				// this area to test against
				// possible regressions
			})
		})
	})
	When("Counting Issues", Label("CountIssues"), func() {
		Context("and using no filter", func() {
			DescribeTable("it returns correct count", func(x int) {
				_ = seeder.SeedDbWithNFakeData(x)
				res, err := db.CountIssues(nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				By("returning the correct count", func() {
					Expect(res).To(BeEquivalentTo(x))
				})
			},
				Entry("when page size is 0", 0),
				Entry("when page size is 1", 1),
				Entry("when page size is 11", 11),
				Entry("when page size is 100", 100),
			)
			Context("and counting issue types", func() {
				var seedCollection *test.SeedCollection
				BeforeEach(func() {
					seedCollection = seeder.SeedDbWithNFakeData(20)
				})
				It("returns the correct count for each issue type", func() {
					vulnerabilityCount := 0
					policyViolationCount := 0
					securityEventCount := 0

					for _, issue := range seedCollection.IssueRows {
						switch issue.Type.String {
						case entity.IssueTypeVulnerability.String():
							vulnerabilityCount++
						case entity.IssueTypePolicyViolation.String():
							policyViolationCount++
						case entity.IssueTypeSecurityEvent.String():
							securityEventCount++
						}
					}

					issueTypeCounts, err := db.CountIssueTypes(nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the correct counts", func() {
						Expect(
							issueTypeCounts.VulnerabilityCount,
						).To(BeEquivalentTo(vulnerabilityCount))
						Expect(
							issueTypeCounts.PolicyViolationCount,
						).To(BeEquivalentTo(policyViolationCount))
						Expect(
							issueTypeCounts.SecurityEventCount,
						).To(BeEquivalentTo(securityEventCount))
					})
				})
			})
		})
		Context("and using a filter", func() {
			Context("and having 20 elements in the Database", func() {
				var seedCollection *test.SeedCollection
				BeforeEach(func() {
					seedCollection = seeder.SeedDbWithNFakeData(20)
				})
				It("does not influence the count when pagination is applied", func() {
					first := 1
					after := ""
					filter := &entity.IssueFilter{
						Paginated: entity.Paginated{
							First: &first,
							After: &after,
						},
					}
					res, err := db.CountIssues(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the correct count", func() {
						Expect(res).To(BeEquivalentTo(20))
					})
				})
				It("does show the correct amount when filtering for a service name", func() {
					var row mariadb.BaseServiceRow
					searchingRow := true
					var issueRows []mariadb.IssueRow

					// get a service that should return at least 1 issue
					for searchingRow {
						row = test.PickOne(seedCollection.ServiceRows)
						issueRows = seedCollection.GetIssueByService(&row)
						searchingRow = len(issueRows) > 0
					}
					filter := &entity.IssueFilter{ServiceCCRN: []*string{&row.CCRN.String}}

					count, err := db.CountIssues(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the correct count", func() {
						Expect(count).To(BeEquivalentTo(len(issueRows)))
					})
				})
				It("does show the correct amount when filtering for a service id", func() {
					var row mariadb.BaseServiceRow
					searchingRow := true
					var issueRows []mariadb.IssueRow

					// get a service that should return at least 1 issue
					for searchingRow {
						row = test.PickOne(seedCollection.ServiceRows)
						issueRows = seedCollection.GetIssueByService(&row)
						searchingRow = len(issueRows) > 0
					}
					filter := &entity.IssueFilter{ServiceId: []*int64{&row.Id.Int64}}

					count, err := db.CountIssues(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the correct count", func() {
						Expect(count).To(BeEquivalentTo(len(issueRows)))
					})
				})
			})
		})
	})
	When("IssueCounts by Severity", Label("IssueCounts"), func() {
		testIssueSeverityCount := func(filter *entity.IssueFilter, counts entity.IssueSeverityCounts) {
			issueSeverityCounts, err := db.CountIssueRatings(filter)

			By("throwing no error", func() {
				Expect(err).To(BeNil())
			})

			By("returning the correct counts", func() {
				Expect(issueSeverityCounts.Critical).To(BeEquivalentTo(counts.Critical))
				Expect(issueSeverityCounts.High).To(BeEquivalentTo(counts.High))
				Expect(issueSeverityCounts.Medium).To(BeEquivalentTo(counts.Medium))
				Expect(issueSeverityCounts.Low).To(BeEquivalentTo(counts.Low))
				Expect(issueSeverityCounts.None).To(BeEquivalentTo(counts.None))
				Expect(issueSeverityCounts.Total).To(BeEquivalentTo(counts.Total))
			})
		}
		Context("and counting issue severities", func() {
			var seedCollection *test.SeedCollection
			var err error
			BeforeEach(func() {
				seedCollection, err = seeder.SeedForIssueCounts()
				Expect(err).To(BeNil())
				err = seeder.RefreshCountIssueRatings()
				Expect(err).To(BeNil())
			})

			It("returns the correct count for all services", func() {
				severityCounts, err := test.LoadIssueCounts(
					test.GetTestDataPath(
						"../mariadb/testdata/issue_counts/issue_counts_per_severity.json",
					),
				)
				Expect(err).To(BeNil())

				filter := &entity.IssueFilter{
					AllServices: true,
				}

				testIssueSeverityCount(filter, severityCounts)
			})
			It("returns the correct count for services in support goups", func() {
				severityCounts, err := test.LoadSupportGroupIssueCounts(
					test.GetTestDataPath(
						"../mariadb/testdata/issue_counts/issue_counts_per_support_group.json",
					),
				)
				Expect(err).To(BeNil())

				for _, sg := range seedCollection.SupportGroupRows {

					filter := &entity.IssueFilter{
						AllServices:      true,
						SupportGroupCCRN: []*string{&sg.CCRN.String},
					}

					strId := fmt.Sprintf("%d", sg.Id.Int64)

					testIssueSeverityCount(filter, severityCounts[strId])
				}
			})
			It("returns the correct count for component version issues", func() {
				severityCounts, err := test.LoadComponentVersionIssueCounts(
					test.GetTestDataPath(
						"../mariadb/testdata/issue_counts/issue_counts_per_component_version.json",
					),
				)
				Expect(err).To(BeNil())

				for _, cv := range seedCollection.ComponentVersionRows {
					filter := &entity.IssueFilter{
						ComponentVersionId: []*int64{&cv.Id.Int64},
					}

					strId := fmt.Sprintf("%d", cv.Id.Int64)

					testIssueSeverityCount(filter, severityCounts[strId])
				}
			})
			It("returns the correct count for services", func() {
				severityCounts, err := test.LoadServiceIssueCounts(
					test.GetTestDataPath(
						"../mariadb/testdata/issue_counts/issue_counts_per_service.json",
					),
				)
				Expect(err).To(BeNil())

				for _, service := range seedCollection.ServiceRows {
					filter := &entity.IssueFilter{
						ServiceId: []*int64{&service.Id.Int64},
					}

					strId := fmt.Sprintf("%d", service.Id.Int64)

					testIssueSeverityCount(filter, severityCounts[strId])
				}
			})
			It("returns the correct count for supportgroup", func() {
				severityCounts, err := test.LoadSupportGroupIssueCounts(
					test.GetTestDataPath(
						"../mariadb/testdata/issue_counts/issue_counts_per_support_group.json",
					),
				)
				Expect(err).To(BeNil())

				for _, sg := range seedCollection.SupportGroupRows {

					filter := &entity.IssueFilter{
						SupportGroupCCRN: []*string{&sg.CCRN.String},
					}

					strId := fmt.Sprintf("%d", sg.Id.Int64)

					testIssueSeverityCount(filter, severityCounts[strId])
				}
			})
			It("returns the correct count for unique filter", Label("ABCDEF"), func() {
				severityCounts, err := test.LoadIssueCounts(
					test.GetTestDataPath(
						"../mariadb/testdata/issue_counts/issue_counts_per_severity.json",
					),
				)
				Expect(err).To(BeNil())
				// Create a new IM that attaches an existing issue to a different component instance
				im := test.NewFakeIssueMatch()
				im.ComponentInstanceId = sql.NullInt64{Int64: 1, Valid: true}
				im.IssueId = sql.NullInt64{Int64: 3, Valid: true}
				im.UserId = sql.NullInt64{Int64: util.SystemUserId, Valid: true}
				_, err = seeder.InsertFakeIssueMatch(im)
				Expect(err).To(BeNil())

				filter := &entity.IssueFilter{
					AllServices: true,
					Unique:      true,
				}

				testIssueSeverityCount(filter, severityCounts)
			})
		})
	})
	When("Insert Issue", Label("InsertIssue"), func() {
		Context("and we have 10 Issues in the database", func() {
			var newIssueRow mariadb.IssueRow
			var newIssue entity.Issue
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
				newIssueRow = test.NewFakeIssue()
				newIssue = newIssueRow.AsIssue()
			})
			It("can insert correctly", func() {
				issue, err := db.CreateIssue(&newIssue)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("sets issue id", func() {
					Expect(issue).NotTo(BeEquivalentTo(0))
				})

				issueFilter := &entity.IssueFilter{
					Id: []*int64{&issue.Id},
				}

				i, err := db.GetIssues(issueFilter, nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning issue", func() {
					Expect(len(i)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(i[0].Issue.PrimaryName).To(BeEquivalentTo(issue.PrimaryName))
					Expect(i[0].Issue.Type.String()).To(BeEquivalentTo(issue.Type.String()))
					Expect(i[0].Issue.Description).To(BeEquivalentTo(issue.Description))
				})
			})
			It("does not insert issue with existing primary name", func() {
				issueRow := seedCollection.IssueRows[0]
				issue := issueRow.AsIssue()
				newIssue, err := db.CreateIssue(&issue)

				By("throwing error", func() {
					Expect(err).ToNot(BeNil())
				})
				By("no issue returned", func() {
					Expect(newIssue).To(BeNil())
				})
			})
		})
	})
	When("Update Issue", Label("UpdateIssue"), func() {
		Context("and we have 10 Issues in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			It("can update issue description correctly", func() {
				issue := seedCollection.IssueRows[0].AsIssue()

				issue.Description = "New Description"
				err := db.UpdateIssue(&issue)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				issueFilter := &entity.IssueFilter{
					Id: []*int64{&issue.Id},
				}

				i, err := db.GetIssues(issueFilter, nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning issue", func() {
					Expect(len(i)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(i[0].Issue.Description).To(BeEquivalentTo(issue.Description))
				})
			})
		})
	})
	When("Delete Issue", Label("DeleteIssue"), func() {
		Context("and we have 10 Issues in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			It("can delete issue correctly", func() {
				issue := seedCollection.IssueRows[0].AsIssue()

				err := db.DeleteIssue(issue.Id, util.SystemUserId)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				issueFilter := &entity.IssueFilter{
					Id: []*int64{&issue.Id},
				}

				i, err := db.GetIssues(issueFilter, nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning no issue", func() {
					Expect(len(i)).To(BeEquivalentTo(0))
				})
			})
		})
	})
	When("Add Component Version to Issue", Label("AddComponentVersionToIssue"), func() {
		Context("and we have 10 Issues in the database", func() {
			var seedCollection *test.SeedCollection
			var newComponentVersionRow mariadb.ComponentVersionRow
			var newComponentVersion entity.ComponentVersion
			var componentVersion *entity.ComponentVersion
			var issue entity.Issue

			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)

				newComponentVersionRow = test.NewFakeComponentVersion()
				newComponentVersionRow.ComponentId = seedCollection.ComponentRows[0].Id
				newComponentVersion = newComponentVersionRow.AsComponentVersion()

				componentVersion, _ = db.CreateComponentVersion(&newComponentVersion)

				issue = seedCollection.IssueRows[0].AsIssue()
			})

			It("adds component version correctly", func() {
				err := db.AddComponentVersionToIssue(issue.Id, componentVersion.Id)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				issueFilter := &entity.IssueFilter{
					Id: []*int64{&issue.Id},
				}

				i, err := db.GetIssues(issueFilter, nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				By("returning issue", func() {
					Expect(i).To(HaveLen(1))
				})
			})

			It("does nothing if it is already added", func() {
				err := db.AddComponentVersionToIssue(issue.Id, componentVersion.Id)
				Expect(err).To(BeNil())

				err = db.AddComponentVersionToIssue(issue.Id, componentVersion.Id)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
			})
		})
	})
	When("Remove Component Version from Issue", Label("RemoveComponentVersionFromIssue"), func() {
		Context("and we have 10 Issues in the database", func() {
			var seedCollection *test.SeedCollection
			var componentVersionIssueRow mariadb.ComponentVersionIssueRow
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
				componentVersionIssueRow = seedCollection.ComponentVersionIssueRows[0]
			})
			It("can remove component version correctly", func() {
				err := db.RemoveComponentVersionFromIssue(
					componentVersionIssueRow.IssueId.Int64,
					componentVersionIssueRow.ComponentVersionId.Int64,
				)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				issueFilter := &entity.IssueFilter{
					ComponentVersionId: []*int64{
						&componentVersionIssueRow.ComponentVersionId.Int64,
					},
				}

				issues, err := db.GetIssues(issueFilter, nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				for _, issue := range issues {
					Expect(
						issue.Issue.Id,
					).ToNot(BeEquivalentTo(componentVersionIssueRow.IssueId.Int64))
				}
			})
		})
	})
})

var _ = Describe("Ordering Issues", Label("IssueOrder"), func() {
	var db *mariadb.SqlDatabase
	var seeder *test.DatabaseSeeder
	var seedCollection *test.SeedCollection

	BeforeEach(func() {
		var err error
		db = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")
	})
	AfterEach(func() {
		dbm.TestTearDown(db)
	})

	testOrder := func(
		order []entity.Order,
		verifyFunc func(res []entity.IssueResult),
	) {
		res, err := db.GetIssues(nil, order)

		By("throwing no error", func() {
			Expect(err).Should(BeNil())
		})

		By("returning the correct number of results", func() {
			Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.IssueRows)))
		})

		By("returning the correct order", func() {
			verifyFunc(res)
		})
	}

	When("with ASC order", Label("IssueASCOrder"), func() {
		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
			seedCollection.GetValidIssueMatchRows()
		})

		It("can order by id", func() {
			sort.Slice(seedCollection.IssueRows, func(i, j int) bool {
				return seedCollection.IssueRows[i].Id.Int64 < seedCollection.IssueRows[j].Id.Int64
			})

			order := []entity.Order{
				{By: entity.IssueId, Direction: entity.OrderDirectionAsc},
			}

			testOrder(order, func(res []entity.IssueResult) {
				for i, r := range res {
					Expect(r.Issue.Id).Should(BeEquivalentTo(seedCollection.IssueRows[i].Id.Int64))
				}
			})
		})

		It("can order by primaryName", func() {
			order := []entity.Order{
				{By: entity.IssuePrimaryName, Direction: entity.OrderDirectionAsc},
			}

			testOrder(order, func(res []entity.IssueResult) {
				prev := ""
				for _, r := range res {
					Expect(r).ShouldNot(BeNil())
					Expect(r.PrimaryName >= prev).Should(BeTrue())
					prev = r.PrimaryName
				}
			})
		})

		It("can order by rating", func() {
			order := []entity.Order{
				{By: entity.IssueVariantRating, Direction: entity.OrderDirectionAsc},
			}

			testOrder(order, func(res []entity.IssueResult) {
				prev := -10
				for _, r := range res {
					variants := seedCollection.GetIssueVariantsByIssueId(r.Issue.Id)
					ratings := lo.Map(variants, func(iv mariadb.IssueVariantRow, _ int) int {
						return test.SeverityToNumerical(iv.Rating.String)
					})
					highestRating := lo.Max(ratings)
					Expect(highestRating >= prev).Should(BeTrue())
				}
			})
		})
	})

	When("with DESC order", Label("IssueDESCOrder"), func() {
		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		It("can order by id", func() {
			sort.Slice(seedCollection.IssueRows, func(i, j int) bool {
				return seedCollection.IssueRows[i].Id.Int64 > seedCollection.IssueRows[j].Id.Int64
			})

			order := []entity.Order{
				{By: entity.IssueId, Direction: entity.OrderDirectionDesc},
			}

			testOrder(order, func(res []entity.IssueResult) {
				for i, r := range res {
					Expect(r.Issue.Id).Should(BeEquivalentTo(seedCollection.IssueRows[i].Id.Int64))
				}
			})
		})

		It("can order by primaryName", func() {
			order := []entity.Order{
				{By: entity.IssuePrimaryName, Direction: entity.OrderDirectionDesc},
			}

			testOrder(order, func(res []entity.IssueResult) {
				prev := "\U0010FFFF"
				for _, r := range res {
					Expect(r).ShouldNot(BeNil())
					Expect(r.PrimaryName <= prev).Should(BeTrue())
					prev = r.PrimaryName
				}
			})
		})

		It("can order by rating", func() {
			order := []entity.Order{
				{By: entity.IssueVariantRating, Direction: entity.OrderDirectionDesc},
			}

			testOrder(order, func(res []entity.IssueResult) {
				prev := 9999
				for _, r := range res {
					variants := seedCollection.GetIssueVariantsByIssueId(r.Issue.Id)
					ratings := lo.Map(variants, func(iv mariadb.IssueVariantRow, _ int) int {
						return test.SeverityToNumerical(iv.Rating.String)
					})
					highestRating := lo.Max(ratings)
					Expect(highestRating <= prev).Should(BeTrue())
				}
			})
		})
	})
})
*/

/*
package mariadb

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	sq "github.com/Masterminds/squirrel"
)

// DbObject
type DbObject[ET entity.Entity] struct {
	Prefix           string
	TableName        string
	Properties       []*Property
	FilterProperties []*FilterProperty
	JoinDefs         []*JoinDef
}

func (do *DbObject[ET]) InsertQuery(entityItem ET) (string, []any, error) {
	columns := lo.Map(do.Properties, func(p *Property, _ int) string {
		return p.GetName()
	})

	values := lo.Map(do.Properties, func(p *Property, _ int) any {
		return p.GetValue(entityItem)
	})

	qb := sq.
		Insert(do.TableName).
		Columns(columns...).
		Values(values...)

	return qb.ToSql()
}

func (do *DbObject[ET]) GetUpdateMap(f any) map[string]any {
	m := make(map[string]any)

	for _, v := range do.Properties {
		val, isUpdatePresent := v.GetUpdateData(f)
		if isUpdatePresent {
			m[v.GetName()] = val
		}
	}

	return m
}

func (do *DbObject[ET]) GetFilterQuery(filter any) string {
	var fl []string
	for _, v := range do.FilterProperties {
		fl = append(fl, v.GetQuery(filter))
	}

	return combineFilterQueries(fl, OP_AND)
}

func (do *DbObject[ET]) GetFilterParameters(
	filter entity.HasPagination,
	withCursor bool,
	cursorFields []Field,
) []any {
	var filterParameters []any
	for _, v := range do.FilterProperties {
		filterParameters = v.AppendParameters(filterParameters, filter)
	}

	if withCursor {
		paginatedX := filter.GetPaginated()
		filterParameters = append(
			filterParameters,
			GetCursorQueryParameters(paginatedX.First, cursorFields)...)
	}

	return filterParameters
}

func (do *DbObject[ET]) Create(db Db, entityItem ET) (ET, error) {
	var zero ET

	l := logrus.WithFields(logrus.Fields{
		do.Prefix: entityItem,
		"event":   fmt.Sprintf("database.Create%s", do.TableName),
	})

	sqlQuery, args, err := do.InsertQuery(entityItem)
	if err != nil {
		return zero, err
	}

	id, err := PerformInsertArgs(db, sqlQuery, args, l)
	if err != nil {
		if strings.HasPrefix(err.Error(), "Error 1062") {
			return zero, database.NewDuplicateEntryDatabaseError(
				fmt.Sprintf("%s element already exists", do.TableName),
			)
		}

		return zero, err
	}

	entityItem.SetId(id)

	return entityItem, nil
}

func (do *DbObject[ET]) Update(db Db, entityItem ET) error {
	l := logrus.WithFields(logrus.Fields{
		do.Prefix: entityItem,
		"event":   fmt.Sprintf("database.Update%s", do.TableName),
	})

	updateValues := do.GetUpdateMap(entityItem)
	qb := sq.
		Update(do.TableName).
		SetMap(updateValues).
		Where(sq.Eq{fmt.Sprintf("%s_id", do.Prefix): entityItem.GetId()})

	sqlQuery, args, err := qb.ToSql()
	if err != nil {
		return err
	}

	_, err = PerformExecArgs(db, sqlQuery, args, l)

	return err
}

func (do *DbObject[ET]) Delete(db Db, id int64, userId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"id":    id,
		"event": fmt.Sprintf("database.Delete%s", do.TableName),
	})

	deletedAtColumn := do.Prefix + "_deleted_at"
	updatedByColumn := do.Prefix + "_updated_by"
	idColumn := do.Prefix + "_id"

	qb := sq.
		Update(do.TableName).
		Set(deletedAtColumn, sq.Expr("NOW()")).
		Set(updatedByColumn, userId).
		Where(sq.Eq{idColumn: id})

	sqlQuery, args, err := qb.ToSql()
	if err != nil {
		return err
	}

	_, err = PerformExecArgs(db, sqlQuery, args, l)

	return err
}

func (do *DbObject[ET]) GetJoins(filter any, order []entity.Order) string {
	return NewJoinResolver(do.JoinDefs).Build(filter, order)
}

// Property
const NoUpdate = false

func NewProperty(name string, access func(any) (any, bool)) *Property {
	return &Property{Name: name, Access: access}
}

type Property struct {
	Name   string
	Access func(any) (any, bool)
}

func (p Property) GetName() string {
	return p.Name
}

func (p Property) GetValue(f any) any {
	val, _ := p.Access(f)
	return val
}

func (p Property) GetUpdateData(f any) (any, bool) {
	return p.Access(f)
}

// FilterProperty
type FilterProperty struct {
	QueryBuilder  func([]any) string
	Param         func(any) []any
	ParamAppender func([]any, any) []any
}

func (fp FilterProperty) AppendParameters(params []any, filter any) []any {
	return fp.ParamAppender(params, filter)
}

func (fp FilterProperty) GetQuery(filter any) string {
	return fp.QueryBuilder(fp.Param(filter))
}

func doNotAppendParameters(params []any, _ any) []any {
	return params
}

func NewFilterProperty(query string, param func(any) []any) *FilterProperty {
	return NewNFilterProperty(query, param, 1)
}

func NewNFilterProperty(query string, param func(any) []any, nparam int) *FilterProperty {
	return &FilterProperty{
		QueryBuilder:  func(filter []any) string { return buildFilterQuery(filter, query, OP_OR) },
		Param:         param,
		ParamAppender: func(params []any, filter any) []any { return buildQueryParametersCount(params, param(filter), nparam) },
	}
}

func NewStateFilterProperty(
	prefix string,
	param func(any) []entity.StateFilterType,
) *FilterProperty {
	return &FilterProperty{
		QueryBuilder:  func(state []any) string { return buildStateFilterQuery(ToStateSlice(state), prefix) },
		Param:         WrapRetSlice(param),
		ParamAppender: doNotAppendParameters,
	}
}

func NewJsonFilterProperty(query string, param func(any) []*entity.Json) *FilterProperty {
	return &FilterProperty{
		QueryBuilder:  func(json []any) string { return buildJsonFilterQuery(ToJsonSlice(json), query, OP_OR) },
		Param:         WrapRetSlice(param),
		ParamAppender: func(params []any, filter any) []any { return buildJsonQueryParameters(params, param(filter)) },
	}
}

func NewCustomFilterProperty(
	queryBuilder func([]any) string,
	param func(any) []any,
) *FilterProperty {
	return &FilterProperty{
		QueryBuilder:  queryBuilder,
		Param:         param,
		ParamAppender: doNotAppendParameters,
	}
}

// Join
type JoinType string

const (
	LeftJoin  JoinType = "LEFT JOIN"
	RightJoin JoinType = "RIGHT JOIN"
	InnerJoin JoinType = "JOIN"
)

func DependentJoin(any, []entity.Order) bool { return false }

type JoinDef struct {
	Name      string
	Type      JoinType
	Table     string
	On        string
	DependsOn []string
	Condition func(any, []entity.Order) bool
}

type JoinResolver struct {
	defs     map[string]*JoinDef
	included map[string]bool
	order    []string
}

func NewJoinResolver(defs []*JoinDef) *JoinResolver {
	r := &JoinResolver{
		defs:     map[string]*JoinDef{},
		included: map[string]bool{},
	}
	for _, d := range defs {
		r.defs[d.Name] = d
	}

	return r
}

func (jr *JoinResolver) require(name string) {
	if jr.included[name] {
		return
	}

	def, ok := jr.defs[name]
	if !ok {
		panic("Unknown join: " + name)
	}

	// resolve dependencies first
	for _, dep := range def.DependsOn {
		jr.require(dep)
	}

	jr.included[name] = true
	jr.order = append(jr.order, name)
}

func (jr *JoinResolver) Build(filter any, order []entity.Order) string {
	for _, def := range jr.defs {
		if def.Condition != nil && def.Condition(filter, order) {
			jr.require(def.Name)
		}
	}

	var result []string

	// This is little tricky part, but we need to deal with that this way
	// until we have stateful join pattern which is created for issue.go
	// with non-uniq tablename 'IM IssueMatch' which join operation
	// depends on filter pattern with precedence for some members (there
	// is if...else if which cannot be replaced by if... and if... what
	// is a mess and misconception
	uniqTableName := make(map[string]struct{})

	for _, name := range jr.order {
		j := jr.defs[name]

		if _, ok := uniqTableName[j.Table]; ok {
			continue
		}

		uniqTableName[j.Table] = struct{}{}

		joinSQL := fmt.Sprintf("%s %s ON %s",
			j.Type,
			j.Table,
			j.On,
		)

		result = append(result, joinSQL)
	}

	return strings.Join(result, "\n")
}

// DB helpers
func EnsurePagination[T entity.HasPagination](filter T) T {
	first := 1000
	after := ""

	px := filter.GetPaginated()

	if px.First == nil {
		px.First = &first
	}

	if px.After == nil {
		px.After = &after
	}

	return filter
}

func PerformExecArgs(db Db, query string, args []any, l *logrus.Entry) (sql.Result, error) {
	res, err := db.Exec(query, args...)
	if err != nil {
		msg := err.Error()
		l.WithFields(logrus.Fields{
			"error": err,
			"query": query,
			"args":  args,
		}).Error(msg)

		return nil, fmt.Errorf("%s", msg)
	}

	return res, nil
}

func PerformInsertArgs(db Db, query string, args []any, l *logrus.Entry) (int64, error) {
	res, err := PerformExecArgs(db, query, args, l)
	if err != nil {
		return -1, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		msg := "Error while getting last insert id"
		l.WithFields(logrus.Fields{
			"error": err,
		}).Error(msg)

		return -1, fmt.Errorf("%s", msg)
	}

	l.WithFields(logrus.Fields{
		"id": id,
	}).Debug("Successfully performed insert")

	return id, nil
}

// Helpers

// WrapAccess turns a type-specific data access into a generic data access
func WrapAccess[T any, TRet any](access func(T) (TRet, bool)) func(any) (any, bool) {
	return func(val any) (any, bool) {
		typedVal, ok := val.(T)
		if !ok {
			panic(fmt.Sprintf("WrapAccess: expected %T but got %T", *new(T), val))
		}

		return access(typedVal)
	}
}

// WrapBuilder turns a type-specific builder function into a generic builder function
func WrapBuilder[T any](build func([]T) string) func([]any) string {
	return func(values []any) string {
		typed := make([]T, len(values))

		for i, v := range values {
			tv, ok := v.(T)
			if !ok {
				panic(fmt.Sprintf(
					"WrapBuilderSlice: expected %T but got %T",
					*new(T), v,
				))
			}

			typed[i] = tv
		}

		return build(typed)
	}
}

// WrapRetSlice turns a type-specific accessor into a generic one
func WrapRetSlice[T any, E any](fn func(T) []E) func(any) []any {
	return func(input any) []any {
		val, ok := input.(T)
		if !ok {
			panic(fmt.Sprintf("WrapRetSlice: expected %T but got %T", *new(T), input))
		}

		res := fn(val)

		out := make([]any, len(res))
		for i := range res {
			out[i] = res[i]
		}

		return out
	}
}

// WrapRetState turns a type-specific accessor into a generic one for StateFilter slice
func WrapRetState[T any](fn func(T) []entity.StateFilterType) func(any) []entity.StateFilterType {
	return func(input any) []entity.StateFilterType {
		val, ok := input.(T)
		if !ok {
			panic(fmt.Sprintf("WrapRetState: expected %T but got %T", *new(T), input))
		}

		res := fn(val)

		out := make([]entity.StateFilterType, len(res))

		copy(out, res)

		return out
	}
}

// WrapRetJson turns a type-specific accessor into a generic one for Json slice
func WrapRetJson[T any](fn func(T) []*entity.Json) func(any) []*entity.Json {
	return func(input any) []*entity.Json {
		val, ok := input.(T)
		if !ok {
			panic(fmt.Sprintf("WrapRetJson: expected %T but got %T", *new(T), input))
		}

		res := fn(val)

		out := make([]*entity.Json, len(res))
		copy(out, res)

		return out
	}
}

// WrapJoinCondition turns a type-specific join planner condition using filter and order
func WrapJoinCondition[T any](joinCond func(T, []entity.Order) bool) func(any, []entity.Order) bool {
	return func(filter any, order []entity.Order) bool {
		typedFilter, ok := filter.(T)
		if !ok {
			panic(fmt.Sprintf("WrapJoinCondition: expected %T but got %T", *new(T), filter))
		}

		return joinCond(typedFilter, order)
	}
}

func ToStateSlice(in []any) []entity.StateFilterType {
	out := make([]entity.StateFilterType, len(in))
	for i := range in {
		s, ok := in[i].(entity.StateFilterType)
		if !ok {
			panic(
				fmt.Sprintf(
					"ToStateSlice: expected %T but got %T",
					new(entity.StateFilterType),
					in[i],
				),
			)
		}

		out[i] = s
	}

	return out
}

func ToJsonSlice(in []any) []*entity.Json {
	out := make([]*entity.Json, len(in))
	for i := range in {
		s, ok := in[i].(*entity.Json)
		if !ok {
			panic(fmt.Sprintf("ToJsonSlice: expected %T but got %T", new(*entity.Json), in[i]))
		}

		out[i] = s
	}

	return out
}

func ValueOrDefault[T any](p *T, def T) T {
	if p == nil {
		return def
	}

	return *p
}*/

// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	sq "github.com/Masterminds/squirrel"
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

func (f *dummyEntityFilter) Get() any {
	return f
}

func (f *dummyEntityFilter) Ensure() entity.Filter {
	if f == nil {
		return &dummyEntityFilter{}
	}

	return f
}

type dummyResult = entity.ComponentResult

var _ = Describe("DbObject", Label("database", "DbObject"), func() {
	const (
		dummyId                      = 7
		dummyString                  = "ttt"
		anyVal                       = false
		aggregated                   = true
		notAggregated                = false
		dummyCursorFieldVal          = 9
		defaultFirstEntriesForCursor = 1000
		dummySelectItem              = "dummySelection"
		dummySelectQuery             = "SELECT dummySelection"
	)

	dummySelectBuilder := sq.Select(dummySelectItem)
	dummyCursorField := mariadb.Field{Name: entity.OrderByField(4), Value: dummyCursorFieldVal, Order: entity.OrderDirectionDesc}

	When("Insert query is called", func() {
		Context("Two properties are there in the object", func() {
			testObject := mariadb.DbObject[*dummyEntity, *dummyEntityFilter, dummyResult]{
				TableName: "DummyTable",
				Properties: []*mariadb.Property[*dummyEntity]{
					mariadb.NewProperty("id", func(de *dummyEntity) (any, bool) { return de.Id, anyVal }),
					mariadb.NewProperty("a", func(de *dummyEntity) (any, bool) { return de.A, anyVal }),
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
			var testObject mariadb.DbObject[*dummyEntity, *dummyEntityFilter, dummyResult]
			BeforeEach(func() {
				testObject = mariadb.DbObject[*dummyEntity, *dummyEntityFilter, dummyResult]{
					TableName: "DummyTable",
					FilterProperties: []*mariadb.FilterProperty[*dummyEntityFilter]{
						mariadb.NewFilterProperty("DT.dummytable_id = ?", func(filter *dummyEntityFilter) any { return filter.Id }),
						mariadb.NewFilterProperty("DT.dummytable_a = ?", func(filter *dummyEntityFilter) any { return filter.A }),
					},
				}
			})
			It("builds filter for one filter item", func() {
				def := dummyEntityFilter{A: []*string{lo.ToPtr(dummyString)}}
				By("returning correct filter in query", func() {
					qb := testObject.AddFilter(dummySelectBuilder, &def)
					query, params, err := qb.ToSql()
					Expect(err).To(BeNil())
					Expect(params).To(ConsistOf([]any{lo.ToPtr(dummyString)}))
					Expect(query).To(BeEquivalentTo(dummySelectQuery + " WHERE (   DT.dummytable_a = ? )"))
				})
				By("returning correct parameter with cursor in non aggregated query", func() {
					qb := testObject.AddFilter(dummySelectBuilder, &def)
					qb = testObject.AddCursor(qb, &def, []mariadb.Field{dummyCursorField})
					query, params, err := qb.ToSql()
					Expect(err).To(BeNil())
					Expect(params).To(ConsistOf([]any{lo.ToPtr(dummyString), dummyCursorFieldVal}))
					Expect(query).To(BeEquivalentTo(dummySelectQuery + " WHERE (   DT.dummytable_a = ? ) AND (  componentinstance_namespace < ?  ) LIMIT 1000"))
				})
				By("returning correct parameter with cursor in aggregated object", func() {
					testObject.Aggregated = true
					qb := testObject.AddFilter(dummySelectBuilder, &def)
					qb = testObject.AddCursor(qb, &def, []mariadb.Field{dummyCursorField})
					query, params, err := qb.ToSql()
					Expect(err).To(BeNil())
					Expect(params).To(ConsistOf([]any{lo.ToPtr(dummyString), dummyCursorFieldVal}))
					Expect(query).To(BeEquivalentTo(dummySelectQuery + " WHERE (   DT.dummytable_a = ? ) HAVING (  componentinstance_namespace < ?  ) LIMIT 1000"))
				})
			})

			It("builds filter for two filter items", func() {
				id := int64(dummyId)
				def := dummyEntityFilter{Id: []*int64{&id}, A: []*string{lo.ToPtr(dummyString)}}
				By("returning correct filter query string", func() {
					qb := testObject.AddFilter(dummySelectBuilder, &def)
					query, params, err := qb.ToSql()
					Expect(err).To(BeNil())
					Expect(params).To(ConsistOf([]any{lo.ToPtr(int64(dummyId)), lo.ToPtr(dummyString)}))
					Expect(query).To(BeEquivalentTo(dummySelectQuery + " WHERE (   DT.dummytable_id = ? ) AND (   DT.dummytable_a = ? )"))
				})
			})

			It("builds empty filter query string for no filter", func() {
				def := dummyEntityFilter{}
				qb := testObject.AddFilter(dummySelectBuilder, &def)
				query, params, err := qb.ToSql()
				Expect(err).To(BeNil())
				Expect(params).To(BeEmpty())
				Expect(query).To(BeEquivalentTo(dummySelectQuery))
			})
		})
	})
	When("Joins query is build", func() {
		Context("no joins are there in the object", func() {
			testObject := mariadb.DbObject[*dummyEntity, *dummyEntityFilter, dummyResult]{
				TableName: "DummyTable",
			}
			It("gets no joins in query as there is no join defined", func() {
				def := dummyEntityFilter{A: []*string{lo.ToPtr(dummyString)}}
				qb := testObject.AddJoins(dummySelectBuilder, &def, nil)
				query, params, err := qb.ToSql()
				Expect(err).To(BeNil())
				Expect(params).To(BeEmpty())
				Expect(query).To(BeEquivalentTo(dummySelectQuery))
			})
		})
		Context("Two basic joins are there in the object", func() {
			testObject := mariadb.DbObject[*dummyEntity, *dummyEntityFilter, dummyResult]{
				TableName: "DummyTable",
				JoinDefs: []*mariadb.JoinDef[*dummyEntityFilter]{
					{
						Name:      "X",
						Type:      mariadb.LeftJoin,
						Table:     "Xabc X",
						On:        "DT.dummytable_id = X.xabc_dummytable_id",
						Condition: func(f *dummyEntityFilter, _ *mariadb.Order) bool { return len(f.A) > 0 },
					},
					{
						Name:      "Y",
						Type:      mariadb.RightJoin,
						Table:     "Yabc Y",
						On:        "DT.dummytable_id = Y.yabc_dummytable_id",
						Condition: func(f *dummyEntityFilter, _ *mariadb.Order) bool { return len(f.B) > 0 },
					},
				},
			}
			It("gets no joins in query when there is an empty filter in use", func() {
				def := dummyEntityFilter{}
				qb := testObject.AddJoins(dummySelectBuilder, &def, nil)
				query, params, err := qb.ToSql()
				Expect(err).To(BeNil())
				Expect(params).To(BeEmpty())
				Expect(query).To(BeEquivalentTo(dummySelectQuery))
			})
			It("gets one join with X when A mapping condition is met (there is at least one element to filter)", func() {
				def := dummyEntityFilter{A: []*string{lo.ToPtr(dummyString)}}
				qb := testObject.AddJoins(dummySelectBuilder, &def, nil)
				query, params, err := qb.ToSql()
				Expect(err).To(BeNil())
				Expect(params).To(BeEmpty())
				Expect(query).To(BeEquivalentTo(dummySelectQuery + " LEFT JOIN Xabc X ON DT.dummytable_id = X.xabc_dummytable_id"))
			})
			It("gets one join with Y when B mapping condition is met (there is at least one element to filter)", func() {
				def := dummyEntityFilter{B: []*int{lo.ToPtr(10)}}
				qb := testObject.AddJoins(dummySelectBuilder, &def, nil)
				query, params, err := qb.ToSql()
				Expect(err).To(BeNil())
				Expect(params).To(BeEmpty())
				Expect(query).To(BeEquivalentTo(dummySelectQuery + " RIGHT JOIN Yabc Y ON DT.dummytable_id = Y.yabc_dummytable_id"))
			})
			It("gets two joins with X and Y when A and B mapping conditions are met (there are at least one element for each filter)", func() {
				def := dummyEntityFilter{A: []*string{lo.ToPtr(dummyString)}, B: []*int{lo.ToPtr(10)}}
				qb := testObject.AddJoins(dummySelectBuilder, &def, nil)
				query, params, err := qb.ToSql()
				Expect(err).To(BeNil())
				Expect(params).To(BeEmpty())
				Expect(query).To(BeEquivalentTo(
					dummySelectQuery +
						" LEFT JOIN Xabc X ON DT.dummytable_id = X.xabc_dummytable_id" +
						" RIGHT JOIN Yabc Y ON DT.dummytable_id = Y.yabc_dummytable_id",
				))
			})
		})
		Context("Join when dependency is there in the object", func() {
			testObject := mariadb.DbObject[*dummyEntity, *dummyEntityFilter, dummyResult]{
				TableName: "DummyTable",
				JoinDefs: []*mariadb.JoinDef[*dummyEntityFilter]{
					{
						Name:      "X",
						Type:      mariadb.LeftJoin,
						Table:     "Xabc X",
						On:        "DT.dummytable_id = X.xabc_dummytable_id",
						Condition: mariadb.DependentJoin[*dummyEntityFilter],
					},
					{
						Name:      "Y",
						Type:      mariadb.RightJoin,
						Table:     "Yabc Y",
						On:        "X.xabc_yabc_id = Y.yabc_id",
						DependsOn: []string{"X"},
						Condition: func(f *dummyEntityFilter, _ *mariadb.Order) bool { return len(f.B) > 0 },
					},
				},
			}
			It("gets two joins with X (and dependent Y) when B mapping condition is met (there is at least one element to filter)", func() {
				def := dummyEntityFilter{B: []*int{lo.ToPtr(10)}}
				qb := testObject.AddJoins(dummySelectBuilder, &def, nil)
				query, params, err := qb.ToSql()
				Expect(err).To(BeNil())
				Expect(params).To(BeEmpty())
				Expect(query).To(BeEquivalentTo(
					dummySelectQuery +
						" LEFT JOIN Xabc X ON DT.dummytable_id = X.xabc_dummytable_id" +
						" RIGHT JOIN Yabc Y ON X.xabc_yabc_id = Y.yabc_id",
				))
			})
		})
		Context("Two joins when the same dependency are there in the object", func() {
			testObject := mariadb.DbObject[*dummyEntity, *dummyEntityFilter, dummyResult]{
				TableName: "DummyTable",
				JoinDefs: []*mariadb.JoinDef[*dummyEntityFilter]{
					{
						Name:      "X",
						Type:      mariadb.LeftJoin,
						Table:     "Xabc X",
						On:        "DT.dummytable_id = X.xabc_dummytable_id",
						Condition: mariadb.DependentJoin[*dummyEntityFilter],
					},
					{
						Name:      "Y",
						Type:      mariadb.RightJoin,
						Table:     "Yabc Y",
						On:        "X.xabc_yabc_id = Y.yabc_id",
						DependsOn: []string{"X"},
						Condition: func(f *dummyEntityFilter, _ *mariadb.Order) bool { return len(f.B) > 0 },
					},
					{
						Name:      "Z",
						Type:      mariadb.RightJoin,
						Table:     "Zabc Z",
						On:        "X.xabc_zabc_id = Z.zabc_id",
						DependsOn: []string{"X"},
						Condition: func(f *dummyEntityFilter, _ *mariadb.Order) bool { return len(f.A) > 0 },
					},
				},
			}
			It("gets two joins and one dependent join when A and B mapping conditions are met (there is at least one element for each filter)", func() {
				def := dummyEntityFilter{A: []*string{lo.ToPtr(dummyString)}, B: []*int{lo.ToPtr(10)}}
				qb := testObject.AddJoins(dummySelectBuilder, &def, nil)
				query, params, err := qb.ToSql()
				Expect(err).To(BeNil())
				Expect(params).To(BeEmpty())
				Expect(query).To(BeEquivalentTo(
					dummySelectQuery +
						" LEFT JOIN Xabc X ON DT.dummytable_id = X.xabc_dummytable_id" +
						" RIGHT JOIN Yabc Y ON X.xabc_yabc_id = Y.yabc_id" +
						" RIGHT JOIN Zabc Z ON X.xabc_zabc_id = Z.zabc_id",
				))
			})
		})
		Context("Two joins when the same table are there in the object", func() {
			testObject := mariadb.DbObject[*dummyEntity, *dummyEntityFilter, dummyResult]{
				TableName: "DummyTable",
				JoinDefs: []*mariadb.JoinDef[*dummyEntityFilter]{
					{
						Name:      "X_left",
						Type:      mariadb.LeftJoin,
						Table:     "Xabc X",
						On:        "DT.dummytable_id = X.xabc_dummytable_id",
						Condition: func(f *dummyEntityFilter, _ *mariadb.Order) bool { return len(f.B) > 0 },
					},
					{
						Name:      "X_right",
						Type:      mariadb.RightJoin,
						Table:     "Xabc X",
						On:        "DT.dummytable_id = X.xabc_dummytable_id",
						Condition: func(f *dummyEntityFilter, _ *mariadb.Order) bool { return len(f.B) > 0 },
					},
				},
			}
			It("gets one join even when both join defs have the same condition but only first one will be included to prevent join table name duplication", func() {
				def := dummyEntityFilter{B: []*int{lo.ToPtr(10)}}
				qb := testObject.AddJoins(dummySelectBuilder, &def, nil)
				query, params, err := qb.ToSql()
				Expect(err).To(BeNil())
				Expect(params).To(BeEmpty())
				Expect(query).To(BeEquivalentTo(
					dummySelectQuery +
						" LEFT JOIN Xabc X ON DT.dummytable_id = X.xabc_dummytable_id",
				))
			})
		})
	})
})
